package rfid

import (
	"context"
	"encoding/json"
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

const subSpeciesCollectionName = "subspecies"

type SubspeciesOptionalField struct {
	SubSpecies *string `bson:"subSpecies,omitempty" json:"subSpecies,omitempty"`
}

type Subspecies struct {
	NameIdField
	SpeciesField
	AliasesField
	NotesField
	LastUpdatedField
	AclField // TODO: handle EVERYWHERE
}

func (subsp Subspecies) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := subsp
	err := decodeItem(&out, encoded)
	return out, err
}

func (subsp Subspecies) EntryTypeField() *string {
	return nil
}

func (subsp Subspecies) CollectionName() string {
	return subSpeciesCollectionName
}

func initializeSubspecies(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(subSpeciesCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		newSimpleIndex("species", "species", false, false, false),
		aliasesIndexModel,
		//Notes (no index) (maybe later with tags?)
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}

	inserted, updated := 0, 0
	for _, subsp := range []Subspecies{
		// White Beech
		{
			NameIdField:  NameIdField{"white beech"},
			SpeciesField: SpeciesField{"beech"},
			AliasesField: AliasesField{},
			NotesField:   NotesField{Notes: []Note{{Time: ogTime, Note: "something to do with light, fixme"}}}, // TODO: something to do with light?
		},
		// Brown Beech
		{
			NameIdField:  NameIdField{"brown beech"},
			SpeciesField: SpeciesField{"beech"},
			AliasesField: AliasesField{},
			NotesField:   NotesField{Notes: []Note{{Time: ogTime, Note: "something to do with light, fixme"}}}, // TODO: something to do with light?
		},
	} {
		var existing Subspecies
		err = coll.FindOne(ctx, bson.D{{"_id", subsp.Name}}).Decode(&existing)
		if err != nil {
			if err != mongo.ErrNoDocuments {
				return err
			}
			// if not exists, add it to the db
			_, err = coll.InsertOne(ctx, subsp)
			if err != nil {
				return err
			}
			inserted++
			continue
		}
		// If exists, ensure it is the same as it was. Add notes if necessary
		update := false
		if subsp.Species != subsp.Species {
			update = true
		}
		finalAliases := utils.Set[string]{}
		finalAliases.Add(existing.Aliases...)
		finalAliases.Add(subsp.Aliases...)
		if len(finalAliases) != len(existing.Aliases) {
			update = true
			subsp.Aliases = finalAliases.ToSlice()
		}

		// Notes
		finalNotes := []Note{}
		copy(finalNotes, existing.Notes)
		for _, note := range subsp.Notes {
			if !sliceutils.Contains(finalNotes, note) {
				finalNotes = append(finalNotes, note)
				update = true
			}
		}
		subsp.Notes = finalNotes

		// Update if necessary
		if update {
			err = coll.FindOneAndReplace(ctx, bson.D{{"_id", subsp.Name}}, subsp).Err()
			if err != nil {
				return err
			}
			updated++
		}
	}
	// Add test entry
	existingEntry := Subspecies{}
	testItem := Subspecies{
		NameIdField:      NameIdField{testEntryStringId},
		SpeciesField:     SpeciesField{testEntryStringId},
		AliasesField:     AliasesField{[]string{"testSubSpecies", "example subspecies"}},
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
	if inserted+updated > 0 {
		println(fmt.Sprintf(`Subspecies: inserted %d, updated %d`, inserted, updated))
	}
	return err
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

	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(subSpeciesCollectionName)
		toInsert := Subspecies{
			NameIdField:      NameIdField{req.Name},
			SpeciesField:     req.SpeciesField,
			AliasesField:     req.AliasesField, // TODO: ensure none exist elsewhere
			NotesField:       req.NotesField,
			LastUpdatedField: LastUpdatedField{unixTimeForNow()},
			AclField:         spec.AclField, // Use parent perms
		}

		_, err = coll.InsertOne(r.Context(), toInsert)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		bsOut, err := json.Marshal(toInsert)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		return w.Write(bsOut)
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}

type updateSubspeciesRequest struct {
	Notes AllEntries[Note] `json:"notes,omitempty"`
	AliasesField
	PermsOnRequest // TODO: handle in typescript and handler!
}

func updateSubspeciesHandler(w http.ResponseWriter, r *http.Request) {
	urlEncodedSpeciesName := r.PathValue("id")
	speciesName, err := url.QueryUnescape(urlEncodedSpeciesName) // TODO: ensure ok
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
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(subSpeciesCollectionName)
		existing, err := GetSpeciesNameInTxn(ctx, speciesName) // TODO: get species specifically
		if err != nil {
			stat := http.StatusInternalServerError
			if err == mongo.ErrNoDocuments {
				stat = http.StatusNotFound
			}
			return DbTxnStdErr(w, err.Error(), stat)
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
		//if err = minimalPermsBetween(existing.Perms, req.Perms).ValidateUserCanWrite(ctx); err != nil { // TODO: PUT PERMS UPDATER ON THE STRUCTS?
		//	return DbTxnStdErr(w, "bad overlapping perms for user: "+err.Error(), http.StatusUnauthorized)
		//}
		upd, err := NewMods().
			updateAliasesIfNeeded(req.Aliases, existing.Aliases).
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
		bsonId := bson.D{{"_id", speciesName}}
		err = coll.FindOneAndUpdate(ctx, bsonId, upd).Err()
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		err = coll.FindOne(ctx, bsonId).Decode(&existing)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		bsOut, err := json.Marshal(existing)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		return w.Write(bsOut)
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}
