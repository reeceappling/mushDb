package rfid

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/reeceappling/goUtils/v2/utils"
	"github.com/reeceappling/goUtils/v2/utils/slices"
	"github.com/reeceappling/mushDb/rfid/pics"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const LcSourceType = "lc"

type LiquidCulture struct { // TODO: LIQUID CULTURE SYRINGE???
	EntryType            string                  `bson:"entryType" json:"entryType"`
	Id                   MainCollectionId        `bson:"_id" json:"_id"`
	PcRun                *alternateCollectionId  `bson:"pcRun,omitempty" json:"pcRun,omitempty"` // likely won't exist for pre-existing or purchased
	Recipe               alternateCollectionId   `bson:"recipe" json:"recipe"`                   // likely won't exist for pre-existing or purchased
	CreationDate         unixTime                `bson:"creationDate" json:"creationDate"`
	Species              *string                 `bson:"species,omitempty" json:"species,omitempty"`
	SubSpecies           *string                 `bson:"subSpecies,omitempty" json:"subSpecies,omitempty"`
	Innoc                *alternateCollectionId  `bson:"innoc,omitempty" json:"innoc,omitempty"`
	GenSinceSpore        *Generation             `bson:"genSpore,omitempty" json:"genSpore,omitempty"`
	GenSinceFruitOrSpore *Generation             `bson:"genFruitOrSpore,omitempty" json:"genFruitOrSpore,omitempty"`
	TransfersOut         []alternateCollectionId `bson:"transfersOut,omitempty" json:"transfersOut,omitempty"`
	ParentType           *string                 `bson:"parentType,omitempty" json:"parentType,omitempty"`
	Parent               *MainCollectionId       `bson:"parent,omitempty" json:"parent,omitempty"`
	Pics                 []PicWithNotes          `bson:"pics,omitempty" json:"pics,omitempty"`
	Projects             []string                `bson:"projects,omitempty" json:"projects,omitempty"`
	ConfirmedClean       *bool                   `bson:"confirmedClean,omitempty" json:"confirmedClean,omitempty"` // TODO: change so that we know exactly what confirmed it
	Contaminations       []Contamination         `bson:"contamination,omitempty" json:"contamination,omitempty"`
	KnownFruitable       *bool                   `bson:"knownFruitable,omitempty" json:"knownFruitable,omitempty"`
	Disposed             *unixTime               `bson:"disposed,omitempty" json:"disposed,omitempty"`
	MostRecentImage      *PicWithNotes           `bson:"mostRecentImage,omitempty" json:"mostRecentImage,omitempty"`
	Sales                []alternateCollectionId `bson:"sales,omitempty" json:"sales,omitempty"`
	Notes                []Note                  `bson:"notes,omitempty" json:"notes,omitempty"`
	LastUpdated          unixTime                `bson:"lastUpdated" json:"lastUpdated"`
}

func (l LiquidCulture) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := l
	err := decodeItem(&out, encoded)
	return out, err
}

func (l LiquidCulture) clean() CollectionItem {
	out := l
	// TODO: Change species
	// TODO: change subspecies
	// TODO: remove parentType and Parent
	// TODO: remove projects
	// TODO: remove pic notes
	// TODO: remove mostRecentImage notes
	// TODO: remove notes
	return out
}

func (l LiquidCulture) DbId() string {
	return l.Id.dbIdStr()
}

func (l LiquidCulture) projects() []string {
	return l.Projects
}

func (l LiquidCulture) GeneticInfoAsParent() (GeneticParentInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (l LiquidCulture) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return l.GenSinceSpore, l.GenSinceFruitOrSpore
}

func (l LiquidCulture) SourceType() string {
	return LcSourceType
}

func (l LiquidCulture) setTransferParent(ctx mongo.SessionContext, xfer Transfer) error {
	upd := pushToArray("transfersOut", xfer.Id)
	res, err := ctx.Client().Database(dbName).Collection(mainCollectionName).UpdateByID(ctx, l.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("Parent not found for transfer update. Should never happen!")
	}
	return nil
}

