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
	"slices"
)

const (
	FruitSourceType = "fruit"
	fruitsCollName  = "fruits"
)

type Fruit struct { // KnownFruitable is always true for this, // creation date field is id
	MainCollectionIdField // TODO: was alt
	CreationDateField     // This is harvest date
	SpeciesField
	SubspeciesOptionalField
	GenSporeField
	TransfersOutField                    // handled by new Transfer. Can only be clone to plate (sporeprint handled another way)
	Prints            []MainCollectionId `bson:"prints,omitempty" json:"prints,omitempty"` // TODO: used to be alt ids
	ParentTypeField
	// parent can be "store, outside, or a mainCollectionId (box/bag)"
	MainCollectionOptionalParentField /* TODO: new, make sure fixed everywhere */ // NONEXISTENT MEANS FROM STORE or outside???
	PicsField
	DisposedField
	MostRecentImageField
	NotesField
	LastUpdatedField
	//PermsField
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

func (f Fruit) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := f
	err := decodeItem(&out, encoded)
	return out, err
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

func (f Fruit) SourceType() string {
	return FruitSourceType
}

func (f Fruit) setTransferParent(ctx mongo.SessionContext, xfer Transfer) error {
	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
	if err != nil {
		return err
	}
	res, err := ctx.Client().Database(dbName).Collection(fruitsCollName).UpdateByID(ctx, f.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer
	}
	return nil
}

func (f Fruit) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
	// Transferring TO a fruit is not a thing
	return errors.New("fruits are invalid transfer children, must be created from a fruiter, or imported")
}

func (f Fruit) EntryTypeField() *string {
	return nil
}

//func (f Fruit) altId() AlternateCollectionId {
//	return AlternateCollectionId(f.Id)
//}
//
//func (f Fruit) id() []byte {
//	return f.Id[:]
//}

func (f Fruit) addSporePrint(ctx mongo.SessionContext, printId MainCollectionId) error {
	// update fruit
	res, err := ctx.Client().Database(dbName).Collection(fruitsCollName).UpdateByID(ctx, f.Id, pushToArrayInline("prints", printId))
	if err != nil {
		return err
	}
	if res.ModifiedCount != 1 {
		return errors.New("invalid result") // TODO: ok?
	}
	return nil
}

func (f Fruit) addSale(ctx mongo.SessionContext, printId AlternateCollectionId) error {
	// TODO; get rid of?
	// update fruit
	//res, err := f.Collection(ctx).UpdateByID(ctx, f.Id, pushToArrayInline("prints", printId)) // TODO: ADD A SALE IF POSSIBLE
	//if err != nil {
	//	return err
	//}
	//if res.ModifiedCount != 1 {
	//	return errors.New("invalid result") // TODO: ok?
	//}
	//return nil
	panic("implement me")
	return nil
}

func (f Fruit) CollectionName() string {
	return fruitsCollName
}

