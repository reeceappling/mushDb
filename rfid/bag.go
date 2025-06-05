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

const BagSourceType = "bag"

type Bag struct {
	EntryType            string                  `bson:"entryType" json:"entryType"`
	Id                   MainCollectionId        `bson:"_id" json:"_id"`
	Substrate            alternateCollectionId   `bson:"recipe" json:"recipe"`
	PcRun                *alternateCollectionId  `bson:"pcRun,omitempty" json:"pcRun,omitempty"` // this may not exist for pre-existing bags
	FilterSize           string                  `bson:"filterSize" json:"filterSize"`
	CreationDate         unixTime                `bson:"creationDate" json:"creationDate"`
	GenSinceSpore        *Generation             `bson:"genSpore,omitempty" json:"genSpore,omitempty"`
	GenSinceFruitOrSpore *Generation             `bson:"genFruitOrSpore,omitempty" json:"genFruitOrSpore,omitempty"`
	SealDate             *unixTime               `bson:"sealDate,omitempty" json:"sealDate,omitempty"`             // set on transfer in
	KnownFruitable       *bool                   `bson:"knownFruitable,omitempty" json:"knownFruitable,omitempty"` // set on transfer in, or once fruited
	Species              *string                 `bson:"species,omitempty" json:"species,omitempty"`               // set on transfer in
	SubSpecies           *string                 `bson:"subSpecies,omitempty" json:"subSpecies,omitempty"`         // set on transfer in
	Innoc                *alternateCollectionId  `bson:"innoc,omitempty" json:"innoc,omitempty"`                   // Set on transfer in. Innoc from LC or grain jar only
	TransfersOut         []alternateCollectionId `bson:"transfersOut,omitempty" json:"transfersOut,omitempty"`     // Set on transfer out
	ParentType           *string                 `bson:"parentType,omitempty" json:"parentType,omitempty"`
	Parent               *MainCollectionId       `bson:"parent,omitempty" json:"parent,omitempty"` // Set on transfer in
	Projects             []string                `bson:"projects,omitempty" json:"projects,omitempty"`
	Pics                 []PicWithNotes          `bson:"pics,omitempty" json:"pics,omitempty"`                   // Updated independently
	Contaminations       []Contamination         `bson:"contamination,omitempty" json:"contamination,omitempty"` // Updated independently
	MostRecentImage      *PicWithNotes           `bson:"mostRecentImage,omitempty" json:"mostRecentImage,omitempty"`
	Flushes              []PicWithNotes          `bson:"flushes,omitempty" json:"flushes,omitempty"` // Updated independently
	Sale                 *alternateCollectionId  `bson:"sale,omitempty" json:"sale,omitempty"`
	Disposed             *unixTime               `bson:"disposed,omitempty" json:"disposed,omitempty"`
	Notes                []Note                  `bson:"notes,omitempty" json:"notes,omitempty"` // Updated independently
	LastUpdated          unixTime                `bson:"lastUpdated" json:"lastUpdated"`
}

func (b Bag) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := b
	err := decodeItem(&out, encoded)
	return out, err
}

func (b Bag) clean() CollectionItem {
	out := b
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

func (b Bag) DbId() string {
	return b.Id.dbIdStr()
}

func (b Bag) projects() []string {
	return b.Projects
}

func (b Bag) GeneticInfoAsParent() (GeneticParentInfo, error) {
	return GeneticParentInfo{
		Species:               b.Species,
		Subspecies:            b.SubSpecies,
		KnownFruitable:        b.KnownFruitable,
		GensSinceSpore:        b.GenSinceSpore,
		GensSinceFruitOrSpore: b.GenSinceFruitOrSpore,
	}, nil
}

func (b Bag) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return b.GenSinceSpore, b.GenSinceFruitOrSpore
}

func (b Bag) SourceType() string {
	return BagSourceType
}

func (b Bag) setTransferParent(ctx mongo.SessionContext, xfer Transfer) error {
	upd := pushToArray("transfersOut", xfer.Id)
	res, err := ctx.Client().Database(dbName).Collection(mainCollectionName).UpdateByID(ctx, b.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("Parent not found for transfer update. Should never happen!")
	}
	return nil
}

