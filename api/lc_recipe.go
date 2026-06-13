package api

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

// needed for lc, lcSyringe

type LcRecipe struct {
	AlternateCollectionIdField `bson:"inline"`
	NameField                  `bson:"inline"`
	LiquidsField               `bson:"inline"` // TapWater, DistilledWater, GrainWater (Oat,
	NutrientsField             `bson:"inline"`
	StandardField              `bson:"inline"` // Whether or not this is a standard recipe
	SugarsField                `bson:"inline"`
	AdditivesField             `bson:"inline"`
	NotesField                 `bson:"inline"`
	LastUpdatedField           `bson:"inline"`
	AclField                   `bson:"inline"`
}

type LcRecipeField struct {
	Recipe AlternateCollectionId `bson:"recipe" json:"recipe"`
}

func (field LcRecipeField) Get(ctx context.Context) (out LcRecipe, err error) {
	err = DbFrom(ctx).Collection(LcRecipesCollectionName).FindOne(ctx, bson.M{
		"_id": field.Recipe,
	}).Decode(&out)
	return out, err
}

func initializeLcRecipes(ctx context.Context) error {
	// Indices
	coll := DbFrom(ctx).Collection(LcRecipesCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		newSimpleIndex("name", "name", false, false, false),
		//newSimpleIndex("liquids", "liquids.name", false, false, false),
		//newSimpleIndex("nutrients", "nutrients.nutrient", false, false, false),
		standardIndexModel,
		//newSimpleIndex("sugars", "sugars.type", false, false, false),
		//newSimpleIndex("additives", "additives.additive", false, false, false),
		//Notes (no index for now unless tags)
		projectsIndexModel,
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	allWater := LiquidsField{[]Liquid{Water.AsLiquid()}}
	allLME := NutrientsField{[]NutrientMeasurement{{
		Nutrient: LME,
		Amount:   0.667,
		Unit:     "g/pt",
	}}}
	// Add builtin
	basicEntries := []*LcRecipe{
		{
			AlternateCollectionIdField: AlternateCollectionIdField{altCollIdForint(idMeaLC)},
			NameField:                  NameField{"Basic LME LC"}, // Light Malt Extract LC
			LiquidsField:               allWater,
			NutrientsField:             allLME,
			SugarsField:                SugarsField{},
			AdditivesField:             AdditivesField{},
			StandardField:              StandardField{true},
			NotesField: NotesField{[]Note{
				builtInNote("0.667g nutes per pint jar"),
			}},
			AclField: allCanReadAcl(nil),
		},
		{
			AlternateCollectionIdField: AlternateCollectionIdField{altCollIdForint(idMeaSugLC)},
			NameField:                  NameField{"Basic Sugary LME LC"}, // Sugary LME LC
			LiquidsField:               allWater,
			NutrientsField:             allLME,
			SugarsField: SugarsField{[]SugarMeasurement{{
				Type:   Honey,
				Amount: 2.0,
				Unit:   "drops/pt",
			}}},
			AdditivesField: AdditivesField{},
			StandardField:  StandardField{true},
			NotesField:     NotesField{[]Note{}},
			AclField:       allCanReadAcl(nil),
		},
	}
	err = addBasicAltEntries(ctx, basicEntries...)
	if err != nil {
		return err
	}
	// Add test entries
	testItem := &LcRecipe{
		AlternateCollectionIdField: AlternateCollectionIdField{exAltId},
		NameField:                  NameField{"testLcRecipeName"},
		StandardField:              StandardField{false},
		LiquidsField:               allWater,
		NutrientsField: NutrientsField{[]NutrientMeasurement{
			{
				Nutrient: LME,
				Amount:   1,
				Unit:     "kg",
			},
			{
				Nutrient: Potato,
				Amount:   8,
				Unit:     "ug",
			},
		}},
		SugarsField: SugarsField{[]SugarMeasurement{
			newSugarMeasurement(Honey, 1, "large drop per quart jar"),
		}},
		AdditivesField: AdditivesField{[]AdditiveMeasurement{
			newAdditiveMeasurement(Vermiculite, 0.25, "tsp"),
			newAdditiveMeasurement(Gypsum, 0.7, "coverage of jar bottom"),
		}},
		NotesField:       NotesField{exampleNotes()},
		LastUpdatedField: LastUpdatedField{exampleTime},
		AclField:         allCanWriteAcl(),
	}
	return addTestAltEntries(ctx, testItem)
}

type createLcRecipeRequest struct {
	NameField
	StandardField // If this is a standard recipe
	LiquidsField
	NutrientsField
	SugarsField
	AdditivesField
	NotesField
}

func createLcRecipeHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctx, now := request.UnixTime(r.Context())
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	req := createLcRecipeRequest{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	id := newAlternateCollectionId()
	if err = errors.Join(
		req.LiquidsField.Validate(),
		req.NutrientsField.Validate(),
		req.SugarsField.Validate(),
		req.AdditivesField.Validate(),
	); err != nil {
		http.Error(w, "invalid request: "+err.Error(), http.StatusBadRequest)
	}
	toInsert := LcRecipe{
		AlternateCollectionIdField: AlternateCollectionIdField{id},
		NameField:                  req.NameField,
		StandardField:              req.StandardField,
		LiquidsField:               req.LiquidsField,
		NutrientsField:             req.NutrientsField,
		SugarsField:                req.SugarsField,
		AdditivesField:             req.AdditivesField,
		NotesField:                 req.NotesField,
		LastUpdatedField:           LastUpdatedField{now},
		AclField:                   allCanWriteAcl(), // TODO: or allCanRead + user write
	}
	finishCreateAlternateEntry(ctx, toInsert, w)
}

type updateLcRecipeRequest struct {
	NameField
	StandardField
	NotesUpdateField
	PermsOnRequest `json:"acl"`
}

func (req updateLcRecipeRequest) modsFor(existing *LcRecipe, aclField AclField) (bson.D, error) {
	return NewMods().
		updateNameIfNeeded(req.Name, existing.Name).
		updateStandardIfNeeded(req.Standard, existing.Standard).
		updateNotesIfNeeded(req, existing).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateLcRecipeHandler(w http.ResponseWriter, r *http.Request) {
	req := updateLcRecipeRequest{}
	_, id, err := altCollIdFromRequest(r, w) // TODO: use this everywhere
	if err != nil {
		return
	}
	if err = ReadSimpleStructuredBody(r, w, &req); err != nil { // TODO: use this everywhere
		return // Writes already if err
	}

	ctx, db := Db(r)
	coll := db.Collection(LcRecipesCollectionName)
	existing, err := GetAltCollectionItem(r.Context(), id, &LcRecipe{})
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
