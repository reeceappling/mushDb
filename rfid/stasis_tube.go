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

const StasisTubeSourceType = "stasisTube"

type StasisTube struct { // TODO: instructions somewhere?
	EntryType            string                  `bson:"entryType" json:"entryType"`
	Id                   MainCollectionId        `bson:"_id" json:"_id"`
	PcRun                alternateCollectionId   `bson:"pcRun" json:"pcRun"` // probably won't exist for pre-existing tubes (imports=="unknown")
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

func (s StasisTube) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := s
	err := decodeItem(&out, encoded)
	return out, err
}

func (s StasisTube) clean() CollectionItem {
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

func (s StasisTube) DbId() string {
	return s.Id.dbIdStr()
}

func (s StasisTube) projects() []string {
	return s.Projects
}

func (s StasisTube) GeneticInfoAsParent() (GeneticParentInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (s StasisTube) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return s.GenSinceSpore, s.GenSinceFruitOrSpore
}

func (s StasisTube) SourceType() string {
	return StasisTubeSourceType
}

func (s StasisTube) setTransferParent(ctx mongo.SessionContext, xfer Transfer) error {
	upd := pushToArray("transfersOut", xfer.Id)
	res, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(mainCollectionName).UpdateByID(ctx, s.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("Parent not found for transfer update. Should never happen!")
	}
	return nil
}

func (s StasisTube) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
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
		return errors.New("parent not found for transfer update. Should never happen") // TODO: MAKE VAR
	}
	return nil
}

func (s StasisTube) EntryTypeField() *string {
	return utils.Pointer(StasisTubeSourceType)
}

func (s StasisTube) CollectionName() string {
	return mainCollectionName
}

func (s StasisTube) id() []byte {
	return s.Id[:]
}

func (s StasisTube) knownFruitable() bool {
	return *s.KnownFruitable // TODO: ensure not nil
}

func (s StasisTube) children(ctx context.Context) ([]geneticSource, error) { // TODO: same as some others (slant, MSS, etc)
	return childrenOnlyToPlate(ctx, s.TransfersOut)
}

func (s StasisTube) idAsStr() string {
	return s.Id.dbIdStr()
}

func initializeStasisTubes(ctx context.Context) error {
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(mainCollectionName)
	// If test agar batch does not exist, then create it
	existingEntry := StasisTube{}
	testId := mainCollIdForint(idTestStasis)
	testItem := StasisTube{
		EntryType:            *existingEntry.EntryTypeField(),
		Id:                   testId,
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

type createStasisTubeRequest struct {
	PcRun      Base58Str
	WriteTagTo *string
}

func createStasisTubeHandler(w http.ResponseWriter, r *http.Request) {
	data := createStasisTubeRequest{}
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
	runId, err := data.PcRun.toAltCollectionId()
	if err != nil {
		http.Error(w, "failed to resolve agar batch ID: "+err.Error(), http.StatusBadRequest)
		return
	}
	now := unixTime(time.Now().UnixMilli())
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(mainCollectionName)
		_, err = coll.
			InsertOne(ctx, StasisTube{
				EntryType:    "stasisTube",
				Id:           id,
				PcRun:        runId,
				CreationDate: now,
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

type updateStasisTubeRequest struct {
	KnownFruitable *bool
	Sale           *alternateCollectionId
	Disposed       *unixTime
	Projects       []string
	Notes          AllEntries[Note]
	Images         SplitEntries[picWithNotesForm, PicWithNotesLessLocation]
	Contams        SplitEntries[contamForm, ContaminationLessLocation]
	WriteTagTo     *string
}

func (upr updateStasisTubeRequest) reform() resolvedUpdateStasisTubeRequest {
	return resolvedUpdateStasisTubeRequest{
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

type resolvedUpdateStasisTubeRequest struct {
	KnownFruitable *bool
	Sale           *alternateCollectionId
	Disposed       *unixTime
	Projects       []string
	Notes          AllEntries[Note]
	Images         SplitEntries[picWithNotesForm, PicWithNotes]
	Contams        SplitEntries[contamForm, Contamination]
	WriteTagTo     *string
}

func updateStasisTubeHandler(w http.ResponseWriter, r *http.Request) {
	data := updateStasisTubeRequest{}
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
		//var maxSize
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
			newFileNameWithPrefixPath, err := pics.SaveFile(r.Context(), fieldBytes, "stasisTube", string(b58Id), "img")
			if err != nil {
				http.Error(w, "failed to save new picture: "+err.Error(), http.StatusBadRequest)
				return
			}
			picsSaved = append(picsSaved, newFileNameWithPrefixPath)
			newPics[num] = newFileNameWithPrefixPath
		case "newContam":
			newFileNameWithPrefixPath, err := pics.SaveFile(r.Context(), fieldBytes, "stasisTube", string(b58Id), "contam")
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
		// go get current stasisTube
		current := StasisTube{}
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
		// Compare PROJECTS
		upd = setProjectsIfUnequal(upd, out.Projects, current.Projects)
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
		return w.Write([]byte(b58Id))
	})
	if err != nil {
		// TODO: WRITE ERR
	}
}

type importStasisTubeRequest struct {
	Created        unixTime
	Species        string
	Subspecies     *string
	KnownFruitable *bool
	Generation     *int
	// pic as "img"
	WriteTagTo *string
}

func importStasisTubeHandler(w http.ResponseWriter, r *http.Request) {
	data := importStasisTubeRequest{}
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
		}
		// Process file
		fieldBytes, err := multipartToImageBytes(p, w)
		if err != nil {
			// Already wrote
			return
		}
		newFileNameWithPrefixPath, errr := pics.SaveFile(r.Context(), fieldBytes, "stasisTube", string(b58id), "img")
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
	pix := []PicWithNotes{}
	if importedPic != nil {
		pix = []PicWithNotes{*importedPic}
	}
	out := StasisTube{
		EntryType:            "stasisTube",
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
		return w.Write([]byte(id.asBase58()))
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}
