package rfid

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

const (
	substrateBatchCollectionName = "substrateBatches"
	substrateBatchType           = "substrateBatch"
)

type SubstrateBatch struct { // TODO: use this
	AlternateCollectionIdField
	// Initial wetness is quantified on each bag/box
	CreationDateField // Date of first hydration
	SubstrateRecipeField
	NotesField
	LastUpdatedField
	AclField // TODO: handle EVERYWHERE
}

func (recipe SubstrateBatch) CollectionName() string {
	return substrateBatchCollectionName
}

func (recipe SubstrateBatch) EntryTypeField() *string { // TODO: make these not pointers
	panic("substrate batch has no entry type field")
}

func (recipe SubstrateBatch) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := &SubstrateBatch{}
	err := decodeItem(out, encoded)
	return *out, err
}

// TODO; USE!
func initializeSubstrateBatches(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(substrateBatchCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		creationDateIndexModel,
		newSimpleIndex("recipe", "recipe", false, false, true),
		//Notes (no index unless tags)
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
				{
					Time: ogTime,
					Note: "test coir batch",
				},
			}},
			//PermsField:       PermsField{Perms: defaultPerms},
			LastUpdatedField: LastUpdatedField{LastUpdated: ogTime}, // TODO: ok?
		},
		// HWFP
		{
			AlternateCollectionIdField: altCollIdFieldForint(idWoodPellets),
			CreationDateField:          createdDate,
			SubstrateRecipeField:       SubstrateRecipeField{Substrate: altCollIdForint(idWoodPellets)},
			NotesField: NotesField{[]Note{
				{
					Time: ogTime,
					Note: "test hwfp batch",
				},
			}},
			//PermsField:       PermsField{Perms: defaultPerms},
			LastUpdatedField: LastUpdatedField{LastUpdated: ogTime}, // TODO: ok?
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
	// Add test entry // TODO: is this necessary?
	existingEntry := SubstrateBatch{}
	testItem := SubstrateBatch{
		AlternateCollectionIdField: altCollIdFieldForint(idTestingOnly),
		CreationDateField:          CreationDateField{exampleTime},
		SubstrateRecipeField:       SubstrateRecipeField{altCollIdForint(idTestingOnly)},
		NotesField:                 NotesField{exampleNotes()},
		LastUpdatedField:           LastUpdatedField{exampleTime},
		//PermsField:                 PermsField{},
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
	resolvedUserPerms, err := GetAuthInfo(r.Context())
	if err != nil {
		http.Error(w, "Failed to get auth info", http.StatusUnauthorized)
		return
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		db := ctx.Client().Database(dbName)
		err = db.Collection(substrateRecipesCollectionName).FindOne(ctx, bson.D{{"_id", req.Substrate}}).Err()
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return DbTxnStdErr(w, "recipe not found", http.StatusNotFound)
			}
			return DbTxnStdErr(w, "error getting substrate recipe: "+err.Error(), http.StatusInternalServerError)
		}
		acl, err := newAlwaysReadableAcl(ctx, resolvedUserPerms, nil, nil)
		if err != nil {
			return DbTxnStdErr(w, "failed to resolve new ACL: "+err.Error(), http.StatusInternalServerError)
		}
		toInsert := SubstrateBatch{
			AlternateCollectionIdField: AlternateCollectionIdField{id},
			CreationDateField:          CreationDateField{CreationDate: unixTimeForNow()},
			SubstrateRecipeField:       SubstrateRecipeField{Substrate: req.Substrate}, // TODO: validate
			NotesField:                 req.NotesField,
			LastUpdatedField:           LastUpdatedFieldForNow(),
			AclField:                   acl,
		}
		_, err = db.Collection(substrateBatchCollectionName).InsertOne(r.Context(), toInsert)
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

type updateSubstrateBatchRequest struct {
	Notes          AllEntries[Note] `json:"notes"`
	PermsOnRequest                  // TODO: handle in typescript and handler!
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
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(substrateBatchCollectionName)
		existing, err := GetAltCollectionItemInTxn(ctx, id, SubstrateBatch{})
		if err != nil {
			stat := http.StatusInternalServerError
			if errors.Is(err, mongo.ErrNoDocuments) {
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
		upd, err := NewMods().
			updateNotesIfNeeded(req.Notes, existing.Notes).
			updatePermsIfNeeded(aclField.ACL, existing.ACL).
			updateLastUpdatedIfNeeded().
			Finalized()
		if err != nil {
			return DbTxnStdErr(w, "error resolving updates list: "+err.Error(), http.StatusInternalServerError)
		}
		bsonId := bson.D{{"_id", existing.Id}}
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

type SubstrateBatchField struct {
	SubstrateBatch AlternateCollectionId `bson:"substrateBatch" json:"substrateBatch"`
}

func (field SubstrateBatchField) Get(ctx context.Context) (out SubstrateBatch, err error) {
	err = ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(substrateBatchCollectionName).FindOne(ctx, bson.M{
		"_id": field.SubstrateBatch,
	}).Decode(&out)
	return out, err
}

func (field SubstrateBatchField) asOptional() SubstrateBatchOptionalField {
	return SubstrateBatchOptionalField{&field.SubstrateBatch}
}

type SubstrateBatchOptionalField struct { // TODO: MOVE
	SubstrateBatch *AlternateCollectionId `bson:"substrateBatch,omitempty" json:"substrateBatch,omitempty"`
}
