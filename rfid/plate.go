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
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"net/http"
	"slices"
)

type Plate struct { // TODO: CACHE RESPONSES?!!!!!
	MainCollectionIdField             `bson:"inline"`
	AgarBatchField                    `bson:"inline"` // TODO: will be empty for preexisting
	CreationDateField                 `bson:"inline"`
	SpeciesOptionalField              `bson:"inline"`
	SubspeciesOptionalField           `bson:"inline"`
	InnocField                        `bson:"inline"`
	GenerationsFields                 `bson:"inline"`
	TransfersOutField                 `bson:"inline"`
	ParentTypeField                   `bson:"inline"` // TODO: NEW! HANDLE! nil == mainCollectionType, can also be MSS or clone! // TODO: INDEX????
	MainCollectionOptionalParentField `bson:"inline"` // TODO: was binary, b58 clientside? // TODO: can be from any MainCollection, or a fruit (alt) cloning/lcSyringe/sporeSwab
	PicsField                         `bson:"inline"`
	ContaminationsField               `bson:"inline"`
	KnownFruitableField               `bson:"inline"` // TODO: handle being yes if clone, among other yeses
	SaleField                         `bson:"inline"`
	DisposedField                     `bson:"inline"`
	MostRecentImageField              `bson:"inline"`
	NotesField                        `bson:"inline"`
	LastUpdatedField                  `bson:"inline"`
	AclField                          `bson:"inline"` // TODO: handle EVERYWHERE
}

func (p Plate) IdValue() any {
	return p.Id.dbIdStr() // TODO: ensure ok
}

