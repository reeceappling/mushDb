package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/reeceappling/mushDb/api/env"
	"github.com/reeceappling/mushDb/api/request"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
)

// used in substrateBatch, species, (bag, fruiting chamber, get it either provided or from the substrateBatch)

var _ CollectionItem = &SubstrateRecipe{}

type SubstrateRecipeField struct {
	Substrate AlternateCollectionId `bson:"recipe" json:"recipe"`
}

func (field SubstrateRecipeField) Get(ctx context.Context) (out SubstrateRecipe, err error) {
	err = DbFrom(ctx).Collection(SubstrateRecipesCollectionName).FindOne(ctx, bson.M{
		IDfld: field.Substrate,
	}).Decode(&out)
	return out, err
}

type SubstrateRecipe struct {
	AlternateCollectionIdField `bson:"inline"`
	NameField                  `bson:"inline"`
	StandardField              `bson:"inline"`
	AliasesField               `bson:"inline"` // must be unique everywhere
	NotesField                 `bson:"inline"` // ingredients in notes
	LastUpdatedField           `bson:"inline"`
	AclField                   `bson:"inline"`
}

func initializeSubstrates(ctx context.Context) error {
	// Indices
	coll := DbFrom(ctx).Collection(SubstrateRecipesCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		newSimpleIndex("name", "name", false, false, true),
		standardIndexModel,
		aliasesIndexModel,
		//Notes (no index unless tags)
		//LastUpdated
		projectsIndexModel,
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	return env.IfNotProd(ctx, func() error {
		basicEntries := []*SubstrateRecipe{
			// Coir
			{
				AlternateCollectionIdField: altCollIdFieldForint(idCoir),
				NameField:                  NameField{"Coir-Builtin"},
				AliasesField:               AliasesField{[]string{}},
				StandardField:              StandardField{true},
				NotesField: NotesField{[]Note{
					newNote(ogTime, "roughly 40g dry coir, 1 cup H20 per quart"),
				}},
				AclField: allCanReadAcl(nil),
			},
			// Coir and Vermiculite
			{
				AlternateCollectionIdField: altCollIdFieldForint(idCoirVerm),
				NameField:                  NameField{"CVG-Builtin"},
				AliasesField:               AliasesField{[]string{"Coir with Vermiculite"}},
				StandardField:              StandardField{true},
				NotesField: NotesField{[]Note{
					newNote(ogTime, "Recipe: roughly 40g dry coir, up to 1/2 cup vermiculite, 1 cup H20 per quart"),
					newNote(ogTime, "Vermiculite helps to keep more moisture in the substrate over time"),
				}},
				AclField: allCanReadAcl(nil),
			},
			{
				AlternateCollectionIdField: altCollIdFieldForint(idWoodPellets),
				NameField:                  NameField{"HWFP-Builtin"},
				AliasesField:               AliasesField{[]string{"Hardwood Fuel Pellets"}},
				StandardField:              StandardField{true},
				NotesField: NotesField{[]Note{
					newNote(ogTime, "Roughly equal parts wood pellets and water (maybe less water. Do less at first to ensure field capacity)"),
				}},
				AclField: allCanReadAcl(nil),
			},
		}
		err = addBasicAltEntries(ctx, basicEntries...)
		if err != nil {
			return err
		}

		// Add test entry
		testItem := &SubstrateRecipe{
			AlternateCollectionIdField: altCollIdFieldForint(idTestingOnly),
			NameField:                  NameField{testEntryStringId},
			StandardField:              StandardField{false},
			AliasesField:               AliasesField{[]string{"testSubstrate", "example substrate"}},
			NotesField:                 NotesField{exampleNotes()},
			LastUpdatedField:           LastUpdatedField{exampleTime},
			AclField:                   allCanWriteAcl(),
		}
		return addTestAltEntries(ctx, testItem)
	})
}

type createSubstrateRecipeRequest struct {
	NameField
	AliasesField
	StandardField // If this is a standard recipe
	NotesField
	// Initial perms are read by all and write only by owner
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
	ctx, now := request.UnixTime(r.Context())
	toInsert := SubstrateRecipe{
		AlternateCollectionIdField: AlternateCollectionIdField{id},
		NameField:                  req.NameField,
		AliasesField:               req.AliasesField,
		StandardField:              req.StandardField,
		NotesField:                 req.NotesField,
		LastUpdatedField:           LastUpdatedField{now},
		AclField:                   allCanReadAcl(GetUserEmailPtr(ctx)),
	}
	finishCreateAlternateEntry(ctx, toInsert, w)
}

type updateSubstrateRecipeRequest struct {
	NameField
	AliasesField
	StandardField
	NotesUpdateField
	PermsOnRequest `json:"acl"`
}

func (req updateSubstrateRecipeRequest) modsFor(existing *SubstrateRecipe, aclField AclField) (bson.D, error) {
	return NewMods().
		updateNameIfNeeded(req.Name, existing.Name).
		updateAliasesIfNeeded(req.Aliases, existing.Aliases). // TODO: make sure no duplicates
		updateStandardIfNeeded(req.Standard, existing.Standard).
		updateNotesIfNeeded(req, existing).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
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
	ctx, db := Db(r)
	coll := db.Collection(SubstrateRecipesCollectionName)
	existing, err := GetAltCollectionItemOutsideTxn(ctx, id, SubstrateRecipe{})
	if err != nil {
		stat := http.StatusInternalServerError
		if err == mongo.ErrNoDocuments {
			stat = http.StatusNotFound
		}
		dbErr(w, err.Error(), stat)
		return
	}
	finishAltCollItemUpdate(ctx, w, coll, req.modsFor, &existing, req.PermsOnRequest)
}

func deleteSubstrateRecipeHandler(w http.ResponseWriter, r *http.Request) {
	// used in batch, species, box, bag!
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
	// Ensure not used as a standard substrate for a species
	err = db.Collection(SpeciesCollectionName).FindOne(ctx, bson.M{"standardSubstrate": id}).Err()
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			http.Error(w, "failed to check for substrate recipe usage in species. "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// At least one item exists, fail
		http.Error(w, "at least one species utilizes the item you are attempting to delete.", http.StatusConflict)
		return
	}
	for collName, key := range map[string]string{
		SpeciesCollectionName:         "standardSubstrate",
		SubstrateBatchCollectionName:  "recipe",
		FruitingChamberCollectionName: "recipe",
		BagsCollectionName:            "recipe",
	} {
		err = db.Collection(collName).FindOne(ctx, bson.M{key: id}).Err()
		if err != nil {
			if !errors.Is(err, mongo.ErrNoDocuments) {
				http.Error(w, "failed to check for substrateRecipe usage in "+collName+". "+err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			// At least one item exists, fail
			http.Error(w, "at least one of "+collName+" utilizes the item you are attempting to delete.", http.StatusConflict)
			return
		}
	}

	// Delete if not found elsewhere!
	deleteResult, err := db.Collection(SubstrateRecipesCollectionName).DeleteOne(ctx, bson.M{IDfld: id})
	if err != nil {
		http.Error(w, "failed to delete: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if deleteResult.DeletedCount == 0 {
		http.Error(w, "failed to delete id "+idStr+" from substrate recipes. Id not found", http.StatusNotFound)
		return
	}
	_, err = w.Write([]byte(idStr))
	handleWriteErr(err, w)
}
