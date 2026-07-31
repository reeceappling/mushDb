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
	"net/url"
)

// required for all mainCollectionItems, as well as subspecies

type Species struct {
	NameIdField       `bson:"inline"` // THIS IS THE COMMON NAME
	ScientificName    string          `bson:"scientificName" json:"scientificName"`
	AliasesField      `bson:"inline"`
	StandardSubstrate AlternateCollectionId `bson:"standardSubstrate" json:"standardSubstrate"`
	Subspecies        []string              `bson:"subspecies,omitempty" json:"subspecies,omitempty"`
	NotesField        `bson:"inline"`
	LastUpdatedField  `bson:"inline"`
	AclField          `bson:"inline"`
	DefaultAcl        ACL `bson:"defaultAcl" json:"defaultAcl"` // Only used when importing main entry types or creating a subspecies

}

const shiitakeName = "Shiitake"
const shiitakeSciName = "Lentinula Edodes"

var shiitakeNotes = NotesField{[]Note{
	newNote(ogTime, "Colonization conditions: Hardwood sawdust with 15% bran, or pegs into a log. Indirect sun 8-12hrs. 80-90% humidity, 60-68degF, and regular FAE"),
	newNote(ogTime, "Fruiting conditions: Needs roughly 3mo incubation, damage block to encourage fruiting."),
	newNote(ogTime, "Best Agar: LMEA"),
}}

const (
	TestSpeciesName  = "TestSpeciesName"
	SpeciesNameBeech = "Beech"
)

func initializeSpecies(ctx context.Context) error {
	// Indices
	coll := DbFrom(ctx).Collection(SpeciesCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		newSimpleIndex("scientificName", "scientificName", false, false, true),
		aliasesIndexModel,
		newSimpleIndex("standardSubstrate", "standardSubstrate", false, true, false),
		// Subspecies (no index)
		//Notes (no index) (maybe later with tags?)
		projectsIndexModel,
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}

	// Add test entry
	return env.IfNotProd(ctx, func() error {
		woodPelletsId := altCollIdForint(idWoodPellets)
		// TODO: ensure does not completely overwrite if there are changes....
		defaultAcl := allCanWriteAcl().ACL // TODO: ensure correct...
		basicEntries := []*Species{
			// King Oyster
			{
				NameIdField:       NameIdField{"King Oyster"},
				ScientificName:    "Pleurotus Eryngii",
				AliasesField:      AliasesField{},
				StandardSubstrate: woodPelletsId,
				Subspecies:        nil,
				NotesField: NotesField{[]Note{
					newNote(ogTime, "Colonization conditions: FIXME"),
					newNote(ogTime, "Fruiting conditions: Prefer higher humidity than pinks FIXME"),
					newNote(ogTime, "Best Agar: LMEA"),
				}},
				AclField:   allCanWriteAcl(),
				DefaultAcl: defaultAcl,
			},

			// Pink Oyster
			{
				NameIdField:       NameIdField{"Pink Oyster"},
				ScientificName:    "Pleurotus Djamor",
				AliasesField:      AliasesField{},
				StandardSubstrate: woodPelletsId,
				Subspecies:        nil,
				NotesField: NotesField{[]Note{
					newNote(ogTime, "Colonization conditions: FIXME"),
					newNote(ogTime, "Fruiting conditions: Prefer higher FAE than kings FIXME"),
					newNote(ogTime, "Best Agar: LMEA"),
				}},
				AclField:   allCanWriteAcl(),
				DefaultAcl: defaultAcl,
			},
			// Enoki
			{
				NameIdField:       NameIdField{"Enoki"},
				ScientificName:    "Flammulina filiformis",
				AliasesField:      AliasesField{},
				StandardSubstrate: woodPelletsId,
				Subspecies:        nil,
				NotesField: NotesField{[]Note{
					newNote(ogTime, "Colonization conditions: FIXME"),
					newNote(ogTime, "Fruiting conditions: Grow in a high-CO2 environment, with the only light being high-up in the enclosure to ensure they grow tall and thin, FAE==0, humidity=70+"),
					newNote(ogTime, "Best Agar: LMEA"),
				}},
				AclField:   allCanWriteAcl(),
				DefaultAcl: defaultAcl,
			},
			// Shiitake
			{
				NameIdField:       NameIdField{shiitakeName},
				ScientificName:    shiitakeSciName,
				AliasesField:      AliasesField{}, // TODO: FIX!
				StandardSubstrate: woodPelletsId,
				Subspecies:        []string{}, // TODO: FIX!
				NotesField:        shiitakeNotes,
				AclField:          allCanWriteAcl(),
				DefaultAcl:        defaultAcl,
			},
			// Maitake, Hen of the Woods
			{
				NameIdField:       NameIdField{"Maitake"},
				ScientificName:    "Grifola frondosa",
				AliasesField:      AliasesField{[]string{"hen of the woods"}},
				Subspecies:        nil,
				StandardSubstrate: woodPelletsId,
				NotesField: NotesField{[]Note{
					newNote(ogTime, "Colonization conditions: FIXME"),
					newNote(ogTime, "Fruiting conditions: 50-70degF (64-66 is ideal). >90% humidity. Cold shock to begin fruiting"),
					newNote(ogTime, "Best Agar: LMEA"),
				}},
				AclField:   allCanWriteAcl(),
				DefaultAcl: defaultAcl,
			},
			// Beech
			{
				NameIdField:       NameIdField{SpeciesNameBeech},
				ScientificName:    "Hypsizygus tessulatus",
				AliasesField:      AliasesField{[]string{"Hypsizygus tessellatus", "Shimeji"}},
				StandardSubstrate: woodPelletsId,
				Subspecies:        []string{"White Beech", "Brown Beech"},
				NotesField: NotesField{[]Note{
					newNote(ogTime, "Fruiting conditions: 90-100RH, 50-60degF, plenty of light, cold shock to begin"),
					newNote(ogTime, "50-60DegF, 80-90RH, FAE"),
					newNote(ogTime, "Best Agar: LMEA"),
					newNote(ogTime, "Can be white (patented) subspecies or brown"),
				}},
				AclField:   allCanWriteAcl(),
				DefaultAcl: defaultAcl,
			},
		}
		err = addBasicAltEntries(ctx, basicEntries...) // return here if we dont want test entries
		if err != nil {
			return err
		}

		testItem := &Species{
			NameIdField:       NameIdField{TestSpeciesName},
			ScientificName:    "examplius speciesus",
			AliasesField:      AliasesField{[]string{"testSpecies", "example species"}},
			Subspecies:        []string{}, // TODO: ADD EXAMPLE SUBSPECIES!
			StandardSubstrate: exAltId,
			NotesField:        NotesField{exampleNotes()},
			LastUpdatedField:  LastUpdatedField{exampleTime},
			AclField:          allCanReadAcl(nil),
			DefaultAcl:        allCanWriteAcl().ACL,
		}
		return addTestAltEntries(ctx, testItem)
	})
}