func (p Plate) CanTransferTo(dst geneticSource) error {
	if !slices.Contains([]string{BagSourceType, GrainJarSourceType, LcSourceType, PlateSourceType, PlugSourceType, SlantSourceType, StasisTubeSourceType}, dst.SourceType()) {
		return errors.New("plates cannot transfer to " + dst.SourceType())
	}
	return nil
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

func (p Plate) setTransferParent(ctx context.Context, xfer Transfer) (error, func() error) {
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(PlatesCollectionName)
	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
	if err != nil {
		return err, nil
	}
	res, err := coll.UpdateByID(ctx, p.Id, upd)
	if err != nil {
		return err, nil
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer, nil
	}
	return nil, func() error {
		return coll.FindOneAndReplace(ctx, bson.D{{"_id", p.Id}}, p).Err()
	}
}

func (p Plate) setTransferChild(ctx context.Context, xfer Transfer, from geneticSource) error {
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
		withPerms(from.Permissions()). // TODO: USE THIS IN A LOT OF PLACES!!!!!
		withLastUpdated(xfer.LastUpdated).
		Finalized()
	if err != nil {
		return ErrFailedToFinalizeMods
	}
	res, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(PlatesCollectionName).UpdateByID(ctx, p.Id, upd)
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

func initializePlates(ctx context.Context) error {
	println("1")
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	println("2")
	coll := db.Collection(PlatesCollectionName)
	//_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
	//	//newSimpleIndex("agarBatch", "agarBatch", false, true, false),
	//	//creationDateIndexModel,
	//	//newSimpleIndex("species", "species", false, true, false),
	//	//newSimpleIndex("subSpecies", "subSpecies", false, true, false),
	//	//newSimpleIndex("innoc", "innoc", false, true, false),
	//	//newSimpleIndex("genSinceSpore", "genSpore", true, true, false),
	//	//newSimpleIndex("genSinceFruitOrSpore", "genFruitOrSpore", true, true, false),
	//	//transfersOutIndexModel,
	//	//newSimpleIndex("parent", "parent", false, true, false),         // TODO: nil is store or outside?
	//	//newSimpleIndex("parentType", "parentType", false, true, false), // TODO: nil is store or outside?
	//	//
	//	////Pics (no index)
	//	//// TODO: Contams
	//	//newSimpleIndex("knownFruitable", "knownFruitable", false, true, false),
	//	//saleIndexModel,
	//	//disposedIndexModel,
	//	//// MostRecentImage
	//	////Notes (no index) (maybe later with tags?)
	//	//lastUpdatedIndexModel,
	//	// TODO: projectsIndexModel,
	//})
	//if err != nil {
	//	return err
	//}
	// If test agar batch does not exist, then create it
	testId := mainCollIdForint(idTestPlate) // 0th id, b58==1
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
	// TODO: INSERT IN MAP COLLECTION?
	println("3")
	result, err := coll.ReplaceOne(ctx, bson.D{{"_id", testId}}, testItem, &options.ReplaceOptions{Upsert: utils.Pointer(true)})
	if err != nil {
		println(err.Error())
		return err
	}
	println("4")
	res := coll.FindOne(ctx, bson.D{{"_id", testId}})
	if res.Err() != nil {
		println(res.Err().Error())
		return res.Err()
	}
	println("5")
	raw, err := res.Raw()
	if err != nil {
		println(err.Error())
		return err
	}
	println("6")
	println(raw.String())
	err = res.Decode(&testItem)
	if err != nil {
		println("failed to decode test item", err.Error())
	}
	println("7")
	println(result.UpsertedID)
	//result, err := coll.ReplaceOne(ctx, bson.D{{"_id", testId}}, testItem)
	//if err != nil {
	//	println(err.Error())
	//	return err
	//}
	//result.UpsertedID =
	//println("ITEM PLACED IN COLLECTION!!!!!!!! --------------------------------------")
	//switch result.UpsertedID.(type) {
	//case primitive.ObjectID:
	//	println(result.UpsertedID)
	//default:
	//	println("not object id: " + reflect.TypeOf(result.UpsertedID).Name())
	//
	//}
	//id, ok := result.UpsertedID.(type)
	//if !ok {
	//	return errors.New("id was not main")
	//}
	//println("id", id)
	return err
}

type createPlateRequest struct {
	AgarBatch AlternateCollectionId `json:"agarBatch"`
	WriteTagToField
}

func createPlateHandler(w http.ResponseWriter, r *http.Request) {
	data := createPlateRequest{}
	id, err := newCollectionId(r.Context(), PlatesCollectionName)
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
	ctx := r.Context()
	client := r.Context().Value(mongoClientContextKey).(*mongo.Client)
	coll := client.Database(dbName).Collection(PlatesCollectionName)
	toInsert := Plate{
		MainCollectionIdField: MainCollectionIdField{id},
		AgarBatchField:        AgarBatchField{AgarBatch: &data.AgarBatch},
		CreationDateField:     CreationDateField{now},
		LastUpdatedField:      LastUpdatedField{now},
		// No Perms here for basic plates
		AclField: allCanWriteAcl(),
	}
	_, err = toInsert.AgarBatchField.Get(ctx)
	if err != nil && !errors.Is(err, ErrMissingOptionalField) {
		http.Error(w, "agar batch field missing: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishCreateMainCollectionEntry(ctx, coll, &toInsert, w) // TODO: use in all main creates
}

type updatePlateRequest struct {
	KnownFruitableField
	SaleField
	DisposedField
	Notes   AllEntries[Note]
	Images  SplitEntries[picWithNotesForm, PicWithNotesLessLocation]
	Contams SplitEntries[contamForm, ContaminationLessLocation]
	WriteTagToField
	PermsOnRequest // TODO: handle in typescript and handler!
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
		PermsOnRequest:      upr.PermsOnRequest,
	}
}

func (mods resolvedUpdatePlateRequest) modsFor(existing *Plate, aclField AclField) (bson.D, error) {
	return NewMods().
		updateKnownFruitableIfNeeded(mods.KnownFruitable, existing.KnownFruitable).
		updateSaleIfNeeded(mods.Sale, existing.Sale).
		updateDisposedIfNeeded(mods.Disposed, existing.Disposed).
		updateNotesIfNeeded(mods.Notes, existing.Notes).
		updatePicsIfNeeded(mods.Images, existing.Pics).
		updateContamsIfNeeded(mods.Contams, existing.Contaminations).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
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
	PermsOnRequest // TODO: handle in typescript and handler!
}

func updatePlateHandler(w http.ResponseWriter, r *http.Request) {
	println("RECEIVED UPDATE PLATE REQUEST") // TODO: THIS
	data := updatePlateRequest{}
	b58Id := Base58Str(r.PathValue("id"))
	id, err := b58Id.toMainCollectionId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	println("CREATED READER") // TODO: THIS
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
	println("GETTING IMAGES") // TODO: THIS
	newPics, newContams, _, err := getMultipartImages(r.Context(), "jar", w, reader, b58Id)
	if err != nil {
		// Already wrotw
		return
	}

	// CHECK THAT ALL NEW PICS EXIST
	// PROCESS ALL NEW PICS AND CONTAMS
	println("REFORMING") // TODO: THIS
	out := data.reform()
	for i, _ := range data.Images.New {
		loc, exists := newPics[i]
		if !exists {
			println("FAILED TO GET LOCATION") // TODO: THIS
			http.Error(w, fmt.Sprintf("error, location for new picture index %d not found (should never happen)", i), http.StatusInternalServerError)
			return
		}
		out.Images.New[i].Location = imageLocation(loc)
	}
	println("PARSING CONTAMS") // TODO: THIS
	for i, _ := range data.Contams.New {
		if loc, exists := newContams[i]; exists {
			finalLoc := imageLocation(loc)
			out.Contams.New[i].Location = &finalLoc
		}
	}
	println("Starting TX") // TODO: THIS
	/* TODO:
	* Our responsiveness last year was very slow
	* FEEDBACK FROM KATE:
	* Move teams immediately...
	* Prepare thoughts for talking to megan.
	* Make her realize that it was in a situation that was mostly out of my control, but am still a great engineer.
	* Make her confident in testing skills, people skills, decision making skills.
	* Build emotional bank account back up.
	*
	 */
	ctx := r.Context()
	client := ctx.Value(mongoClientContextKey).(*mongo.Client)
	coll := client.Database(dbName).Collection(PlatesCollectionName)
	existing := Plate{}
	err = coll.FindOne(ctx, bson.D{{"_id", id}}).Decode(&existing)
	if err != nil {
		// TODO: an issue here?
		http.Error(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishMainCollItemUpdate(ctx, w, coll, out.modsFor, &existing, out.PermsOnRequest)
}

func handleUpdateMods[T any, U MainCollectionId | AlternateCollectionId | string](ctx context.Context, w http.ResponseWriter, coll *mongo.Collection, existing T, id U, upd bson.D, err error) {
	if err != nil {
		println("mod creation failure: " + err.Error())
		dbErr(w, "error creating txn:"+err.Error(), http.StatusInternalServerError)
		return
	}
	if len(upd) == 0 {
		println("no changes made") // TODO: del
		dbErr(w, "no changes made", http.StatusBadRequest)
		return
	}
	// write updates to db
	println("updating") // TODO: del
	bsonId := bson.D{{"_id", id}}
	err = coll.FindOneAndUpdate(ctx, bsonId, upd).Err()
	if err != nil {
		println("failed to update") // TODO: del
		dbErr(w, "failed to write update to db: "+err.Error(), http.StatusInternalServerError)
		return
	}
	println("finding") // TODO: del
	err = coll.FindOne(ctx, bsonId).Decode(&existing)
	if err != nil {
		println("failed to find") // TODO: del
		dbErr(w, "failed to write update to db: "+err.Error(), http.StatusInternalServerError)
		return
	}
	println("marshalling") // TODO: del
	bsOut, err := json.Marshal(existing)
	if err != nil {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	println("Wrote plate update:", string(bsOut))
	_, err = w.Write(bsOut)
	handleWriteErr(err, w)
}

type importPlateRequest struct {
	CreationDateField
	SpeciesField
	SubspeciesOptionalField
	KnownFruitableField
	Generation *int
	// pic as "img"
	WriteTagToField
	PermsOnRequest // TODO: handle in typescript and handler!
}

func importPlateHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data := importPlateRequest{}
	id, err := newCollectionId(r.Context(), PlatesCollectionName)
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
	//	http.Error(w, "email cannot write perms: "+err.Error(), http.StatusBadRequest)
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

	ctx, db := Db(r)
	coll := db.Collection(PlatesCollectionName)

	toInsert := Plate{
		MainCollectionIdField:   MainCollectionIdField{id},
		CreationDateField:       data.CreationDateField,
		SpeciesOptionalField:    data.SpeciesField.AsOptional(),
		SubspeciesOptionalField: data.SubspeciesOptionalField,
		GenerationsFields:       GenerationsFieldFor(gen),
		PicsField:               PicsField{pix},
		KnownFruitableField:     data.KnownFruitableField,
		MostRecentImageField:    MostRecentImageField{importedPic},
		LastUpdatedField:        LastUpdatedField{unixTimeForNow()},
	}
	finishImportMainCollectionEntry(ctx, coll, &toInsert, data.PermsOnRequest, w)
}
