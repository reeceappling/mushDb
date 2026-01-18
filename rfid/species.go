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
	"net/url"
	"reflect"
	sliceutils "slices"
)

type Species struct {
	NameIdField       `bson:"inline"` // THIS IS THE COMMON NAME
	ScientificName    string          `bson:"scientificName" json:"scientificName"`
	AliasesField      `bson:"inline"`
	StandardSubstrate AlternateCollectionId `bson:"standardSubstrate,omitempty" json:"standardSubstrate,omitempty"`
	NotesField        `bson:"inline"`
	LastUpdatedField  `bson:"inline"`
	AclField          `bson:"inline"` // TODO: handle EVERYWHERE
}

func (sp Species) EntryTypeField() *string {
	return nil
}

const shiitakeName = "Shiitake"
const shiitakeSciName = "Lentinula edodes"

var shiitakeNotes = NotesField{[]Note{{
	Time: ogTime,
	Note: "Colonization conditions: Hardwood sawdust with 15% bran, or pegs into a log. Indirect sun 8-12hrs. 80-90% humidity, 60-68degF, and regular FAE",
},
	{
		Time: ogTime,
		Note: "Fruiting conditions: Needs roughly 3mo incubation, damage block to encourage fruiting.",
	},
	{
		Time: ogTime,
		Note: "Best Agar: LMEA",
	},
}}

