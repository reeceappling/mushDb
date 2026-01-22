package rfid

// TODO: JAR RECIPE BATCH (soaked? simmered/time?)

import (
	"context"
	"encoding/json"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
)

const JarRecipesCollectionName = "jarRecipes"

type JarRecipeField struct {
	Recipe *AlternateCollectionId `bson:"recipe,omitempty" json:"recipe,omitempty"`
}

type JarRecipeRequiredField struct {
	Recipe AlternateCollectionId `bson:"recipe" json:"recipe"`
}

func (field JarRecipeRequiredField) Get(ctx context.Context) (out JarRecipe, err error) {
	err = ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(JarRecipesCollectionName).
		FindOne(ctx, bson.M{"_id": field.Recipe}).Decode(&out)
	return
}

func (field JarRecipeField) Get(ctx context.Context) (out JarRecipe, err error) {
	if field.Recipe == nil {
		return out, ErrMissingOptionalField
	}
	err = ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(JarRecipesCollectionName).
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
	AclField                   `bson:"inline"`
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
	return JarRecipesCollectionName
}

func initializeJarRecipes(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(JarRecipesCollectionName)
	err := createIndexes(ctx, coll,
		[]mongo.IndexModel{
			newSimpleIndex("name", "name", false, false, false),
			//newSimpleIndex("grains", "grains.grain", false, false, false),
			standardIndexModel,
			//newSimpleIndex("nutrients", "nutrients.nutrient", false, false, false),
			//newSimpleIndex("sugars", "sugars.type", false, false, false),
			//newSimpleIndex("additives", "additives.additive", false, false, false),
			//Notes (no index unless tags)
			projectsIndexModel,
			lastUpdatedIndexModel,
		})
	if err != nil {
		return err
	}

	// Built-ins
	basicEntries := []*JarRecipe{
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
	}
	err = addBasicAltEntries(ctx, basicEntries...)
	if err != nil {
		return err
	}
	// If test jar recipe does not exist, then create it
	testItem := &JarRecipe{
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
	// TODO: add built-in entries
	return addTestAltEntries(ctx, testItem)
}

type createJarRecipeRequest struct {
	NameField
	Grains        []GrainPercentage `bson:"grains" json:"grains"`
	StandardField                   // If this is a standard recipe
	NutrientsField
	SugarsField
	AdditivesField
	NotesField
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
	ctx, db := Db(r)
	coll := db.Collection(JarRecipesCollectionName)
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
		AclField:                   allCanWriteAcl(),
	}
	finishCreateAlternateEntry(ctx, coll, toInsert, w)
}

type updateJarRecipeRequest struct {
	NameField
	StandardField
	Notes AllEntries[Note] `json:"notes"`
	PermsOnRequest
}

func (req updateJarRecipeRequest) modsFor(existing *JarRecipe, aclField AclField) (bson.D, error) {
	return NewMods().
		updateNameIfNeeded(req.Name, existing.Name).
		updateStandardIfNeeded(req.Standard, existing.Standard).
		updateNotesIfNeeded(req.Notes, existing.Notes).
		// TODO: for perms updates, disallow removing self???
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
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
	ctx, db := Db(r)
	coll := db.Collection(JarRecipesCollectionName)
	existing := JarRecipe{}
	err = coll.FindOne(ctx, bson.D{{"_id", id}}).Decode(&existing)
	if err != nil {
		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishAltCollItemUpdate(ctx, w, coll, req.modsFor, &existing, req.PermsOnRequest)
}
