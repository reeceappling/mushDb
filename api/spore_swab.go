package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/reeceappling/goUtils/v2/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"slices"
)

// TODO: fromSporePrint,

type SporeSwab struct {
	MainCollectionIdField `bson:"inline"`
	// Parent is always either sporePrint, fruit, or purchased
	MainCollectionOptionalParentField `bson:"inline"` // won't exist for pre-existing or purchased
	ParentTypeField                   `bson:"inline"` // sporePrint, fruit, or missing (purchased/other)
	CreationDateField                 `bson:"inline"` // Swab or receive date
	SpeciesField                      `bson:"inline"`
	SubspeciesOptionalField           `bson:"inline"`
	SaleField                         `bson:"inline"` // TODO: was sales! singular now
	DisposedField                     `bson:"inline"`
	TransfersOutField                 `bson:"inline"`
	NotesField                        `bson:"inline"`
	LastUpdatedField                  `bson:"inline"`
	AclField                          `bson:"inline"`
}

func (sw SporeSwab) Innoculatable() bool {
	return false
}

func (sw SporeSwab) CanTransferTo(dst geneticSource) error {
	validSwabDestinations := []string{PlateSourceType, SlantSourceType} // TODO: any more?
	if !slices.Contains(validSwabDestinations, dst.SourceType()) {
		return errors.New("sporeSwabs cannot transfer to " + dst.SourceType())
	}
	return errors.New("swabs cannot be transferred (unsure if this is ok)")
}

func (sw SporeSwab) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
	return errors.New("sporeSwabs cannot be destinations of transfers")
}

func (sw SporeSwab) GeneticInfoAsParent() (GeneticParentInfo, error) {
	return GeneticParentInfo{
		SpeciesOptionalField:    sw.SpeciesField.AsOptional(),
		SubspeciesOptionalField: sw.SubspeciesOptionalField,
		GenerationsFields:       GenerationsFieldFor(utils.Pointer(Generation(0))),
	}, nil
}

func (sw SporeSwab) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return utils.Pointer(Generation(0)), utils.Pointer(Generation(0))
}

func (sw SporeSwab) id() []byte {
	return []byte(sw.Id.dbIdStr())
}

func initializeSporeSwabs(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(SporeSwabCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		//newSimpleIndex("parent", "parent", false, false, false),
		// TODO: parentType?
		newSimpleIndex("creationDate", "creationDate", true, false, false), // TODO: INDEX CREATION DATES EVERYWHERE!
		newSimpleIndex("species", "species", false, false, false),
		newSimpleIndex("subspecies", "subspecies", false, true, false),
		//saleIndexModel,
		//disposedIndexModel,
		//transfersOutIndexModel,
		//Notes (no index unless tags)
		projectsIndexModel,
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	testItem := &SporeSwab{
		MainCollectionIdField:             MainCollectionIdField{exSwabId},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&exSporePrint},
		CreationDateField:                 exampleTime.asCreationDate(),
		SpeciesField:                      SpeciesField{testEntryStringId},
		SubspeciesOptionalField:           SubspeciesOptionalField{&testEntryStringId},
		SaleField:                         SaleField{&exAltId},
		DisposedField:                     DisposedField{&exampleTime},
		NotesField:                        NotesField{exampleNotes()},
		LastUpdatedField:                  LastUpdatedField{exampleTime},
		AclField:                          allCanWriteAcl(),
	}
	return addTestMainEntries(ctx, testItem)
}

type createSporeSwabRequest struct {
	MainCollectionParentField // TODO; required
	// TODO: DERIVE PARENT TYPE!
	// SporePrintId MainCollectionId // TODO: make this just parentId, used to be SporePrintId. Ensure handled on ts side
	NotesField
	WriteTagToField
}

// TODO: multi-swab creation request?

