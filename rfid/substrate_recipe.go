package rfid

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"reflect"
	sliceutils "slices"
)

const substrateRecipesCollectionName = "substrateRecipes"

type SubstrateRecipeField struct {
	Substrate AlternateCollectionId `bson:"recipe" json:"recipe"`
}

func (field SubstrateRecipeField) Get(ctx context.Context) (out SubstrateRecipe, err error) {
	err = ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(pcRunCollectionName).FindOne(ctx, bson.M{
		"_id": field.Substrate,
	}).Decode(&out)
	return out, err
}

type SubstrateRecipe struct {
	AlternateCollectionIdField
	NameField
	StandardField
	AliasesField // TODO: make sure no duplicates
	NotesField   // TODO: ingredients in notes
	LastUpdatedField
}

func (recipe SubstrateRecipe) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := recipe
	err := decodeItem(&out, encoded)
	return out, err
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
			NameField:                  NameField{"Coir"},
			AliasesField:               AliasesField{[]string{}},
			AlternateCollectionIdField: altCollIdFieldForint(idCoir),
			StandardField:              StandardField{true},
			NotesField: NotesField{[]Note{
				{
					Time: ogTime,
					Note: "roughly 40g dry coir, 1 cup H20 per quart",
				},
			}},
		},
		// Coir and Vermiculite
		{
			NameField:                  NameField{"CVG"},
			AliasesField:               AliasesField{[]string{"Coir with Vermiculite"}},
			AlternateCollectionIdField: altCollIdFieldForint(idCoirVerm),
			StandardField:              StandardField{true},
			NotesField: NotesField{[]Note{
				{
					Time: ogTime,
					Note: "Recipe: roughly 40g dry coir, up to 1/2 cup vermiculite, 1 cup H20 per quart",
				},
				{
					Time: ogTime,
					Note: "Vermiculite helps to keep more moisture in the substrate over time",
				},
			}},
		},
		{
			NameField:                  NameField{"HWFP"},
			AliasesField:               AliasesField{[]string{"Hardwood Fuel Pellets"}},
			AlternateCollectionIdField: altCollIdFieldForint(idWoodPellets),
			StandardField:              StandardField{true},
			NotesField: NotesField{[]Note{
				{
					Time: ogTime,
					Note: "Roughly equal parts wood pellets and water (maybe less water. Do less at first to ensure field capacity)",
				},
			}},
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
	// Add test entry
	existingEntry := SubstrateRecipe{}
	testItem := SubstrateRecipe{
		AlternateCollectionIdField: altCollIdFieldForint(idTestingOnly),
		NameField:                  NameField{testEntryStringId},
		StandardField:              StandardField{false},
		AliasesField:               AliasesField{[]string{"testSubstrate", "example substrate"}}, // TODO: search by aliases?
		NotesField:                 NotesField{exampleNotes()},
		LastUpdatedField:           LastUpdatedField{exampleTime},
	}
	err = coll.FindOne(ctx, bson.D{{"_id", exAltId}}).Decode(&existingEntry)
	if err == nil {
		if reflect.DeepEqual(existingEntry, testItem) {
			return nil
		}
	}
	err = testExistingEntry(ctx, coll, exAltId, testItem, existingEntry)
	if inserted+updated > 0 {
		println(fmt.Sprintf(`SubstrateRecipe: inserted %d, updated %d`, inserted, updated))
	}
	return err
}

type createSubstrateRecipeRequest struct {
	NameField
	AliasesField
	StandardField // If this is a standard recipe
	NotesField
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
		toInsert := SubstrateRecipe{
			AlternateCollectionIdField: AlternateCollectionIdField{id},
			NameField:                  req.NameField,
			AliasesField:               req.AliasesField,
			StandardField:              req.StandardField,
			NotesField:                 req.NotesField,
			LastUpdatedField:           LastUpdatedFieldForNow(),
		}
		_, err = coll.InsertOne(r.Context(), toInsert)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		bsOut, err := json.Marshal(toInsert)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		return w.Write(bsOut)
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}

type updateSubstrateRecipeRequest struct {
	NameField
	AliasesField
	StandardField
	Notes AllEntries[Note] `json:"notes"`
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
		existing, err := GetAltCollectionItemInTxn(ctx, id, SubstrateRecipe{})
		if err != nil {
			stat := http.StatusInternalServerError
			if err == mongo.ErrNoDocuments {
				stat = http.StatusNotFound
			}
			return DbTxnStdErr(w, err.Error(), stat)
		}
		upd, err := NewMods().
			updateNameIfNeeded(req.Name, existing.Name).
			updateAliasesIfNeeded(req.Aliases, existing.Aliases). // TODO: make sure no duplicates?
			updateStandardIfNeeded(req.Standard, existing.Standard).
			updateNotesIfNeeded(req.Notes, existing.Notes).
			updateLastUpdatedIfNeeded().
			Finalized()
		if err != nil {
			return DbTxnStdErr(w, "error resolving updates list: "+err.Error(), http.StatusInternalServerError)
		}
		bsonId := bson.D{{"_id", existing.Id}}
		err = coll.FindOneAndUpdate(ctx, bsonId, upd).Err()
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		err = coll.FindOne(ctx, bsonId).Decode(&existing)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		bsOut, err := json.Marshal(existing)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		return w.Write(bsOut)
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}
