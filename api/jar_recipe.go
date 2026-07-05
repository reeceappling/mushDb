package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/reeceappling/mushDb/api/env"
	"github.com/reeceappling/mushDb/api/request"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"log"
	"net/http"
)

// needed for grainJars

const JarRecipesCollectionName = "jarRecipes"

type JarRecipeField struct {
	Recipe *AlternateCollectionId `bson:"recipe,omitempty" json:"recipe,omitempty"`
}

type JarRecipeRequiredField struct {
	Recipe AlternateCollectionId `bson:"recipe" json:"recipe"`
}

func (field JarRecipeRequiredField) Get(ctx context.Context) (out JarRecipe, err error) {
	println("searching for recipe: ", field.Recipe.AsBase58(), string(field.Recipe[:]))
	out = JarRecipe{}
	coll := DbFrom(ctx).Collection(JarRecipesCollectionName)
	findFilter := BsonFindByIdFilterUnordered(field.Recipe) // TODO: BsonFindByIdFilterOrdered(field.Recipe) /* TODO: ??? bson.M{"_id": field.Recipe}*/
	err = coll.FindOne(ctx, findFilter).Decode(&out)
	return out, err
}

func (field JarRecipeField) asRequiredField() (out JarRecipeRequiredField, err error) {
	if field.Recipe == nil {
		return out, ErrMissingOptionalField
	}
	return JarRecipeRequiredField{*field.Recipe}, nil
}

// TODO: move
func checkIdTypeWithRaw[T bson.M | bson.D](ctx context.Context, collection *mongo.Collection, filter T) {
	var rawDoc bson.Raw
	err := collection.FindOne(ctx, filter).Decode(&rawDoc)
	if err != nil {
		log.Fatal(err)
	}

	// Lookup looks up an element in the raw document by key
	idElement, err := rawDoc.LookupErr("_id")
	if err != nil {
		fmt.Println("_id field does not exist in this document")
		return
	}

	// idElement.Type returns a bsontype.Type (e.g., bsontype.ObjectID, bsontype.String)
	fmt.Printf("Exact BSON Type: %s (Hex byte value: %x)\n", idElement.Type, idElement.Type)
}

// TODO: move
func checkIdTypeWithRawOnCursor(cursor *mongo.Cursor) error {
	var rawDoc bson.Raw
	err := cursor.Decode(&rawDoc)
	if err != nil {
		println("failed to decode document from cursor: " + err.Error())
		return errors.Join(err, errors.New("failed to decode document from cursor"))
	}

	// Lookup looks up an element in the raw document by key
	idElement, err := rawDoc.LookupErr("_id")
	if err != nil {
		println("_id field does not exist in this document: " + err.Error())
		return errors.Join(err, errors.New("_id field does not exist in this document"))
	}

	// idElement.Type returns a bsontype.Type (e.g., bsontype.ObjectID, bsontype.String)

	println("item: ", rawDoc.String())
	println("id", idElement.String(), "value", string(idElement.Value))
	//idElType := idElement.Type
	//println(fmt.Sprintf("Exact BSON Type: %s (Hex byte value: %x)\n", idElType, idElType))
	//println(fmt.Sprintf("As string: %s. Value: %s\n", idElement.String(), string(idElement.Value)))
	return nil
}

func (field JarRecipeField) Get(ctx context.Context) (out JarRecipe, err error) {
	f, err := field.asRequiredField()
	if err != nil {
		return out, err
	}
	return f.Get(ctx)
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
	Grain      Grain `bson:"grain" json:"grain"`
	Percentage int   `bson:"percentage" json:"percentage"`
}

func (recipe JarRecipe) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := recipe
	err := decodeItem(&out, encoded)
	return out, err
}

func (recipe JarRecipe) CollectionName() string {
	return JarRecipesCollectionName
}

