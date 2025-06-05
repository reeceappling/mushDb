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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const FruitingChamberSourceType = "fruitingChamber"

type FruitingChamber struct {
	EntryType            string                  `bson:"entryType" json:"entryType"`
	Id                   MainCollectionId        `bson:"_id" json:"_id"`
	Substrate            alternateCollectionId   `bson:"recipe" json:"recipe"`
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
	Flushes              []PicWithNotes          `bson:"flushes,omitempty" json:"flushes,omitempty"`
	KnownFruitable       *bool                   `bson:"knownFruitable,omitempty" json:"knownFruitable,omitempty"`
	MostRecentImage      *PicWithNotes           `bson:"mostRecentImage,omitempty" json:"mostRecentImage,omitempty"`
	Sale                 *alternateCollectionId  `bson:"sale,omitempty" json:"sale,omitempty"` // TODO: NEW! HANDLE THIS EVERYWHERE
	Disposed             *unixTime               `bson:"disposed,omitempty" json:"disposed,omitempty"`
	Notes                []Note                  `bson:"notes,omitempty" json:"notes,omitempty"`
	LastUpdated          unixTime                `bson:"lastUpdated" json:"lastUpdated"`
}

func (f FruitingChamber) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := f
	err := decodeItem(&out, encoded)
	return out, err
}

func (f FruitingChamber) clean() CollectionItem {
	out := f
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

func (f FruitingChamber) DbId() string {
	return f.Id.dbIdStr()
}

func (f FruitingChamber) projects() []string {
	return f.Projects
}

func (f FruitingChamber) GeneticInfoAsParent() (GeneticParentInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (f FruitingChamber) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return f.GenSinceSpore, f.GenSinceFruitOrSpore
}

func (f FruitingChamber) SourceType() string {
	return FruitingChamberSourceType
}

func (f FruitingChamber) setTransferParent(ctx mongo.SessionContext, xfer Transfer) error {
	upd := pushToArray("transfersOut", xfer.Id)
	res, err := ctx.Client().Database(dbName).Collection(mainCollectionName).UpdateByID(ctx, f.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("Parent not found for transfer update. Should never happen!")
	}
	return nil
}

func (f FruitingChamber) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
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
	res, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(mainCollectionName).UpdateByID(ctx, f.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("Parent not found for transfer update. Should never happen!") // TODO: MAKE VAR
	}
	return nil
}

func (f FruitingChamber) EntryTypeField() *string {
	return utils.Pointer(FruitingChamberSourceType)
}

func (f FruitingChamber) CollectionName() string {
	return mainCollectionName
}

func (f FruitingChamber) id() []byte {
	return f.Id[:]
}

func (f FruitingChamber) knownFruitable() bool {
	return *f.KnownFruitable // TODO: ensure not nil
}

// TODO: dont clean in here?
func newFruitFromFruiter(ctx context.Context, frtr fruiter, pics []PicWithNotes, notes ...Note) error {
	var latestPic *PicWithNotes = nil
	if len(pics) != 0 {
		latest := pics[0]
		for i := 1; i < len(pics); i++ {
			if pics[i].Time > latest.Time {
				latest = pics[i]
			}
		}
		latestPic = &latest
	}
	out := frtr.basicFruit()
	out.Pics = pics
	out.MostRecentImage = latestPic
	out.Notes = notes
	out.LastUpdated = unixTime(time.Now().UnixMilli())
	_, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(fruitsCollName).InsertOne(ctx, out)
	return err
}

func (f FruitingChamber) newFruit(ctx context.Context, pics []PicWithNotes, notes ...Note) error {
	// TODO: same as with bag. Probably turn into one shared fxn
	return newFruitFromFruiter(ctx, f, pics, notes...)
}

