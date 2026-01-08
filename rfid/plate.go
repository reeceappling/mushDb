package rfid

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/reeceappling/goUtils/v2/utils"
	"github.com/reeceappling/mushDb/rfid/pics"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"reflect"
	"slices"
)

const (
	PlatesCollectionName = "plates" // TODO: USE
	PlateSourceType      = "plate"
)

type Plate struct { // TODO: CACHE RESPONSES?!!!!!
	MainCollectionIdField
	AgarBatchField // TODO: will be empty for preexisting
	CreationDateField
	SpeciesOptionalField
	SubspeciesOptionalField
	InnocField
	GenerationsFields
	TransfersOutField
	ParentTypeField                   // TODO: NEW! HANDLE! nil == mainCollectionType, can also be MSS or clone! // TODO: INDEX????
	MainCollectionOptionalParentField // TODO: was binary, b58 clientside? // TODO: can be from any MainCollection, or a fruit (alt) cloning/lcSyringe/sporeSwab
	PicsField
	ContaminationsField
	KnownFruitableField // TODO: handle being yes if clone, among other yeses
	SaleField
	DisposedField
	MostRecentImageField
	NotesField
	LastUpdatedField
	//PermsField
}

func (p Plate) CanTransferTo(dst geneticSource) error {
	if !slices.Contains([]string{BagSourceType, GrainJarSourceType, LcSourceType, PlateSourceType, PlugSourceType, SlantSourceType, StasisTubeSourceType}, dst.SourceType()) {
		return errors.New("plates cannot transfer to " + dst.SourceType())
	}
	return nil
}

func (p Plate) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := p
	err := decodeItem(&out, encoded)
	return out, err
}

func (p Plate) GeneticInfoAsParent() (GeneticParentInfo, error) {
	return GeneticParentInfo{
		SpeciesOptionalField:    SpeciesOptionalField{p.Species},
		SubspeciesOptionalField: p.SubspeciesOptionalField,
		KnownFruitableField:     p.KnownFruitableField,
		GenerationsFields:       p.GenerationsFields,
	}, nil
}

func (p Plate) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return p.GenSinceSpore, p.GenSinceFruitOrSpore
}

func (p Plate) SourceType() string {
	return PlateSourceType
}

func (p Plate) setTransferParent(ctx mongo.SessionContext, xfer Transfer) error {
	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
	if err != nil {
		return err
	}
	res, err := ctx.Client().Database(dbName).Collection(mainCollectionName).UpdateByID(ctx, p.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer
	}
	return nil
}

func (p Plate) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
	parentInfo, genSpore, genFruitSpore, err := childGensForParent(from)
	if err != nil {
		return err
	}
	upd, err := xfer.
		PicsModsForChild().
		withInnoc(xfer).
		withParentType(&xfer.FromType).
		withParent(utils.Pointer(from.DbId())).
		withGens(genSpore, genFruitSpore).
		withSpecies(parentInfo.Species).
		withSubspecies(parentInfo.SubSpecies).
		withKnownFruitable(parentInfo.KnownFruitable).
		withLastUpdated(xfer.LastUpdated).
		Finalized()
	if err != nil {
		return ErrFailedToFinalizeMods
	}
	res, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(mainCollectionName).UpdateByID(ctx, p.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer
	}
	return nil
}

func (p Plate) EntryTypeField() *string {
	return utils.Pointer(PlateSourceType)
}

func (p Plate) CollectionName() string {
	return mainCollectionName
}

