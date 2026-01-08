package rfid

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"reflect"
	sliceutils "slices"
)

const lcRecipesCollectionName = "lcRecipes"

// TODO: AgarRecipe, JarRecipe, LcRecipe now have additiveMeasurements. Account for those everywhere
// TODO: ensure standard LC recipes are accessible

type LcRecipe struct {
	AlternateCollectionIdField
	NameField    // TODO: ENSURE THIS IS PROPERLY SET EVERYWHERE AND INDEXED
	LiquidsField // TapWater, DistilledWater, GrainWater (Oat,
	NutrientsField
	StandardField // Whether or not this is a standard recipe
	SugarsField
	AdditivesField
	NotesField
	LastUpdatedField
}

type LcRecipeField struct {
	Recipe AlternateCollectionId `bson:"recipe" json:"recipe"`
}

func (field LcRecipeField) Get(ctx context.Context) (out LcRecipe, err error) {
	err = ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(lcRecipesCollectionName).FindOne(ctx, bson.M{
		"_id": field.Recipe,
	}).Decode(&out)
	return out, err
}

func (recipe LcRecipe) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := recipe
	err := decodeItem(&out, encoded)
	return out, err
}

func (recipe LcRecipe) EntryTypeField() *string {
	return nil
}

func (recipe LcRecipe) CollectionName() string {
	return lcRecipesCollectionName
}

func initializeLcRecipes(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(lcRecipesCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		newSimpleIndex("name", "name", false, false, false),
		newSimpleIndex("liquids", "liquids.name", false, false, false),
		newSimpleIndex("nutrients", "nutrients.nutrient", false, false, false),
		standardIndexModel,
		newSimpleIndex("sugars", "sugars.type", false, false, false),
		newSimpleIndex("additives", "additives.additive", false, false, false),
		//Notes (no index for now unless tags)
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	inserted, updated := 0, 0
	allWater := LiquidsField{[]liquid{Water.AsLiquid()}}
	allLME := NutrientsField{[]nutrientMeasurement{{
		Nutrient: LME,
		Amount:   0.667,
		Unit:     "g/pt",
	}}}
	for _, recipe := range []LcRecipe{
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
		},
	} {
		var existing LcRecipe
		err := coll.FindOne(ctx, bson.D{{"_id", recipe.Id}}).Decode(&existing)
		if err != nil {
			if err != mongo.ErrNoDocuments {
				return err
			}
			// if not exists, add it to the db
			_, err = coll.InsertOne(ctx, recipe)
			if err != nil {
				return err
			}
			inserted++
			continue
		}
		// If exists, ensure it is the same as it was. Add notes if necessary
		update := false
		if len(recipe.Liquids) != len(existing.Liquids) {
			update = true
		} else {
			// if any liquids are different, replace all liquids
			for i, liq := range recipe.Liquids {
				if liq != existing.Liquids[i] {
					update = true
					break
				}
			}
		}
		// Nutrients
		if !update {
			if len(recipe.Nutrients) != len(existing.Nutrients) {
				update = true
			} else {
				// if any nutrients are different, replace
				for i, nut := range recipe.Nutrients {
					if nut != existing.Nutrients[i] {
						update = true
						break
					}
				}
			}
		}

		// Sugars
		if !update {
			if len(recipe.Sugars) != len(existing.Sugars) {
				update = true
			} else {
				// if any sugars are different, replace
				for i, s := range recipe.Sugars {
					if s != existing.Sugars[i] {
						update = true
						break
					}
				}
			}
		}

		// Notes
		finalNotes := []Note{}
		copy(finalNotes, existing.Notes)
		for _, note := range recipe.Notes {
			if !sliceutils.Contains(finalNotes, note) {
				finalNotes = append(finalNotes, note)
				update = true
			}
		}
		recipe.Notes = finalNotes

		// Update if necessary
		if update {
			err = coll.FindOneAndReplace(ctx, bson.D{{"_id", recipe.Id}}, recipe).Err()
			if err != nil {
				return err
			}
			updated++
		}
	}
	// Add test entry
	existingEntry := LcRecipe{}
	testItem := LcRecipe{
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
	err = coll.FindOne(ctx, bson.D{{"_id", exAltId}}).Decode(&existingEntry)
	if err == nil {
		if reflect.DeepEqual(existingEntry, testItem) {
			return nil
		}
	}
	err = testExistingEntry(ctx, coll, exAltId, testItem, existingEntry)
	if inserted+updated > 0 {
		println(fmt.Sprintf(`LC recipes: inserted %d, updated %d`, inserted, updated))
	}
	return err
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
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(lcRecipesCollectionName)
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
		}
		_, err = coll.InsertOne(r.Context(), toInsert)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		bs, err := json.Marshal(toInsert)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		return w.Write(bs)
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}

type updateLcRecipeRequest struct {
	NameField
	StandardField
	Notes AllEntries[Note] `json:"notes"`
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
	existing, err := GetAltCollectionItem(r.Context(), id, LcRecipe{})
	if err != nil {
		stat := http.StatusInternalServerError
		if err == mongo.ErrNoDocuments {
			stat = http.StatusNotFound
		}
		http.Error(w, err.Error(), stat)
		return
	}
	upd, err := NewMods().
		updateNameIfNeeded(req.Name, existing.Name).
		updateStandardIfNeeded(req.Standard, existing.Standard).
		updateNotesIfNeeded(req.Notes, existing.Notes).
		updateLastUpdatedIfNeeded().
		Finalized()
	if err != nil {
		http.Error(w, "error creating txn:"+err.Error(), http.StatusInternalServerError)
		return
	}
	if len(upd) == 0 {
		http.Error(w, "no changes made", http.StatusBadRequest)
		return
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		// TODO: turn everything in this txn into its own func????
		coll := ctx.Client().Database(dbName).Collection(lcRecipesCollectionName)
		bsonId := bson.D{{"_id", existing.Id}}
		err = coll.FindOneAndUpdate(ctx, bsonId, upd).Err()
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		err = coll.FindOne(ctx, bsonId).Decode(&existing)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		bs, err = json.Marshal(existing)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		return w.Write([]byte(b58Id))
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}
