package api

import (
	"context"
	"encoding/json"
	"github.com/reeceappling/goUtils/v2/utils"
	"github.com/reeceappling/mushDb/api/request"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
)

// needed for grain jars

// TODO: mixed-grain batches are covered through jarRecipe

type GrainBatch struct {
	AlternateCollectionIdField `bson:"inline"`
	SoakTimeHours              *int            `bson:"soakTimeHrs,omitempty" json:"soakTimeHrs,omitempty"`
	BoilTimeMins               *int            `bson:"boilTimeMins,omitempty" json:"boilTimeMins,omitempty"`
	DryTimeHours               *int            `bson:"dryTimeHours,omitempty" json:"dryTimeHours,omitempty"`
	CreationDateField          `bson:"inline"` // Date of first hydration
	JarRecipeRequiredField     `bson:"inline"`
	NotesField                 `bson:"inline"`
	LastUpdatedField           `bson:"inline"`
	AclField                   `bson:"inline"`
}

type GrainBatchField struct {
	GrainBatch AlternateCollectionId `bson:"grainBatch" json:"grainBatch"`
}

func (field GrainBatchField) Get(ctx context.Context) (out GrainBatch, err error) {
	var result GrainBatch
	err = DbFrom(ctx).Collection(GrainBatchCollectionName).FindOne(ctx, bson.M{
		"_id": field.GrainBatch,
	}).Decode(&result)
	return result, err
}

func (field GrainBatchField) asOptional() GrainBatchOptionalField {
	return GrainBatchOptionalField{&field.GrainBatch}
}

type GrainBatchOptionalField struct {
	GrainBatch *AlternateCollectionId `bson:"grainBatch,omitempty" json:"grainBatch,omitempty"`
}

func initializeGrainBatches(ctx context.Context) error {
	// Indices
	coll := DbFrom(ctx).Collection(GrainBatchCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		//newSimpleIndex("wetness", "wetness", false, true, false),
		newSimpleIndex("creationDate", "creationDate", true, false, false),
		newSimpleIndex("recipe", "recipe", false, false, false),
		//Notes (no index unless tags)
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}

	testItem := GrainBatch{
		AlternateCollectionIdField: AlternateCollectionIdField{exAltId},
		SoakTimeHours:              utils.Pointer(8),
		BoilTimeMins:               utils.Pointer(30),
		DryTimeHours:               utils.Pointer(4),
		CreationDateField:          CreationDateField{},
		JarRecipeRequiredField:     JarRecipeRequiredField{Recipe: exAltId},
		NotesField:                 NotesField{exampleNotes()},
		LastUpdatedField:           LastUpdatedField{exampleTime},
	}
	err = addTestAltEntries(ctx, testItem)
	println("test Grain Batch:", exAltId.AsBase58())
	return err
}

type createGrainBatchRequest struct {
	JarRecipeRequiredField
	NotesField
	PermsOnRequest `json:"acl"` // Nil means allCanWrite
}

// TODO: separate endpoints for updating soak, boil, and dry times
func createGrainBatchHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req := createGrainBatchRequest{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err = req.Get(r.Context())
	if err != nil {
		http.Error(w, "invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}
	id := newAlternateCollectionId()
	ctx := r.Context()
	// Validate fields
	_, err = req.JarRecipeRequiredField.Get(ctx)
	if err != nil {
		dbErr(w, "Jar Recipe validation failure: "+err.Error(), http.StatusBadRequest)
		return
	}
	user, _ := GetAuthInfo(ctx)
	acl, err := req.AclForUser(ctx, user) // TODO: is this ok? or do we want allCanRead?
	if err != nil {
		dbErr(w, "ACL creation failure: "+err.Error(), http.StatusBadRequest)
		return
	}
	// create new batch
	ctx, now := request.UnixTime(r.Context()) // TODO: no more r.Context below
	toInsert := &GrainBatch{
		AlternateCollectionIdField: AlternateCollectionIdField{id},
		CreationDateField:          CreationDateField{now},
		JarRecipeRequiredField:     req.JarRecipeRequiredField,
		NotesField:                 req.NotesField,
		LastUpdatedField:           LastUpdatedField{now},
		AclField:                   acl, // TODO: allCanWrite???
	}
	finishCreateAlternateEntry(ctx, toInsert, w)
}

type updateGrainBatchRequest struct {
	SoakTimeHours *int `bson:"soakTimeHrs,omitempty" json:"soakTimeHrs,omitempty"`
	BoilTimeMins  *int `bson:"boilTimeMins,omitempty" json:"boilTimeMins,omitempty"`
	DryTimeHours  *int `bson:"dryTimeHrs,omitempty" json:"dryTimeHours,omitempty"`
	NotesUpdateField
	PermsOnRequest `json:"acl"`
}

func (req updateGrainBatchRequest) modsFor(existing *GrainBatch, acl AclField) (bson.D, error) {
	return NewMods().
		updateTimeIfNoLongerNil("soakTimeHours", req.SoakTimeHours, existing.SoakTimeHours).
		updateTimeIfNoLongerNil("boilTimeMins", req.BoilTimeMins, existing.BoilTimeMins).
		updateTimeIfNoLongerNil("dryTimeHours", req.DryTimeHours, existing.DryTimeHours).
		updateNotesIfNeeded(req, existing).
		updatePermsIfNeeded(acl.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateGrainBatchHandler(w http.ResponseWriter, r *http.Request) {
	b58Id := Base58Str(r.PathValue("id"))
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req := updateGrainBatchRequest{}
	err = json.Unmarshal(bs, &req)
	if err != nil {
		http.Error(w, "failed to unmarshal body: "+err.Error(), http.StatusBadRequest)
		return
	}
	//PrettyPrintJson("new notes!", req.Notes.New) // TODO: del
	id, err := b58Id.toAltCollectionId()
	if err != nil {
		http.Error(w, "Invalid id! "+err.Error(), http.StatusBadRequest)
		return
	}
	ctx, db := Db(r)
	coll := db.Collection(GrainBatchCollectionName)
	existing, err := GetAltCollectionItem(ctx, id, &GrainBatch{})
	if err != nil {
		stat := http.StatusInternalServerError
		if err == mongo.ErrNoDocuments {
			stat = http.StatusNotFound
		}
		dbErr(w, err.Error(), stat)
		return
	}
	finishAltCollItemUpdate(ctx, w, coll, req.modsFor, existing, PermsOnRequest{}) // TODO: fix perms!
}
