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

const SlantSourceType = "slant"

type Slant struct {
	// TODO: GENERATION(s)?
	EntryType            string                  `bson:"entryType" json:"entryType"`
	Id                   MainCollectionId        `bson:"_id" json:"_id"`
	Agar                 *alternateCollectionId  `bson:"agar,omitempty" json:"agar,omitempty"` // May not exist for pre-existing specimens
	CreationDate         unixTime                `bson:"creationDate" json:"creationDate"`
	Species              *string                 `bson:"species,omitempty" json:"species,omitempty"`
	SubSpecies           *string                 `bson:"subSpecies,omitempty" json:"subSpecies,omitempty"`
	Innoc                *alternateCollectionId  `bson:"innoc,omitempty" json:"innoc,omitempty"`
	GenSinceSpore        *Generation             `bson:"genSpore,omitempty" json:"genSpore,omitempty"`
	GenSinceFruitOrSpore *Generation             `bson:"genFruitOrSpore,omitempty" json:"genFruitOrSpore,omitempty"`
	TransfersOut         []alternateCollectionId `bson:"transfersOut,omitempty" json:"transfersOut,omitempty"`
	ParentType           *string                 `bson:"parentType,omitempty" json:"parentType,omitempty"` // TODO: NEW! HANDLE! nil == mainCollectionType, can also be MSS or clone! // TODO: INDEX????
	Parent               *MainCollectionId       `bson:"parent,omitempty" json:"parent,omitempty"`
	Projects             []string                `bson:"projects,omitempty" json:"projects,omitempty"`
	Pics                 []PicWithNotes          `bson:"pics,omitempty" json:"pics,omitempty"`
	Contaminations       []Contamination         `bson:"contamination,omitempty" json:"contamination,omitempty"`
	KnownFruitable       *bool                   `bson:"knownFruitable,omitempty" json:"knownFruitable,omitempty"`
	Sale                 *alternateCollectionId  `bson:"sale,omitempty" json:"sale,omitempty"`
	Disposed             *unixTime               `bson:"disposed,omitempty" json:"disposed,omitempty"`
	MostRecentImage      *PicWithNotes           `bson:"mostRecentImage,omitempty" json:"mostRecentImage,omitempty"`
	Notes                []Note                  `bson:"notes,omitempty" json:"notes,omitempty"`
	LastUpdated          unixTime                `bson:"lastUpdated" json:"lastUpdated"`
}

func (s Slant) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := s
	err := decodeItem(&out, encoded)
	return out, err
}

func (s Slant) clean() CollectionItem {
	out := s
	// TODO: Change species
	// TODO: change subspecies
	// TODO: remove parentType and Parent
	// TODO: remove projects
	// TODO: remove pic notes
	// TODO: remove mostRecentImage notes
	// TODO: remove flushes notes
	// TODO: remove notes
	return out
}

func (s Slant) DbId() string {
	return s.Id.dbIdStr()
}

func (s Slant) projects() []string {
	return s.Projects
}

func (s Slant) GeneticInfoAsParent() (GeneticParentInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (s Slant) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return s.GenSinceSpore, s.GenSinceFruitOrSpore
}

func (s Slant) SourceType() string {
	return SlantSourceType
}

func (s Slant) setTransferParent(ctx mongo.SessionContext, xfer Transfer) error {
	upd := pushToArray("transfersOut", xfer.Id)
	res, err := ctx.Client().Database(dbName).Collection(mainCollectionName).UpdateByID(ctx, s.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("Parent not found for transfer update. Should never happen!")
	}
	return nil
}

func (s Slant) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
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
	res, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(mainCollectionName).UpdateByID(ctx, s.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("Parent not found for transfer update. Should never happen!") // TODO: MAKE VAR
	}
	return nil
}

func (s Slant) EntryTypeField() *string {
	return utils.Pointer(SlantSourceType)
}

func (s Slant) CollectionName() string {
	return mainCollectionName
}

