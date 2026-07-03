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
	"slices"
)

type SubspeciesOptionalField struct {
	Subspecies *string `bson:"subspecies,omitempty" json:"subspecies,omitempty"`
}

func (s SubspeciesOptionalField) RequireNoSubspecies() error {
	if s.Subspecies != nil {
		return errors.New("subspecies field should not be populated")
	}
	return nil
}

type Subspecies struct {
	NameIdField      `bson:"inline"`
	SpeciesField     `bson:"inline"`
	AliasesField     `bson:"inline"`
	NotesField       `bson:"inline"`
	LastUpdatedField `bson:"inline"`
	AclField         `bson:"inline"`
	DefaultAcl       ACL `bson:"defaultAcl" json:"defaultAcl"` // Only used when importing mainCollectionItems
}

const TestSubspeciesName = "TestSubspecies"

func initializeSubspecies(ctx context.Context) error {
	// Indices
	coll := DbFrom(ctx).Collection(SubspeciesCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		newSimpleIndex("species", "species", false, false, false),
		aliasesIndexModel,
		//Notes (no index) (maybe later with tags?)
		//projectsIndexModel, // TODO: why are projects here? probably dont want this
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
			AliasesField: AliasesField{[]string{"Bunapi-shimeji"}},
			NotesField: NotesField{Notes: []Note{
				newNote(ogTime, "something to do with light, fixme"),
			}},
			AclField:   allCanWriteAcl(),
			DefaultAcl: allCanWriteAcl().ACL,
		},
		// Brown Beech
		{
			NameIdField:  NameIdField{"Brown Beech"},
			SpeciesField: SpeciesField{"Beech"},
			AliasesField: AliasesField{[]string{"Buna-shimeji"}},
			NotesField: NotesField{Notes: []Note{
				newNote(ogTime, "something to do with light, fixme"),
			}},
			AclField:   allCanWriteAcl(),
			DefaultAcl: allCanWriteAcl().ACL,
		},
	}
	err = addBasicAltEntries(ctx, basicEntries...)
	if err != nil {
		return err
	}
	return env.IfNotProd(ctx, func() error { // TODO: ensure ok
		// Add test entry
		testItem := &Subspecies{
			NameIdField:      NameIdField{TestSubspeciesName},
			SpeciesField:     SpeciesField{TestSpeciesName},
			AliasesField:     AliasesField{[]string{"testSubSpecies", "example subspecies"}},
			NotesField:       NotesField{exampleNotes()},
			LastUpdatedField: LastUpdatedField{exampleTime},
			AclField:         allCanReadAcl(nil),
			DefaultAcl:       allCanWriteAcl().ACL,
		}
		return addTestAltEntries(ctx, testItem)
	})
}

type createSubspeciesRequest struct {
	NameField
	SpeciesField
	AliasesField
	NotesField
	// ACL/DefaultACL are initially inherited from parent species
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
	ctx, now := request.UnixTime(r.Context())
	toInsert := Subspecies{
		NameIdField:      NameIdField{req.Name},
		SpeciesField:     req.SpeciesField,
		AliasesField:     req.AliasesField, // TODO: ensure none exist elsewhere (should just work via mongo)
		NotesField:       req.NotesField,
		LastUpdatedField: LastUpdatedField{now},
		AclField:         spec.AclField,   // Use parent perms
		DefaultAcl:       spec.DefaultAcl, // use parent default acl
	}
	// Create species update modifications
	var tempSpecUpd *Mods
	var speciesUpdate bson.D
	if spec.Subspecies == nil {
		tempSpecUpd = NewMods().Set("subspecies", []string{req.Name})
	} else {
		if slices.Contains(spec.Subspecies, req.Name) {
			// Name already exists, do nothing
			http.Error(w, "subspecies already exists on species", http.StatusBadRequest)
			return
		}
		tempSpecUpd = NewMods().Push("subspecies", req.Name)
	}
	speciesUpdate, err = tempSpecUpd.withLastUpdated(now).Finalized()
	if err != nil {
		http.Error(w, "failed to create species update: "+err.Error(), http.StatusBadRequest)
		return
	}
	if len(speciesUpdate) == 0 {
		http.Error(w, "no changes made to species, should be impossible!", http.StatusInternalServerError)
		return
	}
	// Do DB stuff
	_, err = newTxn(ctx, func(ctx mongo.SessionContext) (any, error) {
		db := mongo.SessionFromContext(ctx).Client().Database(dbName)
		// Update species with subspecies

		bsonId := BsonFindFilter("_id", spec.Name)
		err = db.Collection(SpeciesCollectionName).FindOneAndUpdate(ctx, bsonId, speciesUpdate).Err()
		if err != nil {
			dbErr(w, "failed to write update to db: "+err.Error(), http.StatusInternalServerError)
			return nil, err
		}
		// Insert subspecies
		if _, e := db.Collection(SubspeciesCollectionName).InsertOne(ctx, toInsert); e != nil {
			dbErr(w, e.Error(), http.StatusInternalServerError)
			return nil, e
		}
		return nil, nil
	})
	if err != nil {
		// Already wrote in all cases except success
		println("failed in txn: " + err.Error())
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
	PermsOnRequest `json:"acl"`
	DefaultAcl     PermsOnRequest
}

func (req updateSubspeciesRequest) modsFor(existing *Subspecies, aclField AclField) (bson.D, error) {
	return NewMods().
		updateAliasesIfNeeded(req.Aliases, existing.Aliases).
		updateNotesIfNeeded(req, existing).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateDefaultAclIfNeeded(req.DefaultAcl, existing.DefaultAcl).
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
	// TODO: ensure ok! updating perms to allow for creating user as well...
	user, _ := GetAuthInfo(r.Context())
	finalDefaultAcl, err := req.DefaultAcl.AclForUser(r.Context(), user)
	if err != nil {
		http.Error(w, "failed to create default acl: "+err.Error(), http.StatusInternalServerError)
		return
	}
	req.DefaultAcl = finalDefaultAcl.ACL.AsPermsOnRequest()

	ctx, db := Db(r)
	coll := db.Collection(SubspeciesCollectionName)
	existing, err := GetSubspeciesByNameInTxn(ctx, subspeciesName) // TODO; does this need to be in txn?
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