func initializePlates(ctx context.Context) error {
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(mainCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		newSimpleIndex("agarBatch", "agarBatch", false, true, false),
		creationDateIndexModel,
		newSimpleIndex("species", "species", false, true, false),
		newSimpleIndex("subSpecies", "subSpecies", false, true, false),
		newSimpleIndex("innoc", "innoc", false, true, false),
		newSimpleIndex("genSinceSpore", "genSpore", true, true, false),
		newSimpleIndex("genSinceFruitOrSpore", "genFruitOrSpore", true, true, false),
		transfersOutIndexModel,
		newSimpleIndex("parent", "parent", false, true, false),         // TODO: nil is store or outside?
		newSimpleIndex("parentType", "parentType", false, true, false), // TODO: nil is store or outside?

		//Pics (no index)
		// TODO: Contams
		newSimpleIndex("knownFruitable", "knownFruitable", false, true, false),
		saleIndexModel,
		disposedIndexModel,
		// MostRecentImage
		//Notes (no index) (maybe later with tags?)
		lastUpdatedIndexModel,
		// TODO: projectsIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	existingEntry := Plate{}
	testId := mainCollIdForint(idTestPlate)
	testItem := Plate{
		MainCollectionIdField:   MainCollectionIdField{testId},
		AgarBatchField:          AgarBatchField{&exAltId},
		CreationDateField:       CreationDateField{exampleTime},
		SpeciesOptionalField:    SpeciesOptionalField{&testEntryStringId},
		SubspeciesOptionalField: SubspeciesOptionalField{&testEntryStringId},
		InnocField:              InnocField{&exAltId}, // TODO: multiple?
		GenerationsFields: GenerationsFields{
			GenSporeField:        GenSporeField{&exGenSinceSpore},
			GenSinceFruitOrSpore: &exGenSinceFruitSpore,
		},
		TransfersOutField:                 TransfersOutField{exAlts},
		ParentTypeField:                   ParentTypeField{&exParentType},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&testId}, // TODO: ok? // TODO: multiple?
		PicsField:                         PicsField{exPics},
		ContaminationsField:               ContaminationsField{exContams},
		KnownFruitableField:               KnownFruitableField{exBool},
		SaleField:                         SaleField{&exAltId},
		DisposedField:                     DisposedField{&exampleTime},
		MostRecentImageField:              MostRecentImageField{&exPics[0]},
		NotesField:                        NotesField{exampleNotes()},
		LastUpdatedField:                  LastUpdatedField{exampleTime},
	}
	err := coll.FindOne(ctx, bson.D{{"_id", testId}}).Decode(&existingEntry)
	if err == nil {
		if reflect.DeepEqual(existingEntry, testItem) {
			return nil
		}
	}
	return testExistingEntry(ctx, coll, testId, testItem, existingEntry)
}

type createPlateRequest struct {
	AgarBatch AlternateCollectionId `json:"agarBatch"`
	WriteTagToField
}

