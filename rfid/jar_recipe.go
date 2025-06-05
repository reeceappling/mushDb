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

const jarRecipesCollectionName = "jarRecipes"

type JarRecipe struct {
	Id          alternateCollectionId `bson:"_id" json:"_id"`
	Name        string                `bson:"name" json:"name"`
	Grain       grain                 `bson:"grain" json:"grain"`
	Standard    bool                  `bson:"standard" json:"standard"`                       // If this is a standard recipe
	Nutrients   []nutrientMeasurement `bson:"nutrients,omitempty" json:"nutrients,omitempty"` // Per grain jar
	Sugars      []sugarMeasurement    `bson:"sugars,omitempty" json:"sugars,omitempty"`       // Per grain jar
	Additives   []additiveMeasurement `bson:"additives,omitempty" json:"additives,omitempty"`
	Notes       []Note                `bson:"notes,omitempty" json:"notes,omitempty"`
	LastUpdated unixTime              `bson:"lastUpdated" json:"lastUpdated"`
}

func (recipe JarRecipe) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := recipe
	err := decodeItem(&out, encoded)
	return out, err
}

func (recipe JarRecipe) clean() CollectionItem {
	return recipe
}

func (recipe JarRecipe) EntryTypeField() *string {
	return nil
}

func (recipe JarRecipe) CollectionName() string {
	return jarRecipesCollectionName
}

func initializeJarRecipes(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(jarRecipesCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		newSimpleIndex("name", "name", false, false, false),
		newSimpleIndex("grain", "grain", false, false, false),
		newSimpleIndex("nutrients", "nutrients.nutrient", false, false, false),
		newSimpleIndex("sugars", "sugars.type", false, false, false),
		newSimpleIndex("additives", "additives.additive", false, false, false),
		newSimpleIndex("standard", "standard", true, false, false),
		//Notes (no index unless tags)
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}

	// Built-ins
	inserted, updated := 0, 0
	for _, recipe := range []JarRecipe{
		{
			Name:     "Popcorn",
			Id:       alternateCollectionId(altCollIdForint(idJarPop)),
			Grain:    Popcorn,
			Standard: true,
			Notes: []Note{
				builtInNote("Typically pretty expensive comparably to oats"),
			},
		},
		{
			Name:     "Oats",
			Id:       alternateCollectionId(altCollIdForint(idJarOat)),
			Grain:    Oats,
			Standard: true,
			Notes:    nil,
		},
		{
			Name:     "Oats with standard additives",
			Id:       alternateCollectionId(altCollIdForint(idJarOatWithVermGypsum)),
			Grain:    Oats,
			Standard: true,
			Sugars: []sugarMeasurement{
				newSugarMeasurement(Honey, 1, "large drop per quart jar"),
			},
			Additives: []additiveMeasurement{
				newAdditiveMeasurement(Vermiculite, 0.25, "tsp"),
				newAdditiveMeasurement(Gypsum, 0.7, "coverage of jar bottom"),
			},
			Notes: []Note{
				builtInNote("Gypsum helps grains to not stick, and adds calcium and sulfur"),
				builtInNote("Vermioculite should be added after oats are boiled"),
			},
		},
	} {
		var existing JarRecipe
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
		if recipe.Grain != existing.Grain {
			update = true
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
	if inserted+updated > 0 { // TODO: ok?
		println(fmt.Sprintf(`Jar recipes: inserted %d, updated %d`, inserted, updated))
	}
	return nil
}

type createJarRecipeRequest struct {
	Name      string                `bson:"name" json:"name"`
	Grain     grain                 `bson:"grain" json:"grain"`
	Standard  bool                  `bson:"standard" json:"standard"`                       // If this is a standard recipe
	Nutrients []nutrientMeasurement `bson:"nutrients,omitempty" json:"nutrients,omitempty"` // Per grain jar
	Sugars    []sugarMeasurement    `bson:"sugars,omitempty" json:"sugars,omitempty"`       // Per grain jar
	Additives []additiveMeasurement `bson:"additives,omitempty" json:"additives,omitempty"`
	Notes     []Note                `bson:"notes,omitempty" json:"notes,omitempty"`
}

func createJarRecipeHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req := createJarRecipeRequest{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id := newAlternateCollectionId()
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(jarRecipesCollectionName)
		res, err := coll.InsertOne(r.Context(), JarRecipe{
			Id:          alternateCollectionId(id),
			Name:        req.Name,
			Grain:       req.Grain,
			Standard:    req.Standard,
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
			http.Error(w, "id fail, should never happen", http.StatusInternalServerError)
			return nil, nil
		}
		return w.Write(altId.base58Bytes())
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}

type updateJarRecipeRequest struct {
	Name     string           `json:"name"`
	Standard bool             `json:"standard"`
	Notes    AllEntries[Note] `json:"notes"`
}

func updateJarRecipeHandler(w http.ResponseWriter, r *http.Request) { // TODO: txn?
	b58Id := Base58Str(r.PathValue("id"))
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req := updateJarRecipeRequest{}
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
	existing, err := GetAltCollectionItem(r.Context(), id.String(), JarRecipe{})
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
		coll := ctx.Client().Database(dbName).Collection(jarRecipesCollectionName)
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