type createSpeciesRequest struct {
	NameField
	ScientificName string `json:"scientificName"`
	AliasesField
	SubstrateRecipeField
	NotesField
	PermsOnRequest `json:"acl"`
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
	ctx, now := request.UnixTime(r.Context())
	// Validate
	_, err = req.SubstrateRecipeField.Get(ctx)
	if err != nil {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user, _ := GetAuthInfo(ctx)
	finalAcl, err := req.PermsOnRequest.AclForUser(ctx, user)
	if err != nil {
		dbErr(w, "failed to create final ACL: "+err.Error(), http.StatusInternalServerError)
		return
	}
	toInsert := &Species{
		NameIdField:       NameIdField{req.Name},
		ScientificName:    req.ScientificName,
		AliasesField:      req.AliasesField,
		StandardSubstrate: req.Substrate,
		NotesField:        req.NotesField,
		LastUpdatedField:  LastUpdatedField{now},
		AclField:          finalAcl,
		DefaultAcl:        finalAcl.ACL,
	}
	// Validate new aliases
	ctx, db := Db(r)
	coll := db.Collection(SpeciesCollectionName) // TODO: validate working!
	if err = validateAliasesNameUnused(ctx, coll, req.Name, req.Aliases); err != nil {
		http.Error(w, "aliases or name already in use: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishCreateAlternateEntry(ctx, toInsert, w)
}

type updateSpeciesRequest struct {
	Substrate AlternateCollectionId `json:"substrate"`
	NotesUpdateField
	AliasesField
	PermsOnRequest `json:"acl"`
	DefaultAcl     PermsOnRequest // TODO: handle in TS!
}

func (req updateSpeciesRequest) modsFor(existing *Species, aclField AclField) (bson.D, error) {
	return NewMods().
		UpdateValueIfNeeded("standardSubstrate", req.Substrate, existing.StandardSubstrate). // TODO: validate ok
		updateNotesIfNeeded(req, existing).
		updateAliasesIfNeeded(req.Aliases, existing.Aliases).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateDefaultAclIfNeeded(req.DefaultAcl, existing.DefaultAcl).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateSpeciesHandler(w http.ResponseWriter, r *http.Request) {
	urlEncodedSpeciesName := r.PathValue("id")
	speciesName, err := url.QueryUnescape(urlEncodedSpeciesName)
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
	// Add user to acls as needed // TODO: ensure ok!
	user, _ := GetAuthInfo(r.Context())
	finalDefaultAcl, err := req.DefaultAcl.AclForUser(r.Context(), user)
	if err != nil {
		http.Error(w, "failed to create default acl: "+err.Error(), http.StatusInternalServerError)
		return
	}
	req.DefaultAcl = finalDefaultAcl.ACL.AsPermsOnRequest()

	ctx, db := Db(r)
	coll := db.Collection(SpeciesCollectionName)

	existing, err := GetSpeciesNameInTxn(ctx, speciesName) // TODO: get species specifically. Outside txn?
	if err != nil {
		stat := http.StatusInternalServerError
		if errors.Is(err, mongo.ErrNoDocuments) {
			stat = http.StatusNotFound
		}
		dbErr(w, err.Error(), stat)
		return
	}

	// Validate substrate exists if changed
	if existing.StandardSubstrate.AsBase58() != req.Substrate.AsBase58() {
		err = db.Collection(SubstrateRecipesCollectionName).FindOne(ctx, BsonFindFilter(IDfld, req.Substrate)).Err()
		if err != nil {
			dbErr(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	err = validateAliasesUnused(ctx, coll, existing.Name, existing.Aliases, req.Aliases)
	if err != nil {
		http.Error(w, "At least one new alias already exists as an alias or name on another entry, or there was an error querying: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishStringIdAltCollItemUpdate(ctx, w, coll, req.modsFor, &existing, req.PermsOnRequest)
}

//func getSpecies(ctx context.Context, speciesName string, subspeciesName *string) (Species, *Subspecies, error) {
//	// TODO: DO THIS! ALSO ALLOW SEARCHING VIA SCIENTIFIC NAME OR ALIASES
//	// TODO: maybe use a trie???
//	panic("not yet implemented")
//}

func getSpeciesAndSubspecies(ctx context.Context, speciesName string, subspeciesName *string) (Species, *Subspecies, error) {
	sp := Species{}
	db := DbFrom(ctx)
	err := db.Collection(SpeciesCollectionName).FindOne(ctx, BsonFindFilter(IDfld, speciesName)).Decode(&sp)
	if err != nil {
		return sp, nil, err
	}
	if subspeciesName == nil {
		return sp, nil, nil
	}
	subsp := Subspecies{}
	err = db.Collection(SubspeciesCollectionName).FindOne(ctx, BsonFindFilter(IDfld, *subspeciesName)).Decode(&subsp)
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

func (s SpeciesOptionalField) RequireNoSpecies() error {
	if s.HasSpecies() {
		return errors.New("species field should not be populated")
	}
	return nil
}
func (s SpeciesOptionalField) HasSpecies() bool {
	return s.Species != nil
}
func (s SpeciesOptionalField) Get(ctx context.Context) (*Species, error) {
	var out *Species = nil
	var err error = nil
	if s.HasSpecies() {
		err = DbFrom(ctx).Collection(SpeciesCollectionName).FindOne(ctx, BsonFindFilter(IDfld, *s.Species)).Decode(out)
	} else {
		err = ErrMissingOptionalField
	}
	return out, err
}

func deleteSpeciesHandler(w http.ResponseWriter, r *http.Request) {
	idEncoded := r.PathValue("id")
	if idEncoded == "" {
		http.Error(w, "Empty id for delete request", http.StatusBadRequest)
		return
	}
	speciesName, err := UrlDecodeString(idEncoded)
	if err != nil {
		http.Error(w, "failed to decode species name from url: "+err.Error(), http.StatusBadRequest)
		return
	}
	// Validate not used in other places...
	ctx := r.Context()
	db := DbFrom(ctx)
	// TODO: ensure to also delete subspecies???!?!?!?!?
	// ensure species not used anywhere else first
	for _, collName := range []string{BagsCollectionName, FruitsCollName, FruitingChamberCollectionName, GrainJarCollectionName, LCCollectionName, LcSyringeCollectionName, MssCollectionName, PlatesCollectionName, PlugsCollectionName, SlantsCollectionName, StasisTubeCollectionName, WaterJarsCollectionName} {
		err = db.Collection(collName).FindOne(ctx, bson.M{"species": speciesName}).Err()
		if err != nil {
			if !errors.Is(err, mongo.ErrNoDocuments) {
				http.Error(w, "failed to check for species usage in "+collName+". "+err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			// At least one item exists, fail
			http.Error(w, "at least one "+collName+" utilizes the item you are attempting to delete.", http.StatusExpectationFailed)
			return
		}
	}

	result, err := DbFrom(ctx).Collection(SpeciesCollectionName).DeleteOne(ctx, bson.M{IDfld: speciesName})
	if err != nil {
		http.Error(w, "failed to delete species "+speciesName+". "+err.Error(), http.StatusInternalServerError)
		return
	}
	if result.DeletedCount == 0 {
		http.Error(w, "species was not deleted, it was not found", http.StatusNotFound)
		return
	}
	_, err = w.Write([]byte(speciesName)) // TODO: ok?
	handleWriteErr(err, w)
}
