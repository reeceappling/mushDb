package rfid

import (
	"context"
	"encoding/json"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
)

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
	err = ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(LcRecipesCollectionName).FindOne(ctx, bson.M{
		"_id": field.Recipe,
	}).Decode(&out)
	return out, err
}

func (recipe LcRecipe) EntryTypeField() *string {
	return nil
}

func initializeLcRecipes(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(LcRecipesCollectionName)
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
	allWater := LiquidsField{[]liquid{Water.AsLiquid()}}
	allLME := NutrientsField{[]nutrientMeasurement{{
		Nutrient: LME,
		Amount:   0.667,
		Unit:     "g/pt",
	}}}
	basicEntries := []*LcRecipe{
		// LME LC - Light Malt Extract LC
		{
			AlternateCollectionIdField: AlternateCollectionIdField{altCollIdForint(idMeaLC)},
			LiquidsField:               allWater,
			NutrientsField:             allLME,
			SugarsField:                SugarsField{},
			AdditivesField:             AdditivesField{},
			StandardField:              StandardField{true},
			NotesField: NotesField{[]Note{
				builtInNote("0.667g nutes per pint jar"),
			}},
			AclField: allCanReadAcl(),
		},
		// Sugary LME LC
		{
			AlternateCollectionIdField: AlternateCollectionIdField{altCollIdForint(idMeaSugLC)},
			LiquidsField:               allWater,
			NutrientsField:             allLME,
			SugarsField: SugarsField{[]sugarMeasurement{{
				Type:   Honey,
				Amount: 2.0,
				Unit:   "drops/pt",
			}}},
			AdditivesField: AdditivesField{},
			StandardField:  StandardField{true},
			NotesField:     NotesField{[]Note{}},
			AclField:       allCanReadAcl(),
		},
	}
	err = addBasicAltEntries(ctx, basicEntries...)
	if err != nil {
		return err
	}
	// Add test entry
	testItem := &LcRecipe{
		AlternateCollectionIdField: AlternateCollectionIdField{exAltId},
		NameField:                  NameField{"testJarRecipeName"},
		StandardField:              StandardField{false},
		NutrientsField: NutrientsField{[]nutrientMeasurement{
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
		SugarsField: SugarsField{[]sugarMeasurement{
			newSugarMeasurement(Honey, 1, "large drop per quart jar"),
		}},
		AdditivesField: AdditivesField{[]additiveMeasurement{
			newAdditiveMeasurement(Vermiculite, 0.25, "tsp"),
			newAdditiveMeasurement(Gypsum, 0.7, "coverage of jar bottom"),
		}},
		NotesField:       NotesField{exampleNotes()},
		LastUpdatedField: LastUpdatedField{exampleTime},
	}
	// TODO: add built-in entries
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
	ctx, db := Db(r)
	coll := db.Collection(LcRecipesCollectionName)
	toInsert := LcRecipe{
		AlternateCollectionIdField: AlternateCollectionIdField{id},
		NameField:                  req.NameField,
		StandardField:              req.StandardField,
		LiquidsField:               req.LiquidsField,
		NutrientsField:             req.NutrientsField,
		SugarsField:                req.SugarsField,
		AdditivesField:             req.AdditivesField,
		NotesField:                 req.NotesField,
		LastUpdatedField:           LastUpdatedField{unixTimeForNow()},
		AclField:                   allCanWriteAcl(),
	}
	finishCreateAlternateEntry(ctx, coll, toInsert, w)
}

type updateLcRecipeRequest struct {
	NameField
	StandardField
	Notes AllEntries[Note] `json:"notes"`
	PermsOnRequest
}

func (req updateLcRecipeRequest) modsFor(existing *LcRecipe, aclField AclField) (bson.D, error) {
	return NewMods().
		updateNameIfNeeded(req.Name, existing.Name).
		updateStandardIfNeeded(req.Standard, existing.Standard).
		updateNotesIfNeeded(req.Notes, existing.Notes).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateLcRecipeHandler(w http.ResponseWriter, r *http.Request) {
	b58Id := Base58Str(r.PathValue("id"))
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req := updateLcRecipeRequest{}
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