func initializeJarRecipes(ctx context.Context) error {
	// Indices
	coll := DbFrom(ctx).Collection(JarRecipesCollectionName)
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

	// Built-ins // TODO: ensure these are not completely replaced every time!
	basicEntries := []*JarRecipe{
		{
			NameField:                  NameField{"Popcorn Built-in"},
			AlternateCollectionIdField: AlternateCollectionIdField{altCollIdForint(idJarPop)},
			Grains:                     []GrainPercentage{{Grain: Popcorn, Percentage: 100}},
			StandardField:              StandardField{true},
			NotesField: NotesField{[]Note{
				builtInNote("Typically pretty expensive comparably to oats"),
			}},
		},
		{
			NameField:                  NameField{"Oats Built-in"},
			AlternateCollectionIdField: AlternateCollectionIdField{altCollIdForint(idJarOat)},
			Grains:                     []GrainPercentage{{Grain: Oats, Percentage: 100}},
			StandardField:              StandardField{true},
			NotesField:                 NotesField{},
		},
		{
			NameField:                  NameField{"Oats with standard additives Built-in"},
			AlternateCollectionIdField: AlternateCollectionIdField{altCollIdForint(idJarOatWithVermGypsum)},
			Grains:                     []GrainPercentage{{Grain: Oats, Percentage: 100}},
			StandardField:              StandardField{true},
			SugarsField: SugarsField{[]SugarMeasurement{
				newSugarMeasurement(Honey, 1, "large drop per quart jar"),
			}},
			AdditivesField: AdditivesField{[]AdditiveMeasurement{
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
	// Add test entries
	return env.IfNotProd(ctx, func() error { // TODO: ensure ok
		// If test jar recipe does not exist, then create it
		testItem := &JarRecipe{
			AlternateCollectionIdField: AlternateCollectionIdField{exAltId},
			NameField:                  NameField{"testJarRecipeName"},
			Grains:                     []GrainPercentage{{Grain: BirdSeed, Percentage: 100}},
			StandardField:              StandardField{false},
			NutrientsField: NutrientsField{[]NutrientMeasurement{
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
			SugarsField: SugarsField{[]SugarMeasurement{
				newSugarMeasurement(Honey, 1, "large drop per quart jar"),
			}},
			AdditivesField: AdditivesField{[]AdditiveMeasurement{
				newAdditiveMeasurement(Vermiculite, 0.25, "tsp"),
				newAdditiveMeasurement(Gypsum, 0.7, "coverage of jar bottom"),
			}},
			NotesField:       NotesField{exampleNotes()},
			LastUpdatedField: LastUpdatedField{exampleTime},
		}
		return addTestAltEntries(ctx, testItem)
	})
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
	ctx, now := request.UnixTime(r.Context())
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
	toInsert := JarRecipe{
		AlternateCollectionIdField: AlternateCollectionIdField{newAlternateCollectionId()},
		NameField:                  NameField{req.Name},
		Grains:                     req.Grains,
		StandardField:              StandardField{req.Standard},
		NutrientsField:             NutrientsField{req.Nutrients},
		SugarsField:                SugarsField{req.Sugars},
		AdditivesField:             AdditivesField{req.Additives},
		NotesField:                 NotesField{req.Notes},
		LastUpdatedField:           LastUpdatedField{now},
		AclField:                   allCanWriteAcl(),
	}
	finishCreateAlternateEntry(ctx, toInsert, w)
}

type updateJarRecipeRequest struct {
	NameField
	StandardField
	NotesUpdateField
	PermsOnRequest `json:"acl"`
}

func (req updateJarRecipeRequest) modsFor(existing *JarRecipe, aclField AclField) (bson.D, error) {
	return NewMods().
		updateNameIfNeeded(req.Name, existing.Name).
		updateStandardIfNeeded(req.Standard, existing.Standard).
		updateNotesIfNeeded(req, existing).
		// TODO: for perms updates, disallow removing self???
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateJarRecipeHandler(w http.ResponseWriter, r *http.Request) {
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
	err = coll.FindOne(ctx, BsonFindFilter("_id", id)).Decode(&existing)
	if err != nil {
		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishAltCollItemUpdate(ctx, w, coll, req.modsFor, &existing, req.PermsOnRequest)
}

func deleteJarRecipeHandler(w http.ResponseWriter, r *http.Request) {
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
	// ensure batch not used by any jars first
	for _, collName := range []string{GrainBatchCollectionName, GrainJarCollectionName} {
		err = db.Collection(collName).FindOne(ctx, bson.M{"recipe": id}).Err()
		if err != nil {
			if !errors.Is(err, mongo.ErrNoDocuments) {
				http.Error(w, "failed to check for jar recipe usage in "+collName+" collection. "+err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			// At least one item exists, fail
			http.Error(w, "at least one "+collName+" utilizes the item you are attempting to delete.", http.StatusConflict) // TODO: status ok?
			return
		}
	}

	DeleteCollectionItem(ctx, JarRecipesCollectionName, id, w)
}
