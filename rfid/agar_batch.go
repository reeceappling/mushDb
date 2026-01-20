package rfid

import (
	"context"
	"encoding/json"
	"errors"
	sliceutils "github.com/reeceappling/goUtils/v2/utils/slices"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
)

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

type updateAgarBatchRequest struct {
	Notes AllEntries[Note] `json:"notes"`
	PermsOnRequest
}

func (req updateAgarBatchRequest) modsFor(existing *AgarBatch, acl AclField) (bson.D, error) {
	return NewMods().
		updateNotesIfNeeded(req.Notes, existing.Notes).
		updatePermsIfNeeded(acl.ACL, existing.ACL).
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
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
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
	return addTestAltEntries(ctx, testItem) // TODO: do this everywhere
}

// TODO: move
func addTestAltEntries[T AltCollectionItem[U], U AltCollectionIdType](ctx context.Context, testItems ...T) error {
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(testItems[0].CollectionName())
	_, err := coll.BulkWrite(ctx, sliceutils.Map(testItems, func(item T) mongo.WriteModel {
		return mongo.NewUpdateOneModel().SetFilter(bson.M{"_id": item.DbId()}).SetUpsert(true)
	}))
	// TODO: do something with the result?
	return err
}

// TODO: move
func addBasicAltEntries[T AltCollectionItem[U], U AltCollectionIdType](ctx context.Context, testItems ...T) error {
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(testItems[0].CollectionName())
	_, err := coll.BulkWrite(ctx, sliceutils.Map(testItems, func(item T) mongo.WriteModel {
		return mongo.NewInsertOneModel().SetDocument(item)
	}))
	if err != nil {
		println("error adding basic alt entries: " + err.Error()) // TODO: del
	}
	// TODO: do something with the result?
	return err
}

// TODO: move
func addTestMainEntries[T MainCollectionItem](ctx context.Context, testItems ...T) error {
	_, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).
		Collection(idMapCollectionName).BulkWrite(ctx, sliceutils.Map(testItems, func(item T) mongo.WriteModel {
		return mongo.NewReplaceOneModel().SetReplacement(idMapEntry{
			Id:        item.DbId(),
			EntryType: item.EntryType(),
		}).SetUpsert(true)
	}))
	// TODO: do something with the result?
	if err != nil {
		return err
	}

	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(testItems[0].CollectionName())
	_, err = coll.BulkWrite(ctx, sliceutils.Map(testItems, func(item T) mongo.WriteModel {
		return mongo.NewUpdateOneModel().SetFilter(bson.M{"_id": item.DbId()}).SetUpsert(true)
	}))
	// TODO: do something with the result?
	return err
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
