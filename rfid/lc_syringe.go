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
	"slices"
)

// TODO: new are this, sporeSwab, agarBottle, plugs

// Naming convention "{ParentLCID}-#"

//func parseName() // TODO: ???

type LcSyringe struct {
	MainCollectionIdField `bson:"inline"`
	// Parent is always either purchased (nil), LC, or LcSyringe
	MainCollectionOptionalParentField `bson:"inline"` // TODO: likely won't exist for pre-existing
	CreationDateField                 `bson:"inline"` // create or receive date
	SpeciesField                      `bson:"inline"`
	SubspeciesOptionalField           `bson:"inline"`
	SaleField                         `bson:"inline"`
	GenerationsFields                 `bson:"inline"`
	KnownFruitableField               `bson:"inline"`
	ConfirmedCleanField               `bson:"inline"`
	TransfersOutField                 `bson:"inline"`
	DisposedField                     `bson:"inline"`
	NotesField                        `bson:"inline"`
	LastUpdatedField                  `bson:"inline"`
	AclField                          `bson:"inline"`
}

func (lcs LcSyringe) Innoculatable() bool {
	return false
}

func (lcs LcSyringe) CanTransferTo(dst geneticSource) error {
	if !slices.Contains([]string{BagSourceType, GrainJarSourceType, LcSourceType, PlateSourceType, PlugSourceType, SlantSourceType}, dst.SourceType()) {
		return errors.New("lc syringe cannot transfer to " + dst.SourceType())
	}
	return nil
}

func (sw LcSyringe) setTransferParent(ctx context.Context, xfer Transfer) (error, func() error) {
	// TODO: can this even happen?
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

func (sw LcSyringe) setTransferChild(ctx context.Context, xfer Transfer, from geneticSource) error {
	// TODO: cannot happen
	panic("does not happen")
}

func (sw LcSyringe) GeneticInfoAsParent() (GeneticParentInfo, error) {
	return GeneticParentInfo{
		SpeciesOptionalField:    sw.SpeciesField.AsOptional(),
		SubspeciesOptionalField: sw.SubspeciesOptionalField,
		GenerationsFields:       GenerationsFieldFor(utils.Pointer(Generation(0))),
	}, nil
}

func (sw LcSyringe) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return utils.Pointer(Generation(0)), utils.Pointer(Generation(0))
}

func (sw LcSyringe) EntryTypeField() *string {
	return nil
}

func (sw LcSyringe) altId() MainCollectionId {
	return sw.Id
}

func (sw LcSyringe) id() []byte {
	return []byte(sw.Id.dbIdStr())
}

//func (sp LcSyringe) knownFruitable() bool {
//	return false
//}

func initializeSyringes(ctx context.Context) error { // TODO; this
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(LcSyringeCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		//newSimpleIndex("parent", "parent", false, true, false),
		newSimpleIndex("creationDate", "creationDate", true, false, false), // TODO: INDEX CREATION DATES EVERYWHERE!
		newSimpleIndex("species", "species", false, false, false),
		newSimpleIndex("subspecies", "subspecies", false, true, false),
		//saleIndexModel,
		//newSimpleIndex("genSinceSpore", "genSpore", true, true, false),
		//newSimpleIndex("genSinceFruitOrSpore", "genFruitOrSpore", true, true, false),
		//newSimpleIndex("knownFruitable", "knownFruitable", false, true, false),
		//newSimpleIndex("confirmedClean", "confirmedClean", false, true, false),
		//transfersOutIndexModel,
		//newSimpleIndex("disposed", "disposed", false, true, false),
		//// TODO: Projects?
		//Notes (no index unless tags)
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	testItem := &LcSyringe{
		MainCollectionIdField:             MainCollectionIdField{Id: exLC}, // TODO: FIX!
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&exLC},
		CreationDateField:                 exampleTime.asCreationDate(),
		SpeciesField:                      SpeciesField{testEntryStringId},
		SubspeciesOptionalField:           SubspeciesOptionalField{&testEntryStringId},
		SaleField:                         SaleField{&exAltId},
		DisposedField:                     DisposedField{&exampleTime},
		NotesField:                        NotesField{exampleNotes()},
		LastUpdatedField:                  LastUpdatedField{exampleTime},
		//PermsField:                        PermsField{}, // TODO: fix
	}
	return addTestMainEntries(ctx, testItem)
}

type createLCSyringeRequest struct {
	LC MainCollectionId `json:"lc"`
	NotesField
	WriteTagToField
}

