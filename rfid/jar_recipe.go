package rfid

// TODO: JAR RECIPE BATCH (soaked? simmered/time?)

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

const jarRecipesCollectionName = "jarRecipes"

type JarRecipeField struct {
	Recipe *AlternateCollectionId `bson:"recipe,omitempty" json:"recipe,omitempty"`
}

type JarRecipeRequiredField struct {
	Recipe AlternateCollectionId `bson:"recipe" json:"recipe"`
}

func (field JarRecipeRequiredField) Get(ctx context.Context) (out JarRecipe, err error) {
	err = ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(jarRecipesCollectionName).
		FindOne(ctx, bson.M{"_id": field.Recipe}).Decode(&out)
	return
}

func (field JarRecipeField) Get(ctx context.Context) (out JarRecipe, err error) {
	if field.Recipe == nil {
		return out, ErrMissingOptionalField
	}
	err = ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(jarRecipesCollectionName).
		FindOne(ctx, bson.M{"_id": *field.Recipe}).Decode(&out)
	return
}

type JarRecipe struct {
	AlternateCollectionIdField `bson:"inline"`
	NameField                  `bson:"inline"`
	Grains                     []GrainPercentage `bson:"grains" json:"grains"`
	StandardField              `bson:"inline"`   // If this is a standard recipe
	NutrientsField             `bson:"inline"`   // Per grain jar?
	SugarsField                `bson:"inline"`   // Per grain jar?
	AdditivesField             `bson:"inline"`   // Per grain jar?
	NotesField                 `bson:"inline"`
	LastUpdatedField           `bson:"inline"`
	AclField                   `bson:"inline"` // TODO: handle EVERYWHERE
}

type GrainPercentage struct {
	Grain      grain `bson:"grain" json:"grain"`
	Percentage int   `bson:"percentage" json:"percentage"`
}