func initializeFruits(ctx context.Context) error {
	// Indices
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(fruitsCollName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		// TODO: creationDateIndexModel
		newSimpleIndex("creationDate", "creationDate", false, false, false), // TODO: this is harvest date
		newSimpleIndex("species", "species", false, false, false),
		newSimpleIndex("subSpecies", "subSpecies", false, true, false),
		// TODO: genSpore
		transfersOutIndexModel,
		// TODO: prints
		newSimpleIndex("parent", "parent", false, true, false),         // TODO: nil is store or outside?
		newSimpleIndex("parentType", "parentType", false, true, false), // TODO: nil is store or outside?
		newSimpleIndex("prints", "prints", false, true, false),
		newSimpleIndex("genSpore", "genSpore", true, true, false),
		//Pics (no index)
		newSimpleIndex("disposed", "disposed", false, true, false),
		//MostRecentImage (no index)
		//Notes (no index) (maybe later with tags?)
		lastUpdatedIndexModel,
		// TODO: projectsIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	existingEntry := Fruit{}
	testItem := Fruit{
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
	err = coll.FindOne(ctx, bson.D{{"_id", exAltId}}).Decode(&existingEntry)
	if err == nil {
		if reflect.DeepEqual(existingEntry, testItem) {
			return nil
		}
	}
	return testExistingEntry(ctx, coll, exAltId, testItem, existingEntry)
}

type createFruitRequest struct {
	ParentId    MainCollectionId
	ParentType  string
	HarvestDate unixTime
	NotesField
	Pics []PicWithNotesLessLocation // newPic-1
}

func (req createFruitRequest) reform() createFruitResolved {
	return createFruitResolved{
		MainCollectionParentField: MainCollectionParentField{req.ParentId},
		ParentType:                req.ParentType,
		HarvestDate:               req.HarvestDate,
		NotesField:                NotesField{req.Notes},
		PicsField: PicsField{sliceutils.Map(req.Pics, func(i PicWithNotesLessLocation) PicWithNotes {
			return PicWithNotes{
				Time:       i.Time,
				Location:   "",
				NotesField: NotesField{i.Notes},
			}
		})},
	}
}

type createFruitResolved struct {
	MainCollectionParentField        // TODO: used to be ParentId
	ParentType                string // TODO: swap out for normal parentType
	HarvestDate               unixTime
	NotesField
	PicsField // newPic-1
	//PermsField // TODO: FIX THIS
}

func createFruitHandler(w http.ResponseWriter, r *http.Request) { // TODO: DO FORMAT WITH DATA FIRST!
	data := createFruitRequest{}
	id, err := newMainCollectionId(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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

	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		db := ctx.Client().Database(dbName)
		// go get parent
		parent, err := GetMainCollectionItemInTxn(ctx, data.ParentId, nil)
		if err != nil {
			return DbTxnStdErr(w, "parent not found: "+err.Error(), http.StatusNotFound)
		}
		//parentPerms := parent.Permissions()
		//if err = parentPerms.ValidateUserCanWrite(ctx); err != nil {
		//	return DbTxnStdErr(w, "user not able to modify parent entry: "+err.Error(), http.StatusBadRequest)
		//}
		parentGenetics, err := parent.GeneticInfoAsParent()
		if err != nil {
			return DbTxnStdErr(w, "parent genetics error: "+err.Error(), http.StatusInternalServerError)
		}
		var mri *PicWithNotes = nil
		if len(out.Pics) > 0 {
			mri = &(out.Pics[len(out.Pics)-1])
		}
		if parentGenetics.Species == nil {
			return DbTxnStdErr(w, "parent species was nil", http.StatusInternalServerError)
		}
		toInsert := Fruit{
			MainCollectionIdField:             MainCollectionIdField{id},
			CreationDateField:                 CreationDateField{out.HarvestDate},
			SpeciesField:                      SpeciesField{*parentGenetics.Species},
			SubspeciesOptionalField:           parentGenetics.SubspeciesOptionalField,
			GenSporeField:                     parentGenetics.GenSporeField,
			ParentTypeField:                   ParentTypeField{&out.ParentType},
			MainCollectionOptionalParentField: MainCollectionOptionalParentField{&data.ParentId},
			PicsField:                         PicsField{out.Pics},
			MostRecentImageField:              MostRecentImageField{mri},
			NotesField:                        NotesField{out.Notes},
			LastUpdatedField:                  LastUpdatedField{unixTimeForNow()},
			//PermsField:                        PermsField{parentPerms},
		}
		// Write new fruit to db
		_, err = db.Collection(fruitsCollName).InsertOne(ctx, toInsert)
		if err != nil {
			return DbTxnStdErr(w, "error writing: "+err.Error(), http.StatusInternalServerError)
		}
		bs, err := json.Marshal(toInsert)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		return w.Write(bs)
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}

type updateFruitRequest struct {
	DisposedField
	Notes  AllEntries[Note]
	Images SplitEntries[picWithNotesForm, PicWithNotesLessLocation] //"newPic-1"
	//PermsField
}

func (upr updateFruitRequest) reform() resolvedUpdateFruitRequest {
	return resolvedUpdateFruitRequest{
		DisposedField: upr.DisposedField,
		Notes:         upr.Notes,
		Images: SplitEntries[picWithNotesForm, PicWithNotes]{
			Existing: upr.Images.Existing,
			New: sliceutils.Map(upr.Images.New, func(i PicWithNotesLessLocation) PicWithNotes {
				return i.asPicWithNotes(nil)
			}),
		},
		//PermsField: upr.PermsField,
	}
}

type resolvedUpdateFruitRequest struct {
	DisposedField
	Notes  AllEntries[Note]
	Images SplitEntries[picWithNotesForm, PicWithNotes] //"newPic-1"
	//PermsField
}

func (out resolvedUpdateFruitRequest) modsFor(existing Fruit) (bson.D, error) {
	return NewMods().
		updateDisposedIfNeeded(out.Disposed, existing.Disposed).
		updateNotesIfNeeded(out.Notes, existing.Notes).
		updatePicsIfNeeded(out.Images, existing.Pics).
		//updatePermsIfNeeded(out.Perms, existing.Perms).
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
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(fruitsCollName)
		// go get current plate
		existing := &Fruit{MainCollectionIdField: MainCollectionIdField{id}}
		err = Refresh(ctx, existing)
		if err != nil {
			return DbTxnStdErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		}

		//if err = minimalPermsBetween(existing.Perms, data.Perms).ValidateUserCanWrite(ctx); err != nil { // TODO: PERMS VALIDATION FOR UPDATE
		//	return DbTxnStdErr(w, "cannot modify: "+err.Error(), http.StatusUnauthorized)
		//}
		upd, err := out.modsFor(*existing)
		return handleUpdateMods(ctx, w, coll, existing, id, upd, err) // TODO: DO THIS EVERYWHERE!
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}

type importFruitRequest struct {
	ParentType string // "store" or "outside" // TODO: FIX?
	SpeciesField
	SubspeciesOptionalField
	NotesField
	//PermsField // TODO: new, use
	// image as "img"
}

func importFruitHandler(w http.ResponseWriter, r *http.Request) { // TODO: REDO?????
	data := importFruitRequest{}
	id, err := newMainCollectionId(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
			now := unixTimeForNow()
			importedPic = &PicWithNotes{
				Time:       now,
				Location:   imageLocation(newFileNameWithPrefixPath),
				NotesField: NotesField{[]Note{}},
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

	var gen = (*Generation)(utils.Pointer(0))
	pix := []PicWithNotes{}
	if importedPic != nil {
		pix = []PicWithNotes{*importedPic}
	}
	//sp, subsp, err := getSpeciesAndSubspecies(r.Context(), data.Species, data.SubSpecies)
	//if err != nil {
	//	http.Error(w, "failed to get species or subspecies: "+err.Error(), http.StatusInternalServerError) // TODO: normalize
	//}
	//finalPerms := minimalPermsBetween(data.Perms, sp, subsp)
	//if err = finalPerms.ValidateUserCanWrite(r.Context()); err != nil {
	//	http.Error(w, "user cannot write with the provided perms: "+err.Error(), http.StatusBadRequest)
	//	return
	//}
	now := unixTimeForNow()

	out := Fruit{
		MainCollectionIdField:   MainCollectionIdField{id},
		CreationDateField:       CreationDateField{unixTimeForNow()},
		SpeciesField:            data.SpeciesField,
		SubspeciesOptionalField: data.SubspeciesOptionalField,
		GenSporeField:           GenSporeField{gen},
		ParentTypeField:         ParentTypeField{&data.ParentType},
		PicsField:               PicsField{pix},
		MostRecentImageField:    MostRecentImageField{importedPic},
		NotesField:              NotesField{data.Notes},
		LastUpdatedField:        LastUpdatedField{now},
		//PermsField:              data.PermsField,
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		// TODO: insert map entry
		coll := ctx.Client().Database(dbName).Collection(fruitsCollName)
		_, err = coll.InsertOne(ctx, out)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		bs, err := json.Marshal(out)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		return w.Write(bs)
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}