func (b Bag) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
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
		{"sealDate", xfer.LastUpdated},
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
		upd = append(upd, bson.E{"$set", bson.D{{"subSpecies", *parentInfo.Subspecies}}})
	}
	res, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(mainCollectionName).UpdateByID(ctx, b.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("Parent not found for transfer update. Should never happen!") // TODO: MAKE VAR
	}
	return nil
}

func (b Bag) EntryTypeField() *string {
	return utils.Pointer(BagSourceType)
}

func (b Bag) CollectionName() string {
	return mainCollectionName
}

func (b Bag) id() []byte {
	return b.Id[:]
}

func (b Bag) knownFruitable() bool {
	return *b.KnownFruitable // TODO: ensure not nil
}

func (b Bag) newFruit(ctx context.Context, pics []PicWithNotes, notes ...Note) error {
	return newFruitFromFruiter(ctx, b, pics, notes...)
}

func (b Bag) basicFruit() Fruit {
	var gen *Generation = nil
	if b.GenSinceSpore != nil {
		gen = utils.Pointer((*b.GenSinceSpore) + 1) // TODO: +1 here ok?
	}
	return Fruit{
		Id:            alternateCollectionId(primitive.NewObjectID()),
		Species:       *b.Species,
		SubSpecies:    b.SubSpecies,
		Parent:        &b.Id,
		GenSinceSpore: gen,
		ParentType:    utils.Pointer("bag"),
		Projects:      b.Projects,
		LastUpdated:   unixTime(time.Now().UnixMilli()),
	}
}

func (b Bag) children(ctx context.Context) ([]geneticSource, error) { // TODO: needed?
	return childrenAreOnlyFruits(ctx, b.TransfersOut)
}

func (b Bag) idAsStr() string {
	return b.Id.dbIdStr()
}

