package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/reeceappling/goUtils/v2/utils"
	sliceutils "github.com/reeceappling/goUtils/v2/utils/slices"
	"github.com/reeceappling/mushDb/api/env"
	"github.com/reeceappling/mushDb/api/pics"
	"github.com/reeceappling/mushDb/api/request"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"slices"
)

// required for
// newSporeSwab, newSporePrint, clone(plate, slant)

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
func (f Fruit) Innoculatable() error {
	return errors.New("fruits are not innoculatable")
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

func (f Fruit) setTransferChild(_ mongo.SessionContext, _ Transfer, _ geneticSource) error {
	// Transferring TO a fruit is not a thing
	return errors.New("fruits are invalid transfer children, must be created from a fruiter, or imported")
}

func (f Fruit) addSporePrint(ctx mongo.SessionContext, printId MainCollectionId) error {
	ctx, now := request.UnixTimeInTxn(ctx)
	// update fruit
	upd, err := NewMods().Push("prints", printId).withLastUpdated(now).Finalized()
	if err != nil {
		return err
	}
	res, err := mongo.SessionFromContext(ctx).Client().Database(dbName).Collection(FruitsCollName).UpdateByID(ctx, f.Id, upd) // TODO: validate working fine!
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("invalid result") // TODO: ok?
	}
	return nil
}
func (f Fruit) createSporePrintInTxn(ctx mongo.SessionContext, pics PicsField, notes NotesField, id MainCollectionId) (*SporePrint, error) {
	ctx, now := request.UnixTimeInTxn(ctx)

	var mri *PicWithNotes = nil
	if len(pics.Pics) > 0 {
		lastPic := pics.Pics[len(pics.Pics)-1]
		mri = &lastPic
	}
	toInsert := SporePrint{
		MainCollectionIdField:             MainCollectionIdField{id},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&f.Id},
		CreationDateField:                 CreationDateField{now},
		SpeciesField:                      f.SpeciesField,
		SubspeciesOptionalField:           f.SubspeciesOptionalField,
		PicsField:                         pics,
		MostRecentImageField:              MostRecentImageField{mri},
		NotesField:                        notes,
		LastUpdatedField:                  LastUpdatedField{now},
		// Do not check permissions, just pass parent perms to child
		AclField: f.AclField,
	}
	db := mongo.SessionFromContext(ctx).Client().Database(dbName)
	err := addToIdMapCollection(ctx, &toInsert)
	if err != nil {
		return nil, err
	}
	// Update fruit with new print id
	err = f.addSporePrint(ctx, id)
	if err != nil {
		return nil, errors.Join(errors.New("failed to add spore print to parent fruit"), err)
	}
	_, err = db.Collection(SporePrintCollectionName).InsertOne(ctx, toInsert)
	if err != nil {
		return nil, errors.Join(errors.New("failed to insert new spore print"), err)
	}
	return &toInsert, nil
}

func (f Fruit) createSporeSwabInTxn(ctx mongo.SessionContext, notes NotesField, id MainCollectionId) (*SporeSwab, error) {
	ctx, now := request.UnixTimeInTxn(ctx)
	// writeTagTo is done after this func is called
	toInsert := SporeSwab{
		MainCollectionIdField:             MainCollectionIdField{id},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&f.Id},
		ParentTypeField:                   ParentTypeField{utils.Pointer("fruit")},
		CreationDateField:                 CreationDateField{now},
		SpeciesField:                      f.SpeciesField,
		SubspeciesOptionalField:           f.SubspeciesOptionalField,
		NotesField:                        notes,
		LastUpdatedField:                  LastUpdatedField{now},
		// Do not check permissions, just pass parent perms to child
		AclField: f.AclField,
	}
	db := mongo.SessionFromContext(ctx).Client().Database(dbName)
	err := addToIdMapCollection(ctx, &toInsert)
	if err != nil {
		return nil, err
	}
	// Update fruit with new print id
	//err = parent.addSporeSwab(ctx, id)
	//if err != nil {
	//	return nil, errors.Join(errors.New("failed to add spore swab to parent fruit"), err)
	//}
	// TODO: add transfer to parent for swab! should swabs have their own field on fruits?
	_, err = db.Collection(SporeSwabCollectionName).InsertOne(ctx, toInsert)
	if err != nil {
		return nil, errors.Join(errors.New("failed to insert new spore print"), err)
	}
	return &toInsert, nil
}

//func (f Fruit) addSale(ctx mongo.SessionContext, printId AlternateCollectionId) error {
//	return errors.New("not implemented, implement me")
//}