func createSporeSwabHandler(w http.ResponseWriter, r *http.Request) { // TODO: TEST ALL BRANCHES!
	data := createSporeSwabRequest{}
	defer r.Body.Close()
	// Process text (or object)
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read Data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	// PARSE INTO CORRECT DATA FORMAT
	err = json.Unmarshal(bs, &data)
	if err != nil {
		http.Error(w, "failed to unmarshal Data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	id := NextMainCollectionId()
	ctx := r.Context()
	parentItem, err := GetMainCollectionItemWithId(ctx, data.Parent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	parent, err := parentItem.GeneticInfoAsParent()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !slices.Contains([]string{FruitSourceType, SporePrintSourceType}, parentItem.SourceType()) {
		http.Error(w, "sporeSwab parent must be fruit or print, invalid: "+parentItem.SourceType(), http.StatusInternalServerError)
		return
	}
	if parent.Species == nil {
		http.Error(w, "species nil, should never happen", http.StatusInternalServerError)
		return
	}

	now := unixTimeForNow()
	parentType := parentItem.SourceType() // TODO: ensure ok
	out := SporeSwab{
		MainCollectionIdField:             MainCollectionIdField{id},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&data.Parent},
		ParentTypeField:                   ParentTypeField{&parentType},
		CreationDateField:                 now.asCreationDate(),
		SpeciesField:                      SpeciesField{*parent.Species},
		SubspeciesOptionalField:           parent.SubspeciesOptionalField,
		NotesField:                        NotesField{data.Notes},
		LastUpdatedField:                  LastUpdatedField{now},
		// Do not check permissions, just pass parent perms to child
		AclField: AclField{parentItem.Permissions()}, // note: do NOT add user. They can be created by any non-guest, but not necessarily viewable by them...
	}

	var swabOut *SporeSwab
	_, er := newTxn(ctx, func(sessCtx mongo.SessionContext) (any, error) {
		db := mongo.SessionFromContext(sessCtx).Client().Database(dbName)
		var parentFruit *Fruit
		switch parentItem.SourceType() {
		case FruitSourceType: // Goes directly to swab
			e := db.Collection(FruitsCollName).
				FindOne(sessCtx, BsonFindFilter("_id", data.Parent)).
				Decode(parentFruit)
			if e != nil {
				return nil, e
			}
		case SporePrintSourceType: // Goes directly to swab
			var parentPrint *SporePrint
			e := db.Collection(SporePrintCollectionName).
				FindOne(sessCtx, BsonFindFilter("_id", data.Parent)).
				Decode(parentFruit)
			if e != nil {
				return nil, e
			}
			idOut := NextMainCollectionId()
			swab := SporeSwab{
				MainCollectionIdField:             MainCollectionIdField{idOut},
				MainCollectionOptionalParentField: MainCollectionOptionalParentField{&data.Parent},
				ParentTypeField:                   ParentTypeField{utils.Pointer("sporePrint")},
				CreationDateField:                 CreationDateField{now},
				SpeciesField:                      SpeciesField{*parent.Species},
				SubspeciesOptionalField:           parent.SubspeciesOptionalField,
				NotesField:                        NotesField{}, // TODO: ???
				LastUpdatedField:                  LastUpdatedField{now},
				AclField:                          parentPrint.AclField,
			}
			xfer := Transfer{
				AlternateCollectionIdField: AlternateCollectionIdField{newAlternateCollectionId()},
				From:                       parentPrint.Id,
				To:                         idOut,
				FromType:                   "sporePrint",
				ToType:                     "sporeSwab",
				CreationDateField:          CreationDateField{now},
				Reason:                     xferReasonReady,
				NotesField:                 NotesField{}, // TODO: FIX!
				LastUpdatedField:           LastUpdatedField{now},
				AclField:                   parentPrint.AclField,
			}
			err := addToIdMapCollection(sessCtx, &swab)
			if err != nil {
				return nil, err
			}
			// Update print with new swab id
			// Update xfers out and lastUpdated on parent
			upd, err := NewMods().Push("transfersOut", xfer.Id).withLastUpdated(now).Finalized()
			if err != nil {
				return nil, err
			}
			_, err = db.Collection(SporePrintCollectionName).UpdateByID(ctx, parentPrint.Id, upd)
			if err != nil {
				return nil, err
			}
			_, err = db.Collection(SporeSwabCollectionName).InsertOne(ctx, &swab)
			if err != nil {
				return nil, errors.Join(errors.New("failed to insert new spore print"), err)
			}
			_, err = db.Collection(TransfersCollName).InsertOne(ctx, &xfer)
			if err != nil {
				return nil, errors.Join(errors.New("failed to insert new spore print"), err)
			}
			return &swab, nil
		case BagSourceType, FruitingChamberSourceType, PlateSourceType, PlugSourceType, LcSourceType, SlantSourceType: // TODO: creates intermediary fruit
			// Create transfer. Add spore swab to spore print
			parentFruit, err = FruitFromSourceInTxn(sessCtx, parentItem)
			if err != nil {
				return nil, err
			}
		default: // TODO: ERROR. MssSourceType, WaterJar, etc
			return nil, errors.New("invalid source for new swab: " + parentItem.SourceType())
		}
		swabOut, err = swabFromFruitInTxn(sessCtx, parentFruit, out.NotesField)
		if err != nil {
			return nil, err
		}
		return nil, err
	})
	if er != nil {
		http.Error(w, er.Error(), http.StatusInternalServerError)
		return
	}
	bsOut, err := json.Marshal(swabOut)
	if err != nil {
		http.Error(w, "failed to marshal result: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(bsOut)
	handleWriteErr(err, w)
}

type updateSporeSwabRequest struct {
	SaleField
	DisposedField
	NotesUpdateField
	PermsOnRequest `json:"acl"`
}

func (req updateSporeSwabRequest) modsFor(existing *SporeSwab, aclField AclField) (bson.D, error) {
	return NewMods().
		updateSaleIfNeeded(req.Sale, existing.Sale).
		updateDisposedIfNeeded(req, existing).
		updateNotesIfNeeded(req, existing).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateSporeSwabHandler(w http.ResponseWriter, r *http.Request) {
	data := updateSporeSwabRequest{}
	idStr, err := UrlDecodeString(r.PathValue("id"))
	if err != nil {
		http.Error(w, "failed to url decode string: "+err.Error(), http.StatusInternalServerError)
		return
	}
	mainCollId, err := StandardizeMainCollectionId(idStr)
	if err != nil {
		println("failed to standardize main collection id: " + err.Error()) // TODO: del
		http.Error(w, "failed to standardize main collection id: "+err.Error(), http.StatusBadRequest)
		return
	}
	id := *mainCollId

	out := data
	ctx, db := Db(r)
	coll := db.Collection(SporeSwabCollectionName)

	// go get current sporeSwab
	existing := SporeSwab{}
	err = coll.FindOne(ctx, BsonFindFilter("_id", id)).Decode(&existing)
	if err != nil {
		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishMainCollItemUpdate(ctx, w, out.modsFor, &existing, out.PermsOnRequest)
}

type importSporeSwabRequest struct {
	CreationDateField
	SpeciesField
	SubspeciesOptionalField
	NotesField
}

func importSporeSwabHandler(w http.ResponseWriter, r *http.Request) {
	data := importSporeSwabRequest{}
	id := NextMainCollectionId()
	defer r.Body.Close()

	// Process text (or object)
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "unable to read Data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	// PARSE INTO CORRECT DATA FORMAT
	err = json.Unmarshal(bs, &data)
	if err != nil {
		http.Error(w, "unable to unmarshal json form Data: "+err.Error(), http.StatusBadRequest)
		return
	}
	//if err = Data.Perms.ValidateUserCanWrite(r.Context()); err != nil {
	//	http.Error(w, "email cannot write with these perms: "+err.Error(), http.StatusBadRequest)
	//	return
	//}

	finalPerms, err := ImportFinalPerms(r.Context(), data.Species, data.Subspecies)
	if err != nil {
		http.Error(w, "failed to get species and/or subspecies: "+err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	toInsert := SporeSwab{
		MainCollectionIdField:   MainCollectionIdField{id},
		CreationDateField:       data.CreationDateField,
		SpeciesField:            data.SpeciesField,
		SubspeciesOptionalField: data.SubspeciesOptionalField,
		NotesField:              data.NotesField,
		LastUpdatedField:        LastUpdatedFieldForNow(),
		AclField:                AclField{finalPerms},
	}
	finishImportMainCollectionEntry(ctx, &toInsert, w)
}

func swabFromFruitInTxn(ctx mongo.SessionContext, parent *Fruit, notes NotesField) (*SporeSwab, error) {
	id := NextMainCollectionId()
	now := unixTimeForNow()
	// TODO: writeTagTo?
	toInsert := SporeSwab{
		MainCollectionIdField:             MainCollectionIdField{id},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&parent.Id},
		ParentTypeField:                   ParentTypeField{utils.Pointer("fruit")},
		CreationDateField:                 now.asCreationDate(),
		SpeciesField:                      parent.SpeciesField,
		SubspeciesOptionalField:           parent.SubspeciesOptionalField,
		NotesField:                        notes,
		LastUpdatedField:                  LastUpdatedField{now},
		// Do not check permissions, just pass parent perms to child
		AclField: parent.AclField,
	}
	db := mongo.SessionFromContext(ctx).Client().Database(dbName)
	err := addToIdMapCollection(ctx, &toInsert)
	if err != nil {
		return nil, err
	}
	// Update fruit with new print id
	//err = parent.addSporeSwab(ctx, id)
	//if err != nil {
	//	return nil, errors.Join(errors.New("failed to add spore print to parent fruit"), err)
	//}
	// TODO: add transfer to parent for swab! should swabs have their own field on fruits?
	_, err = db.Collection(SporeSwabCollectionName).InsertOne(ctx, toInsert)
	if err != nil {
		return nil, errors.Join(errors.New("failed to insert new spore print"), err)
	}
	return &toInsert, nil
}
