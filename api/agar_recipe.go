package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/reeceappling/mushDb/api/env"
	"github.com/reeceappling/mushDb/api/request"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"time"
)

// required for: agarBatch

func TimeFromId(id AlternateCollectionId) time.Time { // TODO: USE AND MOVE
	return primitive.ObjectID(id).Timestamp()
}

type AgarRecipe struct {
	AlternateCollectionIdField `bson:"inline"` // CreationDate is embedded?
	NameField                  `bson:"inline"`
	LiquidsField               `bson:"inline"`
	Agar                       int             `bson:"agar" json:"agar"` // agar grams per 1L
	StandardField              `bson:"inline"` // If this is a standard recipe
	NutrientsField             `bson:"inline"`
	SugarsField                `bson:"inline"`
	AdditivesField             `bson:"inline"`
	AntibioticsField           `bson:"inline"`
	NotesField                 `bson:"inline"`
	LastUpdatedField           `bson:"inline"`
	AclField                   `bson:"inline"`
}

type updateAgarRecipeRequest struct {
	NameField
	StandardField
	NotesUpdateField
	PermsOnRequest `json:"acl"`
}

func (req updateAgarRecipeRequest) modsFor(existing *AgarRecipe, aclField AclField) (bson.D, error) {
	return NewMods().
		updateNameIfNeeded(req.Name, existing.Name).
		updateStandardIfNeeded(req.Standard, existing.Standard).
		updateNotesIfNeeded(req, existing).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateAgarRecipeHandler(w http.ResponseWriter, r *http.Request) {
	b58Id := Base58Str(r.PathValue("id"))
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req := updateAgarRecipeRequest{}
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
	coll := db.Collection(AgarRecipesCollectionName)

	existing, err := GetAltCollectionItem(ctx, id, &AgarRecipe{})
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

const LmeaName = "LMEA"
const PdaName = "PDA"
const WaterAgarName = "Water Agar"
const GrainWaterAgarName = "Grainwater Agar"
const AntibioticAgarName = "Antibiotic Agar"

func initializeAgarRecipes(ctx context.Context) error {
	db := DbFrom(ctx)
	coll := db.Collection(AgarRecipesCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		newSimpleIndex("name", "name", false, false, true), // TODO: unique (last) may need to be true (do we want names to be unique or not?)
		//newSimpleIndex("liquids", "liquids.name", false, false, false),
		//newSimpleIndex("agar", "agar", true, false, false),
		standardIndexModel,
		//newSimpleIndex("nutrients", "nutrients.nutrient", false, false, false),
		//newSimpleIndex("sugars", "sugars.type", false, false, false),
		//newSimpleIndex("additives", "additives.additive", false, false, false),
		//newSimpleIndex("antibiotics", "antibiotics", false, false, false),
		//Notes (not indexed)
		projectsIndexModel,
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// Add built-in entries
	builtinTime := RequiredTimeField{Time: ogTime}
	basicEntryAcl := allCanReadAcl(nil)
	basicEntries := []*AgarRecipe{
		{
			AlternateCollectionIdField: AlternateCollectionIdField{altCollIdForint(idLmea)},
			NameField:                  NameField{"LMEA"},
			LiquidsField:               LiquidsField{[]Liquid{Water.AsLiquid()}},
			Agar:                       20,
			NutrientsField: NutrientsField{[]NutrientMeasurement{
				nutMmt(LME, 20, "g"),
			}},
			SugarsField:   SugarsField{},
			StandardField: StandardField{true},
			NotesField: NotesField{Notes: []Note{
				{Note: "Light Malt Extract Agar", RequiredTimeField: builtinTime},
			}},
			AclField: basicEntryAcl,
		},
		{
			AlternateCollectionIdField: AlternateCollectionIdField{altCollIdForint(idPda)},
			NameField:                  NameField{"PDA"},
			LiquidsField:               LiquidsField{[]Liquid{Water.AsLiquid()}},
			Agar:                       20,
			NutrientsField: NutrientsField{[]NutrientMeasurement{
				nutMmt(Potato, 18, "g"),
			}},
			SugarsField: SugarsField{[]SugarMeasurement{
				sugMmt(Dextrose, 1, "g"),
			}},
			StandardField: StandardField{true},
			NotesField: NotesField{Notes: []Note{
				{Note: "Potato Dextrose Agar", RequiredTimeField: builtinTime},
			}},
			AclField: basicEntryAcl,
		},
		{
			AlternateCollectionIdField: AlternateCollectionIdField{altCollIdForint(idWaterAgar)},
			NameField:                  NameField{"Water Agar"},
			LiquidsField:               LiquidsField{[]Liquid{Water.AsLiquid()}},
			Agar:                       20,
			NutrientsField:             NutrientsField{},
			SugarsField:                SugarsField{},
			StandardField:              StandardField{true},
			NotesField:                 NotesField{},
			AclField:                   basicEntryAcl,
		},
		{
			AlternateCollectionIdField: AlternateCollectionIdField{altCollIdForint(idGrainWaterAgar)},
			NameField:                  NameField{GrainWaterAgarName},
			LiquidsField: LiquidsField{[]Liquid{
				GrainWater.AsLiquid().withPct(50.0),
				DistilledWater.AsLiquid().withPct(50.0),
			}},
			Agar:           20,
			NutrientsField: NutrientsField{},
			SugarsField:    SugarsField{},
			StandardField:  StandardField{true},
			NotesField: NotesField{[]Note{
				builtInNote("Grain water also acts as a nutrient source"),
			}},
			AclField: basicEntryAcl,
		},
		{
			AlternateCollectionIdField: AlternateCollectionIdField{altCollIdForint(idAntibioticAgar)},
			NameField:                  NameField{"Antibiotic Built-In"},
			LiquidsField:               LiquidsField{[]Liquid{DistilledWater.AsLiquid()}},
			Agar:                       20,
			NutrientsField: NutrientsField{[]NutrientMeasurement{
				nutMmt(LME, 20, "g"),
			}},
			SugarsField:      SugarsField{},
			AntibioticsField: AntibioticsField{[]Antibiotic{Doxycycline}},
			StandardField:    StandardField{true},
			NotesField: NotesField{[]Note{
				builtInNote("50mg doxycycline per ?????"),
			}},
			AclField: basicEntryAcl,
		},
	}

	err = addBasicAltEntries(ctx, basicEntries...)
	if err != nil {
		return err
	}
	// Add test entries if dev
	return env.IfNotProd(ctx, func() error {
		testItem := &AgarRecipe{
			AlternateCollectionIdField: AlternateCollectionIdField{exAltId},
			NameField:                  NameField{testEntryStringId},
			LiquidsField: LiquidsField{[]Liquid{
				Water.AsLiquid().withPct(40.0),
				DistilledWater.AsLiquid().withPct(60.0),
			}},
			Agar:          20,
			StandardField: StandardField{false},
			NutrientsField: NutrientsField{[]NutrientMeasurement{
				nutMmt(LME, 19, "g"),
				nutMmt(Potato, 2, "g"),
			}},
			SugarsField: SugarsField{[]SugarMeasurement{
				sugMmt(Dextrose, 1, "g"),
				sugMmt(Honey, 2, "g"),
			}},
			AdditivesField: AdditivesField{[]AdditiveMeasurement{
				{
					Additive: Vermiculite,
					Amount:   0.2,
					Unit:     "lb",
				},
				{
					Additive: Perlite,
					Amount:   0.7,
					Unit:     "tons",
				},
				{
					Additive: Gypsum,
					Amount:   1,
					Unit:     "pinch",
				},
			}},
			AntibioticsField: AntibioticsField{[]Antibiotic{Doxycycline, HydrogenPeroxide}},
			NotesField:       NotesField{exampleNotes()},
			LastUpdatedField: LastUpdatedField{exampleTime},
			AclField:         AclField{testAcl},
		}

		// Add test entries
		return addTestAltEntries(ctx, testItem) // TODO: remove once done testing...
	})
}

func sugMmt(t Sugar, amt float64, unit string) SugarMeasurement {
	return SugarMeasurement{
		Type:   t,
		Amount: amt,
		Unit:   unit,
	}
}
func nutMmt(t Nutrient, amt float64, unit string) NutrientMeasurement {
	return NutrientMeasurement{
		Nutrient: t,
		Amount:   amt,
		Unit:     unit,
	}
}

type createAgarRecipeRequest struct {
	NameField
	StandardField
	Agar int `json:"agar"` // agar g/L
	LiquidsField
	NutrientsField
	SugarsField
	AdditivesField
	AntibioticsField // make sure to put dosages in notes
	NotesField
}

func createAgarRecipeHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	req := createAgarRecipeRequest{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	id := newAlternateCollectionId()
	ctx, now := request.UnixTime(r.Context())

	toInsert := AgarRecipe{
		AlternateCollectionIdField: AlternateCollectionIdField{id},
		NameField:                  req.NameField,
		LiquidsField:               req.LiquidsField,
		Agar:                       req.Agar,
		StandardField:              req.StandardField,
		NutrientsField:             req.NutrientsField,
		SugarsField:                req.SugarsField,
		AdditivesField:             req.AdditivesField,
		AntibioticsField:           req.AntibioticsField,
		NotesField:                 req.NotesField,
		LastUpdatedField:           LastUpdatedField{now},
		AclField:                   allCanReadAcl(GetUserEmailPtr(ctx)),
	}
	finishCreateAlternateEntry(ctx, toInsert, w)
}

// TODO: USE!
func getAgarRecipeByName(ctx context.Context, name string) (AgarRecipe, error) {
	out := AgarRecipe{}
	err := DbFrom(ctx).
		Collection(AgarRecipesCollectionName).
		FindOne(ctx, bson.M{"name": name}).
		Decode(&out)
	return out, err
}

type AgarRecipeField struct {
	AgarRecipe AlternateCollectionId `bson:"agarRecipe" json:"agarRecipe"`
}

func (field AgarRecipeField) Get(ctx context.Context) (out AgarRecipe, err error) {
	err = DbFrom(ctx).Collection(AgarRecipesCollectionName).FindOne(ctx, bson.M{
		IDfld: field.AgarRecipe,
	}).Decode(&out)
	return out, err
}

func deleteAgarRecipeHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id") // TODO: recipe by name?
	if idStr == "" {
		http.Error(w, "Empty id for delete request", http.StatusBadRequest)
		return
	}
	id, err := Base58Str(idStr).toAltCollectionId()
	if err != nil {
		http.Error(w, "Invalid ID to delete: "+err.Error(), http.StatusBadRequest)
		return
	}
	// Validate not used in other places...
	ctx := r.Context()
	db := DbFrom(ctx)
	// ensure recipe not used by any batches first
	err = db.Collection(AgarBatchCollectionName).FindOne(ctx, bson.M{"agarRecipe": id}).Err()
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			http.Error(w, "failed to check for agarRecipe usage in agarBatch collection. "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// At least one item exists, fail
		http.Error(w, "at least one agarBatch utilizes the item you are attempting to delete.", http.StatusConflict) // TODO: status ok?
		return
	}

	// Delete if not found elsewhere!
	deleteResult, err := db.Collection(AgarRecipesCollectionName).DeleteOne(ctx, bson.M{IDfld: id})
	if err != nil {
		http.Error(w, "failed to delete: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if deleteResult.DeletedCount == 0 {
		http.Error(w, "failed to delete id "+idStr+" from agar recipes. Id not found", http.StatusNotFound)
		return
	}
	_, err = w.Write([]byte(idStr))
	handleWriteErr(err, w)
}
