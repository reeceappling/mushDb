package api

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"net/url"
)

type SubspeciesOptionalField struct {
	SubSpecies *string `bson:"subspecies,omitempty" json:"subspecies,omitempty"`
}

type Subspecies struct {
	NameIdField      `bson:"inline"`
	SpeciesField     `bson:"inline"`
	AliasesField     `bson:"inline"`
	NotesField       `bson:"inline"`
	LastUpdatedField `bson:"inline"`
	AclField         `bson:"inline"`
	DefaultAcl       *ACL `bson:"defaultAcl,omitempty" json:"defaultAcl,omitempty"` // Only used when importing mainCollectionItems
}

func initializeSubspecies(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(SubspeciesCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		newSimpleIndex("species", "species", false, false, false),
		aliasesIndexModel,
		//Notes (no index) (maybe later with tags?)
		projectsIndexModel, // TODO: why are projects here?
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}

	basicEntries := []Subspecies{
		// White Beech
		{
			NameIdField:  NameIdField{"White Beech"},
			SpeciesField: SpeciesField{"Beech"},
			AliasesField: AliasesField{},
			NotesField: NotesField{Notes: []Note{
				newNote(ogTime, "something to do with light, fixme"),
			}},
			AclField: allCanWriteAcl(),
		},
		// Brown Beech
		{
			NameIdField:  NameIdField{"Brown Beech"},
			SpeciesField: SpeciesField{"Beech"},
			AliasesField: AliasesField{},
			NotesField: NotesField{Notes: []Note{
				newNote(ogTime, "something to do with light, fixme"),
			}},
			AclField: allCanWriteAcl(),
		},
	}
	err = addBasicAltEntries(ctx, basicEntries...)
	if err != nil {
		return err
	}
	// Add test entry
	testItem := &Subspecies{
		NameIdField:      NameIdField{testEntryStringId},
		SpeciesField:     SpeciesField{testEntryStringId},
		AliasesField:     AliasesField{[]string{"testSubSpecies", "example subspecies"}},
		NotesField:       NotesField{exampleNotes()},
		LastUpdatedField: LastUpdatedField{exampleTime},
		AclField:         allCanReadAcl(),
	}
	return addTestAltEntries(ctx, testItem)
}

type createSubspeciesRequest struct {
	NameField
	SpeciesField
	AliasesField
	NotesField
}

func createSubspeciesHandler(w http.ResponseWriter, r *http.Request) {
	req := createSubspeciesRequest{}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	spec, _, err := getSpeciesAndSubspecies(r.Context(), req.Species, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx, db := Db(r)
	coll := db.Collection(SubspeciesCollectionName)
	toInsert := Subspecies{
		NameIdField:      NameIdField{req.Name},
		SpeciesField:     req.SpeciesField,
		AliasesField:     req.AliasesField, // TODO: ensure none exist elsewhere (should just work via mongo)
		NotesField:       req.NotesField,
		LastUpdatedField: LastUpdatedField{unixTimeForNow()},
		AclField:         spec.AclField,   // Use parent perms
		DefaultAcl:       spec.DefaultAcl, // use parent default acl
	}
	_, err = coll.InsertOne(ctx, toInsert)
	if err != nil {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	bsOut, err := json.Marshal(toInsert)
	if err != nil {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(bsOut)
	if err != nil {
		handleWriteErr(err, w)
	}
}

type updateSubspeciesRequest struct {
	NotesUpdateField
	AliasesField
	PermsOnRequest
	DefaultEntryPermsOnRequest PermsOnRequest // TODO: handle in TS
}

func (req updateSubspeciesRequest) modsFor(existing *Subspecies, aclField AclField) (bson.D, error) {
	return NewMods().
		updateAliasesIfNeeded(req.Aliases, existing.Aliases).
		updateNotesIfNeeded(req, existing).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateDefaultEntryPermsIfNeeded(req.DefaultEntryPermsOnRequest, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateSubspeciesHandler(w http.ResponseWriter, r *http.Request) {
	urlEncodedSubspeciesName := r.PathValue("id")
	subspeciesName, err := url.QueryUnescape(urlEncodedSubspeciesName) // TODO: ensure ok
	if err != nil {
		http.Error(w, "failed to decode subspecies name from url: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req := updateSubspeciesRequest{}
	err = json.Unmarshal(bs, &req)
	if err != nil {
		http.Error(w, "failed to unmarshal body: "+err.Error(), http.StatusBadRequest)
		return
	}
	ctx, db := Db(r)
	coll := db.Collection(SubspeciesCollectionName)
	existing, err := GetSubspeciesNameInTxn(ctx, subspeciesName)
	if err != nil {
		stat := http.StatusInternalServerError
		if err == mongo.ErrNoDocuments {
			stat = http.StatusNotFound
		}
		dbErr(w, err.Error(), stat)
		return
	}
	// TODO: validate aliases are not replicas? (should be done by mongo)
	finishStringIdAltCollItemUpdate(ctx, w, coll, req.modsFor, &existing, req.PermsOnRequest) // TODO: use on species, project, user(?)
}