func (s Slant) id() []byte {
	return s.Id[:]
}

func (s Slant) knownFruitable() bool {
	return *s.KnownFruitable // TODO: ensure not nil
}

func (s Slant) children(ctx context.Context) ([]geneticSource, error) {
	return childrenOnlyToPlate(ctx, s.TransfersOut)
}

func (s Slant) idAsStr() string {
	return s.Id.dbIdStr()
}

func initializeSlants(ctx context.Context) error {
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(mainCollectionName)
	// If test agar batch does not exist, then create it
	existingEntry := Slant{}
	testId := mainCollIdForint(idTestSlant)
	testItem := Slant{
		EntryType:            *existingEntry.EntryTypeField(),
		Id:                   testId,
		Agar:                 &exAltId,
		CreationDate:         exampleTime,
		Species:              &exampleSpecies,
		SubSpecies:           exampleSubspecies,
		Innoc:                &exAltId,
		GenSinceSpore:        &exGenSinceSpore,
		GenSinceFruitOrSpore: &exGenSinceFruitSpore,
		TransfersOut:         exAlts,
		ParentType:           &exParentType,
		Parent:               &exPlate,
		Projects:             exProjects,
		Pics:                 exPics,
		Contaminations:       exContams,
		KnownFruitable:       exBool,
		Sale:                 &exAltId,
		Disposed:             &exampleTime,
		MostRecentImage:      &exPics[0],
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

type createSlantRequest struct {
	agarBatch  Base58Str
	writeTagTo *string
}

func createSlantHandler(w http.ResponseWriter, r *http.Request) {
	data := createSlantRequest{}
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
	err = writeRfidTagIfNecessary(r.Context(), data.writeTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}

	batchId, err := data.agarBatch.toAltCollectionId()
	if err != nil {
		http.Error(w, "failed to resolve agar batch ID: "+err.Error(), http.StatusBadRequest)
		return
	}
	batch := alternateCollectionId(batchId)
	now := unixTime(time.Now().UnixMilli())
	item := Slant{
		EntryType:    "slant",
		Id:           id,
		Agar:         &batch,
		CreationDate: now,
		LastUpdated:  now,
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(mainCollectionName)
		_, err = coll.InsertOne(ctx, item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, nil
		}
		_, err = w.Write([]byte(b58id))
		return nil, err
	})

	if err != nil {
		handleWriteErr(err, w)
	}
}

type updateSlantRequest struct {
	KnownFruitable *bool
	Sale           *alternateCollectionId
	Disposed       *unixTime
	Projects       []string
	Notes          AllEntries[Note]
	Images         SplitEntries[picWithNotesForm, PicWithNotesLessLocation]
	Contams        SplitEntries[contamForm, ContaminationLessLocation]
	WriteTagTo     *string
}

func (upr updateSlantRequest) reform() resolvedUpdateSlantRequest {
	return resolvedUpdateSlantRequest{
		KnownFruitable: upr.KnownFruitable,
		Sale:           upr.Sale,
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
		WriteTagTo: upr.WriteTagTo,
	}
}

type resolvedUpdateSlantRequest struct {
	KnownFruitable *bool
	Sale           *alternateCollectionId
	Disposed       *unixTime
	Projects       []string
	Notes          AllEntries[Note]
	Images         SplitEntries[picWithNotesForm, PicWithNotes]
	Contams        SplitEntries[contamForm, Contamination]
	WriteTagTo     *string
}

func updateSlantHandler(w http.ResponseWriter, r *http.Request) {
	data := updateSlantRequest{}
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
		p, errRead := reader.NextPart()
		if errRead != nil {
			if errRead != io.EOF {
				http.Error(w, errRead.Error(), http.StatusInternalServerError)
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
		num, errConv := strconv.Atoi(parts[1])
		if errConv != nil {
			err = errConv
			http.Error(w, "failed to parse image number! "+err.Error(), http.StatusBadRequest)
			return
		}
		fieldBytes, errr := multipartToImageBytes(p, w)
		if errr != nil {
			// Already wrote
			return
		}
		switch parts[0] {
		case "newPic":
			newFileNameWithPrefixPath, err := pics.SaveFile(r.Context(), fieldBytes, "slant", string(b58Id), "img") // TODO: FIX THIS EVERYWHERE!
			if err != nil {
				http.Error(w, "failed to save new picture: "+err.Error(), http.StatusBadRequest)
				return
			}
			picsSaved = append(picsSaved, newFileNameWithPrefixPath)
			newPics[num] = newFileNameWithPrefixPath
		case "newContam":
			newFileNameWithPrefixPath, err := pics.SaveFile(r.Context(), fieldBytes, "slant", string(b58Id), "contam") // TODO: FIX THIS EVERYWHERE!
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
		current := Slant{}
		err = coll.FindOne(ctx, bson.D{{"_id", id}}).Decode(&current)
		if err != nil {
			http.Error(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
			return nil, nil
		}
		if speciesIsSpecial(ctx, current.Species) && !userIsAdmin(ctx) { // TODO: DO THIS EVERYWHERE!
			http.Error(w, "not permitted to modify", http.StatusForbidden)
			return nil, nil
		}
		upd := bson.D{}
		// Compare PROJECTS
		upd = setProjectsIfUnequal(upd, out.Projects, current.Projects)
		// Compare KF
		upd = setUnsetUnequalPointers("knownFruitable", out.KnownFruitable, current.KnownFruitable, upd)
		// Compare SALE
		upd = setUnsetUnequalPointers("sale", out.Sale, current.Sale, upd)
		// Compare DISPOSED
		upd = setUnsetUnequalPointers("disposed", out.Disposed, current.Disposed, upd)
		// Do note changes
		mods, err := WithNotesUpdate(bson.D{}, out.Notes, current.Notes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return nil, nil
		}

		// Compare Images
		mods, err = WithExistingEntriesChange(mods, "pics", out.Images.Existing, current.Pics, compareImageUpdate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return nil, nil
		}
		mods = append(mods, pushToArray("pics", out.Images.New...)...)

		// Compare Contams
		mods, err = WithExistingEntriesChange(mods, "contamination", out.Contams.Existing, current.Contaminations, compareContamUpdate)
		mods = append(mods, pushToArray("contamination", out.Contams.New...)...)
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
		_, err = w.Write([]byte(b58Id))
		return nil, err
	})
	if err != nil {
		// TODO: WHEN WRITE FAIL
	}
}

type importSlantRequest struct {
	Created        unixTime
	Species        string
	Subspecies     *string
	KnownFruitable *bool
	Generation     *int
	// pic as "img"
	WriteTagTo *string
}

func importSlantHandler(w http.ResponseWriter, r *http.Request) {
	data := importSlantRequest{}
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
		newFileNameWithPrefixPath, errSave := pics.SaveFile(r.Context(), fieldBytes, "slant", string(b58id), "img")
		if errSave != nil {
			err = errSave
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
	var gen *Generation = nil
	if data.Generation != nil {
		gen = (*Generation)(data.Generation)
	}
	pix := []PicWithNotes{}
	if importedPic != nil {
		pix = []PicWithNotes{*importedPic}
	}
	out := Slant{
		EntryType:            "slant",
		Id:                   id,
		CreationDate:         data.Created,
		Species:              &data.Species,
		SubSpecies:           data.Subspecies,
		GenSinceSpore:        gen,
		GenSinceFruitOrSpore: gen,
		Pics:                 pix,
		KnownFruitable:       data.KnownFruitable,
		MostRecentImage:      importedPic,
		LastUpdated:          unixTime(time.Now().UnixMilli()),
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(mainCollectionName)
		_, err = coll.InsertOne(ctx, out)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, nil
		}
		_, err = w.Write([]byte(id.asBase58()))
		return nil, err
	})
	if err != nil {
		// TODO: handle write failure
	}
}
