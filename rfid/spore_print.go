package rfid

import (
	"context"
	"encoding/json"
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

const (
	SporePrintSourceType     = "sporePrint"
	sporePrintCollectionName = "sporePrints"
	sporePrintIdPrefix       = "sp"
)

type SporePrint struct { // TODO: add most recently updated field, add creation date field
	Id alternateCollectionId `bson:"_id" json:"_id"`
	// Parent is always either fruit, or purchased
	Parent          *alternateCollectionId `bson:"parent,omitempty" json:"parent,omitempty"` // TODO: handle now a pointer       // TODO: likely won't exist for pre-existing
	Projects        []string               `bson:"projects,omitempty" json:"projects,omitempty"`
	PrintDate       unixTime               `bson:"printDate" json:"printDate"` // TODO: likely won't exist for pre-existing or purchased
	Species         string                 `bson:"species" json:"species"`
	Subspecies      *string                `bson:"subSpecies,omitempty" json:"subSpecies,omitempty"`
	Pics            []PicWithNotes         `bson:"pics,omitempty" json:"pics,omitempty"`
	Sale            *alternateCollectionId `bson:"sale,omitempty" json:"sale,omitempty"`
	Disposed        *unixTime              `bson:"disposed,omitempty" json:"disposed,omitempty"`               // TODO: MAKE SURE UNIXTIME DIDNT BREAK ANYTHING!            // TODO: THIS IS NEW! HANDLE EVERYWHERE!!!!!
	MostRecentImage *PicWithNotes          `bson:"mostRecentImage,omitempty" json:"mostRecentImage,omitempty"` // TODO: handle me
	Notes           []Note                 `bson:"notes,omitempty" json:"notes,omitempty"`
	LastUpdated     unixTime               `bson:"lastUpdated" json:"lastUpdated"`
}

func (sp SporePrint) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := sp
	err := decodeItem(&out, encoded)
	return out, err
}

