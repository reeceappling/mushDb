package rfid

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/reeceappling/goUtils/v2/utils"
	sliceutils "github.com/reeceappling/goUtils/v2/utils/slices"
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

const (
	FruitSourceType = "fruit"
	fruitsCollName  = "fruits"
)

type Fruit struct { // KnownFruitable is always true for this, // creation date field is id
	Id            alternateCollectionId   `bson:"_id" json:"_id"`
	HarvestDate   unixTime                `bson:"harvestDate" json:"harvestDate"`
	Species       string                  `bson:"species" json:"species"`
	SubSpecies    *string                 `bson:"subSpecies,omitempty" json:"subSpecies,omitempty"`
	GenSinceSpore *Generation             `bson:"genSpore,omitempty" json:"genSpore,omitempty"`
	TransfersOut  []alternateCollectionId `bson:"transfersOut,omitempty" json:"transfersOut,omitempty"` // handled by new Transfer. Can only be clone to plate (sporeprint handled another way)
	Prints        []alternateCollectionId `bson:"prints,omitempty" json:"prints,omitempty"`
	ParentType    *string                 `bson:"parentType,omitempty" json:"parentType,omitempty"`
	// parent can be "store, outside, or a mainCollectionId"
	Parent          *MainCollectionId `bson:"parent,omitempty" json:"parent,omitempty"` // NONEXISTENT MEANS FROM STORE or outside???
	Projects        []string          `bson:"projects,omitempty" json:"projects,omitempty"`
	Pics            []PicWithNotes    `bson:"pics" json:"pics"`
	Disposed        *unixTime         `bson:"disposed,omitempty" json:"disposed,omitempty"`
	MostRecentImage *PicWithNotes     `bson:"mostRecentImage,omitempty" json:"mostRecentImage,omitempty"`
	Notes           []Note            `bson:"notes"  json:"notes"`
	LastUpdated     unixTime          `bson:"lastUpdated" json:"lastUpdated"`
}

func (f Fruit) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := f
	err := decodeItem(&out, encoded)
	return out, err
}

func (f Fruit) clean() CollectionItem {
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

func (f Fruit) DbId() string {
	return alternateCollectionId(f.Id).String()
}

func (f Fruit) projects() []string {
	return f.Projects
}

func (f Fruit) GeneticInfoAsParent() (GeneticParentInfo, error) {
	return GeneticParentInfo{
		Species:               &f.Species,
		Subspecies:            f.SubSpecies,
		KnownFruitable:        utils.Pointer(true),
		GensSinceSpore:        f.GenSinceSpore,
		GensSinceFruitOrSpore: utils.Pointer(Generation(0)),
	}, nil
}

func (f Fruit) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return f.GenSinceSpore, (*Generation)(utils.Pointer(0))
}

func (f Fruit) SourceType() string {
	return FruitSourceType
}

func (f Fruit) setTransferParent(ctx mongo.SessionContext, xfer Transfer) error {
	upd := pushToArray("transfersOut", xfer.Id)
	res, err := ctx.Client().Database(dbName).Collection(fruitsCollName).UpdateByID(ctx, f.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("Parent not found for transfer update. Should never happen!")
	}
	return nil
}

func (f Fruit) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
	return errors.New("fruits are invalid transfer children")
}

func (f Fruit) EntryTypeField() *string {
	return nil
}

func (f Fruit) altId() alternateCollectionId {
	return alternateCollectionId(f.Id)
}

func (f Fruit) id() []byte {
	return f.Id[:]
}

func (f Fruit) knownFruitable() bool {
	return true
}

func (f Fruit) addSporePrint(ctx mongo.SessionContext, printId alternateCollectionId) error {
	// update fruit
	res, err := ctx.Client().Database(dbName).Collection(fruitsCollName).UpdateByID(ctx, f.Id, pushToArray("prints", printId))
	if err != nil {
		return err
	}
	if res.ModifiedCount != 1 {
		return errors.New("invalid result") // TODO: ok?
	}
	return nil
}

func (f Fruit) CollectionName() string {
	return fruitsCollName
}