func (l LiquidCulture) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
	parentInfo, err := from.GeneticInfoAsParent()
	if err != nil {
		return err
	}
	if parentInfo.Species == nil {
		return errors.New("parent must have a species")
	}
	var genSpore *Generation = nil
	var genFruitSpore *Generation = nil
	switch from.SourceType() {
	case MssSourceType:
		genSpore = utils.Pointer(Generation(0))
		genFruitSpore = utils.Pointer(Generation(0))
	case FruitSourceType:
		genSpore = parentInfo.GensSinceSpore
		genFruitSpore = utils.Pointer(Generation(0))
	default:
		genSpore = parentInfo.GensSinceSpore.Next()
		genFruitSpore = parentInfo.GensSinceFruitOrSpore.Next()
	}
	upd := bson.D{bson.E{"$set", bson.D{
		{"innoc", xfer.Id},
		{"lastUpdated", xfer.LastUpdated},
		{"parentType", xfer.FromType},
		{"parent", from.DbId()},            // TODO: ENSURE OK!
		{"genSpore", genSpore},             // TODO: ensure works with ptr
		{"genFruitOrSpore", genFruitSpore}, // TODO: ensure works with ptr
		{"species", *parentInfo.Species},
		{"projects", from.projects()},
		{"lastUpdated", xfer.LastUpdated},
	}}}

	pics := []PicWithNotes{}
	if xfer.ToImage != nil {
		pic := PicWithNotes{
			Time:     xfer.Date,
			Location: *xfer.ToImage,
			Notes:    []Note{},
		}
		pics = []PicWithNotes{pic}
		upd = append(upd, bson.E{"$set", bson.D{{"mostRecentImage", pic}}})
	}
	upd = append(upd, bson.E{"$set", bson.D{{"pics", pics}}})
	if parentInfo.KnownFruitable != nil {
		upd = append(upd, bson.E{"$set", bson.D{{"knownFruitable", *parentInfo.KnownFruitable}}})
	}

	if parentInfo.Subspecies != nil {
		upd = append(upd, bson.E{"$set", bson.D{{"subSpecies", *parentInfo.Subspecies}}}) // TODO: ensure ok
	}
	res, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(mainCollectionName).UpdateByID(ctx, l.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("Parent not found for transfer update. Should never happen!") // TODO: MAKE VAR
	}
	return nil
}

func (l LiquidCulture) EntryTypeField() *string {
	return utils.Pointer(LcSourceType)
}

func (l LiquidCulture) CollectionName() string {
	return mainCollectionName
}

func (l LiquidCulture) id() []byte {
	return l.Id[:]
}

func (l LiquidCulture) knownFruitable() bool {
	return *l.KnownFruitable // TODO: ensure not nil
}

func (l LiquidCulture) children(ctx context.Context) ([]geneticSource, error) {
	//TODO implement me
	// TODO: can go anywhere (in theory) except MSS
	panic("implement me")
}

func (l LiquidCulture) idAsStr() string {
	return l.Id.dbIdStr()
}

func initializeLCs(ctx context.Context) error {
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(mainCollectionName)
	// If test agar batch does not exist, then create it
	existingEntry := LiquidCulture{}
	testId := mainCollIdForint(idTestLC)
	testItem := LiquidCulture{
		EntryType:            *existingEntry.EntryTypeField(),
		Id:                   testId,
		PcRun:                &exAltId,
		Recipe:               exAltId,
		CreationDate:         exampleTime,
		Species:              &exampleSpecies,
		SubSpecies:           exampleSubspecies,
		Innoc:                &exAltId,
		GenSinceSpore:        &exGenSinceSpore,
		GenSinceFruitOrSpore: &exGenSinceFruitSpore,
		TransfersOut:         exAlts,
		ParentType:           &exParentType,
		Parent:               &exPlate,
		Pics:                 exPics,
		Projects:             exProjects,
		ConfirmedClean:       exBool,
		Contaminations:       exContams,
		KnownFruitable:       exBool,
		Disposed:             &exampleTime,
		MostRecentImage:      &exPics[0],
		Sales:                exAlts,
		Notes:                exampleNotes(),
		LastUpdated:          exampleTime,
	}
	err := coll.FindOne(ctx, bson.D{{"_id", testId}}).Decode(&existingEntry)
	if err == nil {
		if reflect.DeepEqual(existingEntry, testItem) {
			return nil
		}
	}
	return renameMe(ctx, coll, testId, testItem, existingEntry)
}

