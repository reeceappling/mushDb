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

// TODO: newFromPCdwater ????
// TODO: newFromSporePrint (typical but requires PC-d water to not be referenced)
// TODO: add sterilizedWaterJar table and page

type MSS struct {
	// ALWAYS assume contaminated
	MainCollectionIdField   `bson:"inline"`
	CreationDateField       `bson:"inline"`
	WaterJarOptionalField   `bson:"inline"` // TODO: HANDLE THIS EVERYWHERE! NOT YET DONE IN TS
	SpeciesField            `bson:"inline"`
	SubspeciesOptionalField `bson:"inline"`
	// NOTE: parentType is always either sporePrint or purchased
	MainCollectionOptionalParentField `bson:"inline"` // no parent means purchased, traded-for, or imported
	TransfersOutField                 `bson:"inline"`
	SaleField                         `bson:"inline"`
	DisposedField                     `bson:"inline"`
	NotesField                        `bson:"inline"`
	LastUpdatedField                  `bson:"inline"`
	AclField                          `bson:"inline"`
}

func (M MSS) Innoculatable() bool {
	return false // TODO: ensure ok
}

func (M MSS) CanTransferTo(dst geneticSource) error {
	if dst.SourceType() != PlateSourceType {
		return errors.New("mss transfers must go to plates")
	}
	if !dst.Innoculatable() {
	}
	return nil
}

func (M MSS) GeneticInfoAsParent() (GeneticParentInfo, error) {
	return GeneticParentInfo{
		SpeciesOptionalField:    SpeciesOptionalField{&M.Species},
		SubspeciesOptionalField: M.SubspeciesOptionalField,
		KnownFruitableField:     KnownFruitableField{utils.Pointer(false)},
		GenerationsFields: GenerationsFields{
			GenSporeField:        GenSporeField{utils.Pointer(Generation(0))},
			GenSinceFruitOrSpore: utils.Pointer(Generation(0)),
		},
	}, nil
}

func (M MSS) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return utils.Pointer(Generation(0)), utils.Pointer(Generation(0))
}

//func (M MSS) setTransferParent(ctx context.Context, xfer Transfer) error {
//	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(MssCollectionName)
//	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
//	if err != nil {
//		return err
//	}
//	res, err := coll.UpdateByID(ctx, M.Id, upd)
//	if err != nil {
//		return err
//	}
//	if res.ModifiedCount == 0 {
//		return ErrNoParentModifiedForTransfer
//	}
//	return nil
//}

func (M MSS) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
	return errors.New("mss cannot be a child in a normal transfer. Must be created manually from spore print or imported")
}

func (M MSS) EntryTypeField() *string {
	return utils.Pointer(MssSourceType)
}

func initializeMSS(ctx context.Context) error {
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(MssCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		creationDateIndexModel,
		newSimpleIndex("species", "species", false, false, false),
		newSimpleIndex("subspecies", "subspecies", false, true, false),
		//newSimpleIndex("parent", "parent", false, true, false),
		//transfersOutIndexModel,
		//saleIndexModel,
		//disposedIndexModel,
		//Notes (no index unless tags)
		projectsIndexModel,
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	testId := mainCollIdForint(idTestMSS)
	testItem := &MSS{
		MainCollectionIdField:             MainCollectionIdField{testId},
		CreationDateField:                 CreationDateField{exampleTime},
		SpeciesField:                      SpeciesField{testEntryStringId},
		SubspeciesOptionalField:           SubspeciesOptionalField{&testEntryStringId},
		TransfersOutField:                 TransfersOutField{exAlts},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&exSporePrint},
		DisposedField:                     DisposedField{&exampleTime},
		NotesField:                        NotesField{exampleNotes()},
		LastUpdatedField:                  LastUpdatedField{exampleTime},
	}
	return addTestMainEntries(ctx, testItem)
}

type createMssRequest struct {
	WaterJarOptionalField // TODO: HANDLE THIS! Allow creation with or without
	SporePrintId          AlternateCollectionId
	NotesField
	WriteTagToField
	// Uses parent perms, then email can modify if they have the perms for parent
}