func (f Fruit) children(ctx context.Context) ([]geneticSource, error) {
	out := []geneticSource{}
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	xferColl := db.Collection(transfersCollName)
	mainColl := db.Collection(mainCollectionName)
	for _, xferId := range f.TransfersOut {
		var xfer Transfer
		var dish Plate
		if err := xferColl.FindOne(ctx, bson.D{{"_id", xferId}}).Decode(&xfer); err != nil {
			return nil, errors.Join(errors.New("failed to retrieve and decode transfer by id"), err)
		}
		if err := mainColl.FindOne(ctx, bson.D{{"_id", xfer.To}}).Decode(&dish); err != nil {
			return nil, errors.Join(errors.New("failed to retrieve and decode dish"), err)
		}
		out = append(out, dish)
	}
	return out, nil
}

func (f Fruit) idAsStr() string {
	bs := [12]byte(f.Id)
	return string(bs[:])
}

func initializeFruits(ctx context.Context) error {
	// Indices
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(fruitsCollName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		newSimpleIndex("species", "species", false, false, false),
		newSimpleIndex("subSpecies", "subSpecies", false, true, false),
		newSimpleIndex("transfersOut", "transfersOut", false, true, false),
		newSimpleIndex("harvestDate", "harvestDate", true, false, false),
		newSimpleIndex("prints", "prints", false, true, false),
		newSimpleIndex("parent", "parent", false, false, false),         // TODO: nil is store or outside?
		newSimpleIndex("parentType", "parentType", false, false, false), // TODO: nil is store or outside?
		newSimpleIndex("genSpore", "genSpore", true, true, false),
		newSimpleIndex("projects", "projects", false, false, false),
		//Pics (no index)
		newSimpleIndex("disposed", "disposed", false, true, false),
		//MostRecentImage (no index)
		//Notes (no index) (maybe later with tags?)
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	existingEntry := Fruit{}
	testItem := Fruit{
		Id:              exAltId,
		HarvestDate:     exampleTime,
		Species:         "beech",
		SubSpecies:      utils.Pointer("brown beech"),
		GenSinceSpore:   &exGenSinceSpore,
		TransfersOut:    exAlts,
		Prints:          exAlts,
		ParentType:      &exParentType,
		Parent:          &exPlate,
		Projects:        exProjects,
		Pics:            exPics,
		Disposed:        &exampleTime,
		MostRecentImage: &exPics[0],
		Notes:           exampleNotes(),
		LastUpdated:     exampleTime,
	}
	err = coll.FindOne(ctx, bson.D{{"_id", exAltId}}).Decode(&existingEntry)
	if err == nil {
		if reflect.DeepEqual(existingEntry, testItem) {
			return nil
		}
	}
	return renameMe(ctx, coll, exAltId, testItem, existingEntry)
}

type createFruitRequest struct {
	ParentId    Base58Str
	ParentType  string
	HarvestDate unixTime
	Projects    []string
	Notes       []Note                     `json:"notes,omitempty"`
	Pics        []PicWithNotesLessLocation // newPic-1
}

func (req createFruitRequest) reform() createFruitResolved {
	return createFruitResolved{
		ParentId:    req.ParentId,
		ParentType:  req.ParentType,
		HarvestDate: req.HarvestDate,
		Projects:    req.Projects,
		Notes:       req.Notes,
		Pics: sliceutils.Map(req.Pics, func(i PicWithNotesLessLocation) PicWithNotes {
			return PicWithNotes{
				Time:     i.Time,
				Location: "",
				Notes:    i.Notes,
			}
		}),
	}
}

type createFruitResolved struct {
	ParentId    Base58Str
	ParentType  string
	HarvestDate unixTime
	Projects    []string
	Notes       []Note         `json:"notes,omitempty"`
	Pics        []PicWithNotes // newPic-1
}

func createFruitHandler(w http.ResponseWriter, r *http.Request) { // TODO: DO FORMAT WITH DATA FIRST!
	data, dataParsed := createFruitRequest{}, false
	id := newAlternateCollectionId()
	b58id := id.base58()
	defer r.Body.Close()
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
	newPics := map[int]string{}
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
		fileName := p.FileName()

		if isFile := fileName != ""; isFile {
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
				newFileNameWithPrefixPath, err := pics.SaveFile(r.Context(), fieldBytes, "fruit", string(b58id), "img")
				if err != nil {
					http.Error(w, "failed to save new picture: "+err.Error(), http.StatusBadRequest)
					return
				}
				picsSaved = append(picsSaved, newFileNameWithPrefixPath)
				newPics[num] = newFileNameWithPrefixPath
			default:
				http.Error(w, "invalid picture name", http.StatusBadRequest)
				return
			}
		} else {
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
			dataParsed = true
		}

		// Go to next part or break
		p, err = reader.NextPart()
		if err != nil {
			if err != io.EOF {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			break
		}
	}

	if !dataParsed {
		http.Error(w, "no data found on form!", http.StatusBadRequest)
		return
	}
	// CHECK THAT ALL NEW PICS EXIST
	// PROCESS ALL NEW PICS AND CONTAMS
	out := data.reform()

	for i, _ := range data.Pics {
		loc, exists := newPics[i]
		if !exists {
			http.Error(w, fmt.Sprintf("error, location for picture index %d not found (should never happen)", i), http.StatusInternalServerError)
			return
		}
		out.Pics[i].Location = imageLocation(loc)
	}

	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		db := ctx.Client().Database(dbName)
		// go get parent
		parentId, err := data.ParentId.toMainCollectionId()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return nil, nil
		}
		parent, err := GetMainCollectionItemInTxn(ctx, parentId, nil)
		if err != nil {
			http.Error(w, "parent not found: "+err.Error(), http.StatusNotFound)
			return nil, nil
		}
		parentGenetics, err := parent.GeneticInfoAsParent()
		if err != nil {
			http.Error(w, "parent genetics error: "+err.Error(), http.StatusInternalServerError)
			return nil, nil
		}
		var mri *PicWithNotes = nil
		if len(out.Pics) > 0 {
			mri = &(out.Pics[len(out.Pics)-1])
		}
		if speciesIsSpecial(r.Context(), parentGenetics.Species) && !userIsAdmin(r.Context()) { // TODO: DO THIS EVERYWHERE!
			http.Error(w, "not permitted to modify", http.StatusForbidden)
			return nil, nil
		}
		// Write new fruit to db
		res, err := db.Collection(fruitsCollName).InsertOne(ctx, Fruit{
			Id:              alternateCollectionId(id),
			HarvestDate:     out.HarvestDate,
			Species:         *parentGenetics.Species,
			SubSpecies:      parentGenetics.Subspecies,
			GenSinceSpore:   parentGenetics.GensSinceSpore,
			ParentType:      &out.ParentType,
			Parent:          &parentId,
			Projects:        out.Projects,
			Pics:            out.Pics,
			MostRecentImage: mri,
			Notes:           out.Notes,
			LastUpdated:     unixTime(time.Now().UnixMilli()),
		})
		if err != nil {
			http.Error(w, "error writing: "+err.Error(), http.StatusInternalServerError)
			return nil, nil
		}
		idOut, ok := res.InsertedID.(alternateCollectionId)
		if !ok {
			http.Error(w, "could not find written id", http.StatusInternalServerError)
			return nil, nil
		}
		return w.Write(idOut.base58Bytes())
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}

