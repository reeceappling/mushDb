package api

// needed directly for:
// AgarBatch, Bag, Jar, LC, Plugs, Slant, StasisTube, WaterJar

// needed for but provided from another:
// Plate, MSS,

// TODO: ? includeStasisTubes, includeAgarBatches, includePlugsJar
// TODO: newAgarBatch (on agar batch page)

// TODO: fully test this functionality first!!!!

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/reeceappling/mushDb/api/request"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
)

type PCRun struct {
	AlternateCollectionIdField `bson:"inline"`
	CreationDateField          `bson:"inline"`
	RunTimeMinutes             int `bson:"runtimeMinutes" json:"runtimeMinutes"`
	NotesField                 `bson:"inline"`
	LastUpdatedField           `bson:"inline"`
	AclField                   `bson:"inline"`
}

var impPcRun = exAltId

func initializePCRun(ctx context.Context) error {
	// Indices
	coll := DbFrom(ctx).Collection(PcRunCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		creationDateIndexModel,
		// TODO: newSimpleIndex("runtimeMinutes","runtimeMinutes", true, false, false),
		//Notes (no index unless tags)
		projectsIndexModel,
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test run does not exist, then create it
	testItem := &PCRun{ // TODO: this is a
		AlternateCollectionIdField: impPcRun.asIdField(),
		CreationDateField:          CreationDateField{exampleTime},
		RunTimeMinutes:             60,
		NotesField: NotesField{[]Note{{
			RequiredTimeField: RequiredTimeField{exampleTime},
			Note:              "PC Run specified for Imports, also utilized for testing and development",
		}}},
		LastUpdatedField: LastUpdatedField{exampleTime},
		AclField:         allCanReadAcl(nil),
	}
	return addTestAltEntries(ctx, testItem)
}

type createPcRunRequest struct {
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
	ctx := r.Context()
	ctx, now := request.UnixTime(r.Context()) // TODO: no more r.Context below
	toInsert := PCRun{
		AlternateCollectionIdField: AlternateCollectionIdField{id},
		CreationDateField:          CreationDateField{now},
		RunTimeMinutes:             req.RunTimeMinutes,
		NotesField:                 req.NotesField,
		LastUpdatedField:           LastUpdatedField{now},
		AclField:                   allCanReadAcl(GetUserEmailPtr(ctx)),
	}
	finishCreateAlternateEntry(ctx, toInsert, w)
}

type updatePcRunRequest struct {
	NotesUpdateField
	PermsOnRequest `json:"acl"`
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
	err = DbFrom(ctx).Collection(PcRunCollectionName).FindOne(ctx, bson.M{
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
type pcRunOptional interface {
	pcRunId() *AlternateCollectionId
	HasPcRun() error
}

func (field PcRunOptionalField) pcRunId() *AlternateCollectionId {
	return field.PcRun
}
func (field PcRunOptionalField) HasPcRun() error {
	if field.PcRun == nil {
		return errors.New("must have pc run")
	}
	return nil
}

func (field PcRunOptionalField) Get(ctx context.Context) (out PCRun, err error) {
	if field.PcRun == nil {
		err = ErrMissingOptionalField
		return
	}
	return PcRunField{*field.PcRun}.Get(ctx)
}
