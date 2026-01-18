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
	"reflect"
	"slices"
)

type AgarRecipe struct {
	AlternateCollectionIdField `bson:"inline"`
	NameField                  `bson:"inline"`
	LiquidsField               `bson:"inline"`
	Agar                       int             `bson:"agar" json:"agar"` // agar grams per 1L
	StandardField              `bson:"inline"` // If this is a standard recipe
	NutrientsField             `bson:"inline"`
	SugarsField                `bson:"inline"`
	AdditivesField             `bson:"inline"`
	AntibioticsField           `bson:"inline"`
	NotesField                 `bson:"inline"`
	LastUpdatedField           `bson:"inline"`
	AclField                   `bson:"inline"` // TODO: handle EVERYWHERE
}

func (recipe AgarRecipe) EntryTypeField() *string {
	return nil
}

//type NewAgarRecipeRequest struct {
//	NameField
//	StandardField
//	LiquidsField
//	AgarGPerL int
//	NutrientsField
//	SugarsField
//	AdditivesField
//	AntibioticsField          // TODO: make sure to put dosages in notes
//	Notes            []string // TODO: these are just strings?
//	// TODO: ACL?
//}

//func (req NewAgarRecipeRequest) asRecipe() AgarRecipe {
//	now := time.Now()
//	return AgarRecipe{
//		AlternateCollectionIdField: AlternateCollectionIdField{newAlternateCollectionId()},
//		NameField:                  req.NameField,
//		Agar:                       req.AgarGPerL,
//		LiquidsField:               req.LiquidsField,
//		NutrientsField:             req.NutrientsField,
//		SugarsField:                req.SugarsField,
//		AdditivesField:             req.AdditivesField,
//		AntibioticsField:           req.AntibioticsField,
//		StandardField:              req.StandardField,
//		NotesField:                 NotesField{stringsToNotes(req.Notes, now)},
//		LastUpdatedField:           LastUpdatedField{unixTimeFor(now)},
//		AclField:                   AclField{}, // TODO: ACL?
//	}
//}

func newAgarRecipe(ctx mongo.SessionContext, req AgarRecipe) utils.Result[AgarRecipe] {
	res, err := ctx.Client().Database(dbName).Collection(AgarRecipesCollectionName).InsertOne(ctx, req)
	if err != nil {
		return utils.ErroredResult[AgarRecipe](err)
	}
	if res == nil {
		return utils.ErroredResult[AgarRecipe](errors.New("empty response when adding recipe"))
	}
	id, ok := res.InsertedID.(AlternateCollectionId)
	if !ok {
		return utils.ErroredResult[AgarRecipe](errors.New("failed to resolve new agar recipe ID"))
	}
	req.Id = id
	return utils.SuccessfulResult(req)
}

type updateAgarRecipeRequest struct {
	NameField
	StandardField
	Notes          AllEntries[Note] `json:"notes"`
	PermsOnRequest                  // TODO: handle in typescript and handler!
}

