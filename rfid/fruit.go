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
	"slices"
	"time"
)

// TODO: required for
// TODO: newSporeSwab, newSporePrint, clone(plate, slant)
// TODO: newSporeSwab

type Fruit struct { // KnownFruitable is always true for this, // creation date field is id
	MainCollectionIdField   `bson:"inline"`
	CreationDateField       `bson:"inline"` // This is harvest date
	SpeciesField            `bson:"inline"`
	SubspeciesOptionalField `bson:"inline"`
	GenSporeField           `bson:"inline"`
	TransfersOutField       `bson:"inline"`    // handled by new Transfer. Can only be clone to plate (sporeprint handled another way)
	Prints                  []MainCollectionId `bson:"prints,omitempty" json:"prints,omitempty"`
	ParentTypeField         `bson:"inline"`
	// parent can be "store, outside, or a mainCollectionId (box/bag)"
	MainCollectionOptionalParentField `bson:"inline"` // NONEXISTENT MEANS FROM STORE or outside
	PicsField                         `bson:"inline"`
	DisposedField                     `bson:"inline"`
	MostRecentImageField              `bson:"inline"`
	NotesField                        `bson:"inline"`
	LastUpdatedField                  `bson:"inline"`
	AclField                          `bson:"inline"`
}

func (f Fruit) CanTransferTo(dst geneticSource) error {
	if !slices.Contains([]string{PlateSourceType, SlantSourceType, StasisTubeSourceType}, dst.SourceType()) {
		return errors.New("fruit cannot transfer to " + dst.SourceType() + " via this endpoint")
	}
	return nil
}
func (f Fruit) Innoculatable() bool {
	return false
}

func (f Fruit) GeneticInfoAsParent() (GeneticParentInfo, error) {
	return GeneticParentInfo{
		SpeciesOptionalField:    SpeciesOptionalField{&f.Species},
		SubspeciesOptionalField: f.SubspeciesOptionalField,
		KnownFruitableField:     KnownFruitableField{utils.Pointer(true)},
		GenerationsFields: GenerationsFields{
			GenSporeField:        GenSporeField{f.GenSinceSpore},
			GenSinceFruitOrSpore: utils.Pointer(Generation(0)),
		},
	}, nil
}

func (f Fruit) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return f.GenSinceSpore, (*Generation)(utils.Pointer(0))
}

//func (f Fruit) setTransferParent(ctx context.Context, xfer Transfer) error {
//	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(FruitsCollName)
//	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
//	if err != nil {
//		return err
//	}
//	res, err := coll.UpdateByID(ctx, f.Id, upd)
//	if err != nil {
//		return err
//	}
//	if res.ModifiedCount == 0 {
//		return ErrNoParentModifiedForTransfer
//	}
//	return nil
//}

func (f Fruit) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
	// Transferring TO a fruit is not a thing
	return errors.New("fruits are invalid transfer children, must be created from a fruiter, or imported")
}

//func (f Fruit) altId() AlternateCollectionId {
//	return AlternateCollectionId(f.Email)
//}
//
//func (f Fruit) id() []byte {
//	return f.Email[:]
//}

func (f Fruit) addSporePrint(ctx mongo.SessionContext, printId MainCollectionId) error {
	// update fruit
	res, err := mongo.SessionFromContext(ctx).Client().Database(dbName).Collection(FruitsCollName).UpdateByID(ctx, f.Id, pushToArrayInline("prints", printId))
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("invalid result") // TODO: ok?
	}
	return nil
}

func (f Fruit) addSale(ctx mongo.SessionContext, printId AlternateCollectionId) error {
	// TODO; get rid of?
	// update fruit
	//res, err := f.Collection(ctx).UpdateByID(ctx, f.Email, pushToArrayInline("prints", printId)) // TODO: ADD A SALE IF POSSIBLE
	//if err != nil {
	//	return err
	//}
	//if res.ModifiedCount != 1 {
	//	return errors.New("invalid result") // TODO: ok?
	//}
	//return nil
	return errors.New("not implemented, implement me")
}