func createMssHandler(w http.ResponseWriter, r *http.Request) { // Only called from spore print page
	data := createMssRequest{}
	id := NextMainCollectionId()
	//id, err := newMainCollectionId(r.Context(), MssCollectionName)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
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
	ctx, db := Db(r)
	// Validate parent // TODO: move into txn?
	parent := SporePrint{}
	err = db.Collection(SporePrintCollectionName).FindOne(ctx, bsonFindFilter("_id", data.SporePrintId)).Decode(&parent)
	if err != nil {
		dbErr(w, "failed to find sporePrint: "+err.Error(), http.StatusBadRequest)
		return
	}
	err = writeRfidTagIfNecessary(ctx, data.WriteTagTo, id)
	if err != nil {
		dbErr(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	now := unixTimeForNow()
	toInsert := &MSS{
		MainCollectionIdField:             MainCollectionIdField{id},
		CreationDateField:                 CreationDateField{now},
		SpeciesField:                      SpeciesField{parent.Species},
		SubspeciesOptionalField:           parent.SubspeciesOptionalField,
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&parent.Id},
		NotesField:                        NotesField{data.Notes},
		LastUpdatedField:                  LastUpdatedField{now},
		AclField:                          parent.AclField, // NOTE: do NOT ensure email is authorized to write on parent, they will just be blocked from viewing.
	}
	finishCreateMainCollectionEntry(ctx, toInsert, w)
}

type importMssRequest struct {
	CreationDateField
	SpeciesField
	SubspeciesOptionalField
	NotesField
	WriteTagToField
	PermsOnRequest // TODO: use species/subspecies perms instead? Remove this from both sides
	// image as "img"
	// No ParentType/Parent because these are assumed to be from outside sources
}

func importMssHandler(w http.ResponseWriter, r *http.Request) {
	data := importMssRequest{}
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
	id := NextMainCollectionId()
	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}

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
	// Add user to the acl as a writer
	finalPerms.Users[user.Email] = true

	ctx, db := Db(r)
	coll := db.Collection(MssCollectionName)
	toInsert := MSS{
		MainCollectionIdField:   MainCollectionIdField{id},
		CreationDateField:       data.CreationDateField,
		SpeciesField:            data.SpeciesField,
		SubspeciesOptionalField: data.SubspeciesOptionalField,
		NotesField:              data.NotesField,
		LastUpdatedField:        LastUpdatedField{unixTimeForNow()},
		AclField:                AclField{finalPerms},
	}
	finishImportMainCollectionEntry(ctx, coll, &toInsert, data.PermsOnRequest, w)
}

type updateMssRequest struct {
	NotesUpdateField
	DisposedField
	SaleField
	WriteTagToField
	PermsOnRequest
}

func (req updateMssRequest) modsFor(existing *MSS, aclField AclField) (bson.D, error) {
	return NewMods().
		updateSaleIfNeeded(req.Sale, existing.Sale).
		updateDisposedIfNeeded(req, existing).
		updateNotesIfNeeded(req, existing).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateMssHandler(w http.ResponseWriter, r *http.Request) {
	data := updateMssRequest{}
	b58Id := Base58Str(r.PathValue("id"))
	defer r.Body.Close()
	id, err := b58Id.toMainCollectionId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
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
	ctx, db := Db(r)
	coll := db.Collection(MssCollectionName)

	// go get current entry
	existing := MSS{}
	err = coll.FindOne(ctx, bsonFindFilter("_id", id)).Decode(&existing)
	if err != nil {
		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	//Validation
	if data.Sale != nil && (existing.Sale == nil || *existing.Sale != *data.Sale) {
		if err = db.Collection(SalesCollectionName).FindOne(ctx, bsonFindFilter("_id", data.Sale)).Err(); err != nil {
			dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
			return
		}
	}
	finishMainCollItemUpdate(ctx, w, coll, data.modsFor, &existing, data.PermsOnRequest)
}