func (sp SporePrint) clean() CollectionItem {
	out := sp
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

func (sp SporePrint) GeneticInfoAsParent() (GeneticParentInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (sp SporePrint) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return utils.Pointer(Generation(0)), utils.Pointer(Generation(0))
}

func (sp SporePrint) SourceType() string {
	return SporePrintSourceType
}

func (sp SporePrint) EntryTypeField() *string {
	return nil
}

func (sp SporePrint) altId() alternateCollectionId {
	return alternateCollectionId(sp.Id)
}

func (sp SporePrint) id() []byte {
	return sp.Id[:]
}

func (sp SporePrint) knownFruitable() bool {
	return false
}

func (sp SporePrint) prefix() string {
	return sporePrintIdPrefix
}

func (sp SporePrint) children(ctx context.Context) ([]geneticSource, error) {
	//TODO implement me
	panic("implement me")
}

func (sp SporePrint) CollectionName() string {
	return sporePrintCollectionName
}

func initializeSporePrints(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(sporePrintCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		newSimpleIndex("parent", "parent", false, false, false), // TODO: unique?
		newSimpleIndex("printDate", "printDate", true, false, false),
		newSimpleIndex("species", "species", false, false, false),
		newSimpleIndex("subSpecies", "subSpecies", false, true, false),
		newSimpleIndex("projects", "projects", false, false, false),
		//Notes (no index unless tags)
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	existingEntry := SporePrint{}
	testItem := SporePrint{
		Id:              exAltId,
		Parent:          &exAltId,
		Projects:        exProjects,
		PrintDate:       exampleTime,
		Species:         exampleSpecies,
		Subspecies:      exampleSubspecies,
		Pics:            exPics,
		Sale:            &exAltId,
		Disposed:        &exampleTime,
		MostRecentImage: utils.Pointer(exPics[0]),
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

type createSporePrintRequest struct {
	FruitId Base58Str `bson:"fruitId" json:"fruitId"`
	Notes   []Note
	Pics    []PicWithNotesLessLocation //"newPic-1"
}

func (upr createSporePrintRequest) reform() resolvedCreateSporePrintRequest {
	return resolvedCreateSporePrintRequest{
		FruitId: upr.FruitId,
		Notes:   upr.Notes,
		Pics: slices.Map(upr.Pics, func(i PicWithNotesLessLocation) PicWithNotes {
			return i.asPicWithNotes(nil)
		}),
	}
}

type resolvedCreateSporePrintRequest struct {
	FruitId Base58Str      `bson:"fruitId" json:"fruitId"`
	Notes   []Note         `json:"notes,omitempty"`
	Pics    []PicWithNotes `json:"pics,omitempty"`
}

func createSporePrintHandler(w http.ResponseWriter, r *http.Request) {
	data := createSporePrintRequest{}
	id := newAlternateCollectionId()
	b58Id := id.base58()
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
	// Get any images
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
			newFileNameWithPrefixPath, err := pics.SaveFile(r.Context(), fieldBytes, "sporePrint", string(b58Id), "img")
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
	}
	// CHECK THAT ALL NEW PICS EXIST
	// PROCESS ALL NEW PICS AND CONTAMS
	out := data.reform()
	for i, _ := range data.Pics {
		loc, exists := newPics[i]
		if !exists {
			http.Error(w, fmt.Sprintf("error, location for new picture index %d not found (should never happen)", i), http.StatusInternalServerError)
			return
		}
		out.Pics[i].Location = imageLocation(loc)
	}

	_, txErr := doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		db := ctx.Client().Database(dbName)
		parent := Fruit{}
		err = db.Collection(fruitsCollName).FindOne(ctx, bson.D{{"_id", id}}).Decode(&parent)
		if err != nil {
			http.Error(w, "failed to get parent fruit: "+err.Error(), http.StatusInternalServerError)
			return nil, err
		}
		now := unixTime(time.Now().UnixMilli())
		spid := alternateCollectionId(id)
		var mri *PicWithNotes = nil
		if len(out.Pics) > 0 {
			lastPic := out.Pics[len(out.Pics)-1]
			mri = &lastPic
		}
		_, err = db.Collection(sporePrintCollectionName).InsertOne(ctx, SporePrint{
			Id:              spid,
			Parent:          &parent.Id,
			Projects:        parent.Projects, // TODO: always retain projects?
			PrintDate:       now,
			Species:         parent.Species,
			Subspecies:      parent.SubSpecies,
			Pics:            out.Pics,
			MostRecentImage: mri,
			Notes:           out.Notes,
			LastUpdated:     now,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, err
		}
		// Update fruit with new print id
		err = parent.addSporePrint(ctx, spid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, err
		}
		return w.Write([]byte(b58Id))
	})
	if txErr != nil {
		// TODO: handle writing err
	}
}

type updateSporePrintRequest struct {
	Sale     *alternateCollectionId
	Disposed *unixTime
	Projects []string
	Notes    AllEntries[Note]
	Pics     SplitEntries[picWithNotesForm, PicWithNotesLessLocation]
}

func (upr updateSporePrintRequest) reform() resolvedUpdateSporePrintRequest {
	return resolvedUpdateSporePrintRequest{
		Sale:     upr.Sale,
		Disposed: upr.Disposed,
		Projects: upr.Projects,
		Notes:    upr.Notes,
		Pics: SplitEntries[picWithNotesForm, PicWithNotes]{
			Existing: upr.Pics.Existing,
			New: slices.Map(upr.Pics.New, func(i PicWithNotesLessLocation) PicWithNotes {
				return i.asPicWithNotes(nil)
			}),
		},
	}
}

type resolvedUpdateSporePrintRequest struct {
	Sale     *alternateCollectionId
	Disposed *unixTime
	Projects []string
	Notes    AllEntries[Note]
	Pics     SplitEntries[picWithNotesForm, PicWithNotes]
}

func updateSporePrintHandler(w http.ResponseWriter, r *http.Request) {
	data := updateSporePrintRequest{}
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
	// Get any images
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
		//var maxSize int64 = 32 << 20 // TODO: IS THIS OK? DO WE NEED THIS?
		//buf := bufio.NewReader(p)  // TODO: ensure close?
		//lmt := io.MultiReader(buf, io.LimitReader(p, maxSize-511)) // TODO: ensure close?
		//fieldBytes, err := io.ReadAll(lmt)
		//if err != nil && err != io.EOF {
		//	http.Error(w, err.Error(), http.StatusInternalServerError)
		//	return
		//}
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
			newFileNameWithPrefixPath, err := pics.SaveFile(r.Context(), fieldBytes, "sporePrint", string(b58Id), "img")
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
	}

	// CHECK THAT ALL NEW PICS EXIST
	// PROCESS ALL NEW PICS AND CONTAMS
	out := data.reform()
	for i, _ := range data.Pics.New {
		loc, exists := newPics[i]
		if !exists {
			http.Error(w, fmt.Sprintf("error, location for new picture index %d not found (should never happen)", i), http.StatusInternalServerError)
			return
		}
		out.Pics.New[i].Location = imageLocation(loc)
	}

	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(sporePrintCollectionName)
		// go get current sporePrint
		current := SporePrint{}
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
		// Compare SALE
		upd = setUnsetUnequalPointers("sale", out.Sale, current.Sale, upd)
		// Compare DISPOSED
		upd = setUnsetUnequalPointers("disposed", out.Disposed, current.Disposed, upd)
		// Compare PROJECTS
		upd = setProjectsIfUnequal(upd, out.Projects, current.Projects)
		// Do note changes
		mods, err := WithNotesUpdate(bson.D{}, out.Notes, current.Notes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return nil, nil
		}

		// Compare Images
		mods, err = WithExistingEntriesChange(mods, "pics", out.Pics.Existing, current.Pics, compareImageUpdate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		mods = append(mods, pushToArray("pics", out.Pics.New...)...)

		// write updates to db
		res := coll.FindOneAndUpdate(ctx, bson.D{{"_id", id}}, mods)
		if err = res.Err(); err != nil {
			http.Error(w, "failed to write update to db: "+err.Error(), http.StatusInternalServerError)
			return nil, nil
		}
		return w.Write([]byte(b58Id))
	})
	if err != nil {
		// TODO: WRITE ERR
	}
}