func initializeFruits(ctx context.Context) error {
	// Indices
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(FruitsCollName)
	err := createIndexes(ctx, coll,
		[]mongo.IndexModel{
			// TODO: creationDateIndexModel
			newSimpleIndex("creationDate", "creationDate", false, false, false), // TODO: this is harvest date
			newSimpleIndex("species", "species", false, false, false),
			newSimpleIndex("subspecies", "subspecies", false, true, false),
			transfersOutIndexModel,
			newSimpleIndex("parent", "parent", false, true, false),         // TODO: nil is store or outside?
			newSimpleIndex("parentType", "parentType", false, true, false), // TODO: nil is store or outside?
			//TODO: newSimpleIndex("prints", "prints", false, true, false),
			//newSimpleIndex("genSpore", "genSpore", true, true, false),
			//Pics (no index)
			//newSimpleIndex("disposed", "disposed", false, true, false),
			//MostRecentImage (no index)
			//Notes (no index) (maybe later with tags?)
			projectsIndexModel,
			lastUpdatedIndexModel,
		})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	testItem := &Fruit{
		MainCollectionIdField:             MainCollectionIdField{exFruitId},
		CreationDateField:                 CreationDateField{exampleTime},
		SpeciesField:                      SpeciesField{testEntryStringId},
		SubspeciesOptionalField:           SubspeciesOptionalField{utils.Pointer(testEntryStringId)},
		GenSporeField:                     GenSporeField{&exGenSinceSpore},
		TransfersOutField:                 TransfersOutField{exAlts},
		Prints:                            []MainCollectionId{exSporePrint},
		ParentTypeField:                   ParentTypeField{&exParentType},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&exPlate},
		PicsField:                         PicsField{exPics},
		DisposedField:                     DisposedField{&exampleTime},
		MostRecentImageField:              MostRecentImageField{&exPics[0]},
		NotesField:                        NotesField{exampleNotes()},
		LastUpdatedField:                  LastUpdatedField{exampleTime},
	}
	return addTestMainEntries(ctx, testItem)
}

type createFruitRequest struct {
	ParentId   MainCollectionId
	ParentType string
	NotesField
	Pics []PicWithNotesLessLocation // newPic-1
	PermsOnRequest
}

func (req createFruitRequest) reform() createFruitResolved {
	return createFruitResolved{
		MainCollectionParentField: MainCollectionParentField{req.ParentId},
		ParentType:                req.ParentType,
		NotesField:                NotesField{req.Notes},
		PicsField: PicsField{sliceutils.Map(req.Pics, func(i PicWithNotesLessLocation) PicWithNotes {
			return newPicWithNotes(i.Time, i.Notes, "")
		})},
	}
}

type createFruitResolved struct {
	MainCollectionParentField        // TODO: used to be ParentId
	ParentType                string // TODO: swap out for normal parentType
	NotesField
	PicsField // newPic-1
}

