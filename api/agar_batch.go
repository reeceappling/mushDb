package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/reeceappling/mushDb/api/env"
	"github.com/reeceappling/mushDb/api/request"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
)

// Required for
// CreatePlates
// CreateSlants (only if the slants are directly PC'd, which they should generally be)

type AgarBatch struct { // This is >=1 media bottles of the same recipe that went through the same PC cycle
	AlternateCollectionIdField `bson:"inline"`
	// CreationDate is assumed to be the same as on PcRun // Also exists in the ID
	PcRunField       `bson:"inline"`
	AgarRecipeField  `bson:"inline"`
	Color            Colorant `bson:"color" json:"color"`
	NotesField       `bson:"inline"`
	LastUpdatedField `bson:"inline"`
	AclField         `bson:"inline"`
}

//func (ab AgarBatch) Blank() CollectionItem {
//	return &AgarBatch{}
//}

type AgarBatchField struct {
	AgarBatch *AlternateCollectionId `bson:"agarBatch,omitempty" json:"agarBatch,omitempty"`
}

func (field AgarBatchField) Get(ctx context.Context) (out AgarBatch, err error) {
	if field.AgarBatch == nil {
		return out, ErrMissingOptionalField
	}
	err = DbFrom(ctx).Collection(AgarBatchCollectionName).FindOne(ctx, bson.M{
		IDfld: *field.AgarBatch,
	}).Decode(&out)
	return out, err
}

type updateAgarBatchRequest struct {
	NotesUpdateField
	PermsOnRequest `json:"acl"`
}

func (req updateAgarBatchRequest) modsFor(existing *AgarBatch, acl AclField) (bson.D, error) {
	return NewMods().
		updateNotesIfNeeded(req, existing).
		updatePermsIfNeeded(acl.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateAgarBatchHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	_, id, err := altCollIdFromRequest(r, w)
	if err != nil {
		return
	}
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
	db := DbFrom(ctx)
	coll := db.Collection(AgarBatchCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		newSimpleIndex("pcRun", "pcRun", false, true, false),
		newSimpleIndex("agarRecipe", "agarRecipe", false, false, false), // Required for deleting agar batches
		//newSimpleIndex("color", "color", false, false, false),
		projectsIndexModel,
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, and is dev, then create it
	return env.IfNotProd(ctx, func() error {
		testAltId := altCollIdForint(0)
		testItem := AgarBatch{
			AlternateCollectionIdField: AlternateCollectionIdField{Id: exAltId},
			PcRunField:                 PcRunField{testAltId},
			AgarRecipeField:            AgarRecipeField{testAltId},
			Color:                      clearColor,
			NotesField:                 NotesField{exampleNotes()},
			LastUpdatedField:           LastUpdatedField{exampleTime},
			AclField:                   allCanReadAcl(nil),
		}
		println("test Agar Batch:", testAltId.AsBase58())
		return addTestAltEntries(ctx, testItem)
	})
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
	ctx := r.Context()
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
	ctx, now := request.UnixTime(ctx)
	// create new batch
	toInsert := &AgarBatch{
		AlternateCollectionIdField: AlternateCollectionIdField{id},
		PcRunField:                 req.PcRunField,
		AgarRecipeField:            req.AgarRecipeField,
		Color:                      req.Color,
		NotesField:                 req.NotesField,
		LastUpdatedField:           LastUpdatedField{now},
		AclField:                   allCanWriteAcl(),
	}
	finishCreateAlternateEntry(ctx, toInsert, w)
}

//func deleteAgarBatchHandler(w http.ResponseWriter, r *http.Request) {
//	idStr := r.PathValue("id")
//	if idStr == "" {
//		http.Error(w, "Empty id for delete request", http.StatusBadRequest)
//		return
//	}
//	id, err := Base58Str(idStr).toAltCollectionId()
//	if err != nil {
//		http.Error(w, "Invalid ID to delete: "+err.Error(), http.StatusBadRequest)
//		return
//	}
//	// TODO: ensure batches not used anywhere else first
//
//	// Validate not used in other places...
//	ctx := r.Context()
//	db := DbFrom(ctx)
//	for _, collName := range []string{PlatesCollectionName, SlantsCollectionName} {
//		err = db.Collection(collName).FindOne(ctx, bson.M{"agarBatch": id}).Err()
//		if err != nil {
//			if !errors.Is(err, mongo.ErrNoDocuments) {
//				http.Error(w, "failed to check for agarBatch usage in collection "+collName+". "+err.Error(), http.StatusInternalServerError)
//				return
//			}
//		} else {
//			// At least one item exists, fail
//			http.Error(w, "at least one item in collection "+collName+" utilizes the item you are attempting to delete.", http.StatusExpectationFailed)
//			return
//		}
//	}
//	// Delete if not found elsewhere!
//	deleteResult, err := db.Collection(AgarBatchCollectionName).DeleteOne(ctx, bson.M{IDfld: id})
//	if err != nil {
//		http.Error(w, "failed to delete: "+err.Error(), http.StatusInternalServerError)
//		return
//	}
//	if deleteResult.DeletedCount == 0 {
//		http.Error(w, "failed to delete id "+idStr+" from agarBatch collection. Id not found", http.StatusNotFound)
//		return
//	}
//	_, err = w.Write([]byte(idStr))
//	handleWriteErr(err, w)
//}
