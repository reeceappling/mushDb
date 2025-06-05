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

const substrateRecipesCollectionName = "subRecipes"

type SubstrateRecipe struct {
	Id          alternateCollectionId `bson:"_id" json:"_id"`
	Name        string                `bson:"name" json:"name"`         // TODO: ensure indexed and properly handled
	Standard    bool                  `bson:"standard" json:"standard"` // If this is a standard recipe // TODO: account for this
	Aliases     []string              `bson:"aliases,omitempty" json:"aliases,omitempty"`
	Notes       []Note                `bson:"notes,omitempty" json:"notes,omitempty"` // TODO: ingredients in notes
	LastUpdated unixTime              `bson:"lastUpdated" json:"lastUpdated"`
}

func (recipe SubstrateRecipe) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := recipe
	err := decodeItem(&out, encoded)
	return out, err
}

func (recipe SubstrateRecipe) clean() CollectionItem {
	return recipe
}

func (recipe SubstrateRecipe) EntryTypeField() *string {
	return nil
}

func (recipe SubstrateRecipe) CollectionName() string {
	return substrateRecipesCollectionName
}

func initializeSubstrates(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(substrateRecipesCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		newSimpleIndex("name", "name", false, false, true),
		newSimpleIndex("aliases", "aliases", false, true, false),
		newSimpleIndex("standard", "standard", true, false, false),
		//Notes (no index unless tags)
		//LastUpdated
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	inserted, updated := 0, 0
	for _, recipe := range []SubstrateRecipe{
		// Coir
		{
			Name:     "Coir",
			Aliases:  []string{},
			Id:       alternateCollectionId(altCollIdForint(idCoir)), // TODO: this
			Standard: true,
			Notes: []Note{
				{
					Time: ogTime,
					Note: "roughly 40g dry coir, 1 cup H20 per quart",
				},
			},
		},
		// Coir and Vermiculite
		{
			Name:     "CVG",
			Aliases:  []string{"Coir with Vermiculite"},
			Id:       alternateCollectionId(altCollIdForint(idCoirVerm)), // TODO: this
			Standard: true,
			Notes: []Note{
				{
					Time: ogTime,
					Note: "Recipe: roughly 40g dry coir, up to 1/2 cup vermiculite, 1 cup H20 per quart",
				},
				{
					Time: ogTime,
					Note: "Vermiculite helps to keep more moisture in the substrate over time",
				},
			},
		},
		{
			Name:     "HWFP",
			Aliases:  []string{"Hardwood Fuel Pellets"},
			Id:       alternateCollectionId(altCollIdForint(idWoodPellets)), // TODO: this
			Standard: true,
			Notes: []Note{
				{
					Time: ogTime,
					Note: "Roughly equal parts wood pellets and water (maybe less water. Do less at first to ensure field capacity)",
				},
			},
		},
	} {
		var existing SubstrateRecipe
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
		println(fmt.Sprintf(`Substrate recipes: inserted %d, updated %d`, inserted, updated))
	}
	return nil
}

type createSubstrateRecipeRequest struct {
	Name     string   `bson:"name" json:"name"`
	Aliases  []string `json:"aliases,omitempty"`
	Standard bool     `json:"standard"` // If this is a standard recipe
	Notes    []Note   `json:"notes,omitempty"`
}

func createSubstrateRecipeHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	req := createSubstrateRecipeRequest{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	id := newAlternateCollectionId()
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(substrateRecipesCollectionName)
		_, err := coll.InsertOne(r.Context(), SubstrateRecipe{
			Id:          alternateCollectionId(id),
			Name:        req.Name,
			Aliases:     req.Aliases,
			Standard:    req.Standard,
			Notes:       req.Notes,
			LastUpdated: unixTime(time.Now().UnixMilli()),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, nil
		}
		return w.Write(id.base58Bytes())
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}

type updateSubstrateRecipeRequest struct {
	Name     string           `bson:"name" json:"name"`
	Aliases  []string         `json:"aliases,omitempty"`
	Standard bool             `json:"standard"` // If this is a standard recipe
	Notes    AllEntries[Note] `json:"notes"`
}

func updateSubstrateRecipeHandler(w http.ResponseWriter, r *http.Request) {
	b58Id := Base58Str(r.PathValue("id"))
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req := updateSubstrateRecipeRequest{}
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
		coll := ctx.Client().Database(dbName).Collection(substrateRecipesCollectionName)
		existing, err := GetAltCollectionItemInTxn(ctx, id.String(), SubstrateRecipe{})
		if err != nil {
			stat := http.StatusInternalServerError
			if err == mongo.ErrNoDocuments {
				stat = http.StatusNotFound
			}
			http.Error(w, err.Error(), stat)
			return nil, nil
		}
		mods := bson.D{}
		// change name if needed
		if req.Name != existing.Name {
			mods = bson.D{{"$set", bson.D{{"name", req.Name}}}}
		}
		// aliases
		mods = setStringArrayIfUnequal(mods, req.Aliases, existing.Aliases, "aliases")
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
