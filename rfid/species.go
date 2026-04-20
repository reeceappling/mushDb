package rfid

import (
	"context"
	"encoding/json"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"net/url"
)

type Species struct {
	NameIdField       `bson:"inline"` // THIS IS THE COMMON NAME
	ScientificName    string          `bson:"scientificName" json:"scientificName"`
	AliasesField      `bson:"inline"`
	StandardSubstrate AlternateCollectionId `bson:"standardSubstrate,omitempty" json:"standardSubstrate,omitempty"`
	NotesField        `bson:"inline"`
	LastUpdatedField  `bson:"inline"`
	AclField          `bson:"inline"`
	DefaultAcl        *ACL `bson:"defaultAcl,omitempty" json:"defaultAcl,omitempty"` // TODO; NEW!!! // Only used when importing!

}

func (sp Species) EntryTypeField() *string {
	return nil
}

const shiitakeName = "Shiitake"
const shiitakeSciName = "Lentinula edodes"

var shiitakeNotes = NotesField{[]Note{
	newNote(ogTime, "Colonization conditions: Hardwood sawdust with 15% bran, or pegs into a log. Indirect sun 8-12hrs. 80-90% humidity, 60-68degF, and regular FAE"),
	newNote(ogTime, "Fruiting conditions: Needs roughly 3mo incubation, damage block to encourage fruiting."),
	newNote(ogTime, "Best Agar: LMEA"),
}}

func initializeSpecies(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(SpeciesCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		newSimpleIndex("scientificName", "scientificName", false, false, true),
		aliasesIndexModel,
		newSimpleIndex("standardSubstrate", "standardSubstrate", false, true, false),
		//Notes (no index) (maybe later with tags?)
		projectsIndexModel,
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}

	woodPelletsId := altCollIdForint(idWoodPellets)
	basicEntries := []*Species{
		// King Oyster
		{
			NameIdField:       NameIdField{"king oyster"},
			ScientificName:    "Pleurotus Eryngii",
			AliasesField:      AliasesField{},
			StandardSubstrate: woodPelletsId,
			NotesField: NotesField{[]Note{
				newNote(ogTime, "Colonization conditions: FIXME"),
				newNote(ogTime, "Fruiting conditions: Prefer higher humidity than pinks FIXME"),
				newNote(ogTime, "Best Agar: LMEA"),
			}},
			AclField: allCanWriteAcl(),
		},

		// Pink Oyster
		{
			NameIdField:       NameIdField{"Pink Oyster"},
			ScientificName:    "Pleurotus Djamor",
			AliasesField:      AliasesField{},
			StandardSubstrate: woodPelletsId,
			NotesField: NotesField{[]Note{
				newNote(ogTime, "Colonization conditions: FIXME"),
				newNote(ogTime, "Fruiting conditions: Prefer higher FAE than kings FIXME"),
				newNote(ogTime, "Best Agar: LMEA"),
			}},
			AclField: allCanWriteAcl(),
		},
		// Enoki
		{
			NameIdField:       NameIdField{"Enoki"},
			ScientificName:    "Flammulina filiformis",
			AliasesField:      AliasesField{},
			StandardSubstrate: woodPelletsId,
			NotesField: NotesField{[]Note{
				newNote(ogTime, "Colonization conditions: FIXME"),
				newNote(ogTime, "Fruiting conditions: Grow in a high-CO2 environment, with the only light being high-up in the enclosure to ensure they grow tall and thin, FAE==0, humidity=70+"),
				newNote(ogTime, "Best Agar: LMEA"),
			}},
			AclField: allCanWriteAcl(),
		},
		// Shiitake
		{
			NameIdField:       NameIdField{shiitakeName},
			ScientificName:    shiitakeSciName,
			AliasesField:      AliasesField{},
			StandardSubstrate: woodPelletsId,
			NotesField:        shiitakeNotes,
			AclField:          allCanWriteAcl(),
		},
		// Maitake, Hen of the Woods
		{
			NameIdField:       NameIdField{"Maitake"},
			ScientificName:    "Grifola frondosa",
			AliasesField:      AliasesField{[]string{"hen of the woods"}},
			StandardSubstrate: woodPelletsId,
			NotesField: NotesField{[]Note{
				newNote(ogTime, "Colonization conditions: FIXME"),
				newNote(ogTime, "Fruiting conditions: 50-70degF (64-66 is ideal). >90% humidity. Cold shock to begin fruiting"),
				newNote(ogTime, "Best Agar: LMEA"),
			}},
			AclField: allCanWriteAcl(),
		},
		// Beech
		{ // TODO: WHITE AND BROWN
			NameIdField:       NameIdField{"Beech"},
			ScientificName:    "", // TODO: this
			AliasesField:      AliasesField{[]string{"hen of the woods"}},
			StandardSubstrate: woodPelletsId,
			NotesField: NotesField{[]Note{
				newNote(ogTime, "Fruiting conditions: 90-100RH, 50-60degF, plenty of light, cold shock to begin"),
				newNote(ogTime, "50-60DegF, 80-90RH, FAE"),
				newNote(ogTime, "Best Agar: LMEA"),
			}},
			AclField: allCanWriteAcl(),
		},
	}
	err = addBasicAltEntries(ctx, basicEntries...)
	if err != nil {
		return err
	}
	// Add test entry
	testItem := &Species{
		NameIdField:       NameIdField{testEntryStringId},
		ScientificName:    "examplius speciesus",
		AliasesField:      AliasesField{[]string{"testSpecies", "example species"}},
		StandardSubstrate: exAltId,
		NotesField:        NotesField{exampleNotes()},
		LastUpdatedField:  LastUpdatedField{exampleTime},
		AclField:          allCanReadAcl(), // TODO: write?
	}
	return addTestAltEntries(ctx, testItem)
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
		DefaultAcl:        allCanWriteAcl().ACL,
	}
	finishCreateAlternateEntry(ctx, coll, &toInsert, w)
}

type updateSpeciesRequest struct {
	Substrate AlternateCollectionId `json:"standardSubstrate"`
	NotesUpdateField
	AliasesField
	PermsOnRequest
	DefaultEntryPermsOnRequest PermsOnRequest // TODO: handle in TS
}

func (req updateSpeciesRequest) modsFor(existing *Species, aclField AclField) (bson.D, error) {
	return NewMods().
		UpdateValueIfNeeded("standardSubstrate", req.Substrate, existing.StandardSubstrate). // TODO: validate ok
		updateNotesIfNeeded(req, existing).
		updateAliasesIfNeeded(req.Aliases, existing.Aliases).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateDefaultEntryPermsIfNeeded(req.DefaultEntryPermsOnRequest, existing.ACL).
		updateLastUpdatedIfNeeded().
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
	finishStringIdAltCollItemUpdate(ctx, w, coll, req.modsFor, &existing, req.PermsOnRequest)
}

func getSpecies(ctx context.Context, speciesName string, subspeciesName *string) (Species, *Subspecies, error) {
	// TODO: DO THIS! ALSO ALLOW SEARCHING VIA SCIENTIFIC NAME OR ALIASES
	// TODO: maybe use a trie???
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
