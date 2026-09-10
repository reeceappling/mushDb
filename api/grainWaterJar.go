package api

import (
	"context"
	"encoding/json"
	"github.com/reeceappling/mushDb/api/env"
	"github.com/reeceappling/mushDb/api/request"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
)

// TODO: implement in ts, and reveal in endpoints!

type GrainWaterJar struct {
	AlternateCollectionIdField `bson:"inline"`
	GrainBatchField            `bson:"inline"`
	NotesField                 `bson:"inline"`
	CreationDateField          `bson:"inline"`
	LastUpdatedField           `bson:"inline"`
	AclField                   `bson:"inline"`
}

func createGrainWaterJarHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctx, now := request.UnixTime(r.Context())
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	req := createGrainWaterJarRequest{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	if _, err = req.GrainBatchField.Get(ctx); err != nil { // TODO; ensure ok
		http.Error(w, "failed to validate grain batch: "+err.Error(), http.StatusBadRequest)
		return
	}
	id := newAlternateCollectionId()
	toInsert := &GrainWaterJar{
		AlternateCollectionIdField: AlternateCollectionIdField{id},
		GrainBatchField:            req.GrainBatchField,
		NotesField:                 req.NotesField,
		CreationDateField:          CreationDateField{now},
		LastUpdatedField:           LastUpdatedField{now},
		AclField:                   allCanWriteAcl(),
	}
	finishCreateAlternateEntry(ctx, toInsert, w)
}

func initializeGrainWaterJars(ctx context.Context) error { // TODO: USE!
	// Indices
	coll := DbFrom(ctx).Collection(GrainWaterJarCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		// Id index already exists
		//Notes (no index unless tags)
		newSimpleIndex("creationDate", "creationDate", true, false, false),
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}

	return env.IfNotProd(ctx, func() error {

		testItem := GrainWaterJar{
			AlternateCollectionIdField: AlternateCollectionIdField{exAltId},
			GrainBatchField:            GrainBatchField{},   // TODO: ???
			CreationDateField:          CreationDateField{}, // TODO: ???
			NotesField:                 NotesField{exampleNotes()},
			LastUpdatedField:           LastUpdatedField{exampleTime},
		}
		println("test Grain Water Jar:", exAltId.AsBase58())
		return addTestAltEntries(ctx, testItem)
	})
}

type updateGrainWaterJarRequest struct {
	NotesUpdateField
	PermsOnRequest `json:"acl"`
}

func (req updateGrainWaterJarRequest) modsFor(existing *GrainWaterJar, acl AclField) (bson.D, error) {
	return NewMods().
		updateNotesIfNeeded(req, existing).
		updatePermsIfNeeded(acl.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateGrainWaterJarHandler(w http.ResponseWriter, r *http.Request) {
	_, id, err := altCollIdFromRequest(r, w)
	if err != nil {
		return
	}
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req := updateGrainWaterJarRequest{}
	err = json.Unmarshal(bs, &req)
	if err != nil {
		http.Error(w, "failed to unmarshal body: "+err.Error(), http.StatusBadRequest)
		return
	}
	ctx, db := Db(r)
	coll := db.Collection(GrainWaterJarCollectionName)
	existing, err := GetAltCollectionItem(ctx, id, &GrainWaterJar{})
	if err != nil {
		stat := http.StatusInternalServerError
		if err == mongo.ErrNoDocuments {
			stat = http.StatusNotFound
		}
		dbErr(w, err.Error(), stat)
		return
	}
	finishAltCollItemUpdate(ctx, w, coll, req.modsFor, existing, req.PermsOnRequest)
}
