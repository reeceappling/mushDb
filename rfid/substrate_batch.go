package rfid

// TODO: inNewBag

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"reflect"
	sliceutils "slices"
)

type SubstrateBatch struct {
	AlternateCollectionIdField `bson:"inline"`
	// Initial wetness is quantified on each bag/box
	CreationDateField    `bson:"inline"` // Date of first hydration
	SubstrateRecipeField `bson:"inline"`
	NotesField           `bson:"inline"`
	LastUpdatedField     `bson:"inline"`
	AclField             `bson:"inline"`
}

func (batch SubstrateBatch) EntryTypeField() *string { // TODO: make these not pointers
	panic("substrate batch has no entry type field")
}

func initializeSubstrateBatches(ctx context.Context) error { // TODO: overhaul to match others
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(SubstrateBatchCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		creationDateIndexModel,
		newSimpleIndex("recipe", "recipe", false, false, true),
		//Notes (no index unless tags)
		projectsIndexModel,
		lastUpdatedIndexModel,
		//Perms
	})
	if err != nil {
		return err
	}
	inserted, updated := 0, 0
	//defaultPerms := &Perms{}           // TODO: ensure ok
	createdDate := CreationDateField{} // TODO: ensure ok
	for _, entry := range []SubstrateBatch{
		// Coir
		{
			AlternateCollectionIdField: altCollIdFieldForint(idCoir),
			CreationDateField:          createdDate,
			SubstrateRecipeField:       SubstrateRecipeField{Substrate: altCollIdForint(idCoir)},
			NotesField: NotesField{[]Note{
				newNote(ogTime, "test coir batch"),
			}},
			AclField:         allCanWriteAcl(),
			LastUpdatedField: LastUpdatedField{LastUpdated: ogTime},
		},
		// HWFP
		{
			AlternateCollectionIdField: altCollIdFieldForint(idWoodPellets),
			CreationDateField:          createdDate,
			SubstrateRecipeField:       SubstrateRecipeField{Substrate: altCollIdForint(idWoodPellets)},
			NotesField: NotesField{[]Note{
				newNote(ogTime, "test hwfp batch"),
			}},
			AclField:         allCanWriteAcl(),
			LastUpdatedField: LastUpdatedField{LastUpdated: ogTime},
		},
	} {
		var existing SubstrateBatch
		err := coll.FindOne(ctx, bson.D{{"_id", entry.Id}}).Decode(&existing)
		if err != nil {
			if err != mongo.ErrNoDocuments {
				return err
			}
			// if not exists, add it to the db
			_, err = coll.InsertOne(ctx, entry)
			if err != nil {
				return err
			}
			inserted++
			continue
		}
		// If exists, ensure it is the same as it was. Add notes if necessary
		update := false

		// Notes
		finalNotes := []Note{}
		copy(finalNotes, existing.Notes)
		for _, note := range entry.Notes {
			if !sliceutils.Contains(finalNotes, note) {
				finalNotes = append(finalNotes, note)
				update = true
			}
		}
		entry.Notes = finalNotes

		// Update if necessary
		if update {
			err = coll.FindOneAndReplace(ctx, bson.D{{"_id", entry.Id}}, entry).Err()
			if err != nil {
				return err
			}
			updated++
		}
	}
	// Add test entry
	existingEntry := SubstrateBatch{}
	testItem := SubstrateBatch{
		AlternateCollectionIdField: altCollIdFieldForint(idTestingOnly),
		CreationDateField:          CreationDateField{exampleTime},
		SubstrateRecipeField:       SubstrateRecipeField{altCollIdForint(idTestingOnly)},
		NotesField:                 NotesField{exampleNotes()},
		LastUpdatedField:           LastUpdatedField{exampleTime},
		AclField:                   allCanWriteAcl(),
	}
	err = coll.FindOne(ctx, bson.D{{"_id", exAltId}}).Decode(&existingEntry)
	if err == nil {
		if reflect.DeepEqual(existingEntry, testItem) {
			return nil
		}
	}
	err = testExistingEntry(ctx, coll, exAltId, testItem, existingEntry)
	if inserted+updated > 0 {
		println(fmt.Sprintf(`SubstrateRecipe: inserted %d, updated %d`, inserted, updated))
	}
	return err
}

type createSubstrateBatchRequest struct {
	SubstrateRecipeField
	NotesField
}

func createSubstrateBatchHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	req := createSubstrateBatchRequest{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	id := newAlternateCollectionId()

	ctx, db := Db(r)
	coll := db.Collection(SubstrateRecipesCollectionName)
	// TODO: all can read but only user can write perms
	//resolvedUserPerms, err := GetAuthInfo(r.Context())
	//if err != nil {
	//	http.Error(w, "Failed to get auth info", http.StatusUnauthorized)
	//	return
	//}
	//acl, err := newAlwaysReadableAcl(ctx, resolvedUserPerms, nil, nil)
	//if err != nil {
	//	return dbErr(w, "failed to resolve new ACL: "+err.Error(), http.StatusInternalServerError)
	//}
	toInsert := SubstrateBatch{
		AlternateCollectionIdField: AlternateCollectionIdField{id},
		CreationDateField:          CreationDateField{CreationDate: unixTimeForNow()},
		SubstrateRecipeField:       SubstrateRecipeField{Substrate: req.Substrate}, // TODO: validate
		NotesField:                 req.NotesField,
		LastUpdatedField:           LastUpdatedFieldForNow(),
		AclField:                   allCanWriteAcl(), // TODO: all can read but only user can write perms
	}
	// Validate
	_, err = toInsert.SubstrateRecipeField.Get(ctx)
	if err != nil {
		http.Error(w, "failed to validate substrate recipe: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishCreateAlternateEntry(ctx, coll, &toInsert, w)
}

type updateSubstrateBatchRequest struct {
	NotesUpdateField
	PermsOnRequest
}

func (req updateSubstrateBatchRequest) modsFor(existing *SubstrateBatch, aclField AclField) (bson.D, error) {
	return NewMods().
		updateNotesIfNeeded(req, existing).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateSubstrateBatchHandler(w http.ResponseWriter, r *http.Request) {
	b58Id := Base58Str(r.PathValue("id"))
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req := updateSubstrateBatchRequest{}
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
	coll := db.Collection(SubstrateBatchCollectionName)
	existing, err := GetAltCollectionItemOutsideTxn(ctx, id, SubstrateBatch{})
	if err != nil {
		stat := http.StatusInternalServerError
		if errors.Is(err, mongo.ErrNoDocuments) {
			stat = http.StatusNotFound
		}
		dbErr(w, err.Error(), stat)
		return
	}
	finishAltCollItemUpdate(ctx, w, coll, req.modsFor, &existing, req.PermsOnRequest)
}

type SubstrateBatchField struct {
	SubstrateBatch AlternateCollectionId `bson:"substrateBatch" json:"substrateBatch"`
}

func (field SubstrateBatchField) Get(ctx context.Context) (out SubstrateBatch, err error) {
	err = ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(SubstrateBatchCollectionName).FindOne(ctx, bson.M{
		"_id": field.SubstrateBatch,
	}).Decode(&out)
	return out, err
}

func (field SubstrateBatchField) asOptional() SubstrateBatchOptionalField {
	return SubstrateBatchOptionalField{&field.SubstrateBatch}
}

type SubstrateBatchOptionalField struct {
	SubstrateBatch *AlternateCollectionId `bson:"substrateBatch,omitempty" json:"substrateBatch,omitempty"`
}