func createSyringeHandler(w http.ResponseWriter, r *http.Request) {
	data := createLCSyringeRequest{}
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id := <-NextMainCollectionIdChan(r.Context())
	//id, err := newMainCollectionId(r.Context(), LcSyringeCollectionName)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
	err = json.Unmarshal(bs, &data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx, db := Db(r)
	coll := db.Collection(LcSyringeCollectionName)
	// Validate inputs and grab parent
	parent := &LiquidCulture{}
	err = db.Collection(LCCollectionName).FindOne(ctx, bson.D{{"_id", id}}).Decode(&parent)
	if err != nil {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if parent.Species == nil {
		dbErr(w, "Parent LC must be innoculated", http.StatusInternalServerError)
		return
	}
	now := unixTimeForNow()
	toInsert := LcSyringe{
		MainCollectionIdField:             MainCollectionIdField{Id: id},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&parent.Id},
		ConfirmedCleanField:               parent.ConfirmedCleanField, // TODO: is this ok?
		KnownFruitableField:               parent.KnownFruitableField,
		CreationDateField:                 now.asCreationDate(),
		SpeciesField:                      SpeciesField{Species: *parent.Species},
		SubspeciesOptionalField:           parent.SubspeciesOptionalField,
		GenerationsFields:                 parent.GenerationsFields,
		NotesField:                        NotesField{data.Notes},
		LastUpdatedField:                  LastUpdatedField{now},
		// Do not check permissions, just pass parent perms to child
		AclField: parent.AclField,
	}
	finishCreateMainCollectionEntry(ctx, coll, &toInsert, w)
}

type updateSyringeRequest struct {
	SaleField // TODO: validate?
	DisposedField
	ConfirmedClean      *bool `json:"confirmedClean,omitempty"` // TODO: handle in react
	KnownFruitableField       // TODO: handle in react
	Notes               AllEntries[Note]
	PermsOnRequest
}

func (upr updateSyringeRequest) reform() resolvedUpdateSyringeRequest {
	return resolvedUpdateSyringeRequest{
		SaleField:           upr.SaleField,
		DisposedField:       upr.DisposedField,
		ConfirmedClean:      upr.ConfirmedClean,
		KnownFruitableField: upr.KnownFruitableField,
		Notes:               upr.Notes,
		PermsOnRequest:      upr.PermsOnRequest,
	}
}

type resolvedUpdateSyringeRequest struct {
	SaleField
	DisposedField
	ConfirmedClean *bool `json:"confirmedClean,omitempty"`
	KnownFruitableField
	Notes AllEntries[Note]
	PermsOnRequest
}

func (mods resolvedUpdateSyringeRequest) modsFor(existing *LcSyringe, aclField AclField) (bson.D, error) {
	mds := NewMods()
	updatePointerIfNeeded(mds, "confirmedClean", mods.ConfirmedClean, existing.ConfirmedClean)
	return mds.updateSaleIfNeeded(mods.Sale, existing.Sale).
		updateDisposedIfNeeded(mods.Disposed, existing.Disposed).
		updateKnownFruitableIfNeeded(mods.KnownFruitable, existing.KnownFruitable).
		updateNotesIfNeeded(mods.Notes, existing.Notes).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateSyringeHandler(w http.ResponseWriter, r *http.Request) {
	data := updateSyringeRequest{}
	b58Id := Base58Str(r.PathValue("id")) // TODO: ensure ok
	id, err := b58Id.toMainCollectionId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	// CHECK THAT ALL NEW PICS EXIST
	// PROCESS ALL NEW PICS AND CONTAMS
	out := data.reform()
	ctx, db := Db(r)
	coll := db.Collection(LcSyringeCollectionName)
	// go get current LcSyringe
	existing := LcSyringe{}
	err = coll.FindOne(ctx, bson.D{{"_id", id}}).Decode(&existing)
	if err != nil {
		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishMainCollItemUpdate(ctx, w, coll, out.modsFor, &existing, out.PermsOnRequest)
}

type importLcSyringeRequest struct {
	CreationDateField
	SpeciesField
	SubspeciesOptionalField
	KnownFruitableField
	NotesField
	// pic as "img"
	PermsOnRequest
}

// TODO: USE!!!
func importLcSyringeHandler(w http.ResponseWriter, r *http.Request) {
	data := importLcSyringeRequest{}
	id := <-NextMainCollectionIdChan(r.Context())
	//id, err := newMainCollectionId(r.Context(), LcSyringeCollectionName)
	//if err != nil {
	//	http.Error(w, "failed to create new mainCollectionId", http.StatusInternalServerError)
	//	// TODO: err
	//}
	defer r.Body.Close()
	// Process text (or object)
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "unable to read Data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	// PARSE INTO CORRECT DATA FORMAT
	err = json.Unmarshal(bs, &data)
	if err != nil {
		http.Error(w, "unable to unmarshal json form Data: "+err.Error(), http.StatusBadRequest)
		return
	}

	user, err := GetAuthInfo(r.Context())
	if err != nil {
		http.Error(w, "failed to get auth info: "+err.Error(), http.StatusUnauthorized)
		return
	}
	sp, subsp, err := getSpeciesAndSubspecies(r.Context(), data.Species, data.SubSpecies)
	if err != nil {
		http.Error(w, "failed to get species or subspecies: "+err.Error(), http.StatusInternalServerError) // TODO: normalize
		return
	}
	var finalPerms *ACL = nil
	if subsp != nil {
		finalPerms = subsp.DefaultAcl.Clone()
	} else {
		sp.DefaultAcl.Clone()
	}
	// Add user to the acl as a writer
	finalPerms.Users[user.Email] = true

	ctx, db := Db(r)
	coll := db.Collection(LcSyringeCollectionName)
	// TODO: validate species, subspecies,
	toInsert := LcSyringe{
		MainCollectionIdField:   MainCollectionIdField{Id: id},
		CreationDateField:       data.CreationDateField,
		SpeciesField:            data.SpeciesField,
		KnownFruitableField:     data.KnownFruitableField,
		SubspeciesOptionalField: data.SubspeciesOptionalField,
		NotesField:              data.NotesField,
		LastUpdatedField:        LastUpdatedFieldForNow(),
		AclField:                AclField{finalPerms},
	}
	finishImportMainCollectionEntry(ctx, coll, &toInsert, data.PermsOnRequest, w)
}
