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

func (sw SporeSwab) Innoculatable() error {
	return errors.New("sporeSwabs are not innoculatable")
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
	coll := DbFrom(ctx).Collection(SporeSwabCollectionName)
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
	return env.IfNotProd(ctx, func() error { // TODO: ensure ok
		// If test agar batch does not exist, then create it
		testItem := &SporeSwab{
			MainCollectionIdField:             MainCollectionIdField{exSwabId},
			MainCollectionOptionalParentField: MainCollectionOptionalParentField{&exSporePrint},
			CreationDateField:                 CreationDateField{exampleTime},
			SpeciesField:                      SpeciesField{testEntryStringId},
			SubspeciesOptionalField:           SubspeciesOptionalField{&testEntryStringId},
			SaleField:                         SaleField{&exAltId},
			DisposedField:                     DisposedField{&exampleTime},
			NotesField:                        NotesField{exampleNotes()},
			LastUpdatedField:                  LastUpdatedField{exampleTime},
			AclField:                          allCanWriteAcl(),
		}
		return addTestMainEntries(ctx, testItem)
	})
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

	ctx, _ = request.UnixTime(ctx)
	var fr *Fruit
	var swabOut *SporeSwab
	_, er := newTxn(ctx, func(sessCtx mongo.SessionContext) (any, error) {
		var e error = nil
		switch parentItem.SourceType() {
		case FruitingChamberSourceType, BagSourceType, PlateSourceType, SlantSourceType, PlugSourceType, GrainJarSourceType:
			fr, e = FruitFromSourceInTxn(sessCtx, parentItem)
			if e != nil {
				return nil, e
			}
			swabOut, e = fr.createSporeSwabInTxn(sessCtx, data.NotesField)
			if e != nil {
				return nil, e
			}
			// TODO: unsure if we want this to write!!!
			e = writeRfidTagIfNecessary(ctx, data.WriteTagTo, swabOut.Id) // TODO: this should always only occur right before the true writes
			if e != nil {
				return nil, errors.Join(e, errors.New("failed to write tag"))
			}
			return nil, e
		case FruitSourceType:
			var ok bool
			fr, ok = parentItem.(*Fruit)
			if !ok {
				return nil, errors.New("fruit is not a Fruit?")
			}
			swabOut, e = fr.createSporeSwabInTxn(sessCtx, data.NotesField) // TODO: notes?
			if e != nil {
				return nil, e
			}
			e = writeRfidTagIfNecessary(ctx, data.WriteTagTo, swabOut.Id) // TODO: this should always only occur right before the true writes
			if e != nil {
				return nil, errors.Join(e, errors.New("failed to write tag"))
			}
			return nil, e
		case SporePrintSourceType: // Goes directly to swab
			parentPrint, ok := parentItem.(*SporePrint)
			if !ok {
				return nil, errors.New("print is not a print?")
			}
			swabOut, e = parentPrint.createSwabInTxn(sessCtx, data.NotesField, NotesField{}) // TODO: xferNotes
			if e != nil {
				return nil, e
			}
			e = writeRfidTagIfNecessary(ctx, data.WriteTagTo, swabOut.Id) // TODO: this should always only occur right before the true writes
			if e != nil {
				return nil, errors.Join(e, errors.New("failed to write tag"))
			}
			return nil, e
		default:
			e := errors.New("invalid source type: " + parentItem.SourceType())
			http.Error(w, e.Error(), http.StatusBadRequest)
			return nil, e
		}
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
	defer r.Body.Close()
	req := updateSporeSwabRequest{}
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(bs, &req)
	if err != nil {
		http.Error(w, "failed to unmarshal body: "+err.Error(), http.StatusBadRequest)
		return
	}
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

	ctx, db := Db(r)
	coll := db.Collection(SporeSwabCollectionName)

	// go get current sporeSwab
	existing := SporeSwab{}
	err = coll.FindOne(ctx, BsonFindFilter("_id", id)).Decode(&existing)
	if err != nil {
		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishMainCollItemUpdate(ctx, w, req.modsFor, &existing, req.PermsOnRequest)
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

	ctx, now := request.UnixTime(r.Context())
	toInsert := SporeSwab{
		MainCollectionIdField:   id.IdField(),
		CreationDateField:       data.CreationDateField,
		SpeciesField:            data.SpeciesField,
		SubspeciesOptionalField: data.SubspeciesOptionalField,
		NotesField:              data.NotesField,
		LastUpdatedField:        LastUpdatedField{now},
		AclField:                finalPerms.AsField(),
	}
	//// TODO: THIS! err = writeRfidTagIfNecessary(ctx, data.WriteTagTo, id) // TODO: this should always only occur right before the true writes
	//if err != nil {
	//	http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
	//	return
	//}
	finishImportMainCollectionEntry(ctx, &toInsert, w)
}

func deleteSporeSwabHandler(w http.ResponseWriter, r *http.Request) {
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
	item, err := GetMainCollectionItemSpecific[*SporeSwab](ctx, id, &SporeSwab{})
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
		http.Error(w, "Cannot delete innoculated items!", http.StatusConflict) // TODO: type ok?
		return
	}
	if item.TransfersOut != nil && len(item.TransfersOut) > 0 {
		http.Error(w, "Cannot delete items with transfers out", http.StatusConflict) // TODO: type ok?
		return
	}

	// Delete if not found elsewhere!
	DeleteCollectionItem(ctx, item.CollectionName(), id, w)
}

func DeleteCollectionItem[U CollectionId](ctx context.Context, collName string, id U, w http.ResponseWriter) {
	idStr := string(id.AsBase58())
	deleteResult, err := DbFrom(ctx).Collection(collName).DeleteOne(ctx, BsonFindFilter("_id", id)) // TODO: _id always ok here?
	if err != nil {
		http.Error(w, "failed to delete item "+idStr+" from "+collName+": "+err.Error(), http.StatusInternalServerError)
		return
	}

	if deleteResult.DeletedCount == 0 {
		http.Error(w, "failed to delete item "+idStr+" from "+collName+". Id not found", http.StatusNotFound)
		return
	}
	_, err = w.Write([]byte(idStr))
	handleWriteErr(err, w)
}
