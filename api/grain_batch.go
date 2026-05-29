package api

import (
	"context"
	"encoding/json"
	"github.com/reeceappling/goUtils/v2/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
)

// needed for grain jars

// TODO: what about mixed-grain batches???? (covered through jarRecipe)

// const grainBatchesCollectionName = "grainBatches"
type GrainBatch struct { // TODO: use this
	AlternateCollectionIdField
	SoakTimeHours     *int `bson:"soakTimeHrs,omitempty" json:"soakTimeHrs,omitempty"`
	BoilTimeMins      *int `bson:"boilTimeMins,omitempty" json:"boilTimeMins,omitempty"`
	DryTimeHours      *int `bson:"dryTimeHours,omitempty" json:"dryTimeHours,omitempty"`
	CreationDateField      // Date of first hydration
	JarRecipeRequiredField
	NotesField
	LastUpdatedField
	AclField
}

func initializeGrainBatches(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(GrainBatchCollectionName)
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
}

// TODO: separate endpoints for updating soak, boil, and dry times

func createGrainBatchHandler(w http.ResponseWriter, r *http.Request) { // TODO: DO THIS!
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
	ctx, db := Db(r)
	coll := db.Collection(GrainBatchCollectionName)
	// Validate fields
	_, err = req.JarRecipeRequiredField.Get(ctx)

	if err != nil {
		dbErr(w, "Jar Recipe validation failure: "+err.Error(), http.StatusBadRequest)
		return
	}
	// create new batch
	now := unixTimeForNow()
	toInsert := &GrainBatch{
		AlternateCollectionIdField: AlternateCollectionIdField{id},
		CreationDateField:          CreationDateField{now},
		JarRecipeRequiredField:     req.JarRecipeRequiredField,
		NotesField:                 req.NotesField,
		LastUpdatedField:           LastUpdatedField{now},
		AclField:                   allCanWriteAcl(),
	}
	finishCreateAlternateEntry(ctx, coll, toInsert, w)
}

type updateGrainBatchRequest struct {
	SoakTimeHours *int `bson:"soakTimeHrs,omitempty" json:"soakTimeHrs,omitempty"`
	BoilTimeMins  *int `bson:"boilTimeMins,omitempty" json:"boilTimeMins,omitempty"`
	DryTimeHours  *int `bson:"dryTimeHours,omitempty" json:"dryTimeHours,omitempty"`
	NotesUpdateField
	PermsOnRequest
}

func (req updateGrainBatchRequest) modsFor(existing *GrainBatch, acl AclField) (bson.D, error) {
	mods := NewMods()
	mods = updatePointerIfNeeded(mods, "soakTimeHours", req.SoakTimeHours, existing.SoakTimeHours)
	mods = updatePointerIfNeeded(mods, "boilTimeMins", req.BoilTimeMins, existing.BoilTimeMins)
	return updatePointerIfNeeded(mods, "dryTimeHours", req.DryTimeHours, existing.DryTimeHours).
		updateNotesIfNeeded(req, existing).
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
	PrettyPrintJson("new notes!", req.Notes.New) // TODO: del
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
	finishAltCollItemUpdate(ctx, w, coll, req.modsFor, existing, req.PermsOnRequest)
}
