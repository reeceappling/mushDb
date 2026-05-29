package api

import (
	"context"
	"encoding/json"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
)

// Required for
// TODO: CreatePlates
// TODO: CreateSlants (still required if the slants are directly PC'd

type AgarBatch struct { // This is >=1 media bottles of the same recipe that went through the same PC cycle
	AlternateCollectionIdField `bson:"inline"`
	// CreationDate is assumed to be the same as on PcRun
	PcRunField       `bson:"inline"`
	AgarRecipeField  `bson:"inline"`
	Color            Colorant `bson:"color" json:"color"`
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

//type NewAgarBatchRequest struct {
//	PcRunField
//	AgarRecipeField
//	Color *string
//	NotesCreationField
//}

type updateAgarBatchRequest struct {
	NotesUpdateField
	PermsOnRequest
}

func (req updateAgarBatchRequest) modsFor(existing *AgarBatch, acl AclField) (bson.D, error) {
	return NewMods().
		updateNotesIfNeeded(req, existing).
		updatePermsIfNeeded(acl.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

// TODO: MOVE
func ReadSimpleStructuredBody[T any](r *http.Request, w http.ResponseWriter, req *T) error {
	defer r.Body.Close()
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		println("failed to read body: " + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	if err = json.Unmarshal(bytes, &req); err != nil {
		println("bad body format: " + string(bytes))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	return nil
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
	ctx, db := Db(r)
	coll := db.Collection(AgarBatchCollectionName)
	existing, err := GetAltCollectionItemOutsideTxn(ctx, id, AgarBatch{})
	if err != nil {
		stat := http.StatusInternalServerError
		if errors.Is(err, mongo.ErrNoDocuments) {
			stat = http.StatusNotFound
		}
		http.Error(w, err.Error(), stat)
		return
	}
	finishAltCollItemUpdate(ctx, w, coll, req.modsFor, &existing, req.PermsOnRequest)
}

func initializeAgarBatches(ctx context.Context) error {
	// Indices
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(AgarBatchCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		newSimpleIndex("pcRun", "pcRun", false, true, false),
		newSimpleIndex("recipe", "recipe", false, false, false),
		//newSimpleIndex("color", "color", false, false, false),
		projectsIndexModel,
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it

	testAltId := altCollIdForint(0)
	testItem := AgarBatch{
		AlternateCollectionIdField: AlternateCollectionIdField{Id: exAltId},
		PcRunField:                 PcRunField{testAltId},
		AgarRecipeField:            AgarRecipeField{testAltId},
		Color:                      clearColor,
		NotesField:                 NotesField{exampleNotes()},
		LastUpdatedField:           LastUpdatedField{exampleTime},
		AclField:                   allCanReadAcl(),
	}
	err = addTestAltEntries(ctx, testItem)
	println("test Agar Batch:", testAltId.AsBase58())
	return err
}

type createAgarBatchRequest struct {
	Color Colorant `json:"color"`
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
	ctx, db := Db(r)
	coll := db.Collection(AgarBatchCollectionName)
	// Validate fields
	_, err = req.PcRunField.Get(ctx)

	if err != nil {
		dbErr(w, "PcRun validation failure: "+err.Error(), http.StatusBadRequest)
		return
	}
	_, err = req.AgarRecipeField.Get(ctx)
	if err != nil {
		dbErr(w, "Agar recipe validation failure: "+err.Error(), http.StatusBadRequest)
		return
	}
	// create new batch
	toInsert := &AgarBatch{
		AlternateCollectionIdField: AlternateCollectionIdField{id},
		PcRunField:                 req.PcRunField,
		AgarRecipeField:            req.AgarRecipeField,
		Color:                      req.Color,
		NotesField:                 req.NotesField,
		LastUpdatedField:           LastUpdatedField{unixTimeForNow()},
		AclField:                   allCanWriteAcl(),
	}
	finishCreateAlternateEntry(ctx, coll, toInsert, w)
}
