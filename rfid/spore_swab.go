package rfid

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/reeceappling/goUtils/v2/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
)

type SporeSwab struct { // TODO: FIX EVERYTHING IN THIS FILE BELOW THIS POINT!!!!
	MainCollectionIdField `bson:"inline"`
	// Parent is always either sporePrint, or purchased
	MainCollectionOptionalParentField `bson:"inline"` // won't exist for pre-existing or purchased
	CreationDateField                 `bson:"inline"` // Swab or receive date
	SpeciesField                      `bson:"inline"`
	SubspeciesOptionalField           `bson:"inline"`
	SaleField                         `bson:"inline"` // TODO: was sales! singular now
	DisposedField                     `bson:"inline"`
	TransfersOutField                 `bson:"inline"`
	NotesField                        `bson:"inline"`
	LastUpdatedField                  `bson:"inline"`
	AclField                          `bson:"inline"`
}

func (sw SporeSwab) Innoculatable() bool {
	return false
}

func (sw SporeSwab) CanTransferTo(dst geneticSource) error {
	if dst.SourceType() != PlateSourceType {
		return errors.New("sporeSwabs can only transfer to plates")
	}
	return errors.New("fc cannot be transferred (unsure if this is ok)")
}

func (sw SporeSwab) setTransferParent(ctx context.Context, xfer Transfer) (error, func() error) {
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(sw.CollectionName())
	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
	if err != nil {
		return err, nil
	}
	res, err := coll.UpdateByID(ctx, sw.Id, upd)
	if err != nil {
		return err, nil
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer, nil
	}
	return nil, func() error {
		return coll.FindOneAndReplace(ctx, bson.D{{"_id", sw.Id}}, sw).Err()
	}
}

func (sw SporeSwab) setTransferChild(ctx context.Context, xfer Transfer, from geneticSource) error {
	panic("cannot happen stc")
	//// TODO: can this happen????? should always be from a fruit right?
	//// This is a special case because it will always be 0-gen
	//parentInfo, err := from.GeneticInfoAsParent()
	//if err != nil {
	//	return err
	//}
	//if parentInfo.Species == nil {
	//	return errors.New("parent must have a species")
	//}
	//if from.SourceType() != SporePrintSourceType {
	//	errors.New("only fruits are supported as a transfer source type into sporeSwabs")
	//}
	//upd, err := xfer. // TODO: fix this whole thing
	//			PicsModsForChild(). // TODO: fix
	//			withInnoc(xfer).    // TODO: fix
	//			withParent(utils.Pointer(from.DbId())).
	//			withSpecies(parentInfo.Species).
	//			withSubspecies(parentInfo.SubSpecies).
	//			updateLastUpdatedIfNeeded().
	//			Finalized()
	//res, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(sw.CollectionName()).UpdateByID(ctx, sw.Email, upd)
	//if err != nil {
	//	return err
	//}
	//if res.ModifiedCount == 0 {
	//	return ErrNoParentModifiedForTransfer
	//}
	//return nil
}

func (sw SporeSwab) GeneticInfoAsParent() (GeneticParentInfo, error) {
	return GeneticParentInfo{
		SpeciesOptionalField:    sw.SpeciesField.AsOptional(),
		SubspeciesOptionalField: sw.SubspeciesOptionalField,
		GenerationsFields:       GenerationsFieldFor(utils.Pointer(Generation(0))),
	}, nil
}

func (sw SporeSwab) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return utils.Pointer(Generation(0)), utils.Pointer(Generation(0))
}

func (sw SporeSwab) EntryTypeField() *string {
	return nil
}

func (sw SporeSwab) id() []byte {
	return []byte(sw.Id.dbIdStr())
}

func initializeSporeSwabs(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(SporeSwabCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		//newSimpleIndex("parent", "parent", false, false, false),
		newSimpleIndex("creationDate", "creationDate", true, false, false), // TODO: INDEX CREATION DATES EVERYWHERE!
		newSimpleIndex("species", "species", false, false, false),
		newSimpleIndex("subspecies", "subspecies", false, true, false),
		//saleIndexModel,
		//disposedIndexModel,
		//transfersOutIndexModel,
		//Notes (no index unless tags)
		projectsIndexModel,
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	testItem := &SporeSwab{
		MainCollectionIdField:             MainCollectionIdField{exSwabId},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&exSporePrint},
		CreationDateField:                 exampleTime.asCreationDate(),
		SpeciesField:                      SpeciesField{testEntryStringId},
		SubspeciesOptionalField:           SubspeciesOptionalField{&testEntryStringId},
		SaleField:                         SaleField{&exAltId},
		DisposedField:                     DisposedField{&exampleTime},
		NotesField:                        NotesField{exampleNotes()},
		LastUpdatedField:                  LastUpdatedField{exampleTime},
	}
	return addTestMainEntries(ctx, testItem)
}

type createSporeSwabsRequest struct {
	num          int
	SporePrintId MainCollectionId
	NotesField
}