type updateFruitRequest struct {
	Disposed *unixTime `json:"disposed,omitempty"`
	Projects []string  `json:"projects,omitempty"`
	Notes    AllEntries[Note]
	Images   SplitEntries[picWithNotesForm, PicWithNotesLessLocation] //"newPic-1"
}

func (upr updateFruitRequest) reform() resolvedUpdateFruitRequest {
	return resolvedUpdateFruitRequest{
		Disposed: upr.Disposed,
		Projects: upr.Projects,
		Notes:    upr.Notes,
		Images: SplitEntries[picWithNotesForm, PicWithNotes]{
			Existing: upr.Images.Existing,
			New: sliceutils.Map(upr.Images.New, func(i PicWithNotesLessLocation) PicWithNotes {
				return i.asPicWithNotes(nil)
			}),
		},
	}
}

type resolvedUpdateFruitRequest struct {
	Disposed *unixTime `json:"disposed,omitempty"`
	Projects []string  `json:"projects,omitempty"`
	Notes    AllEntries[Note]
	Images   SplitEntries[picWithNotesForm, PicWithNotes] //"newPic-1"
}

func updateFruitHandler(w http.ResponseWriter, r *http.Request) {
	data := updateFruitRequest{}
	b58Id := Base58Str(r.PathValue("id"))
	id, err := b58Id.toAltCollectionId()
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
	// Do pics if exists!
	newPics := map[int]string{}
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
		// Go to next part or error
		p, err := reader.NextPart()
		if err != nil {
			if err != io.EOF {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			break
		}
		fileName := p.FileName()
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
			newFileNameWithPrefixPath, err := pics.SaveFile(r.Context(), fieldBytes, "fruit", string(b58Id), "img")
			if err != nil {
				http.Error(w, "failed to save new picture: "+err.Error(), http.StatusBadRequest)
				return
			}
			picsSaved = append(picsSaved, newFileNameWithPrefixPath)
			newPics[num] = newFileNameWithPrefixPath
		default:
			http.Error(w, "invalid picture name. Should never occur", http.StatusInternalServerError)
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
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(fruitsCollName)
		// go get current plate
		current := Fruit{}
		err = coll.FindOne(ctx, bson.D{{"_id", id}}).Decode(&current)
		if err != nil {
			http.Error(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
			return nil, nil
		}
		if speciesIsSpecial(ctx, &current.Species) && !userIsAdmin(ctx) { // TODO: DO THIS EVERYWHERE!
			http.Error(w, "not permitted to modify", http.StatusForbidden)
			return nil, nil
		}
		upd := bson.D{}
		// Compare Projects
		upd = setProjectsIfUnequal(upd, out.Projects, current.Projects)
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

type importFruitRequest struct {
	ParentType string // "store" or "outside"
	Species    string
	Subspecies *string
	Notes      []Note
	// image as "img"
}

func importFruitHandler(w http.ResponseWriter, r *http.Request) { // TODO: REDO!!!!!!
	data := importFruitRequest{}
	id := newAlternateCollectionId()
	b58id := id.base58()
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
	picsSaved := []string{}
	defer func() {
		if err != nil {
			err = pics.DeleteFiles(r.Context(), picsSaved...)
			if err != nil {
				handleFileDeleteErr(err)
			}
		}
	}()
	// var maxSize int64 = 32 << 20 // TODO: IS THIS OK? DO WE NEED THIS?
	var importedPic *PicWithNotes = nil
	dataProcessed := false
	filesProcessed := 0
	for {
		fileName := p.FileName()
		defer p.Close() // TODO: ensure close?

		if isFile := fileName != ""; isFile {
			if filesProcessed > 0 {
				// TODO: ERROR! DONT CREATE MORE THAN 1 IMAGE!
			}
			// Process file
			fieldBytes, err := multipartToImageBytes(p, w)
			if err != nil {
				// Already wrote
				return
			}
			newFileNameWithPrefixPath, errr := pics.SaveFile(r.Context(), fieldBytes, "fruit", string(b58id), "img")
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
			filesProcessed++
		} else {
			// Process text (or object)
			bs, errr := io.ReadAll(p)
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
			dataProcessed = true
		}

		// Go to next part or break
		p, err = reader.NextPart()
		if err != nil {
			if err != io.EOF {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			break
		}
	}
	if !dataProcessed {
		err = errors.New("no non-image data found on form request")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if speciesIsSpecial(r.Context(), &data.Species) && !userIsAdmin(r.Context()) { // TODO: DO THIS EVERYWHERE!
		http.Error(w, "not permitted to modify", http.StatusForbidden)
		return
	}
	var gen = (*Generation)(utils.Pointer(1)) // TODO: ok?
	pix := []PicWithNotes{}
	if importedPic != nil {
		pix = []PicWithNotes{*importedPic}
	}
	now := unixTime(time.Now().UnixMilli())
	out := Fruit{
		Id:              alternateCollectionId(id),
		HarvestDate:     unixTime(time.Now().UnixMilli()),
		Species:         data.Species,
		SubSpecies:      data.Subspecies,
		GenSinceSpore:   gen,
		ParentType:      &data.ParentType,
		Pics:            pix,
		MostRecentImage: importedPic,
		Notes:           data.Notes,
		LastUpdated:     now,
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(fruitsCollName)
		_, err := coll.InsertOne(ctx, out)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, nil
		}
		return w.Write(id.base58Bytes())
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}
