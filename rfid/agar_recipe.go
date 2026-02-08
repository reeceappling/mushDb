package rfid

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
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
	AclField                   `bson:"inline"`
}

func (recipe AgarRecipe) EntryTypeField() *string {
	return nil
}

type updateAgarRecipeRequest struct {
	NameField
	StandardField
	Notes AllEntries[Note] `json:"notes"`
	PermsOnRequest
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
	//println("base58 reconverted: ", id.asBase58()) // TODO; DEL
	ctx, db := Db(r)
	coll := db.Collection(AgarRecipesCollectionName)

	existing, err := GetAltCollectionItem(ctx, id, &AgarRecipe{}) // TODO: fix this elsewhere? Should this be a pointer?
	if err != nil {
		stat := http.StatusInternalServerError
		if err == mongo.ErrNoDocuments {
			stat = http.StatusNotFound
		}
		dbErr(w, err.Error(), stat)
		return
	}
	finishAltCollItemUpdate(ctx, w, coll, req.modsFor, existing, req.PermsOnRequest)
}

func initializeAgarRecipes(ctx context.Context) error {
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(AgarRecipesCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		newSimpleIndex("name", "name", false, false, false), // TODO: names unique?
		//newSimpleIndex("liquids", "liquids.name", false, false, false),
		//newSimpleIndex("agar", "agar", true, false, false),
		standardIndexModel,
		//newSimpleIndex("nutrients", "nutrients.nutrient", false, false, false),
		//newSimpleIndex("sugars", "sugars.type", false, false, false),
		//newSimpleIndex("additives", "additives.additive", false, false, false),
		//newSimpleIndex("antibiotics", "antibiotics", false, false, false),
		//Notes (not indexed, unless tags)
		projectsIndexModel,
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	basicEntries := []*AgarRecipe{
		{
			AlternateCollectionIdField: AlternateCollectionIdField{altCollIdForint(idLmea)},
			NameField:                  NameField{"LMEA - Light Malt Extract Agar"},
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
			AlternateCollectionIdField: AlternateCollectionIdField{altCollIdForint(idPda)},
			NameField:                  NameField{"PDA - Potato Dextrose Agar"},
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
			AlternateCollectionIdField: AlternateCollectionIdField{altCollIdForint(idWaterAgar)},
			NameField:                  NameField{"Water Agar"},
			LiquidsField:               LiquidsField{[]liquid{Water.AsLiquid()}},
			Agar:                       20,
			NutrientsField:             NutrientsField{},
			SugarsField:                SugarsField{},
			StandardField:              StandardField{true},
			NotesField:                 NotesField{},
			AclField:                   allCanReadAcl(),
		},
		{
			AlternateCollectionIdField: AlternateCollectionIdField{altCollIdForint(idGrainWaterAgar)},
			NameField:                  NameField{"Grain Water Agar"},
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
			AlternateCollectionIdField: AlternateCollectionIdField{altCollIdForint(idAntibioticAgar)},
			NameField:                  NameField{"Antibiotic Agar"},
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
	}
	err = addBasicAltEntries(ctx, basicEntries...)
	if err != nil {
		return err
	}
	testItem := &AgarRecipe{
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
	// TODO: add built-in entries
	return addTestAltEntries(ctx, testItem)
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