func (req updateAgarRecipeRequest) modsFor(existing *AgarRecipe, aclField AclField) (bson.D, error) {
	return NewMods().
		updateNameIfNeeded(req.Name, existing.Name).
		updateStandardIfNeeded(req.Standard, existing.Standard).
		updateNotesIfNeeded(req.Notes, existing.Notes).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
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
	ctx, db := Db(r)
	coll := db.Collection(AgarRecipesCollectionName)

	existing, err := GetAltCollectionItem(ctx, id, AgarRecipe{})
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

func initializeAgarRecipes(ctx context.Context) error {
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(AgarRecipesCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		newSimpleIndex("name", "name", false, false, false), // TODO: names unique?
		//newSimpleIndex("liquids", "liquids.name", false, false, false),
		//newSimpleIndex("agar", "agar", true, false, false),
		standardIndexModel,
		//newSimpleIndex("nutrients", "nutrients.nutrient", false, false, false),
		//newSimpleIndex("sugars", "sugars.type", false, false, false),
		//newSimpleIndex("additives", "additives.additive", false, false, false),
		//newSimpleIndex("antibiotics", "antibiotics", false, false, false),
		//Notes (not indexed, unless tags)
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	inserted, updated := 0, 0
	for _, recipe := range []AgarRecipe{
		{
			NameField:                  NameField{"LMEA - Light Malt Extract Agar"},
			AlternateCollectionIdField: AlternateCollectionIdField{altCollIdForint(idLmea)},
			LiquidsField:               LiquidsField{[]liquid{Water.AsLiquid()}},
			Agar:                       20,
			NutrientsField: NutrientsField{[]nutrientMeasurement{
				{
					Nutrient: LME,
					Amount:   20,
					Unit:     "g",
				},
			}},
			SugarsField:   SugarsField{},
			StandardField: StandardField{true},
			NotesField:    NotesField{},
			AclField:      allCanReadAcl(),
		},
		{
			NameField:                  NameField{"PDA - Potato Dextrose Agar"},
			AlternateCollectionIdField: AlternateCollectionIdField{altCollIdForint(idPda)},
			LiquidsField:               LiquidsField{[]liquid{Water.AsLiquid()}},
			Agar:                       20,
			NutrientsField: NutrientsField{[]nutrientMeasurement{{
				Nutrient: Potato,
				Amount:   18,
				Unit:     "g",
			}}},
			SugarsField: SugarsField{[]sugarMeasurement{{
				Type:   Dextrose,
				Amount: 1,
				Unit:   "g",
			}}},
			StandardField: StandardField{true},
			NotesField:    NotesField{},
			AclField:      allCanReadAcl(),
		},
		{
			NameField:                  NameField{"Water Agar"},
			AlternateCollectionIdField: AlternateCollectionIdField{altCollIdForint(idWaterAgar)},
			LiquidsField:               LiquidsField{[]liquid{Water.AsLiquid()}},
			Agar:                       20,
			NutrientsField:             NutrientsField{},
			SugarsField:                SugarsField{},
			StandardField:              StandardField{true},
			NotesField:                 NotesField{},
			AclField:                   allCanReadAcl(),
		},
		{
			NameField:                  NameField{"Grain Water Agar"},
			AlternateCollectionIdField: AlternateCollectionIdField{altCollIdForint(idGrainWaterAgar)},
			LiquidsField: LiquidsField{[]liquid{
				GrainWater.AsLiquid().withPct(50.0),
				DistilledWater.AsLiquid().withPct(50.0),
			}},
			Agar:           20,
			NutrientsField: NutrientsField{},
			SugarsField:    SugarsField{},
			StandardField:  StandardField{true},
			NotesField: NotesField{[]Note{
				builtInNote("Grain water also acts as a nutrient source"),
			}},
			AclField: allCanReadAcl(),
		},
		{
			NameField:                  NameField{"Antibiotic Agar"},
			AlternateCollectionIdField: AlternateCollectionIdField{altCollIdForint(idAntibioticAgar)},
			LiquidsField:               LiquidsField{[]liquid{DistilledWater.AsLiquid()}},
			Agar:                       20,
			NutrientsField: NutrientsField{[]nutrientMeasurement{
				{
					Nutrient: LME,
					Amount:   20,
					Unit:     "g",
				},
			}},
			SugarsField:      SugarsField{},
			AntibioticsField: AntibioticsField{[]antibiotic{Doxycycline}},
			StandardField:    StandardField{true},
			NotesField: NotesField{[]Note{
				builtInNote("50mg doxycycline per ?????"),
			}},
			AclField: allCanReadAcl(),
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
		if len(recipe.LiquidsField.Liquids) != len(existing.LiquidsField.Liquids) {
			update = true
		} else {
			// if any liquids are different, replace all liquids
			for i, liq := range recipe.LiquidsField.Liquids {
				if liq != existing.LiquidsField.Liquids[i] {
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
			if len(recipe.NutrientsField.Nutrients) != len(existing.NutrientsField.Nutrients) {
				update = true
			} else {
				// if any nutrients are different, replace
				for i, nut := range recipe.NutrientsField.Nutrients {
					if nut != existing.NutrientsField.Nutrients[i] {
						update = true
						break
					}
				}
			}
		}

		// Sugars
		if !update {
			if len(recipe.SugarsField.Sugars) != len(existing.SugarsField.Sugars) {
				update = true
			} else {
				// if any sugars are different, replace
				for i, s := range recipe.SugarsField.Sugars {
					if s != existing.SugarsField.Sugars[i] {
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
	existingEntry := AgarRecipe{}
	testItem := AgarRecipe{
		AlternateCollectionIdField: AlternateCollectionIdField{exAltId},
		NameField:                  NameField{testEntryStringId},
		LiquidsField: LiquidsField{[]liquid{
			Water.AsLiquid().withPct(40.0),
			DistilledWater.AsLiquid().withPct(60.0),
		}},
		Agar:          20,
		StandardField: StandardField{false},
		NutrientsField: NutrientsField{[]nutrientMeasurement{
			{
				Nutrient: LME,
				Amount:   19,
				Unit:     "g",
			},
			{
				Nutrient: Potato,
				Amount:   2,
				Unit:     "g",
			},
		}},
		SugarsField: SugarsField{[]sugarMeasurement{{
			Type:   Dextrose,
			Amount: 1,
			Unit:   "g",
		}, {
			Type:   Honey,
			Amount: 2,
			Unit:   "g",
		}}},
		AdditivesField: AdditivesField{[]additiveMeasurement{
			{
				Additive: Vermiculite,
				Amount:   0.2,
				Unit:     "lb",
			},
			{
				Additive: Perlite,
				Amount:   0.7,
				Unit:     "tons",
			},
			{
				Additive: Gypsum,
				Amount:   1,
				Unit:     "pinch",
			},
		}},
		AntibioticsField: AntibioticsField{[]antibiotic{Doxycycline, HydrogenPeroxide}},
		NotesField:       NotesField{exampleNotes()},
		LastUpdatedField: LastUpdatedField{exampleTime},
		AclField:         AclField{&testAcl},
	}
	err = coll.FindOne(ctx, bson.D{{"_id", exAltId}}).Decode(&existingEntry)
	if err == nil {
		if reflect.DeepEqual(existingEntry, testItem) {
			return nil
		}
	}
	err = testExistingEntry(ctx, coll, exAltId, testItem, existingEntry)
	if inserted+updated > 0 {
		println(fmt.Sprintf(`Agar recipes: inserted %d, updated %d`, inserted, updated))
	}
	return err
}

type createAgarRecipeRequest struct {
	NameField
	StandardField
	Agar int `json:"agar"` // agar g/L
	LiquidsField
	NutrientsField
	SugarsField
	AdditivesField
	AntibioticsField // TODO: make sure to put dosages in notes?
	NotesField
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
	ctx, db := Db(r)
	coll := db.Collection(AgarRecipesCollectionName)

	toInsert := AgarRecipe{
		AlternateCollectionIdField: AlternateCollectionIdField{id},
		NameField:                  req.NameField,
		LiquidsField:               req.LiquidsField,
		Agar:                       req.Agar,
		StandardField:              req.StandardField,
		NutrientsField:             req.NutrientsField,
		SugarsField:                req.SugarsField,
		AdditivesField:             req.AdditivesField,
		AntibioticsField:           req.AntibioticsField,
		NotesField:                 req.NotesField,
		LastUpdatedField:           LastUpdatedField{unixTimeForNow()},
		AclField:                   allCanWriteAcl(),
	}
	finishCreateAlternateEntry(ctx, coll, toInsert, w)
}

// TODO: USE!
func getAgarRecipeByName(ctx context.Context, name string) (AgarRecipe, error) { // TODO: USE ME
	out := AgarRecipe{}
	err := ctx.Value(mongoClientContextKey).(*mongo.Client).
		Database(dbName).
		Collection(AgarRecipesCollectionName).
		FindOne(ctx, bson.M{"name": name}).
		Decode(&out)
	return out, err
}

type AgarRecipeField struct {
	AgarRecipe AlternateCollectionId `bson:"agarRecipe" json:"agarRecipe"` // TODO: FIX, USED TO BE Recipe and recipe
}

func (field AgarRecipeField) Get(ctx context.Context) (out AgarRecipe, err error) {
	err = ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(AgarRecipesCollectionName).FindOne(ctx, bson.M{
		"_id": field.AgarRecipe,
	}).Decode(&out)
	return out, err
}
