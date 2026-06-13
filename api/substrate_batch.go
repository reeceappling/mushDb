package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/reeceappling/mushDb/api/request"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
)

// used in: bag, FruitingChamber

type SubstrateBatch struct {
	AlternateCollectionIdField `bson:"inline"`
	// Initial wetness is quantified on each bag/box
	CreationDateField    `bson:"inline"` // Date of first hydration
	SubstrateRecipeField `bson:"inline"`
	NotesField           `bson:"inline"`
	LastUpdatedField     `bson:"inline"`
	AclField             `bson:"inline"`
}

func initializeSubstrateBatches(ctx context.Context) error {
	// Indices
	coll := DbFrom(ctx).Collection(SubstrateBatchCollectionName)
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
	createdDate := CreationDateField{exampleTime}
	err = addTestAltEntries(ctx, []SubstrateBatch{
		// Coir
		{
			AlternateCollectionIdField: altCollIdFieldForint(idCoir),
			CreationDateField:          createdDate,
			SubstrateRecipeField:       SubstrateRecipeField{Substrate: altCollIdForint(idCoir)},
			NotesField: NotesField{[]Note{
				newNote(ogTime, "test coir batch"),
			}},
			AclField:         allCanReadAcl(nil),
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
			AclField:         allCanReadAcl(nil),
			LastUpdatedField: LastUpdatedField{LastUpdated: ogTime},
		},
	}...)
	// Add test entry
	testItem := SubstrateBatch{
		AlternateCollectionIdField: altCollIdFieldForint(idTestingOnly),
		CreationDateField:          CreationDateField{exampleTime},
		SubstrateRecipeField:       SubstrateRecipeField{altCollIdForint(idTestingOnly)},
		NotesField:                 NotesField{exampleNotes()},
		LastUpdatedField:           LastUpdatedField{exampleTime},
		AclField:                   allCanWriteAcl(),
	}
	return addTestAltEntries(ctx, testItem)
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
	// Validate
	_, err = req.SubstrateRecipeField.Get(r.Context())
	if err != nil {
		http.Error(w, "did not find substrate recipe: "+err.Error(), http.StatusNotFound)
	}
	id := newAlternateCollectionId()

	ctx, now := request.UnixTime(r.Context())
	// Create entry to insert
	toInsert := &SubstrateBatch{
		AlternateCollectionIdField: AlternateCollectionIdField{id},
		CreationDateField:          CreationDateField{now},
		SubstrateRecipeField:       SubstrateRecipeField{Substrate: req.Substrate},
		NotesField:                 req.NotesField,
		LastUpdatedField:           LastUpdatedField{now},
		AclField:                   allCanReadAcl(GetUserEmailPtr(ctx)),
	}
	finishCreateAlternateEntry(ctx, toInsert, w)
}

type updateSubstrateBatchRequest struct {
	NotesUpdateField
	PermsOnRequest `json:"acl"`
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
	err = DbFrom(ctx).Collection(SubstrateBatchCollectionName).FindOne(ctx, bson.M{
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