func createFruitHandler(w http.ResponseWriter, r *http.Request) { // TODO: DO FORMAT WITH DATA FIRST!
	data := createFruitRequest{}
	id := NextMainCollectionId()
	b58Id := id.asBase58()
	defer r.Body.Close()
	newPics, _, _, err := fullMultipartWithNoBreaks(w, r, "fruit", &data, b58Id)
	if err != nil {
		// Already wrote
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
	ctx := r.Context()
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	parent, err := typeForSource(data.ParentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = db.Collection(parent.CollectionName()).FindOne(ctx, bsonFindFilter("_id", data.ParentId)).Decode(&parent)
	if err != nil {
		http.Error(w, "failed to get parent: "+err.Error(), http.StatusInternalServerError)
		return
	}
	parentGenetics, err := parent.GeneticInfoAsParent()
	if err != nil {
		http.Error(w, "parent genetics error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var mri *PicWithNotes = nil
	if len(out.Pics) > 0 {
		mri = &(out.Pics[len(out.Pics)-1])
	}
	if parentGenetics.Species == nil {
		http.Error(w, "parent species was nil", http.StatusInternalServerError)
		return
	}
	toInsert := &Fruit{
		MainCollectionIdField:             MainCollectionIdField{id},
		CreationDateField:                 CreationDateField{unixTime(time.Now().UnixMilli())},
		SpeciesField:                      SpeciesField{*parentGenetics.Species},
		SubspeciesOptionalField:           parentGenetics.SubspeciesOptionalField,
		GenSporeField:                     parentGenetics.GenSporeField,
		ParentTypeField:                   ParentTypeField{&out.ParentType},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&data.ParentId},
		PicsField:                         PicsField{out.Pics},
		MostRecentImageField:              MostRecentImageField{mri},
		NotesField:                        NotesField{out.Notes},
		LastUpdatedField:                  LastUpdatedField{unixTimeForNow()},
		AclField:                          AclField{parent.Permissions()},
	}
	finishCreateMainCollectionEntry(ctx, toInsert, w)
}

type updateFruitRequest struct {
	DisposedField
	NotesUpdateField
	Images SplitEntries[picWithNotesForm, PicWithNotesLessLocation] //"newPic-1"
	PermsOnRequest
}

func (upr updateFruitRequest) reform() resolvedUpdateFruitRequest {
	return resolvedUpdateFruitRequest{
		DisposedField:    upr.DisposedField,
		NotesUpdateField: upr.NotesUpdateField,
		Images: SplitEntries[picWithNotesForm, PicWithNotes]{
			Existing: upr.Images.Existing,
			New: sliceutils.Map(upr.Images.New, func(i PicWithNotesLessLocation) PicWithNotes {
				return i.asPicWithNotes(nil)
			}),
		},
		PermsOnRequest: upr.PermsOnRequest,
	}
}

type resolvedUpdateFruitRequest struct {
	DisposedField
	NotesUpdateField
	Images SplitEntries[picWithNotesForm, PicWithNotes] //"newPic-1"
	PermsOnRequest
}

func (req resolvedUpdateFruitRequest) modsFor(existing *Fruit, aclField AclField) (bson.D, error) {
	return NewMods().
		updateDisposedIfNeeded(req, existing).
		updateNotesIfNeeded(req, existing).
		updatePicsIfNeeded(req.Images, existing.Pics).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateFruitHandler(w http.ResponseWriter, r *http.Request) {
	data := updateFruitRequest{}
	b58Id := Base58Str(r.PathValue("id"))
	id, err := b58Id.toMainCollectionId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	newPics, _, _, err := fullMultipartWithNoBreaks(w, r, "fruit", &data, b58Id)
	if err != nil {
		// Already wrote
		return
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
	ctx := r.Context()
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(FruitsCollName)
	// go get current plate
	existing := &Fruit{MainCollectionIdField: MainCollectionIdField{id}}
	err = Refresh(ctx, db, existing)
	if err != nil {
		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishMainCollItemUpdate(ctx, w, coll, out.modsFor, existing, data.PermsOnRequest)
}

type importFruitRequest struct {
	// TODO: REMOVED ParentType string "store" or "outside" // TODO: FIX?
	SpeciesField
	SubspeciesOptionalField
	NotesField
	PermsOnRequest
	// image as "img"
}

func importFruitHandler(w http.ResponseWriter, r *http.Request) { // TODO: REDO?????
	data := importFruitRequest{}
	id := NextMainCollectionId()
	b58id := id.asBase58()
	r.Body = http.MaxBytesReader(w, r.Body, maxMultipartRequestSize)
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
			if filesProcessed > 0 {
				http.Error(w, "imported fruits can only be sent with one image: "+err.Error(), http.StatusBadRequest)
				return
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
			now := unixTimeForNow()
			importedPic = &PicWithNotes{
				PicWithNotesLessLocation: newPicWithNotesLessLocation(now, []Note{}),
				Location:                 imageLocation(newFileNameWithPrefixPath),
			}
			filesProcessed++
		} else {
			// Process text (or object)
			bs, errr := io.ReadAll(p)
			if errr != nil {
				err = errr
				http.Error(w, "unable to read Data from form: "+err.Error(), http.StatusBadRequest)
				return
			}
			// PARSE INTO CORRECT DATA FORMAT
			err = json.Unmarshal(bs, &data)
			if err != nil {
				http.Error(w, "unable to unmarshal json form Data: "+err.Error(), http.StatusBadRequest)
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
		err = errors.New("no non-image Data found on form request")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var gen = (*Generation)(utils.Pointer(0))
	pix := []PicWithNotes{}
	if importedPic != nil {
		pix = []PicWithNotes{*importedPic}
	}
	// TODO: PERMS ON REQUEST INSTEAD?????
	user, err := GetAuthInfo(r.Context())
	if err != nil {
		http.Error(w, "failed to get auth info: "+err.Error(), http.StatusUnauthorized)
		return
	}
	sp, subsp, err := getSpeciesAndSubspecies(r.Context(), data.Species, data.SubSpecies)
	if err != nil {
		http.Error(w, "failed to get species or subspecies: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var finalPerms *ACL = nil
	if subsp != nil {
		finalPerms = subsp.DefaultAcl.Clone()
	} else {
		finalPerms = sp.DefaultAcl.Clone()
	}
	// Add user to the acl as a writer (since they own this?)
	finalPerms.Users[user.Email] = true

	// TODO: Even if user cannot write, allow them to import???
	now := unixTimeForNow()
	ctx := r.Context()
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(FruitsCollName)
	toInsert := &Fruit{
		MainCollectionIdField:   MainCollectionIdField{id},
		CreationDateField:       CreationDateField{unixTimeForNow()},
		SpeciesField:            data.SpeciesField,
		SubspeciesOptionalField: data.SubspeciesOptionalField,
		GenSporeField:           GenSporeField{gen},
		ParentTypeField:         ParentTypeField{nil},
		PicsField:               PicsField{pix},
		MostRecentImageField:    MostRecentImageField{importedPic},
		NotesField:              NotesField{data.Notes},
		LastUpdatedField:        LastUpdatedField{now},
		AclField:                AclField{finalPerms},
	}
	finishImportMainCollectionEntry(ctx, coll, toInsert, data.PermsOnRequest, w)
}
