package rfid

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/reeceappling/goUtils/v2/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"slices"
	"time"
)

const agarRecipesCollectionName = "agarRecipes"

type AgarRecipe struct {
	Id          alternateCollectionId `bson:"_id" json:"_id"`
	Name        string                `bson:"name" json:"name"`
	Liquids     []liquid              `bson:"liquids" json:"liquids"`   // TapWater, DistilledWater, GrainWater (Oat, etc)
	Agar        int                   `bson:"agar" json:"agar"`         // agar grams per 1L
	Standard    bool                  `bson:"standard" json:"standard"` // If this is a standard recipe
	Nutrients   []nutrientMeasurement `bson:"nutrients,omitempty" json:"nutrients,omitempty"`
	Sugars      []sugarMeasurement    `bson:"sugars,omitempty" json:"sugars,omitempty"`
	Additives   []additiveMeasurement `bson:"additives,omitempty" json:"additives,omitempty"`
	Antibiotics []antibiotic          `bson:"antibiotics,omitempty" json:"antibiotics,omitempty"`
	Notes       []Note                `bson:"notes,omitempty" json:"notes,omitempty"`
	LastUpdated unixTime              `bson:"lastUpdated" json:"lastUpdated"`
}

func (recipe AgarRecipe) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := recipe
	err := decodeItem(&out, encoded)
	return out, err
}

func (recipe AgarRecipe) clean() CollectionItem {
	return recipe
}

func (recipe AgarRecipe) EntryTypeField() *string {
	return nil
}

func (recipe AgarRecipe) CollectionName() string {
	return agarRecipesCollectionName
}

type NewAgarRecipeRequest struct {
	Name        string `json:"name"`
	Standard    *bool  `json:"standard,omitempty"`
	Liquids     []liquid
	AgarGPerL   int
	Nutrients   []nutrientMeasurement
	Sugars      []sugarMeasurement
	Additives   []additiveMeasurement
	Antibiotics []antibiotic
	Notes       []string
}

func (req NewAgarRecipeRequest) asRecipe() AgarRecipe {
	now := time.Now()
	return AgarRecipe{
		Id:          alternateCollectionId(newAlternateCollectionId()),
		Name:        req.Name,
		Agar:        req.AgarGPerL,
		Liquids:     req.Liquids,
		Nutrients:   req.Nutrients,
		Sugars:      req.Sugars,
		Additives:   req.Additives,
		Antibiotics: req.Antibiotics,
		Standard:    utils.Default(req.Standard, false),
		Notes:       stringsToNotes(req.Notes, now),
		LastUpdated: unixTime(now.UnixMilli()),
	}
}

func newAgarRecipe(ctx mongo.SessionContext, req AgarRecipe) utils.Result[alternateCollectionId] {
	res, err := ctx.Client().Database(dbName).Collection(agarRecipesCollectionName).InsertOne(ctx, req)
	if err != nil {
		return utils.ErroredResult[alternateCollectionId](err)
	}
	if res == nil {
		return utils.ErroredResult[alternateCollectionId](errors.New("empty response when adding recipe"))
	}
	out, ok := res.InsertedID.(alternateCollectionId)
	if !ok {
		return utils.ErroredResult[alternateCollectionId](errors.New("failed to resolve new agar recipe ID"))
	}
	return utils.SuccessfulResult(alternateCollectionId(out))
}