func (f FruitingChamber) basicFruit() Fruit {
	var gen *Generation = nil
	if f.GenSinceSpore != nil {
		gen = utils.Pointer((*f.GenSinceSpore) + 1) // TODO: +1 here ok?
	}
	return Fruit{
		Id:            alternateCollectionId(primitive.NewObjectID()),
		Species:       *f.Species,
		SubSpecies:    f.SubSpecies,
		GenSinceSpore: gen,
		ParentType:    utils.Pointer(FruitingChamberSourceType), // TODO: unsure if pointer ok
		Parent:        &f.Id,
		Projects:      f.Projects,
		LastUpdated:   unixTime(time.Now().UnixMilli()),
	}
}

func (f FruitingChamber) children(ctx context.Context) ([]geneticSource, error) {
	return childrenAreOnlyFruits(ctx, f.TransfersOut)
}

func (f FruitingChamber) idAsStr() string {
	return f.Id.dbIdStr()
}

func initializeFruitingChamber(ctx context.Context) error {
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(mainCollectionName)
	// If test agar batch does not exist, then create it
	existingEntry := FruitingChamber{}
	testId := mainCollIdForint(idTestFC)
	xfer := altCollIdForint(idExampleTransfer)
	plateId := mainCollIdForint(idTestPlate)
	testItem := FruitingChamber{
		EntryType:            *existingEntry.EntryTypeField(),
		Id:                   testId,
		Substrate:            exAltId,
		CreationDate:         exampleTime,
		GenSinceSpore:        &exGenSinceSpore,
		GenSinceFruitOrSpore: &exGenSinceFruitSpore,
		KnownFruitable:       exBool,
		Species:              &exampleSpecies,
		SubSpecies:           exampleSubspecies,
		Innoc:                &xfer,
		TransfersOut:         exAlts,
		ParentType:           &exParentType,
		Parent:               &plateId,
		Projects:             exProjects,
		Pics:                 exPics,
		Contaminations:       exContams,
		MostRecentImage:      &exPics[0],
		Flushes:              exPics,
		Sale:                 nil,
		Disposed:             nil,
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

type createFruitingChamberRequest struct { // TODO: THIS!
	Recipe       Base58Str // substrate recipe
	CreationDate unixTime
	Notes        []Note `json:"notes,omitempty"`
	WriteTagTo   *string
}

func createFruitingChamberHandler(w http.ResponseWriter, r *http.Request) {
	data := createFruitingChamberRequest{}
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
	recipeId, err := data.Recipe.toAltCollectionId()
	if err != nil {
		http.Error(w, "failed to resolve substrate recipe ID: "+err.Error(), http.StatusBadRequest)
		return
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(mainCollectionName)
		now := unixTime(time.Now().UnixMilli())
		_, err := coll.InsertOne(ctx, FruitingChamber{
			EntryType:    "fruitingChamber",
			Id:           id,
			Substrate:    alternateCollectionId(recipeId),
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

type importFruitingChamberRequest struct { // TODO: THIS!
	Recipe         Base58Str // Substrate recipe
	CreationDate   unixTime
	Species        string // TODO: VALIDATE ON INSERT
	Subspecies     *string
	Generation     *int
	KnownFruitable *bool
	WriteTagTo     *string
	// image as "img"
}

func importFruitingChamberHandler(w http.ResponseWriter, r *http.Request) { // TODO: THIS!
	data := importFruitingChamberRequest{}
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
		fieldBytes, errr := multipartToImageBytes(p, w)
		if errr != nil {
			// Already wrote
			return
		}
		newFileNameWithPrefixPath, errr := pics.SaveFile(r.Context(), fieldBytes, "bag", string(b58id), "img")
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
	out := FruitingChamber{
		EntryType:            "fruitingChamber",
		Id:                   id,
		Substrate:            alternateCollectionId(recipeId),
		CreationDate:         data.CreationDate,
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

type updateFruitingChamberRequest struct {
	Projects       []string `json:"projects,omitempty"`
	Notes          AllEntries[Note]
	KnownFruitable *bool                                                    `json:"knownFruitable,omitempty"`
	Disposed       *unixTime                                                `json:"disposed,omitempty"`
	Sale           *alternateCollectionId                                   `json:"sale,omitempty"`
	Images         SplitEntries[picWithNotesForm, PicWithNotesLessLocation] //"newPic-1"
	Contams        SplitEntries[contamForm, ContaminationLessLocation]      //"newContam-1"
	Flushes        SplitEntries[picWithNotesForm, PicWithNotesLessLocation] //"newFlush-1"
	WriteTagTo     *string
}

func (upr updateFruitingChamberRequest) reform() resolvedUpdateFruitingChamberRequest {
	return resolvedUpdateFruitingChamberRequest{
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
		Flushes: SplitEntries[picWithNotesForm, PicWithNotes]{
			Existing: upr.Flushes.Existing,
			New: slices.Map(upr.Flushes.New, func(i PicWithNotesLessLocation) PicWithNotes {
				return i.asPicWithNotes(nil)
			}),
		},
	}
}

type resolvedUpdateFruitingChamberRequest struct {
	KnownFruitable *bool
	Sale           *alternateCollectionId
	Disposed       *unixTime
	Projects       []string
	Notes          AllEntries[Note]
	Images         SplitEntries[picWithNotesForm, PicWithNotes]
	Contams        SplitEntries[contamForm, Contamination]
	Flushes        SplitEntries[picWithNotesForm, PicWithNotes]
}

func updateFruitingChamberHandler(w http.ResponseWriter, r *http.Request) {
	data := updateFruitingChamberRequest{}
	b58Id := Base58Str(r.PathValue("id"))
	id, err := b58Id.toMainCollectionId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	r.Body = http.MaxBytesReader(w, r.Body, 32<<20+1024) // TODO: is this max size ok?
	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "unable to open multipart reader: "+err.Error(), http.StatusBadRequest)
		return
	}
	p, err := reader.NextPart()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Process text (or object)
	bs, errr := io.ReadAll(p)
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
	newFlushes := map[int]string{}
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
		p, err = reader.NextPart()
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
			newFileNameWithPrefixPath, err := pics.SaveFile(r.Context(), fieldBytes, "fruitingChamber", string(b58Id), "img")
			if err != nil {
				http.Error(w, "failed to save new picture: "+err.Error(), http.StatusBadRequest)
				return
			}
			picsSaved = append(picsSaved, newFileNameWithPrefixPath)
			newPics[num] = newFileNameWithPrefixPath
		case "newContam":
			newFileNameWithPrefixPath, err := pics.SaveFile(r.Context(), fieldBytes, "fruitingChamber", string(b58Id), "contam")
			if err != nil {
				http.Error(w, "failed to save new contamination: "+err.Error(), http.StatusBadRequest)
				return
			}
			picsSaved = append(picsSaved, newFileNameWithPrefixPath)
			newContams[num] = newFileNameWithPrefixPath
		case "newFlush":
			newFileNameWithPrefixPath, err := pics.SaveFile(r.Context(), fieldBytes, "fruitingChamber", string(b58Id), "flush")
			if err != nil {
				http.Error(w, "failed to save new flush: "+err.Error(), http.StatusBadRequest)
				return
			}
			picsSaved = append(picsSaved, newFileNameWithPrefixPath)
			newFlushes[num] = newFileNameWithPrefixPath
		default:
			http.Error(w, "failed to save new picture: "+err.Error(), http.StatusBadRequest)
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
	for i, _ := range data.Flushes.New {
		loc, exists := newFlushes[i]
		if !exists {
			http.Error(w, fmt.Sprintf("error, location for new flush index %d not found (should never happen)", i), http.StatusInternalServerError)
			return
		}
		out.Flushes.New[i].Location = imageLocation(loc)
	}

	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(mainCollectionName)
		// go get current plate
		current := FruitingChamber{}
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
		// Compare Projects
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

		// Compare Flushes
		mods, err = WithImageChanges(mods, "flushes", out.Flushes, current.Flushes)
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
		if err := res.Err(); err != nil {
			http.Error(w, "failed to write update to db: "+err.Error(), http.StatusInternalServerError)
			return nil, nil
		}
		return w.Write([]byte(b58Id))
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}
