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
	"reflect"
	"slices"
)

// TODO: new are this, sporeSwab, agarBottle, plugs

// Naming convention "{ParentLCID}-#"

//func parseName() // TODO: ???

const (
	lcSyringeSourceType     = "lcSyringe"
	lcSyringeCollectionName = "lcSyringes"
	lcSyringeIdPrefix       = "LCS"
)

type LcSyringe struct {
	MainCollectionIdField `bson:"inline"`
	// Parent is always either purchased (nil), LC, or LcSyringe
	MainCollectionOptionalParentField `bson:"inline"` // TODO: likely won't exist for pre-existing
	CreationDateField                 `bson:"inline"` // create or receive date
	SpeciesField                      `bson:"inline"`
	SubspeciesOptionalField           `bson:"inline"`
	SaleField                         `bson:"inline"`
	GenerationsFields                 `bson:"inline"`
	KnownFruitableField               `bson:"inline"` // TODO: NEW! HANDLE EVERYWHERE!
	ConfirmedCleanField               `bson:"inline"`
	TransfersOutField                 `bson:"inline"`
	DisposedField                     `bson:"inline"`
	NotesField                        `bson:"inline"`
	LastUpdatedField                  `bson:"inline"`
	AclField                          `bson:"inline"` // TODO: handle EVERYWHERE
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

func (sw LcSyringe) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := sw
	err := decodeItem(&out, encoded)
	return out, err
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

func (sw LcSyringe) SourceType() string {
	return lcSyringeSourceType
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

func (sw LcSyringe) prefix() string {
	return lcSyringeIdPrefix
}

func (sw LcSyringe) CollectionName() string {
	return lcSyringeCollectionName
}

func initializeSyringes(ctx context.Context) error { // TODO; this
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(lcSyringeCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		newSimpleIndex("parent", "parent", false, true, false),
		newSimpleIndex("creationDate", "creationDate", true, false, false), // TODO: INDEX CREATION DATES EVERYWHERE!
		newSimpleIndex("species", "species", false, false, false),
		newSimpleIndex("subSpecies", "subSpecies", false, true, false),
		saleIndexModel,
		newSimpleIndex("genSinceSpore", "genSpore", true, true, false),
		newSimpleIndex("genSinceFruitOrSpore", "genFruitOrSpore", true, true, false),
		newSimpleIndex("knownFruitable", "knownFruitable", false, true, false),
		newSimpleIndex("confirmedClean", "confirmedClean", false, true, false),
		transfersOutIndexModel,
		newSimpleIndex("disposed", "disposed", false, true, false),
		//// TODO: Projects?
		//Notes (no index unless tags)
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	existingEntry := LcSyringe{}
	testItem := LcSyringe{
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
	err = coll.FindOne(ctx, bson.D{{"_id", exAltId}}).Decode(&existingEntry)
	if err == nil {
		if reflect.DeepEqual(existingEntry, testItem) {
			return nil
		}
	}
	return testExistingEntry(ctx, coll, exAltId, testItem, existingEntry)
}

type createLCSyringeRequest struct {
	LC MainCollectionId `json:"lc"`
	NotesField
	WriteTagToField
}

// TODO: USE
func createSyringeHandler(w http.ResponseWriter, r *http.Request) {
	data := createLCSyringeRequest{}
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := newCollectionId(r.Context(), lcSyringeCollectionName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
	_, txErr := doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		db := ctx.Client().Database(dbName)
		parent := &LiquidCulture{}
		err = db.Collection(LCCollectionName).FindOne(ctx, bson.D{{"_id", id}}).Decode(&parent)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		if parent.Species == nil {
			return DbTxnStdErr(w, "Parent LC must be innoculated", http.StatusInternalServerError)
		}
		now := unixTimeForNow()
		if err = addToIdMapCollectionInTxn(ctx, id.ToBinaryCollectionId(), lcSyringeSourceType); err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
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
		// TODO: ADD TO MAP
		_, err = db.Collection(lcSyringeCollectionName).InsertOne(ctx, toInsert)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		bsOut, err := json.Marshal(toInsert)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		return w.Write(bsOut)
	})
	if txErr != nil {
		handleWriteErr(txErr, w)
	}
}

type updateSyringeRequest struct {
	SaleField // TODO: validate?
	DisposedField
	ConfirmedClean      *bool `json:"confirmedClean,omitempty"` // TODO: handle in react
	KnownFruitableField       // TODO: handle in react
	Notes               AllEntries[Note]
	PermsOnRequest      // TODO: handle in typescript and handler!
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
	Notes          AllEntries[Note]
	PermsOnRequest // TODO: handle in typescript and handler!
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
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(lcSyringeCollectionName)
		// go get current LcSyringe
		existing := LcSyringe{}
		err = coll.FindOne(ctx, bson.D{{"_id", id}}).Decode(&existing)
		if err != nil {
			return DbTxnStdErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		}
		user, err := GetAuthInfo(ctx)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		if !user.HasPermissionToEdit(existing) {
			return DbTxnStdErr(w, "unauthorized to edit", http.StatusForbidden)
		}
		aclField, err := data.AclFor(ctx, user)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		//if err = minimalPermsBetween(existing.Perms, data.Perms).ValidateUserCanWrite(ctx); err != nil {
		//	return DbTxnStdErr(w, "failed to validate overlapping permissions: "+err.Error(), http.StatusBadRequest)
		//}
		mds := NewMods()
		updatePointerIfNeeded(mds, "confirmedClean", out.ConfirmedClean, existing.ConfirmedClean)
		upd, err := mds.
			updateSaleIfNeeded(out.Sale, existing.Sale).
			updateDisposedIfNeeded(data.Disposed, existing.Disposed).
			updateKnownFruitableIfNeeded(data.KnownFruitable, existing.KnownFruitable).
			updateNotesIfNeeded(data.Notes, existing.Notes).
			updatePermsIfNeeded(aclField.ACL, existing.ACL).
			updateLastUpdatedIfNeeded().
			Finalized()
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
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		bsOut, err := json.Marshal(existing)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		return w.Write(bsOut)
	})
	if err != nil {
		HandleHttpWriteError(err)
	}
}

type importLcSyringeRequest struct {
	CreationDateField
	SpeciesField
	SubspeciesOptionalField
	KnownFruitableField
	NotesField
	// pic as "img"
	PermsOnRequest // TODO: handle in typescript and handler!
}

func importLcSyringeHandler(w http.ResponseWriter, r *http.Request) {
	data := importLcSyringeRequest{}
	id, err := newCollectionId(r.Context(), lcSyringeCollectionName)
	if err != nil {
		http.Error(w, "failed to create new mainCollectionId", http.StatusInternalServerError)
		// TODO: err
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

	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		//finalPerms := data.Perms
		//if data.Perms != nil {
		//	spec, subsp, err := getSpeciesAndSubspecies(ctx, data.Species, data.SubSpecies)
		//	if err != nil {
		//		return DbTxnStdErr(w, "failed to get species or subspecies: "+err.Error(), http.StatusInternalServerError) // TODO: ok?
		//	}
		//	finalPerms = minimalPermsBetween(spec, subsp)
		//	// TODO: add email perms if provided, as well as make email author?
		//	if !finalPerms.Valid() {
		//		// TODO: invalid species/subspecies perm crossover. DO THIS ELSEwHERE
		//		return DbTxnStdErr(w, "invalid species/subspecies perm crossover: "+err.Error(), http.StatusInternalServerError) // TODO: ok?
		//	}
		//}

		perms, err := GetAuthInfo(ctx)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		acl, err := data.AclFor(ctx, perms)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		toInsert := LcSyringe{
			MainCollectionIdField:   MainCollectionIdField{Id: id},
			CreationDateField:       data.CreationDateField,
			SpeciesField:            data.SpeciesField,
			KnownFruitableField:     data.KnownFruitableField,
			SubspeciesOptionalField: data.SubspeciesOptionalField,
			NotesField:              data.NotesField,
			LastUpdatedField:        LastUpdatedFieldForNow(),
			AclField:                acl,
		}
		// TODO: ADD TO MAP
		coll := ctx.Client().Database(dbName).Collection(lcSyringeCollectionName)
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