func initializeFruits(ctx context.Context) error {
	// Indices
	db := DbFrom(ctx)
	coll := db.Collection(FruitsCollName)
	err := createIndexes(ctx, coll,
		[]mongo.IndexModel{
			creationDateIndexModel, // TODO: this is harvest date
			newSimpleIndex("species", "species", false, false, false),
			newSimpleIndex("subspecies", "subspecies", false, true, false),
			//transfersOutIndexModel,
			newSimpleIndex("parent", "parent", false, true, false),         // TODO: nil is store or outside?
			newSimpleIndex("parentType", "parentType", false, true, false), // TODO: nil is store or outside?
			//newSimpleIndex("prints", "prints", false, true, false), // Prints shouldnt need indexing
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
	// Add test entries if dev
	return env.IfNotProd(ctx, func() error {
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
	})
}

type createFruitRequest struct {
	ParentId   MainCollectionId
	ParentType string
	NotesField
	Pics            []PicWithNotesLessLocation // newPic-1
	PermsOnRequest  `json:"acl"`
	WriteTagToField // TODO: add to typescript or delete!
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
	MainCollectionParentField
	ParentType string // TODO: swap out for normal parentType?
	NotesField
	PicsField // newPic-1
}

func createFruitHandler(w http.ResponseWriter, r *http.Request) {
	data := createFruitRequest{}
	id := NextMainCollectionId()
	b58Id := id.AsBase58()
	defer r.Body.Close()
	newPics, _, _, err := fullMultipartWithNoBreaks(w, r, &data, b58Id)
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
		out.Pics[i].Location = ImageLocation(loc)
	}
	ctx := r.Context()
	db := DbFrom(ctx)
	parent, err := typeForEntryType(data.ParentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// TODO: NEXT LINE IS FAILING BECAUSE IT CANNOT UNMARSHAL A MAIN COLLECTION ITEM! UNSURE IF STILL FAILING!
	// parent is not a pointer because the interface's underlying types are each pointers
	err = db.Collection(parent.CollectionName()).FindOne(ctx, BsonFindFilter(IDfld, data.ParentId)).Decode(parent)
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
	ctx, now := request.UnixTime(ctx)
	toInsert := &Fruit{
		MainCollectionIdField:             MainCollectionIdField{id},
		CreationDateField:                 CreationDateField{now},
		SpeciesField:                      SpeciesField{*parentGenetics.Species},
		SubspeciesOptionalField:           parentGenetics.SubspeciesOptionalField,
		GenSporeField:                     parentGenetics.GenSporeField,
		ParentTypeField:                   ParentTypeField{&out.ParentType},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&data.ParentId},
		PicsField:                         PicsField{out.Pics},
		MostRecentImageField:              MostRecentImageField{mri},
		NotesField:                        NotesField{out.Notes},
		LastUpdatedField:                  LastUpdatedField{now},
		AclField:                          AclField{parent.Permissions()},
	}
	err = writeRfidTagIfNecessary(ctx, data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	finishCreateMainCollectionEntry(ctx, toInsert, w)
}

type updateFruitRequest struct {
	DisposedField
	NotesUpdateField
	Images         SplitEntries[picWithNotesForm, PicWithNotesLessLocation] //"newPic-1"
	PermsOnRequest `json:"acl"`
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
		updateMostRecentImageIfNeeded(existing.MostRecentImage, loadMriPics(&req.Images, nil, nil)).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateFruitHandler(w http.ResponseWriter, r *http.Request) {
	idStr, err := UrlDecodeString(r.PathValue("id"))
	if err != nil {
		http.Error(w, "failed to url decode string: "+err.Error(), http.StatusInternalServerError)
		return
	}
	mainCollId, err := StandardizeMainCollectionId(idStr)
	if err != nil {
		http.Error(w, "failed to standardize main collection id: "+err.Error(), http.StatusBadRequest)
		return
	}
	b58Id := mainCollId.AsBase58()
	data := updateFruitRequest{}
	id := *mainCollId
	newPics, _, _, err := fullMultipartWithNoBreaks(w, r, &data, b58Id)
	if err != nil {
		// Already written
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
		out.Images.New[i].Location = ImageLocation(loc)
	}
	ctx := r.Context()
	existing := &Fruit{MainCollectionIdField: MainCollectionIdField{id}} // TODO: FAILING HERE! (double-check, unsure when this was failing as of 7/25/26)
	err = DbFrom(ctx).Collection(FruitsCollName).FindOne(ctx, BsonFindFilter(IDfld, id)).Decode(existing)
	if err != nil {
		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishMainCollItemUpdate(ctx, w, out.modsFor, existing, data.PermsOnRequest)
}

type importFruitRequest struct {
	ParentType string // "store" or "outside" // TODO: ? FIX?
	SpeciesField
	SubspeciesOptionalField
	NotesField
	// image as "img"
	WriteTagToField // TODO: add to typescript or remove!
}

func importFruitHandler(w http.ResponseWriter, r *http.Request) {
	ctx, now := request.UnixTime(r.Context())
	data := importFruitRequest{}
	id := NextMainCollectionId()
	b58id := id.AsBase58()
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
			importedPic = &PicWithNotes{
				PicWithNotesLessLocation: newPicWithNotesLessLocation(now, []Note{}),
				Location:                 ImageLocation(newFileNameWithPrefixPath),
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
	finalPerms, err := ImportFinalPerms(ctx, data.Species, data.Subspecies)
	if err != nil {
		http.Error(w, "failed to get species and/or subspecies: "+err.Error(), http.StatusInternalServerError)
		return
	}
	toInsert := &Fruit{
		MainCollectionIdField:   MainCollectionIdField{id},
		CreationDateField:       CreationDateField{now},
		SpeciesField:            data.SpeciesField,
		SubspeciesOptionalField: data.SubspeciesOptionalField,
		GenSporeField:           GenSporeField{gen},
		ParentTypeField:         ParentTypeField{&data.ParentType}, // TODO: is this ok?
		PicsField:               PicsField{pix},
		MostRecentImageField:    MostRecentImageField{importedPic},
		NotesField:              NotesField{data.Notes},
		LastUpdatedField:        LastUpdatedField{now},
		AclField:                AclField{finalPerms},
	}
	err = writeRfidTagIfNecessary(ctx, data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	finishImportMainCollectionEntry(ctx, toInsert, w)
}

func FruitFromSourceInTxn(ctx mongo.SessionContext, parent geneticSource) (*Fruit, error) {
	id := NextMainCollectionId()
	ctx, now := request.UnixTimeInTxn(ctx)

	genetics, err := parent.GeneticInfoAsParent()
	if err != nil {
		return nil, err
	}
	if genetics.Species == nil {
		return nil, errors.New("parent not innoculated")
	}
	parentId := parent.DbId()
	newFruit := &Fruit{
		MainCollectionIdField:             MainCollectionIdField{id},
		CreationDateField:                 CreationDateField{now},
		SpeciesField:                      SpeciesField{*genetics.Species},
		SubspeciesOptionalField:           SubspeciesOptionalField{genetics.Subspecies},
		GenSporeField:                     GenSporeField{genetics.GenSinceSpore.Next()},
		TransfersOutField:                 TransfersOutField{},
		Prints:                            nil,
		ParentTypeField:                   ParentTypeField{utils.Pointer(parent.SourceType())},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&parentId},
		PicsField:                         PicsField{}, // TODO: ????
		DisposedField:                     DisposedField{},
		MostRecentImageField:              MostRecentImageField{}, // TODO: ????
		NotesField:                        NotesField{},           // TODO: ????
		LastUpdatedField:                  LastUpdatedField{now},
		AclField:                          parent.Permissions().AsField(),
	}

	// add fruit to ids collection and its own collection
	err = createMainCollectionEntryInTxn(ctx, newFruit)
	if err != nil {
		return nil, err
	}
	xferId := newAlternateCollectionId()
	xfer := Transfer{ // TODO: ptr?
		AlternateCollectionIdField: AlternateCollectionIdField{xferId},
		From:                       parentId,
		To:                         id,
		FromType:                   parent.SourceType(),
		ToType:                     "fruit",
		CreationDateField:          CreationDateField{now},
		Reason:                     xferReasonReady, // TODO: what here? This is only used when going from non-fruit to print/swab
		FromImage:                  nil,             // TODO: ?????
		ToImage:                    nil,             // TODO: ?????
		NotesField:                 NotesField{},    // TODO: ?????
		LastUpdatedField:           LastUpdatedField{now},
		AclField:                   parent.Permissions().AsField(),
	}
	// Add transfer
	db := mongo.SessionFromContext(ctx).Client().Database(dbName)
	_, err = db.Collection(TransfersCollName).InsertOne(ctx, xfer)
	if err != nil {
		return nil, err
	}
	// Add transfer to parent
	parentUpd, err := NewMods().addTransferOut(xferId).withLastUpdated(now).Finalized()
	if err != nil {
		return nil, err
	}
	_, err = db.Collection(parent.CollectionName()).
		UpdateByID(ctx, parent.DbId(), parentUpd)
	if err != nil {
		return nil, err
	}
	// Return the new fruit
	return newFruit, nil
}

//func deleteFruitHandler(w http.ResponseWriter, r *http.Request) {
//	idStr := r.PathValue("id")
//	if idStr == "" {
//		http.Error(w, "Empty id for delete request", http.StatusBadRequest)
//		return
//	}
//	id, err := Base58Str(idStr).ToMainCollectionId()
//	if err != nil {
//		http.Error(w, "Invalid ID to delete: "+err.Error(), http.StatusBadRequest)
//		return
//	}
//	// Validate not used in other places...
//	ctx := r.Context()
//	// ensure item does not have any transfers in or out
//	item, err := GetMainCollectionItemSpecific[*Fruit](ctx, id, &Fruit{})
//	if err != nil {
//		if errors.Is(err, mongo.ErrNoDocuments) {
//			http.Error(w, "Item to be deleted not found! Should never happen!: "+err.Error(), http.StatusNotFound)
//		} else {
//			http.Error(w, "Failed to retrieve item to be deleted: "+err.Error(), http.StatusInternalServerError)
//		}
//		return
//	}
//	if item.Parent != nil {
//		// TODO: what if we want to remove it from the parent as well?
//		http.Error(w, "Cannot delete innoculated items!", http.StatusExpectationFailed)
//		return
//	}
//	if item.TransfersOut != nil && len(item.TransfersOut) > 0 {
//		http.Error(w, "Cannot delete items with transfers out", http.StatusExpectationFailed)
//		return
//	}
//
//	// Delete if not found elsewhere!
//	DeleteCollectionItem(ctx, item.CollectionName(), id, w)
//}
