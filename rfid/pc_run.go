package rfid

// TODO: needed directly for:
// TODO: AgarBatch, Bag, Jar, LC, Plugs, Slant, StasisTube, WaterJar

// TODO: needed for but provided from another:
// TODO:  Plate, MSS,

// TODO: ? includeStasisTubes, includeAgarBatches, includePlugsJar
// TODO: newAgarBatch (on agar batch page)

// TODO: fully test this functionality first!!!!

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
)

type PCRun struct {
	AlternateCollectionIdField `bson:"inline"`
	CreationDateField          `bson:"inline"` //TODO: USED TO BE date, is now CreationDate
	RunTimeMinutes             int             `bson:"runtimeMinutes" json:"runtimeMinutes"` // todo; used to just be runtime, also used to be string
	NotesField                 `bson:"inline"`
	LastUpdatedField           `bson:"inline"`
	AclField                   `bson:"inline"`
}

func initializePCRun(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(PcRunCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		creationDateIndexModel,
		// TODO: newSimpleIndex("runtimeMinutes","runtimeMinutes", true, false, false),
		//RunTime (likely no index)    string                `bson:"runtime" json:"runtime"`
		//Notes (no index unless tags)
		projectsIndexModel,
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test run does not exist, then create it
	testItem := &PCRun{
		AlternateCollectionIdField: AlternateCollectionIdField{exAltId},
		CreationDateField:          CreationDateField{exampleTime},
		RunTimeMinutes:             60,
		NotesField:                 NotesField{exampleNotes()},
		LastUpdatedField:           LastUpdatedField{exampleTime},
		AclField:                   allCanReadAcl(),
	}
	return addTestAltEntries(ctx, testItem)
}

type createPcRunRequest struct {
	CreationDateField
	RunTimeMinutes int
	NotesField
}

func createPcRunHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req := createPcRunRequest{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.RunTimeMinutes < 10 {
		http.Error(w, "runtime must be greater than 10 minutes", http.StatusBadRequest)
	}
	id := newAlternateCollectionId()
	ctx, db := Db(r)
	coll := db.Collection(PcRunCollectionName)
	toInsert := PCRun{
		AlternateCollectionIdField: AlternateCollectionIdField{id},
		CreationDateField:          req.CreationDateField,
		RunTimeMinutes:             req.RunTimeMinutes,
		NotesField:                 req.NotesField,
		LastUpdatedField:           LastUpdatedField{unixTimeForNow()},
		AclField:                   allCanWriteAcl(), // TODO: ? acl, err := newAlwaysReadableAcl(ctx, resolvedUserPerms, nil, nil)
	}
	finishCreateAlternateEntry(ctx, coll, toInsert, w)
}

type updatePcRunRequest struct {
	NotesUpdateField
	PermsOnRequest
}

func (req updatePcRunRequest) modsFor(existing *PCRun, aclField AclField) (bson.D, error) {
	return NewMods().
		updateNotesIfNeeded(req, existing).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updatePcRunHandler(w http.ResponseWriter, r *http.Request) {
	b58Id := Base58Str(r.PathValue("id"))
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req := updatePcRunRequest{}
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
	coll := db.Collection(PcRunCollectionName)
	existing, err := GetAltCollectionItem(r.Context(), id, &PCRun{})
	if err != nil {
		stat := http.StatusInternalServerError
		if err == mongo.ErrNoDocuments {
			stat = http.StatusNotFound
		}
		http.Error(w, err.Error(), stat)
		return
	}
	finishAltCollItemUpdate(ctx, w, coll, req.modsFor, existing, req.PermsOnRequest)
}

type PcRunField struct {
	PcRun AlternateCollectionId `bson:"pcRun" json:"pcRun"`
}

func (field PcRunField) Get(ctx context.Context) (out PCRun, err error) {
	err = ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(PcRunCollectionName).FindOne(ctx, bson.M{
		"_id": field.PcRun,
	}).Decode(&out)
	return out, err
}

func (field PcRunField) asOptional() PcRunOptionalField {
	return PcRunOptionalField{&field.PcRun}
}

type PcRunOptionalField struct {
	PcRun *AlternateCollectionId `bson:"pcRun,omitempty" json:"pcRun,omitempty"`
}

func (field PcRunOptionalField) Get(ctx context.Context) (out PCRun, err error) {
	if field.PcRun == nil {
		err = ErrMissingOptionalField
		return
	}
	return PcRunField{*field.PcRun}.Get(ctx)
}