// TODO: REALLY FLESH THIS OUT
func createSporeSwabHandler(w http.ResponseWriter, r *http.Request) { // TODO: NO PICS
	data := createSporeSwabsRequest{}
	defer r.Body.Close()
	// TODO: no pictures, so use other way
	// Process text (or object)
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	// PARSE INTO CORRECT DATA FORMAT
	err = json.Unmarshal(bs, &data)
	if err != nil {
		http.Error(w, "failed to unmarshal data from form: "+err.Error(), http.StatusBadRequest)
		return
	}

	ids, err := generateCollectionIds(r.Context(), SporeSwabCollectionName, data.num)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx, db := Db(r)
	coll := db.Collection(SporeSwabCollectionName)

	parent := SporePrint{}
	err = db.Collection(SporePrintCollectionName).FindOne(ctx, bson.D{{"_id", data.SporePrintId}}).Decode(&parent)
	if err != nil {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}

	now := unixTimeForNow()
	out := make([]interface{}, len(ids))
	idsToMap := make([]MainCollectionId, len(ids))
	for i, id := range ids {
		idsToMap[i] = id
		out[i] = SporeSwab{
			MainCollectionIdField:             MainCollectionIdField{idsToMap[i]},
			MainCollectionOptionalParentField: MainCollectionOptionalParentField{&parent.Id},
			CreationDateField:                 now.asCreationDate(),
			SpeciesField:                      parent.SpeciesField,
			SubspeciesOptionalField:           parent.SubspeciesOptionalField,
			NotesField:                        NotesField{data.Notes},
			LastUpdatedField:                  LastUpdatedField{now},
			// Do not check permissions, just pass parent perms to child
			AclField: parent.AclField,
		}

	}

	// TODO: add new swabs to mappings

	_, err = coll.InsertMany(ctx, out)
	if err != nil {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	bsOut, err := json.Marshal(out)
	if err != nil {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(bsOut)
	if err != nil {
		handleWriteErr(err, w)
	}
}

type updateSporeSwabRequest struct { // TODO: fix everything below this
	SaleField
	DisposedField
	Notes AllEntries[Note]
	PermsOnRequest
}

func (upr updateSporeSwabRequest) reform() resolvedUpdateSporeSwabRequest {
	return resolvedUpdateSporeSwabRequest{
		SaleField:      upr.SaleField,
		DisposedField:  upr.DisposedField,
		Notes:          upr.Notes,
		PermsOnRequest: upr.PermsOnRequest,
	}
}

type resolvedUpdateSporeSwabRequest struct {
	SaleField
	DisposedField
	Notes AllEntries[Note]
	PermsOnRequest
}

func (mods resolvedUpdateSporeSwabRequest) modsFor(existing *SporeSwab, aclField AclField) (bson.D, error) {
	return NewMods().
		updateSaleIfNeeded(mods.Sale, existing.Sale).
		updateDisposedIfNeeded(mods.Disposed, existing.Disposed).
		updateNotesIfNeeded(mods.Notes, existing.Notes).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateSporeSwabHandler(w http.ResponseWriter, r *http.Request) {
	data := updateSporeSwabRequest{}
	b58Id := Base58Str(r.PathValue("id")) // TODO: ensure ok
	id, err := b58Id.toMainCollectionId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	out := data.reform()
	ctx, db := Db(r)
	coll := db.Collection(SporeSwabCollectionName)

	// go get current sporeSwab
	existing := SporeSwab{}
	err = coll.FindOne(ctx, bson.D{{"_id", id}}).Decode(&existing)
	if err != nil {
		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishMainCollItemUpdate(ctx, w, coll, out.modsFor, &existing, out.PermsOnRequest)
}

type importSporeSwabRequest struct {
	CreationDateField
	SpeciesField
	SubspeciesOptionalField
	NotesField
	PermsOnRequest
}

func importSporeSwabHandler(w http.ResponseWriter, r *http.Request) { // TODO: NO IMAGES
	data := importSporeSwabRequest{}
	id, err := newCollectionId(r.Context(), SporeSwabCollectionName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Process text (or object)
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "unable to read data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	// PARSE INTO CORRECT DATA FORMAT
	err = json.Unmarshal(bs, &data)
	if err != nil {
		http.Error(w, "unable to unmarshal json form data: "+err.Error(), http.StatusBadRequest)
		return
	}
	//if err = data.Perms.ValidateUserCanWrite(r.Context()); err != nil {
	//	http.Error(w, "email cannot write with these perms: "+err.Error(), http.StatusBadRequest)
	//	return
	//}
	ctx, db := Db(r)
	coll := db.Collection(SporeSwabCollectionName)
	toInsert := SporeSwab{
		MainCollectionIdField:   MainCollectionIdField{id},
		CreationDateField:       data.CreationDateField,
		SpeciesField:            data.SpeciesField,
		SubspeciesOptionalField: data.SubspeciesOptionalField,
		NotesField:              data.NotesField,
		LastUpdatedField:        LastUpdatedFieldForNow(),
		// TODO: perms? AclField:                acl,
	}
	finishImportMainCollectionEntry(ctx, coll, &toInsert, data.PermsOnRequest, w)
	//finalPerms := data.Perms
	//if data.Perms != nil {
	//	spec, subsp, err := getSpeciesAndSubspecies(ctx, data.Species, data.SubSpecies)
	//	if err != nil {
	//		return dbErr(w, "failed to get species or subspecies: "+err.Error(), http.StatusInternalServerError) // TODO: ok?
	//	}
	//	finalPerms = minimalPermsBetween(spec, subsp)
	//	// TODO: add email perms if provided, as well as make email author?
	//	if !finalPerms.Valid() {
	//		// TODO: invalid species/subspecies perm crossover. DO THIS ELSEwHERE
	//		return dbErr(w, "invalid species/subspecies perm crossover: "+err.Error(), http.StatusInternalServerError) // TODO: ok?
	//	}
	//}
	//perms, err := GetAuthInfo(ctx)
	//if err != nil {
	//	return dbErr(w, err.Error(), http.StatusInternalServerError)
	//}
	//acl, err := data.AclForUser(ctx, perms)
	//if err != nil {
	//	return dbErr(w, err.Error(), http.StatusInternalServerError)
	//}
}
