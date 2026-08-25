package api

import (
	"context"
	"encoding/json"
	"github.com/reeceappling/goUtils/v2/utils"
	"github.com/reeceappling/mushDb/api/env"
	"github.com/reeceappling/mushDb/api/request"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
)

// needed for grain jars

// mixed-grain batches are covered through jarRecipe
// TODO: allow creating grain water jars through this?

type GrainBatch struct {
	AlternateCollectionIdField `bson:"inline"`
	// TODO: what of different grains that have different timings? think over this.
	SoakTimeHours          *int            `bson:"soakTimeHrs,omitempty" json:"soakTimeHrs,omitempty"`
	BoilTimeMins           *int            `bson:"boilTimeMins,omitempty" json:"boilTimeMins,omitempty"`
	DryTimeHours           *int            `bson:"dryTimeHours,omitempty" json:"dryTimeHours,omitempty"`
	CreationDateField      `bson:"inline"` // Date of first hydration. also exists in the id
	JarRecipeRequiredField `bson:"inline"`
	NotesField             `bson:"inline"`
	LastUpdatedField       `bson:"inline"`
	AclField               `bson:"inline"`
}

//func (g *GrainBatch) Blank() CollectionItem {
//	return &GrainBatch{}
//}

type GrainBatchField struct {
	GrainBatch AlternateCollectionId `bson:"grainBatch" json:"grainBatch"`
}

func (field GrainBatchField) Get(ctx context.Context) (out GrainBatch, err error) {
	var result GrainBatch
	err = DbFrom(ctx).Collection(GrainBatchCollectionName).FindOne(ctx, bson.M{
		IDfld: field.GrainBatch,
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

	return env.IfNotProd(ctx, func() error {

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
		println("test Grain Batch:", exAltId.AsBase58())
		return addTestAltEntries(ctx, testItem)
	})
}

type createGrainBatchRequest struct {
	JarRecipeRequiredField
	WetnessField
	BurstGrainsField
	NotesField
	PermsOnRequest `json:"acl"` // Nil means allCanWrite
}

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

	acl := allCanReadAcl(GetUserEmailPtr(ctx))
	// create new batch
	ctx, now := request.UnixTime(r.Context())
	toInsert := &GrainBatch{
		AlternateCollectionIdField: AlternateCollectionIdField{id},
		CreationDateField:          CreationDateField{now},
		JarRecipeRequiredField:     req.JarRecipeRequiredField,
		NotesField:                 req.NotesField,
		LastUpdatedField:           LastUpdatedField{now},
		AclField:                   acl,
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
	req := updateGrainBatchRequest{}
	err = json.Unmarshal(bs, &req)
	if err != nil {
		http.Error(w, "failed to unmarshal body: "+err.Error(), http.StatusBadRequest)
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

//func deleteGrainBatchHandler(w http.ResponseWriter, r *http.Request) {
//	idStr := r.PathValue("id") // recipe by name?
//	if idStr == "" {
//		http.Error(w, "Empty id for delete request", http.StatusBadRequest)
//		return
//	}
//	id, err := Base58Str(idStr).toAltCollectionId()
//	if err != nil {
//		http.Error(w, "Invalid ID to delete: "+err.Error(), http.StatusBadRequest)
//		return
//	}
//	// Validate not used in other places...
//	ctx := r.Context()
//	db := DbFrom(ctx)
//	// ensure batch not used by any jars first
//	err = db.Collection(GrainJarCollectionName).FindOne(ctx, bson.M{"grainBatch": id}).Err()
//	if err != nil {
//		if !errors.Is(err, mongo.ErrNoDocuments) {
//			http.Error(w, "failed to check for agarRecipe usage in agarBatch collection. "+err.Error(), http.StatusInternalServerError)
//			return
//		}
//	} else {
//		// At least one item exists, fail
//		http.Error(w, "at least one grainJar utilizes the item you are attempting to delete.", http.StatusExpectationFailed)
//		return
//	}
//
//	DeleteCollectionItem(ctx, GrainBatchCollectionName, id, w)
//}
