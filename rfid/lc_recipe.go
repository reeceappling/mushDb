package rfid

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	sliceutils "slices"
	"time"
)

const lcRecipesCollectionName = "lcRecipes"

// TODO: AgarRecipe, JarRecipe, LcRecipe now have additiveMeasurements. Account for those everywhere
// TODO: ensure standard LC recipes are accessible

type LCRecipe struct { // TODO: add most recently updated field, add creation date field
	Id          alternateCollectionId `bson:"_id" json:"_id"`
	Name        string                `bson:"name" json:"name"`       // TODO: ENSURE THIS IS PROPERLY SET EVERYWHERE AND INDEXED
	Liquids     []liquid              `bson:"liquids" json:"liquids"` // TapWater, DistilledWater, GrainWater (Oat, etc)
	Nutrients   []nutrientMeasurement `bson:"nutrients" json:"nutrients"`
	Standard    bool                  `bson:"standard" json:"standard"` // If this is a standard recipe // TODO: account for this
	Sugars      []sugarMeasurement    `bson:"sugars,omitempty" json:"sugars,omitempty"`
	Additives   []additiveMeasurement `bson:"additives,omitempty" json:"additives,omitempty"` // TODO: ACCOUNT FOR THIS EVERYWHERE!
	Notes       []Note                `bson:"notes,omitempty" json:"notes,omitempty"`
	LastUpdated unixTime              `bson:"lastUpdated" json:"lastUpdated"`
}

func (recipe LCRecipe) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := recipe
	err := decodeItem(&out, encoded)
	return out, err
}

func (recipe LCRecipe) clean() CollectionItem {
	return recipe
}

func (recipe LCRecipe) EntryTypeField() *string {
	return nil
}

func (recipe LCRecipe) CollectionName() string {
	return lcRecipesCollectionName
}

func initializeLcRecipes(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(lcRecipesCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		newSimpleIndex("name", "name", false, false, false),
		newSimpleIndex("liquids", "liquids.name", false, false, false),
		newSimpleIndex("nutrients", "nutrients.nutrient", false, false, false),
		newSimpleIndex("sugars", "sugars.type", false, false, false),
		newSimpleIndex("additives", "additives.additive", false, false, false),
		newSimpleIndex("standard", "standard", true, false, false),
		//Notes (no index for now unless tags)
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	inserted, updated := 0, 0
	allWater := []liquid{Water.AsLiquid()}
	allLME := []nutrientMeasurement{{
		Nutrient: LME,
		Amount:   0.667,
		Unit:     "g/pt",
	}}
	for _, recipe := range []LCRecipe{
		// LME LC - Light Malt Extract LC
		{
			Id:        alternateCollectionId(altCollIdForint(idMeaLC)), // TODO: this
			Liquids:   allWater,
			Nutrients: allLME,
			Sugars:    nil,
			Additives: nil,
			Standard:  true,
			Notes: []Note{
				builtInNote("0.667g nutes per pint jar"),
			},
		},
		// Sugary LME LC
		{
			Id:        alternateCollectionId(altCollIdForint(idMeaSugLC)), // TODO: this
			Liquids:   allWater,
			Nutrients: allLME,
			Sugars: []sugarMeasurement{{
				Type:   Honey,
				Amount: 2.0,
				Unit:   "drops/pt",
			}},
			Additives: nil,
			Standard:  true,
			Notes:     []Note{},
		},
	} {
		var existing LCRecipe
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
	if inserted+updated > 0 {
		println(fmt.Sprintf(`LC recipes: inserted %d, updated %d`, inserted, updated))
	}
	return nil
}

type createLcRecipeRequest struct {
	Name      string `bson:"name" json:"name"`
	Standard  bool   `bson:"standard" json:"standard"` // If this is a standard recipe
	Liquids   []liquid
	Nutrients []nutrientMeasurement `bson:"nutrients,omitempty" json:"nutrients,omitempty"` // Per grain jar
	Sugars    []sugarMeasurement    `bson:"sugars,omitempty" json:"sugars,omitempty"`       // Per grain jar
	Additives []additiveMeasurement `bson:"additives,omitempty" json:"additives,omitempty"`
	Notes     []Note                `bson:"notes,omitempty" json:"notes,omitempty"`
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
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(lcRecipesCollectionName)
		res, err := coll.InsertOne(r.Context(), LCRecipe{
			Id:          alternateCollectionId(id),
			Name:        req.Name,
			Standard:    req.Standard,
			Liquids:     req.Liquids,
			Nutrients:   req.Nutrients,
			Sugars:      req.Sugars,
			Additives:   req.Additives,
			Notes:       req.Notes,
			LastUpdated: unixTime(time.Now().UnixMilli()),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, nil
		}
		altId, ok := res.InsertedID.(alternateCollectionId)
		if !ok {
			http.Error(w, "bad id out, should never happen", http.StatusInternalServerError)
			return nil, nil
		}
		return w.Write(altId.base58Bytes())
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}

type updateLcRecipeRequest struct {
	Name     string           `json:"name"`
	Standard bool             `json:"standard"`
	Notes    AllEntries[Note] `json:"notes"`
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
	existing, err := GetAltCollectionItem(r.Context(), id.String(), LCRecipe{})
	if err != nil {
		stat := http.StatusInternalServerError
		if err == mongo.ErrNoDocuments {
			stat = http.StatusNotFound
		}
		http.Error(w, err.Error(), stat)
		return
	}
	if !ableToModify(r.Context()) { // TODO: DO THIS EVERYWHERE!
		http.Error(w, "not permitted to modify", http.StatusForbidden)
		return
	}
	mods := bson.D{}
	// change name if needed
	if req.Name != existing.Name {
		mods = bson.D{{"$set", bson.D{{"name", req.Name}}}}
	}
	// change standard if needed
	if req.Standard != existing.Standard {
		mods = bson.D{{"$set", bson.D{{"standard", req.Standard}}}}
	}
	// Do note changes
	mods, err = WithNotesUpdate(bson.D{}, req.Notes, existing.Notes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	if len(mods) == 0 {
		http.Error(w, "no changes made", http.StatusBadRequest)
		return
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(lcRecipesCollectionName)
		result := coll.FindOneAndUpdate(ctx, bson.D{{"_id", existing.Id}}, mods)
		err := result.Err()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, nil
		}
		return w.Write([]byte(b58Id))
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}
