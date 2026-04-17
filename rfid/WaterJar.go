package rfid

import (
	"context"
	"encoding/json"
	"github.com/reeceappling/goUtils/v2/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
)

type WaterJar struct { // TODO: HANDLE THIS EVERYWHERE! DO ALL TYPESCRIPT FOR THIS!
	AlternateCollectionIdField `bson:"inline"`
	PcRunField                 `bson:"inline"` // Creation date assumed to be the same as pc run date
	NotesField                 `bson:"inline"`
	LastUpdatedField           `bson:"inline"`
}

func (wj WaterJar) Permissions() *ACL {
	return nil
}

type WaterJarOptionalField struct {
	WaterSource *AlternateCollectionId `bson:"waterSource,omitempty" json:"waterSource,omitempty"`
}

func (field WaterJarOptionalField) Get(ctx context.Context) (out PCRun, err error) {
	if field.WaterSource == nil {
		err = ErrMissingOptionalField
		return
	}
	return WaterJarField{*field.WaterSource}.Get(ctx)
}

type WaterJarField struct {
	WaterSource AlternateCollectionId `bson:"waterSource" json:"waterSource"`
}

func (field WaterJarField) Get(ctx context.Context) (out PCRun, err error) {
	err = ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(WaterJarsCollectionName).FindOne(ctx, bson.M{
		"_id": field.WaterSource,
	}).Decode(&out)
	return out, err
}

func initializeWaterJars(ctx context.Context) error { // TODO: use this!
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(WaterJarsCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		newSimpleIndex("pcRun", "pcRun", false, false, false),
		//Notes (no index unless tags)
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	testItem := &WaterJar{
		AlternateCollectionIdField: AlternateCollectionIdField{exAltId},
		PcRunField:                 PcRunField{exAltId},
		NotesField:                 NotesField{exampleNotes()},
		LastUpdatedField:           LastUpdatedField{exampleTime},
	}
	return addTestAltEntries(ctx, testItem)
}

type createWaterJarRequest struct {
	PcRunField
	NotesField
}

func createWaterJarHandler(w http.ResponseWriter, r *http.Request) { // TODO: THIS!
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req := createWaterJarRequest{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id := newAlternateCollectionId()
	ctx, db := Db(r)
	coll := db.Collection(WaterJarsCollectionName)
	toInsert := WaterJar{
		AlternateCollectionIdField: AlternateCollectionIdField{id},
		PcRunField:                 req.PcRunField,
		NotesField:                 req.NotesField,
		LastUpdatedField:           LastUpdatedField{unixTimeForNow()},
	}
	finishCreateAlternateEntry(ctx, coll, toInsert, w)
}

type updateWaterJarRequest struct {
	Notes AllEntries[Note] `json:"notes"`
}

func (req updateWaterJarRequest) modsFor(existing *WaterJar, aclField AclField) (bson.D, error) {
	return NewMods().
		updateNotesIfNeeded(req.Notes, existing.Notes).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateWaterJarHandler(w http.ResponseWriter, r *http.Request) {
	b58Id := Base58Str(r.PathValue("id"))
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req := updateWaterJarRequest{}
	err = json.Unmarshal(bs, &req)
	if err != nil {
		http.Error(w, "failed to unmarshal body: "+err.Error(), http.StatusBadRequest)
		return
	}
	id, err := b58Id.toAltCollectionId()
	if err != nil {
		http.Error(w, "Invalid id! "+err.Error(), http.StatusBadRequest)
		return
	}
	ctx, db := Db(r)
	coll := db.Collection(WaterJarsCollectionName)
	existing, err := GetAltCollectionItem(r.Context(), id, &WaterJar{})
	if err != nil {
		stat := http.StatusInternalServerError
		if err == mongo.ErrNoDocuments {
			stat = http.StatusNotFound
		}
		http.Error(w, err.Error(), stat)
		return
	}
	finishAltCollItemUpdate(ctx, w, coll, req.modsFor, existing, PermsOnRequest{BlanketPerm: utils.Pointer(ReadWritePerm(true))})
}