func initializeBags(ctx context.Context) error {
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(mainCollectionName)
	// If test agar batch does not exist, then create it
	existingEntry := Bag{}
	testId := mainCollIdForint(idTestBag)
	testItem := Bag{
		EntryType:            *existingEntry.EntryTypeField(),
		Id:                   testId,
		Substrate:            exAltId,
		PcRun:                &exAltId,
		FilterSize:           "5nm",
		CreationDate:         exampleTime,
		GenSinceSpore:        &exGenSinceSpore,
		GenSinceFruitOrSpore: &exGenSinceFruitSpore,
		SealDate:             &exampleTime,
		KnownFruitable:       exBool,
		Species:              &exampleSpecies,
		SubSpecies:           exampleSubspecies,
		Innoc:                &exAltId,
		TransfersOut:         exAlts,
		ParentType:           &exParentType,
		Parent:               &exPlate,
		Projects:             exProjects,
		Pics:                 exPics,
		Contaminations:       exContams,
		MostRecentImage:      &exPics[0],
		Flushes:              exPics,
		Sale:                 &exAltId,
		Disposed:             &exampleTime,
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

type createBagRequest struct {
	Recipe       Base58Str // substrate recipe
	PcRun        Base58Str
	FilterSize   string
	CreationDate unixTime
	Notes        []Note `json:"notes,omitempty"`
	WriteTagTo   *string
}

func createBagHandler(w http.ResponseWriter, r *http.Request) {
	data := createBagRequest{}
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
	pcRunId, err := data.PcRun.toAltCollectionId()
	if err != nil {
		http.Error(w, "failed to resolve substrate recipe ID: "+err.Error(), http.StatusBadRequest)
		return
	}
	now := unixTime(time.Now().UnixMilli())
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(mainCollectionName)
		_, err := coll.InsertOne(ctx, Bag{
			EntryType:    "bag",
			Id:           id,
			Substrate:    recipeId,
			PcRun:        &pcRunId,
			FilterSize:   data.FilterSize,
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

type updateBagRequest struct {
	KnownFruitable *bool                  `json:"knownFruitable,omitempty"`
	Sale           *alternateCollectionId `json:"sale,omitempty"`
	Disposed       *unixTime              `json:"disposed,omitempty"`
	Projects       []string               `json:"projects,omitempty"`
	Notes          AllEntries[Note]
	Images         SplitEntries[picWithNotesForm, PicWithNotesLessLocation] //"newPic-1"
	Contams        SplitEntries[contamForm, ContaminationLessLocation]      //"newContam-1"
	Flushes        SplitEntries[picWithNotesForm, PicWithNotesLessLocation] //"newFlush-1"
	WriteTagTo     *string
}

func (upr updateBagRequest) reform() resolvedUpdateBagRequest {
	return resolvedUpdateBagRequest{
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

type resolvedUpdateBagRequest struct {
	KnownFruitable *bool
	Sale           *alternateCollectionId
	Disposed       *unixTime
	Projects       []string
	Notes          AllEntries[Note]
	Images         SplitEntries[picWithNotesForm, PicWithNotes]
	Contams        SplitEntries[contamForm, Contamination]
	Flushes        SplitEntries[picWithNotesForm, PicWithNotes]
}

func updateBagHandler(w http.ResponseWriter, r *http.Request) {
	data := updateBagRequest{}
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
			newFileNameWithPrefixPath, err := pics.SaveFile(r.Context(), fieldBytes, "bag", string(b58Id), "img")
			if err != nil {
				http.Error(w, "failed to save new picture: "+err.Error(), http.StatusBadRequest)
				return
			}
			picsSaved = append(picsSaved, newFileNameWithPrefixPath)
			newPics[num] = newFileNameWithPrefixPath
		case "newContam":
			newFileNameWithPrefixPath, err := pics.SaveFile(r.Context(), fieldBytes, "bag", string(b58Id), "contam")
			if err != nil {
				http.Error(w, "failed to save new contamination: "+err.Error(), http.StatusBadRequest)
				return
			}
			picsSaved = append(picsSaved, newFileNameWithPrefixPath)
			newContams[num] = newFileNameWithPrefixPath
		case "newFlush":
			newFileNameWithPrefixPath, err := pics.SaveFile(r.Context(), fieldBytes, "bag", string(b58Id), "flush")
			if err != nil {
				http.Error(w, "failed to save new flush: "+err.Error(), http.StatusBadRequest)
				return
			}
			picsSaved = append(picsSaved, newFileNameWithPrefixPath)
			newFlushes[num] = newFileNameWithPrefixPath
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
		current := Bag{}
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

type importBagRequest struct {
	SealDate       unixTime  // Also becomes creation date
	Recipe         Base58Str // Substrate recipe
	FilterSize     string
	Species        string
	Subspecies     *string
	Generation     *int
	KnownFruitable *bool
	WriteTagTo     *string
	// image as "img"
}

func importBagHandler(w http.ResponseWriter, r *http.Request) { // TODO: COPY FRUITING CHAMBER
	data := importBagRequest{}
	id, err := generateMainCollectionId(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	b58id := id.asBase58()
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
	var importedPic *PicWithNotes = nil
	dataProcessed := false
	filesProcessed := 0
	for {
		fileName := p.FileName()
		defer p.Close()
		if isFile := fileName != ""; isFile {
			if filesProcessed == 1 {
				// TODO: ERROR! DONT CREATE MORE THAN 1 IMAGE!
			}
			// Process file
			fieldBytes, err := multipartToImageBytes(p, w)
			if err != nil {
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
		// TODO: THIS!
	}
	pix := []PicWithNotes{}
	if importedPic != nil {
		pix = []PicWithNotes{*importedPic}
	}
	out := Bag{
		EntryType:            "bag",
		Id:                   id,
		Substrate:            recipeId,
		PcRun:                nil,
		FilterSize:           data.FilterSize,
		CreationDate:         data.SealDate,
		GenSinceSpore:        gen,
		GenSinceFruitOrSpore: gen,
		SealDate:             &data.SealDate,
		KnownFruitable:       data.KnownFruitable,
		Species:              &data.Species,
		SubSpecies:           data.Subspecies,
		Pics:                 pix,
		Contaminations:       nil,
		MostRecentImage:      importedPic,
		Flushes:              nil,
		Sale:                 nil,
		Disposed:             nil,
		Notes:                nil,
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
