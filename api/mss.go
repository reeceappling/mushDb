package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/reeceappling/goUtils/v2/utils"
	"github.com/reeceappling/mushDb/api/env"
	"github.com/reeceappling/mushDb/api/request"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"slices"
)

// TODO: needed for
// MSS-MSS transfers, MSS-plate/slant transfers

// TODO: newFromPCdwater ????
// TODO: newFromSporePrint (typical but requires PC-d water to not be referenced)

type MSS struct {
	// ALWAYS assume contaminated
	MainCollectionIdField   `bson:"inline"`
	CreationDateField       `bson:"inline"`
	WaterJarOptionalField   `bson:"inline"` // TODO: HANDLE THIS EVERYWHERE! NOT YET DONE IN TS!
	SpeciesField            `bson:"inline"`
	SubspeciesOptionalField `bson:"inline"`
	// NOTE: parentType is always either sporePrint or purchased
	MainCollectionOptionalParentField `bson:"inline"` // no parent means purchased, traded-for, or imported
	TransfersOutField                 `bson:"inline"`
	SaleField                         `bson:"inline"`
	DisposedField                     `bson:"inline"`
	// TODO: ADD PICS?
	NotesField       `bson:"inline"`
	LastUpdatedField `bson:"inline"`
	AclField         `bson:"inline"`
}

func (M MSS) Innoculatable() error {
	return errors.New("mss never innoculatable") // TODO: ensure ok
}

func (M MSS) CanTransferTo(dst geneticSource) error {
	if !slices.Contains([]string{PlateSourceType, SlantSourceType /*TODO: MSS?*/, StasisTubeSourceType}, dst.SourceType()) {
		return errors.New("mss transfers cannot go to " + dst.SourceType())
	}
	return nil
}

func (M MSS) GeneticInfoAsParent() (GeneticParentInfo, error) {
	return GeneticParentInfo{
		SpeciesOptionalField:    M.SpeciesField.AsOptional(),
		SubspeciesOptionalField: M.SubspeciesOptionalField,
		KnownFruitableField:     KnownFruitableField{utils.Pointer(false)},
		GenerationsFields: GenerationsFields{
			GenSporeField:        GenSporeField{utils.Pointer(Generation(0))},
			GenSinceFruitOrSpore: utils.Pointer(Generation(0)),
		},
	}, nil
}

func (M MSS) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return utils.Pointer(Generation(0)), utils.Pointer(Generation(0))
}

//func (M MSS) setTransferParent(ctx context.Context, xfer Transfer) error {
//	coll := DbFrom(ctx).Collection(MssCollectionName)
//	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
//	if err != nil {
//		return err
//	}
//	res, err := coll.UpdateByID(ctx, M.Id, upd)
//	if err != nil {
//		return err
//	}
//	if res.ModifiedCount == 0 {
//		return ErrNoParentModifiedForTransfer
//	}
//	return nil
//}

func (M MSS) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
	return errors.New("mss cannot be a child in a normal transfer. Must be created manually from spore print or imported")
}