func (recipe JarRecipe) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := recipe
	err := decodeItem(&out, encoded)
	return out, err
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
		newSimpleIndex("grains", "grains.grain", false, false, false),
		standardIndexModel,
		newSimpleIndex("nutrients", "nutrients.nutrient", false, false, false),
		newSimpleIndex("sugars", "sugars.type", false, false, false),
		newSimpleIndex("additives", "additives.additive", false, false, false),
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
			NameField:                  NameField{"Popcorn"},
			AlternateCollectionIdField: AlternateCollectionIdField{altCollIdForint(idJarPop)},
			Grains:                     []GrainPercentage{{Grain: Popcorn, Percentage: 100}},
			StandardField:              StandardField{true},
			NotesField: NotesField{[]Note{
				builtInNote("Typically pretty expensive comparably to oats"),
			}},
		},
		{
			NameField:                  NameField{"Oats"},
			AlternateCollectionIdField: AlternateCollectionIdField{altCollIdForint(idJarOat)},
			Grains:                     []GrainPercentage{{Grain: Oats, Percentage: 100}},
			StandardField:              StandardField{true},
			NotesField:                 NotesField{},
		},
		{
			NameField:                  NameField{"Oats with standard additives"},
			AlternateCollectionIdField: AlternateCollectionIdField{altCollIdForint(idJarOatWithVermGypsum)},
			Grains:                     []GrainPercentage{{Grain: Oats, Percentage: 100}},
			StandardField:              StandardField{true},
			SugarsField: SugarsField{[]sugarMeasurement{
				newSugarMeasurement(Honey, 1, "large drop per quart jar"),
			}},
			AdditivesField: AdditivesField{[]additiveMeasurement{
				newAdditiveMeasurement(Vermiculite, 0.25, "tsp"),
				newAdditiveMeasurement(Gypsum, 0.7, "coverage of jar bottom"),
			}},
			NotesField: NotesField{[]Note{
				builtInNote("Gypsum helps grains to not stick, and adds calcium and sulfur"),
				builtInNote("Vermioculite should be added after oats are boiled"),
			}},
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
		if len(recipe.Grains) != len(existing.Grains) {
			update = true
		} else {
			for i, _ := range recipe.Grains {
				if recipe.Grains[i] != existing.Grains[i] {
					update = true
					break
				} else {
					if recipe.Grains[i].Percentage != existing.Grains[i].Percentage {
						update = true
						break
					}
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
	// If test jar recipe does not exist, then create it
	existingEntry := JarRecipe{}
	testItem := JarRecipe{
		AlternateCollectionIdField: AlternateCollectionIdField{exAltId},
		NameField:                  NameField{"testJarRecipeName"},
		Grains:                     []GrainPercentage{{Grain: BirdSeed, Percentage: 100}},
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
	if inserted+updated > 0 { // TODO: ok?
		println(fmt.Sprintf(`Jar recipes: inserted %d, updated %d`, inserted, updated))
	}
	return err
}

type createJarRecipeRequest struct {
	NameField
	Grains        []GrainPercentage `bson:"grains" json:"grains"`
	StandardField                   // If this is a standard recipe
	NutrientsField
	SugarsField
	AdditivesField
	NotesField
	UsersThatCanEdit    []Base58Str   `json:"usersThatCanEdit,omitempty"`    // TODO: DO THIS IN TYPESCRIPT!
	ProjectsThatCanEdit []projectName `json:"projectsThatCanEdit,omitempty"` // TODO: DO THIS IN TYPESCRIPT!
}

func (req createJarRecipeRequest) ValidateGrains() error {
	pct := 0
	if len(req.Grains) < 1 {
		return errors.New("no grains in request")
	}
	for i, g := range req.Grains {
		if g.Percentage < 1 {
			return errors.New("grain percentage must be greater than zero")
		}
		pct += req.Grains[i].Percentage
	}
	if pct != 100 {
		return errors.New("invalid grain percentages. Must add to 100")
	}
	return nil
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
	if err = errors.Join( // TODO: do this on all other recipes and stuff
		req.ValidateGrains(),
		req.NutrientsField.Validate(),
		req.SugarsField.Validate(),
		req.AdditivesField.Validate(),
	); err != nil {
		http.Error(w, "invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}
	resolvedUserPerms, err := GetAuthInfo(r.Context())
	if err != nil {
		http.Error(w, "Failed to get auth info", http.StatusUnauthorized)
		return
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(jarRecipesCollectionName)
		acl, err := newAlwaysReadableAcl(ctx, resolvedUserPerms, nil, nil)
		if err != nil {
			return DbTxnStdErr(w, "failed to resolve new ACL: "+err.Error(), http.StatusInternalServerError)
		}
		toInsert := JarRecipe{
			AlternateCollectionIdField: AlternateCollectionIdField{newAlternateCollectionId()},
			NameField:                  NameField{req.Name},
			Grains:                     req.Grains,
			StandardField:              StandardField{req.Standard},
			NutrientsField:             NutrientsField{req.Nutrients},
			SugarsField:                SugarsField{req.Sugars},
			AdditivesField:             AdditivesField{req.Additives},
			NotesField:                 NotesField{req.Notes},
			LastUpdatedField:           LastUpdatedField{unixTimeForNow()},
			AclField:                   acl,
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

type updateJarRecipeRequest struct {
	NameField
	StandardField
	Notes          AllEntries[Note] `json:"notes"`
	PermsOnRequest                  // TODO: handle in typescript and handler!
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

	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(jarRecipesCollectionName)
		existing := JarRecipe{}
		err = coll.FindOne(ctx, bson.D{{"_id", id}}).Decode(&existing)
		if err != nil {
			return DbTxnStdErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		}
		user, err := GetAuthInfo(ctx)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		if !user.HasPermissionToEdit(existing) {
			return DbTxnStdErr(w, "unauthorized to edit", http.StatusForbidden)
		}
		aclField, err := req.AclFor(ctx, user)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		// TODO: make and/or validate grain changes?
		upd, err := NewMods().
			updateNameIfNeeded(req.Name, existing.Name).
			updateStandardIfNeeded(req.Standard, existing.Standard).
			updateNotesIfNeeded(req.Notes, existing.Notes).
			updatePermsIfNeeded(aclField.ACL, existing.ACL).
			updateLastUpdatedIfNeeded().
			Finalized()
		if err != nil {
			return DbTxnStdErr(w, "error creating txn:"+err.Error(), http.StatusInternalServerError)
		}
		if len(upd) == 0 {
			return DbTxnStdErr(w, "no changes made", http.StatusBadRequest)
		}
		bsonId := bson.D{{"_id", existing.Id}}
		// TODO: USE handleUpdateMods
		// TODO: UPDATE PERMS!
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
		return w.Write(bs)
	})
	if err != nil {
		HandleHttpWriteError(err)
	}
}
