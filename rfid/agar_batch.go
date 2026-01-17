package rfid

import (
	"context"
	"encoding/json"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"reflect"
)

// TODO: FIX ALL REQUESTS AND HANDLERS FOR IMPORTS, UPDATES, and CREATES

// TODO: SPORE SWABS?!?!?!?!
// TODO: PEGS?????!!?!?!?! Oak, Poplar, Bamboo
type AgarBatch struct { // This is >=1 media bottles of the same recipe that went through the same PC cycle
	AlternateCollectionIdField `bson:"inline"`
	// CreationDate is assumed to be the same as on PcRun
	PcRunField       `bson:"inline"`
	AgarRecipeField  `bson:"inline"`
	Color            colorant `bson:"color" json:"color"`
	NotesField       `bson:"inline"`
	LastUpdatedField `bson:"inline"`
	AclField         `bson:"inline"`
}

type AgarBatchField struct {
	AgarBatch *AlternateCollectionId `bson:"agarBatch,omitempty" json:"agarBatch,omitempty"`
}

func (field AgarBatchField) Get(ctx context.Context) (out AgarBatch, err error) {
	if field.AgarBatch == nil {
		return out, ErrMissingOptionalField
	}
	err = ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(AgarBatchCollectionName).FindOne(ctx, bson.M{
		"_id": *field.AgarBatch,
	}).Decode(&out)
	return out, err
}

func (batch AgarBatch) EntryTypeField() *string {
	return nil
}

type NewAgarBatchRequest struct {
	PcRunField
	AgarRecipeField
	Color *string
	Notes []string
}

func newAgarBatch(ctx mongo.SessionContext, batch AgarBatch) (*AgarBatch, error) {
	out, err := ctx.Client().Database(dbName).Collection(AgarBatchCollectionName).InsertOne(ctx, batch)
	if err != nil {
		return nil, err
	}
	res, ok := out.InsertedID.(AlternateCollectionId)
	if !ok {
		return nil, errors.New("failed to convert to primitive.ObjectID")
	}
	batch.Id = res
	return &batch, err
}

type updateAgarBatchRequest struct {
	Notes          AllEntries[Note] `json:"notes"`
	PermsOnRequest                  // TODO: handle in typescript and handler!
}

func (req updateAgarBatchRequest) modsFor(existing AgarBatch, acl *ACL) (bson.D, error) {
	return NewMods().
		updateNotesIfNeeded(req.Notes, existing.Notes).
		updatePermsIfNeeded(acl, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateAgarBatchHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	b58Id := Base58Str(r.PathValue("id"))
	req := updateAgarBatchRequest{}
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(bytes, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := b58Id.toAltCollectionId()
	if err != nil {
		http.Error(w, "Invalid id! "+err.Error(), http.StatusBadRequest)
		return
	}

	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(AgarBatchCollectionName)
		existing, err := GetAltCollectionItemInTxn(ctx, id, AgarBatch{})
		if err != nil {
			stat := http.StatusInternalServerError
			if errors.Is(err, mongo.ErrNoDocuments) {
				stat = http.StatusNotFound
			}
			http.Error(w, err.Error(), stat)
			return nil, ErrInTxnAlreadyTriedToWrite
		}
		user, err := GetAuthInfo(ctx)
		if err != nil {
			return dbErr(w, err.Error(), http.StatusInternalServerError)
		}
		if !user.HasPermissionToEdit(existing) {
			return dbErr(w, "unauthorized to edit", http.StatusForbidden)
		}
		aclField, err := req.AclFor(ctx, user)
		if err != nil {
			return dbErr(w, err.Error(), http.StatusInternalServerError)
		}

		upd, err := req.modsFor(existing, aclField.ACL)
		return handleUpdateMods(ctx, w, coll, existing, id, upd, err)
	})

	if err != nil {
		handleWriteErr(err, w)
	}
}

func initializeAgarBatches(ctx context.Context) error {
	// Indices
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(AgarBatchCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		newSimpleIndex("pcRun", "pcRun", false, true, false),
		newSimpleIndex("recipe", "recipe", false, false, false),
		newSimpleIndex("color", "color", false, false, false),
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	testAltId := altCollIdForint(0)
	existingEntry := AgarBatch{}
	testItem := AgarBatch{
		AlternateCollectionIdField: AlternateCollectionIdField{Id: exAltId},
		PcRunField:                 PcRunField{testAltId},
		AgarRecipeField:            AgarRecipeField{testAltId},
		Color:                      clearColor,
		NotesField:                 NotesField{exampleNotes()},
		LastUpdatedField:           LastUpdatedField{exampleTime},
		AclField:                   allCanReadAcl(),
	}
	err = coll.FindOne(ctx, bson.D{{"_id", exAltId}}).Decode(&existingEntry)
	if err == nil {
		if reflect.DeepEqual(existingEntry, testItem) {
			return nil
		}
	}
	return testExistingEntry(ctx, coll, exAltId, testItem, existingEntry)
}

type createAgarBatchRequest struct {
	Color colorant `json:"color"`
	PcRunField
	AgarRecipeField
	NotesField
}

func createAgarBatchHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	req := createAgarBatchRequest{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id := newAlternateCollectionId()
	if !ValidColor(req.Color) {
		http.Error(w, "Invalid agar color!", http.StatusBadRequest)
		return
	}

	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		// Validate fields
		_, err = req.PcRunField.Get(ctx)
		if err != nil {
			return dbErr(w, "PcRun validation failure: "+err.Error(), http.StatusBadRequest)
		}
		_, err = req.AgarRecipeField.Get(ctx)
		if err != nil {
			return dbErr(w, "Agar recipe validation failure: "+err.Error(), http.StatusBadRequest)
		}
		// create new batch
		newBatch := AgarBatch{
			AlternateCollectionIdField: AlternateCollectionIdField{id},
			PcRunField:                 req.PcRunField,
			AgarRecipeField:            req.AgarRecipeField,
			Color:                      req.Color,
			NotesField:                 req.NotesField,
			LastUpdatedField:           LastUpdatedField{unixTimeForNow()},
			AclField:                   allCanWriteAcl(),
		}
		batch, err := newAgarBatch(ctx, newBatch)
		if err != nil {
			return dbErr(w, "Agar batch creation failure: "+err.Error(), http.StatusInternalServerError)
		}
		bs, err := json.Marshal(*batch)
		if err != nil {
			return dbErr(w, "Agar batch marshal failure: "+err.Error(), http.StatusInternalServerError)
		}
		return w.Write(bs)
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}