func createPlateHandler(w http.ResponseWriter, r *http.Request) {
	data := createPlateRequest{}
	id, err := newMainCollectionId(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
	now := unixTimeForNow()
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(mainCollectionName)
		toInsert := Plate{
			MainCollectionIdField: MainCollectionIdField{id},
			AgarBatchField:        AgarBatchField{AgarBatch: &data.AgarBatch},
			CreationDateField:     CreationDateField{now},
			LastUpdatedField:      LastUpdatedField{now},
			// No Perms here for basic plates
		}
		_, err = toInsert.AgarBatchField.Get(ctx)
		if err != nil && !errors.Is(err, ErrMissingOptionalField) {
			return DbTxnStdErr(w, "agar batch field missing: "+err.Error(), http.StatusInternalServerError)
		}
		_, err = coll.InsertOne(ctx, toInsert)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		bsOut, err := json.Marshal(toInsert)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		return w.Write(bsOut)
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}

type updatePlateRequest struct {
	KnownFruitableField
	SaleField
	DisposedField
	Notes   AllEntries[Note]
	Images  SplitEntries[picWithNotesForm, PicWithNotesLessLocation]
	Contams SplitEntries[contamForm, ContaminationLessLocation]
	WriteTagToField
	//PermsField
}

func (upr updatePlateRequest) reform() resolvedUpdatePlateRequest {
	return resolvedUpdatePlateRequest{
		KnownFruitableField: upr.KnownFruitableField,
		SaleField:           upr.SaleField,
		DisposedField:       upr.DisposedField,
		Notes:               upr.Notes,
		Images:              imageUpdates(upr.Images),
		Contams:             contamUpdates(upr.Contams),
		WriteTagToField:     upr.WriteTagToField,
		//PermsField:          upr.PermsField,
	}
}

func (mods resolvedUpdatePlateRequest) modsFor(existing Plate) (bson.D, error) {
	return NewMods().
		updateKnownFruitableIfNeeded(mods.KnownFruitable, existing.KnownFruitable).
		updateSaleIfNeeded(mods.Sale, existing.Sale).
		updateDisposedIfNeeded(mods.Disposed, existing.Disposed).
		updateNotesIfNeeded(mods.Notes, existing.Notes).
		updatePicsIfNeeded(mods.Images, existing.Pics).
		updateContamsIfNeeded(mods.Contams, existing.Contaminations).
		//updatePermsIfNeeded(mods.Perms, existing.Perms).
		updateLastUpdatedIfNeeded().
		Finalized()
}

type resolvedUpdatePlateRequest struct {
	KnownFruitableField
	SaleField
	DisposedField
	Notes   AllEntries[Note]
	Images  SplitEntries[picWithNotesForm, PicWithNotes]
	Contams SplitEntries[contamForm, Contamination]
	WriteTagToField
	//PermsField // TODO: new
}

func updatePlateHandler(w http.ResponseWriter, r *http.Request) {
	data := updatePlateRequest{}
	b58Id := Base58Str(r.PathValue("id"))
	id, err := b58Id.toMainCollectionId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	reader, err := multipartReaderForRequest(r, w, &data)
	if err != nil {
		// Already written
		return
	}
	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	newPics, newContams, _, err := getMultipartImages(r.Context(), "jar", w, reader, b58Id)
	if err != nil {
		// Already wrotw
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
	for i, _ := range data.Contams.New {
		if loc, exists := newContams[i]; exists {
			finalLoc := imageLocation(loc)
			out.Contams.New[i].Location = &finalLoc
		}
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(mainCollectionName)
		// go get current plate
		existing := Plate{}
		err := coll.FindOne(ctx, bson.D{{"_id", id}}).Decode(&existing)
		if err != nil {
			return DbTxnStdErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		}
		//if err = minimalPermsBetween(existing.Perms, data.Perms).ValidateUserCanWrite(ctx); err != nil {
		//	return DbTxnStdErr(w, "failed to validate user permissions, old or new: "+err.Error(), http.StatusBadRequest)
		//}
		upd, err := out.modsFor(existing)
		return handleUpdateMods(ctx, w, coll, existing, id, upd, err) // TODO: use this on all updates!!!!!!!
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}

func handleUpdateMods[T any, U MainCollectionId | AlternateCollectionId](ctx context.Context, w http.ResponseWriter, coll *mongo.Collection, existing T, id U, upd bson.D, err error) (any, error) {
	if err != nil {
		return DbTxnStdErr(w, "error creating txn:"+err.Error(), http.StatusInternalServerError)
	}
	if len(upd) == 0 {
		return DbTxnStdErr(w, "no changes made", http.StatusBadRequest)
	}
	// write updates to db
	bsonId := bson.D{{"_id", id}}
	err = coll.FindOneAndUpdate(ctx, bsonId, upd).Err()
	if err != nil {
		return DbTxnStdErr(w, "failed to write update to db: "+err.Error(), http.StatusInternalServerError)
	}
	err = coll.FindOne(ctx, bsonId).Decode(&existing)
	if err != nil {
		return DbTxnStdErr(w, "failed to write update to db: "+err.Error(), http.StatusInternalServerError)
	}
	bsOut, err := json.Marshal(existing)
	if err != nil {
		return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
	}
	return w.Write(bsOut)
}

type importPlateRequest struct {
	CreationDateField
	SpeciesField
	SubspeciesOptionalField
	KnownFruitableField
	Generation *int
	// pic as "img"
	WriteTagToField
	//PermsField
}

func importPlateHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data := importPlateRequest{}
	id, err := newMainCollectionId(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	b58id := id.asBase58()
	reader, err := multipartReaderForRequest(r, w, &data)
	if err != nil {
		// Already written
		return
	}
	//if err = data.Perms.ValidateUserCanWrite(r.Context()); err != nil {
	//	http.Error(w, "user cannot write perms: "+err.Error(), http.StatusBadRequest)
	//	return
	//}
	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Try to get pic if exists
	picsSaved := []string{} // TODO: do this streamlined with the multipart function
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
		newFileNameWithPrefixPath, errr := pics.SaveFile(r.Context(), fieldBytes, "plate", string(b58id), "img")
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
	}
	var gen *Generation = nil
	if data.Generation != nil {
		gen = (*Generation)(data.Generation)
	}
	pix := []PicWithNotes{}
	if importedPic != nil {
		pix = []PicWithNotes{*importedPic}
	}

	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		//finalPerms := data.Perms
		//if data.Perms != nil {
		//	spec, subsp, err := getSpeciesAndSubspecies(ctx, data.Species, data.SubSpecies)
		//	if err != nil {
		//		return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		//	}
		//	finalPerms = minimalPermsBetween(data.Perms, spec, subsp) // TODO: DO MAXIMAL PERMS WITH data.Perms if not allWrite?
		//}

		toInsert := Plate{
			MainCollectionIdField:   id.IdField(),
			CreationDateField:       data.CreationDateField,
			SpeciesOptionalField:    data.SpeciesField.AsOptional(),
			SubspeciesOptionalField: data.SubspeciesOptionalField,
			GenerationsFields:       GenerationsFieldFor(gen),
			PicsField:               PicsField{pix},
			KnownFruitableField:     data.KnownFruitableField,
			MostRecentImageField:    MostRecentImageField{importedPic},
			LastUpdatedField:        LastUpdatedField{unixTimeForNow()},
			//PermsField:              PermsField{finalPerms},
		}
		coll := ctx.Client().Database(dbName).Collection(mainCollectionName)
		_, err := coll.InsertOne(ctx, toInsert)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		bsOut, err := json.Marshal(toInsert)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		return w.Write(bsOut)
	})

	if err != nil {
		handleWriteErr(err, w)
	}
}