func initializeMSS(ctx context.Context) error {
	db := DbFrom(ctx)
	coll := db.Collection(MssCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		creationDateIndexModel,
		newSimpleIndex("species", "species", false, false, false),
		newSimpleIndex("subspecies", "subspecies", false, true, false),
		//newSimpleIndex("parent", "parent", false, true, false),
		//transfersOutIndexModel,
		//saleIndexModel,
		//disposedIndexModel,
		//Notes (no index unless tags)
		projectsIndexModel,
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	return env.IfNotProd(ctx, func() error {
		// If test agar batch does not exist, then create it
		testId := mainCollIdForint(idTestMSS)
		testItem := &MSS{
			MainCollectionIdField:             MainCollectionIdField{testId},
			CreationDateField:                 CreationDateField{exampleTime},
			WaterJarOptionalField:             WaterJarOptionalField{WaterSource: &exWaterId},
			SpeciesField:                      SpeciesField{testEntryStringId},
			SubspeciesOptionalField:           SubspeciesOptionalField{&testEntryStringId},
			TransfersOutField:                 TransfersOutField{exAlts},
			MainCollectionOptionalParentField: MainCollectionOptionalParentField{&exSporePrint},
			DisposedField:                     DisposedField{&exampleTime},
			NotesField:                        NotesField{exampleNotes()},
			LastUpdatedField:                  LastUpdatedField{exampleTime},
		}
		return addTestMainEntries(ctx, testItem)
	})
}

type createMssRequest struct {
	WaterJarOptionalField // TODO: HANDLE THIS! Allow creation with or without
	SporePrintId          MainCollectionId
	NotesField
	WriteTagToField
	// Uses parent perms, then email can modify if they have the perms for parent
}

func createMssHandler(w http.ResponseWriter, r *http.Request) { // Only called from spore print page
	data := createMssRequest{}
	id := NextMainCollectionId()
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(bs, &data)
	if err != nil {
		http.Error(w, "failed to unmarshal request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	ctx, db := Db(r)
	// Validate parent
	parent := SporePrint{}
	err = db.Collection(SporePrintCollectionName).FindOne(ctx, BsonFindFilter(IDfld, data.SporePrintId)).Decode(&parent)
	if err != nil {
		dbErr(w, "failed to find sporePrint: "+err.Error(), http.StatusBadRequest)
		return
	}
	ctx, now := request.UnixTime(r.Context())
	toInsert := &MSS{
		MainCollectionIdField:             MainCollectionIdField{id},
		CreationDateField:                 CreationDateField{now},
		SpeciesField:                      SpeciesField{parent.Species},
		SubspeciesOptionalField:           parent.SubspeciesOptionalField,
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&parent.Id},
		NotesField:                        NotesField{data.Notes},
		LastUpdatedField:                  LastUpdatedField{now},
		AclField:                          parent.AclField, // do NOT ensure email is authorized to write on parent, they will just be blocked from viewing.
	}
	err = writeRfidTagIfNecessary(ctx, data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	finishCreateMainCollectionEntry(ctx, toInsert, w)
}

type importMssRequest struct {
	CreationDateField
	SpeciesField
	SubspeciesOptionalField
	NotesField
	WriteTagToField
	// image as "img"
	// No ParentType/Parent because these are assumed to be from outside sources
}

func importMssHandler(w http.ResponseWriter, r *http.Request) {
	ctx, now := request.UnixTime(r.Context())
	data := importMssRequest{}
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(bs, &data)
	if err != nil {
		http.Error(w, "failed to unmarshal request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	id := NextMainCollectionId()
	finalPerms, err := ImportFinalPerms(r.Context(), data.Species, data.Subspecies)
	if err != nil {
		http.Error(w, "failed to get species and/or subspecies: "+err.Error(), http.StatusInternalServerError)
		return
	}

	toInsert := MSS{
		MainCollectionIdField:   MainCollectionIdField{id},
		CreationDateField:       data.CreationDateField,
		SpeciesField:            data.SpeciesField,
		SubspeciesOptionalField: data.SubspeciesOptionalField,
		NotesField:              data.NotesField,
		LastUpdatedField:        LastUpdatedField{now},
		AclField:                AclField{finalPerms},
	}
	err = writeRfidTagIfNecessary(ctx, data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	finishImportMainCollectionEntry(ctx, &toInsert, w)
}

type updateMssRequest struct {
	NotesUpdateField
	DisposedField
	SaleField
	PermsOnRequest `json:"acl"`
}

func (req updateMssRequest) modsFor(existing *MSS, aclField AclField) (bson.D, error) {
	return NewMods().
		updateSaleIfNeeded(req.Sale, existing.Sale).
		updateDisposedIfNeeded(req, existing).
		// TODO: pics?
		updateNotesIfNeeded(req, existing).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateMssHandler(w http.ResponseWriter, r *http.Request) {
	data := updateMssRequest{}
	defer r.Body.Close()
	idStr, err := UrlDecodeString(r.PathValue("id"))
	if err != nil {
		http.Error(w, "failed to url decode string: "+err.Error(), http.StatusInternalServerError)
		return
	}
	mainCollId, err := StandardizeMainCollectionId(idStr)
	if err != nil {
		http.Error(w, "failed to standardize main collection id: "+err.Error(), http.StatusBadRequest)
		return
	}
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(bs, &data)
	if err != nil {
		http.Error(w, "failed to unmarshal request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	ctx, db := Db(r)
	coll := db.Collection(MssCollectionName)

	// go get current entry
	existing := MSS{}
	err = coll.FindOne(ctx, BsonFindFilter(IDfld, *mainCollId)).Decode(&existing)
	if err != nil {
		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	//Validation
	if data.Sale != nil && (existing.Sale == nil || *existing.Sale != *data.Sale) {
		if err = db.Collection(SalesCollectionName).FindOne(ctx, BsonFindFilter(IDfld, data.Sale)).Err(); err != nil {
			dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
			return
		}
	}
	finishMainCollItemUpdate(ctx, w, data.modsFor, &existing, data.PermsOnRequest)
}

func deleteMssHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Empty id for delete request", http.StatusBadRequest)
		return
	}
	id, err := Base58Str(idStr).ToMainCollectionId()
	if err != nil {
		http.Error(w, "Invalid ID to delete: "+err.Error(), http.StatusBadRequest)
		return
	}
	// Validate not used in other places...
	ctx := r.Context()
	// ensure item does not have any transfers in or out
	item, err := GetMainCollectionItemSpecific[*MSS](ctx, id, &MSS{})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			http.Error(w, "Item to be deleted not found: "+err.Error(), http.StatusNotFound) // TODO: ok?
		} else {
			http.Error(w, "Failed to retrieve item to be deleted: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if item.Parent != nil {
		// TODO: what if we want to remove it from the parent as well?
		http.Error(w, "Cannot delete innoculated items!", http.StatusConflict)
		return
	}
	if item.TransfersOut != nil && len(item.TransfersOut) > 0 {
		http.Error(w, "Cannot delete items with transfers out", http.StatusConflict)
		return
	}

	// Delete if not found elsewhere!
	DeleteCollectionItem(ctx, item.CollectionName(), id, w)
}