type createLiquidCultureRequest struct {
	Recipe       alternateCollectionId // lc recipe // TODO: change all bast58str to expected type
	CreationDate unixTime
	PcRun        alternateCollectionId
	Notes        []Note `json:"notes,omitempty"`
	WriteTagTo   *string
}

func createLiquidCultureHandler(w http.ResponseWriter, r *http.Request) {
	data := createLiquidCultureRequest{}
	id, err := generateMainCollectionId(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	b58id := id.asBase58()
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(bs, &data)
	if err != nil {
		http.Error(w, "failed to unmarshal request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	now := unixTime(time.Now().UnixMilli())
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(mainCollectionName)
		_, err := coll.InsertOne(ctx, LiquidCulture{
			EntryType:    "lc",
			Id:           id,
			Recipe:       data.Recipe,
			PcRun:        &data.PcRun,
			CreationDate: data.CreationDate,
			Notes:        data.Notes,
			LastUpdated:  now,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, nil
		}
		return w.Write([]byte(b58id))
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}

type importLiquidCultureRequest struct {
	CreationDate   unixTime
	Recipe         Base58Str // lc recipe
	Species        string    // TODO: VALIDATE ON INSERT
	Subspecies     *string
	KnownFruitable *bool
	Generation     *int
	ConfirmedClean *bool
	WriteTagTo     *string
	// image as "img"
}

func importLiquidCultureHandler(w http.ResponseWriter, r *http.Request) {
	data := importLiquidCultureRequest{}
	id, err := generateMainCollectionId(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	b58id := id.asBase58()
	r.Body = http.MaxBytesReader(w, r.Body, 32<<20+1024) // TODO: is this max size ok?
	defer r.Body.Close()
	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "unable to open multipart reader: "+err.Error(), http.StatusBadRequest)
		return
	}
	p1, err := reader.NextPart()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer p1.Close()
	// Process text (or object)
	bs, errr := io.ReadAll(p1)
	if errr != nil {
		err = errr
		http.Error(w, "unable to read data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	// PARSE INTO CORRECT DATA FORMAT
	err = json.Unmarshal(bs, &data)
	if err != nil {
		http.Error(w, "unable to unmarshal json form data: "+err.Error(), http.StatusBadRequest)
		return
	}
	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Try to get pic if exists
	picsSaved := []string{}
	defer func() {
		if err != nil {
			err = pics.DeleteFiles(r.Context(), picsSaved...)
			if err != nil {
				handleFileDeleteErr(err)
			}
		}
	}()
	// Go to next part, if exists to get image
	var importedPic *PicWithNotes = nil
	p, err := reader.NextPart()
	if err != nil {
		if err != io.EOF {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		fileName := p.FileName()
		defer p.Close()
		if fileName != "img" {
			http.Error(w, "invalid image name", http.StatusBadRequest)
			return
		}
		// Process file
		fieldBytes, err := multipartToImageBytes(p, w)
		if err != nil {
			// Already wrote
			return
		}

		newFileNameWithPrefixPath, errr := pics.SaveFile(r.Context(), fieldBytes, "lc", string(b58id), "img")
		if errr != nil {
			err = errr
			http.Error(w, "failed to save file: "+err.Error(), http.StatusBadRequest)
			return
		}
		picsSaved = append(picsSaved, newFileNameWithPrefixPath)
		now := unixTime(time.Now().UnixMilli())
		importedPic = &PicWithNotes{
			Time:     now,
			Location: imageLocation(newFileNameWithPrefixPath),
			Notes:    []Note{},
		}
	}
	if speciesIsSpecial(r.Context(), &data.Species) && !userIsAdmin(r.Context()) { // TODO: DO THIS EVERYWHERE!
		http.Error(w, "not permitted to modify", http.StatusForbidden)
		return
	}
	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var gen *Generation = nil
	if data.Generation != nil {
		gen = (*Generation)(data.Generation)
	}
	recipeId, err := data.Recipe.toAltCollectionId()
	if err != nil {
		http.Error(w, "failed to resolve substrate recipe ID: "+err.Error(), http.StatusBadRequest)
		return
	}
	pix := []PicWithNotes{}
	if importedPic != nil {
		pix = []PicWithNotes{*importedPic}
	}
	out := LiquidCulture{
		EntryType:            "lc",
		Id:                   id,
		PcRun:                nil, // No pc runs on imports
		Recipe:               alternateCollectionId(recipeId),
		CreationDate:         data.CreationDate,
		Species:              &data.Species,
		SubSpecies:           data.Subspecies,
		GenSinceSpore:        gen,
		GenSinceFruitOrSpore: gen,
		Pics:                 pix,
		ConfirmedClean:       data.ConfirmedClean,
		KnownFruitable:       data.KnownFruitable,
		MostRecentImage:      importedPic,
		LastUpdated:          unixTime(time.Now().UnixMilli()),
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(mainCollectionName)
		_, err := coll.InsertOne(ctx, out)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, nil
		}
		return w.Write([]byte(id.asBase58()))
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}

type updateLiquidCultureRequest struct {
	Projects       []string `json:"projects,omitempty"`
	Notes          AllEntries[Note]
	KnownFruitable *bool                   `json:"knownFruitable,omitempty"`
	Disposed       *unixTime               `json:"disposed,omitempty"`
	Sales          []alternateCollectionId `json:"sale,omitempty"`
	ConfirmedClean *bool
	Images         SplitEntries[picWithNotesForm, PicWithNotesLessLocation] //"newPic-1"
	Contams        SplitEntries[contamForm, ContaminationLessLocation]      //"newContam-1"
	WriteTagTo     *string
}

func (upr updateLiquidCultureRequest) reform() resolvedUpdateLiquidCultureRequest {
	return resolvedUpdateLiquidCultureRequest{
		ConfirmedClean: upr.ConfirmedClean,
		KnownFruitable: upr.KnownFruitable,
		Sales:          upr.Sales,
		Disposed:       upr.Disposed,
		Projects:       upr.Projects,
		Notes:          upr.Notes,
		Images: SplitEntries[picWithNotesForm, PicWithNotes]{
			Existing: upr.Images.Existing,
			New: slices.Map(upr.Images.New, func(i PicWithNotesLessLocation) PicWithNotes {
				return i.asPicWithNotes(nil)
			}),
		},
		Contams: SplitEntries[contamForm, Contamination]{
			Existing: upr.Contams.Existing,
			New: slices.Map(upr.Contams.New, func(i ContaminationLessLocation) Contamination {
				return i.asContamination(nil)
			}),
		},
	}
}

type resolvedUpdateLiquidCultureRequest struct {
	KnownFruitable *bool
	Sales          []alternateCollectionId
	Disposed       *unixTime
	Projects       []string
	Notes          AllEntries[Note]
	ConfirmedClean *bool
	Images         SplitEntries[picWithNotesForm, PicWithNotes]
	Contams        SplitEntries[contamForm, Contamination]
}

func updateLiquidCultureHandler(w http.ResponseWriter, r *http.Request) {
	data := updateLiquidCultureRequest{}
	b58Id := Base58Str(r.PathValue("id"))
	id, err := b58Id.toMainCollectionId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	r.Body = http.MaxBytesReader(w, r.Body, 32<<20+1024) // TODO: is this max size ok?
	defer r.Body.Close()
	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "unable to open multipart reader: "+err.Error(), http.StatusBadRequest)
		return
	}
	p1, err := reader.NextPart()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer p1.Close()
	// Process text (or object)
	bs, errr := io.ReadAll(p1)
	if errr != nil {
		err = errr
		http.Error(w, "failed to read data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	// PARSE INTO CORRECT DATA FORMAT
	err = json.Unmarshal(bs, &data)
	if err != nil {
		http.Error(w, "failed to unmarshal data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Get any images
	newPics := map[int]string{}
	newContams := map[int]string{}
	picsSaved := []string{}
	defer func() {
		if err != nil {
			errDel := pics.DeleteFiles(r.Context(), picsSaved...)
			if errDel != nil {
				handleFileDeleteErr(errDel)
			}
		}
	}()
	for {
		// Go to next part or break
		p, err := reader.NextPart()
		if err != nil {
			if err != io.EOF {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			break
		}
		fileName := p.FileName()
		if fileName == "" {
			http.Error(w, "file name is empty for what should have been an image", http.StatusBadRequest)
			return
		}
		// Process file
		parts := strings.Split(fileName, "-")
		if len(parts) != 2 {
			http.Error(w, "invalid image name: "+fileName, http.StatusBadRequest)
			return
		}
		num, errr := strconv.Atoi(parts[1])
		if errr != nil {
			err = errr
			http.Error(w, "failed to parse image number! "+errr.Error(), http.StatusBadRequest)
			return
		}
		fieldBytes, err := multipartToImageBytes(p, w)
		if err != nil {
			// Already wrote
			return
		}
		switch parts[0] {
		case "newPic":
			newFileNameWithPrefixPath, err := pics.SaveFile(r.Context(), fieldBytes, "lc", string(b58Id), "img")
			if err != nil {
				http.Error(w, "failed to save new picture: "+err.Error(), http.StatusBadRequest)
				return
			}
			picsSaved = append(picsSaved, newFileNameWithPrefixPath)
			newPics[num] = newFileNameWithPrefixPath
		case "newContam":
			newFileNameWithPrefixPath, err := pics.SaveFile(r.Context(), fieldBytes, "lc", string(b58Id), "contam")
			if err != nil {
				http.Error(w, "failed to save new contamination: "+err.Error(), http.StatusBadRequest)
				return
			}
			picsSaved = append(picsSaved, newFileNameWithPrefixPath)
			newContams[num] = newFileNameWithPrefixPath
		default:
			http.Error(w, "invalid picture name", http.StatusBadRequest)
			return
		}
	}

	// CHECK THAT ALL NEW PICS EXIST
	// PROCESS ALL NEW PICS AND CONTAMS
	out := data.reform()
	for i, _ := range data.Images.New {
		loc, exists := newPics[i]
		if !exists {
			http.Error(w, fmt.Sprintf("error, location for new picture index %d not found (should never happen)", i), http.StatusInternalServerError)
			return
		}
		out.Images.New[i].Location = imageLocation(loc)
	}
	for i, _ := range data.Contams.New {
		if loc, exists := newContams[i]; exists {
			finalLoc := imageLocation(loc)
			out.Contams.New[i].Location = &finalLoc
		}
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(mainCollectionName)
		// go get current plate
		current := LiquidCulture{}
		err := coll.FindOne(ctx, bson.D{{"_id", id}}).Decode(&current)
		if err != nil {
			http.Error(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
			return nil, nil
		}
		if speciesIsSpecial(ctx, current.Species) && !userIsAdmin(ctx) { // TODO: DO THIS EVERYWHERE!
			http.Error(w, "not permitted to modify", http.StatusForbidden)
			return nil, nil
		}
		upd := bson.D{}
		// Compare KF
		upd = setUnsetUnequalPointers("knownFruitable", out.KnownFruitable, current.KnownFruitable, upd)
		// Compare ConfirmedClean
		upd = setUnsetUnequalPointers("confirmedClean", out.ConfirmedClean, current.ConfirmedClean, upd)
		// Compare SALES
		upd = setSalesIfUnequal(upd, out.Sales, out.Sales)
		// Compare PROJECTS
		upd = setProjectsIfUnequal(upd, out.Projects, current.Projects)
		upd = setUnsetUnequalPointers("disposed", out.Disposed, current.Disposed, upd)
		// Do note changes
		mods, err := WithNotesUpdate(bson.D{}, out.Notes, current.Notes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return nil, nil
		}

		// Compare Images
		mods, err = WithImageChanges(mods, "pics", out.Images, current.Pics)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return nil, nil
		}

		// Compare Contams
		mods, err = WithContamChanges(mods, "contamination", out.Contams, current.Contaminations)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return nil, nil
		}

		if len(mods) == 0 {
			http.Error(w, "no changes made", http.StatusBadRequest)
			return nil, nil
		}

		// write updates to db
		res := coll.FindOneAndUpdate(ctx, bson.D{{"_id", id}}, mods)
		if err = res.Err(); err != nil {
			http.Error(w, "failed to write update to db: "+err.Error(), http.StatusInternalServerError)
			return nil, nil
		}
		return w.Write([]byte(b58Id))
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}