func initializeSpecies(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(SpeciesCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		newSimpleIndex("scientificName", "scientificName", false, false, true),
		aliasesIndexModel,
		//BackupSpecies (likely no index)      *string           `bson:"backupSpecies,omitempty" json:"backupSpecies,omitempty"` // a species that is similar to this one, somehow
		newSimpleIndex("standardSubstrate", "standardSubstrate", false, true, false),
		//Notes (no index) (maybe later with tags?)
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}

	inserted, updated := 0, 0
	woodPelletsId := altCollIdForint(idWoodPellets)
	for _, species := range []Species{
		// King Oyester
		{
			NameIdField:       NameIdField{"king oyster"},
			ScientificName:    "Pleurotus Eryngii",
			AliasesField:      AliasesField{},
			StandardSubstrate: woodPelletsId,
			NotesField: NotesField{[]Note{
				{
					Time: ogTime,
					Note: "Colonization conditions: FIXME",
				},
				{
					Time: ogTime,
					Note: "Fruiting conditions: Prefer higher humidity than pinks FIXME",
				},
				{
					Time: ogTime,
					Note: "Best Agar: LMEA",
				}},
			}},
		// Pink Oyster
		{
			NameIdField:       NameIdField{"Pink Oyster"},
			ScientificName:    "Pleurotus Djamor",
			AliasesField:      AliasesField{},
			StandardSubstrate: woodPelletsId,
			NotesField: NotesField{[]Note{{
				Time: ogTime,
				Note: "Colonization conditions: FIXME",
			},
				{
					Time: ogTime,
					Note: "Fruiting conditions: Prefer higher FAE than kings FIXME",
				},
				{
					Time: ogTime,
					Note: "Best Agar: LMEA",
				}},
			}},
		// Enoki
		{
			NameIdField:       NameIdField{"Enoki"},
			ScientificName:    "Flammulina filiformis",
			AliasesField:      AliasesField{},
			StandardSubstrate: woodPelletsId,
			NotesField: NotesField{[]Note{{
				Time: ogTime,
				Note: "Colonization conditions: FIXME",
			},
				{
					Time: ogTime,
					Note: "Fruiting conditions: Grow in a high-CO2 environment, with the only light being high-up in the enclosure to ensure they grow tall and thin, FAE==0, humidity=70+",
				},
				{
					Time: ogTime,
					Note: "Best Agar: LMEA",
				}},
			}},
		// Shiitake
		{
			NameIdField:       NameIdField{shiitakeName},
			ScientificName:    shiitakeSciName,
			AliasesField:      AliasesField{},
			StandardSubstrate: woodPelletsId,
			NotesField:        shiitakeNotes,
		},
		// Maitake, Hen of the Woods
		{
			NameIdField:       NameIdField{"Maitake"},
			ScientificName:    "Grifola frondosa",
			AliasesField:      AliasesField{[]string{"hen of the woods"}},
			StandardSubstrate: woodPelletsId,
			NotesField: NotesField{[]Note{{
				Time: ogTime,
				Note: "Colonization conditions: FIXME",
			},
				{
					Time: ogTime,
					Note: "Fruiting conditions: 50-70degF (64-66 is ideal). >90% humidity. Cold shock to begin fruiting ",
				},
				{
					Time: ogTime,
					Note: "Resistant to high temps",
				},
				{
					Time: ogTime,
					Note: "Best Agar: LMEA",
				}},
			}},
		// Beech
		{ // TODO: WHITE AND BROWN
			NameIdField:       NameIdField{"Beech"},
			ScientificName:    "", // TODO: this
			AliasesField:      AliasesField{[]string{"hen of the woods"}},
			StandardSubstrate: woodPelletsId,
			NotesField: NotesField{[]Note{
				{
					Time: ogTime,
					Note: "Fruiting conditions: 90-100RH, 50-60degF, plenty of light, cold shock to begin",
				},
				{
					Time: ogTime,
					Note: "50-60DegF, 80-90RH, FAE",
				},
				{
					Time: ogTime,
					Note: "Best Agar: LMEA",
				},
			}},
		},
	} {
		var existing Species
		err = coll.FindOne(ctx, bson.D{{"_id", species.Name}}).Decode(&existing)
		if err != nil {
			if err != mongo.ErrNoDocuments {
				return err
			}
			// if not exists, add it to the db
			_, err = coll.InsertOne(ctx, species)
			if err != nil {
				return err
			}
			inserted++
			continue
		}
		// If exists, ensure it is the same as it was. Add notes if necessary
		update := false
		if species.ScientificName != existing.ScientificName {
			update = true
		}
		finalAliases := utils.Set[string]{}
		finalAliases.Add(existing.Aliases...)
		finalAliases.Add(species.Aliases...)
		if len(finalAliases) != len(existing.Aliases) {
			update = true
			species.Aliases = finalAliases.ToSlice()
		}
		if !update && existing.StandardSubstrate != species.StandardSubstrate {
			update = true
		}

		// Notes
		finalNotes := []Note{}
		copy(finalNotes, existing.Notes)
		for _, note := range species.Notes {
			if !sliceutils.Contains(finalNotes, note) {
				finalNotes = append(finalNotes, note)
				update = true
			}
		}
		species.Notes = finalNotes

		// Update if necessary
		if update {
			err = coll.FindOneAndReplace(ctx, bson.D{{"_id", species.Name}}, species).Err()
			if err != nil {
				return err
			}
			updated++
		}
	}
	// Add test entry
	existingEntry := Species{}
	testItem := Species{
		NameIdField:       NameIdField{testEntryStringId},
		ScientificName:    "examplius speciesus",
		AliasesField:      AliasesField{[]string{"testSpecies", "example species"}},
		StandardSubstrate: exAltId,
		NotesField:        NotesField{exampleNotes()},
		LastUpdatedField:  LastUpdatedField{exampleTime},
	}
	err = coll.FindOne(ctx, bson.D{{"_id", exAltId}}).Decode(&existingEntry)
	if err == nil {
		if reflect.DeepEqual(existingEntry, testItem) {
			return nil
		}
	}
	err = testExistingEntry(ctx, coll, exAltId, testItem, existingEntry)
	if inserted+updated > 0 {
		println(fmt.Sprintf(`Species: inserted %d, updated %d`, inserted, updated))
	}
	return err
}

type createSpeciesRequest struct {
	NameField
	ScientificName string `json:"scientificName"`
	AliasesField
	SubstrateRecipeField // TODO: tag used to be "sub" is now "recipe"
	NotesField
}

func createSpeciesHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	req := createSpeciesRequest{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	//if err = req.Perms.ValidateUserCanWrite(r.Context()); err != nil {
	//	http.Error(w, "can not write with provided perms: "+err.Error(), http.StatusBadRequest)
	//	return
	//}
	ctx, db := Db(r)
	coll := db.Collection(SpeciesCollectionName)
	// Validate
	// TODO: Aliases?
	err = db.Collection(SubstrateRecipesCollectionName).FindOne(ctx, bson.D{{"_id", req.Substrate}}).Err()
	if err != nil {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	toInsert := Species{
		NameIdField:       NameIdField{req.Name},
		ScientificName:    req.ScientificName,
		AliasesField:      req.AliasesField,
		StandardSubstrate: req.Substrate,
		NotesField:        req.NotesField,
		LastUpdatedField:  LastUpdatedField{unixTimeForNow()},
	}
	finishCreateAlternateEntry(ctx, coll, &toInsert, w)
}

type updateSpeciesRequest struct {
	Substrate AlternateCollectionId `json:"standardSubstrate"`
	Notes     AllEntries[Note]      `json:"notes,omitempty"`
	AliasesField
	PermsOnRequest // TODO: handle in typescript and handler!
}

func (out updateSpeciesRequest) modsFor(existing *Species, aclField AclField) (bson.D, error) {
	return NewMods().
		UpdateValueIfNeeded("standardSubstrate", out.Substrate, existing.StandardSubstrate). // TODO: validate ok
		updateNotesIfNeeded(out.Notes, existing.Notes).
		updateAliasesIfNeeded(out.Aliases, existing.Aliases).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		Finalized()
}

func updateSpeciesHandler(w http.ResponseWriter, r *http.Request) {
	urlEncodedSpeciesName := r.PathValue("id")
	speciesName, err := url.QueryUnescape(urlEncodedSpeciesName) // TODO: make sure we are doing this right!
	if err != nil {
		http.Error(w, "failed to decode species name from url: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req := updateSpeciesRequest{}
	err = json.Unmarshal(bs, &req)
	if err != nil {
		http.Error(w, "failed to unmarshal body: "+err.Error(), http.StatusBadRequest)
		return
	}

	ctx, db := Db(r)
	coll := db.Collection(SpeciesCollectionName)

	// TODO: ensure next line works
	existing, err := GetSpeciesNameInTxn(ctx, speciesName) // TODO: get species specifically
	if err != nil {
		stat := http.StatusInternalServerError
		if errors.Is(err, mongo.ErrNoDocuments) {
			stat = http.StatusNotFound
		}
		dbErr(w, err.Error(), stat)
		return
	}

	// Validate substrate exists
	if req.Substrate.asBase58() != existing.StandardSubstrate.asBase58() {
		err = db.Collection(SubstrateRecipesCollectionName).FindOne(ctx, bson.D{{"_id", req.Substrate}}).Err()
		if err != nil {
			dbErr(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	// TODO: FIX!!!!!
	finishAltCollItemUpdate(ctx, w, coll, req.modsFor, &existing, req.PermsOnRequest)
}

func getSpecies(ctx context.Context, speciesName string, subspeciesName *string) (Species, *Subspecies, error) {
	// TODO: DO THIS! ALSO ALLOW SEARCHING VIA SCIENTIFIC NAME OR ALIASES
	panic("not yet implemented")
}

func getSpeciesAndSubspecies(ctx context.Context, speciesName string, subspeciesName *string) (Species, *Subspecies, error) {
	sp := Species{}
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	err := db.Collection(SpeciesCollectionName).FindOne(ctx, bson.D{{"_id", speciesName}}).Decode(&sp)
	if err != nil {
		return sp, nil, err
	}
	if subspeciesName == nil {
		return sp, nil, nil
	}
	subsp := Subspecies{}
	err = db.Collection(SubspeciesCollectionName).FindOne(ctx, bson.D{{"_id", *subspeciesName}}).Decode(&subsp)
	if err != nil {
		return sp, nil, err
	}
	return sp, &subsp, nil
}

type SpeciesField struct {
	Species string `bson:"species" json:"species"`
}

func (field SpeciesField) AsOptional() SpeciesOptionalField {
	val := field.Species
	return SpeciesOptionalField{&val}
}

type SpeciesOptionalField struct {
	Species *string `bson:"species,omitempty" json:"species,omitempty"`
}