type importSporePrintRequest struct {
	Created    unixTime
	Species    string
	Subspecies *string
	Notes      []Note
	// pic as "img"
}

func importSporePrintHandler(w http.ResponseWriter, r *http.Request) {
	data := importSporePrintRequest{}
	id := newAlternateCollectionId()
	b58id := id.base58()
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
	// Try to get pic if exists
	picsSaved := []string{}
	defer func() {
		if err != nil {
			errDel := pics.DeleteFiles(r.Context(), picsSaved...)
			if err != nil {
				handleFileDeleteErr(errDel)
			}
		}
	}()
	now := unixTime(time.Now().UnixMilli())
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
		newFileNameWithPrefixPath, errr := pics.SaveFile(r.Context(), fieldBytes, "sporePrint", string(b58id), "img")
		if errr != nil {
			err = errr
			http.Error(w, "failed to save file: "+err.Error(), http.StatusBadRequest)
			return
		}
		picsSaved = append(picsSaved, newFileNameWithPrefixPath)

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
	pix := []PicWithNotes{}
	if importedPic != nil {
		pix = []PicWithNotes{*importedPic}
	}
	out := SporePrint{
		Id:              alternateCollectionId(id),
		PrintDate:       data.Created,
		Species:         data.Species,
		Subspecies:      data.Subspecies,
		Pics:            pix,
		MostRecentImage: importedPic,
		Notes:           data.Notes,
		LastUpdated:     now,
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(mainCollectionName)
		_, err = coll.InsertOne(ctx, out)
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
