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
)

const MssSourceType = "mss"

type MSS struct {
	// ALWAYS assume contaminated
	EntryTypeStructField // always mss
	MainCollectionIdField
	CreationDateField
	SpeciesField
	SubspeciesOptionalField
	// NOTE: parentType is always either sporePrint or purchased
	AlternateCollectionOptionalParentField // no parent means purchased, traded-for, or imported
	TransfersOutField
	SaleField
	DisposedField
	NotesField
	LastUpdatedField
	PermsField
}

func (M MSS) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := M
	err := decodeItem(&out, encoded)
	return out, err
}

func (M MSS) projects() []projectName {
	return M.Perms.Projects.Ids
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

func (M MSS) SourceType() string {
	return MssSourceType
}

func (M MSS) setTransferParent(ctx mongo.SessionContext, xfer Transfer) error { // TODO: I think these are the same for pretty much everywhere (except maybe sporeprint?), so we should get rid of this
	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
	if err != nil {
		return err
	}
	res, err := ctx.Client().Database(dbName).Collection(mainCollectionName).UpdateByID(ctx, M.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer
	}
	return nil
}

func (M MSS) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
	return errors.New("mss cannot be a child in a normal transfer. Must be created manually from spore print or imported")
}

func (M MSS) EntryTypeField() *string {
	return utils.Pointer(MssSourceType)
}

func (M MSS) CollectionName() string {
	return mainCollectionName
}

func initializeMSS(ctx context.Context) error {
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(mainCollectionName)
	// If test agar batch does not exist, then create it
	existingEntry := MSS{}
	testId := mainCollIdForint(idTestMSS)
	testItem := MSS{
		EntryTypeStructField:                   EntryTypeStructField{*existingEntry.EntryTypeField()},
		MainCollectionIdField:                  MainCollectionIdField{testId},
		CreationDateField:                      CreationDateField{exampleTime},
		SpeciesField:                           SpeciesField{testEntryStringId},
		SubspeciesOptionalField:                SubspeciesOptionalField{&testEntryStringId},
		TransfersOutField:                      TransfersOutField{exAlts},
		AlternateCollectionOptionalParentField: AlternateCollectionOptionalParentField{&exAltId},
		DisposedField:                          DisposedField{&exampleTime},
		NotesField:                             NotesField{exampleNotes()},
		LastUpdatedField:                       LastUpdatedField{exampleTime},
	}
	err := coll.FindOne(ctx, bson.D{{"_id", testId}}).Decode(&existingEntry)
	if err == nil {
		if reflect.DeepEqual(existingEntry, testItem) {
			return nil
		}
	}
	return testExistingEntry(ctx, coll, testId, testItem, existingEntry)
}

type createMssRequest struct {
	SporePrintId AlternateCollectionId
	NotesField
	WriteTagToField
	// Uses parent perms, then user can modify if they have the perms for parent
	// TODO: DESCRIBE RATIONALE
}

func createMssHandler(w http.ResponseWriter, r *http.Request) { // Only called from spore print page
	data := createMssRequest{}
	id, err := generateMainCollectionId(r.Context())
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
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		db := ctx.Client().Database(dbName)
		var parent SporePrint
		err = db.Collection(sporePrintCollectionName).FindOne(ctx, bson.D{{"_id", data.SporePrintId}}).Decode(&parent)
		if err != nil {
			return DbTxnStdErr(w, "failed to find sporePrint: "+err.Error(), http.StatusBadRequest)
		}
		err = writeRfidTagIfNecessary(ctx, data.WriteTagTo, id)
		if err != nil {
			return DbTxnStdErr(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		}
		now := unixTimeForNow()
		toInsert := MSS{
			EntryTypeStructField:                   EntryTypeStructField{"mss"},
			MainCollectionIdField:                  MainCollectionIdField{id},
			CreationDateField:                      CreationDateField{now},
			SpeciesField:                           SpeciesField{parent.Species},
			SubspeciesOptionalField:                parent.SubspeciesOptionalField,
			AlternateCollectionOptionalParentField: AlternateCollectionOptionalParentField{&parent.Id},
			NotesField:                             NotesField{data.Notes},
			LastUpdatedField:                       LastUpdatedField{now},
			PermsField:                             PermsField{parent.Perms}, // NOTE: do NOT ensure user is authorized to write on parent, they will just be blocked from viewing.
		}
		_, err = db.Collection(mainCollectionName).InsertOne(ctx, toInsert)
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

type importMssRequest struct {
	CreationDateField
	SpeciesField
	SubspeciesOptionalField
	NotesField
	WriteTagToField
	PermsField // TODO: new // TODO: SEND THIS OVER. ptr?
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
	id, err := generateMainCollectionId(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	authinfo, err := GetAuthInfo(r.Context())
	if err != nil {
		http.Error(w, "failed to get auth info: "+err.Error(), http.StatusUnauthorized)
		return
	}
	spec, subsp, err := getSpeciesAndSubspecies(r.Context(), data.Species, data.SubSpecies)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	finalPerms := minimalPermsBetween(data.Perms, spec, subsp)
	finalPerms.Users = finalPerms.Users.WithAuthor(authinfo.Id) // TODO: species and subspecies perms?
	toInsert := MSS{
		EntryTypeStructField:    EntryTypeStructField{"mss"},
		MainCollectionIdField:   MainCollectionIdField{id},
		CreationDateField:       data.CreationDateField,
		SpeciesField:            data.SpeciesField,
		SubspeciesOptionalField: data.SubspeciesOptionalField,
		NotesField:              data.NotesField,
		LastUpdatedField:        LastUpdatedField{unixTimeForNow()},
		PermsField:              PermsField{finalPerms},
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(mainCollectionName)
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

type updateMssRequest struct {
	Notes AllEntries[Note]
	DisposedField
	SaleField
	WriteTagToField
	PermsField
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
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		db := ctx.Client().Database(dbName)
		coll := db.Collection(mainCollectionName)
		// go get current plate
		existing := MSS{}
		err := coll.FindOne(ctx, bson.D{{"_id", id}}).Decode(&existing)
		if err != nil {
			return DbTxnStdErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		}
		if err = minimalPermsBetween(data.Perms, existing.Perms).ValidateUserCanWrite(ctx); err != nil {
			return DbTxnStdErr(w, "old or new perms not writeable by user: "+err.Error(), http.StatusBadRequest)
		}
		if data.Sale != nil && (existing.Sale == nil || *existing.Sale != *data.Sale) {
			if err = db.Collection(salesCollectionName).FindOne(ctx, bson.D{{"_id", data.Sale}}).Err(); err != nil {
				return DbTxnStdErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
			}
		}
		upd, err := NewMods().
			updateSaleIfNeeded(data.Sale, existing.Sale).
			updateDisposedIfNeeded(data.Disposed, existing.Disposed).
			updateNotesIfNeeded(data.Notes, existing.Notes).
			updatePermsIfNeeded(data.Perms, existing.Perms).
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
		err = coll.FindOneAndUpdate(ctx, bsonId, upd).Decode(&existing)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		bs, err = json.Marshal(existing)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		return w.Write(bs)
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}