type updateAgarRecipeRequest struct {
	Name     string           `json:"name"`
	Standard bool             `json:"standard"`
	Notes    AllEntries[Note] `json:"notes"`
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
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(agarRecipesCollectionName)
		existing, err := GetAltCollectionItem(ctx, id.String(), AgarRecipe{})
		if err != nil {
			stat := http.StatusInternalServerError
			if err == mongo.ErrNoDocuments {
				stat = http.StatusNotFound
			}
			http.Error(w, err.Error(), stat)
			return nil, nil
		}
		if !ableToModify(ctx) { // TODO: DO THIS EVERYWHERE!
			http.Error(w, "not permitted to modify", http.StatusForbidden)
			return nil, nil
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
			return nil, nil
		}
		if len(mods) == 0 {
			http.Error(w, "no changes made", http.StatusBadRequest)
			return nil, nil
		}
		result := coll.FindOneAndUpdate(ctx, bson.D{{"_id", existing.Id}}, mods)
		err = result.Err()
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

func initializeAgarRecipes(ctx context.Context) error {
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(agarRecipesCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		newSimpleIndex("name", "name", false, false, false),
		newSimpleIndex("liquids", "liquids.name", false, false, false),
		newSimpleIndex("agar", "agar", true, false, false),
		newSimpleIndex("nutrients", "nutrients.nutrient", false, false, false),
		newSimpleIndex("sugars", "sugars.type", false, false, false),
		newSimpleIndex("additives", "additives.additive", false, false, false),
		newSimpleIndex("antibiotics", "antibiotics", false, false, false),
		newSimpleIndex("standard", "standard", true, false, false),
		//Notes (not indexed, unless tags)
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	inserted, updated := 0, 0
	for _, recipe := range []AgarRecipe{
		{
			Name:    "LMEA - Light Malt Extract Agar",
			Id:      altCollIdForint(idLmea),
			Liquids: []liquid{Water.AsLiquid()},
			Agar:    20,
			Nutrients: []nutrientMeasurement{
				{
					Nutrient: LME,
					Amount:   20,
					Unit:     "g",
				},
			},
			Sugars:   nil,
			Standard: true,
			Notes:    nil,
		},
		{
			Name:    "PDA - Potato Dextrose Agar",
			Id:      altCollIdForint(idPda),
			Liquids: []liquid{Water.AsLiquid()},
			Agar:    20,
			Nutrients: []nutrientMeasurement{{
				Nutrient: Potato,
				Amount:   18,
				Unit:     "g",
			}},
			Sugars: []sugarMeasurement{{
				Type:   Dextrose,
				Amount: 1,
				Unit:   "g",
			}},
			Standard: true,
			Notes:    nil,
		},
		{
			Name:      "Water Agar",
			Id:        altCollIdForint(idWaterAgar),
			Liquids:   []liquid{Water.AsLiquid()},
			Agar:      20,
			Nutrients: nil,
			Sugars:    nil,
			Standard:  true,
			Notes:     nil,
		},
		{
			Name: "Grain Water Agar",
			Id:   altCollIdForint(idGrainWaterAgar),
			Liquids: []liquid{
				GrainWater.AsLiquid().withPct(50.0),
				DistilledWater.AsLiquid().withPct(50.0),
			},
			Agar:      20,
			Nutrients: nil,
			Sugars:    nil,
			Standard:  true,
			Notes: []Note{
				builtInNote("Grain water also acts as a nutrient source"),
			},
		},
		{
			Name:    "Antibiotic Agar",
			Id:      altCollIdForint(idAntibioticAgar),
			Liquids: []liquid{DistilledWater.AsLiquid()},
			Agar:    20,
			Nutrients: []nutrientMeasurement{
				{
					Nutrient: LME,
					Amount:   20,
					Unit:     "g",
				},
			},
			Sugars:      nil,
			Antibiotics: []antibiotic{Doxycycline},
			Standard:    true,
			Notes: []Note{
				builtInNote("50mg doxycycline per ?????"),
			},
		},
	} {
		var existing AgarRecipe
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
		if !update && recipe.Agar != existing.Agar {
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
			if !slices.Contains(finalNotes, note) {
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
		println(fmt.Sprintf(`Agar recipes: inserted %d, updated %d`, inserted, updated))
	}
	return nil
}

type createAgarRecipeRequest struct {
	Name        string                `json:"name"`
	Standard    bool                  `json:"standard"`
	Agar        int                   `json:"agar"`
	Liquids     []liquid              `json:"liquids"`
	Nutrients   []nutrientMeasurement `json:"nutrients,omitempty"`
	Sugars      []sugarMeasurement    `json:"sugars,omitempty"`
	Additives   []additiveMeasurement `json:"additives,omitempty"`
	Antibiotics []antibiotic          `json:"antibiotics,omitempty"`
	Notes       []Note                `json:"notes,omitempty"`
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
	entry := AgarRecipe{
		Id:          id,
		Name:        req.Name,
		Liquids:     req.Liquids,
		Agar:        req.Agar,
		Standard:    req.Standard,
		Nutrients:   req.Nutrients,
		Sugars:      req.Sugars,
		Additives:   req.Additives,
		Antibiotics: req.Antibiotics,
		Notes:       req.Notes,
		LastUpdated: unixTime(time.Now().UnixMilli()),
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		result := newAgarRecipe(ctx, entry)
		if result.Err != nil {
			http.Error(w, "Agar batch creation failure: "+result.Err.Error(), http.StatusInternalServerError)
			return nil, nil
		}
		return w.Write(id.base58Bytes())
	})

	if err != nil {
		handleWriteErr(err, w)
	}
}
