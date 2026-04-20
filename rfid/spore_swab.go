package rfid

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
	}
	return addTestMainEntries(ctx, testItem)
}

type createSporeSwabsRequest struct {
	num int
	MainCollectionParentField
	// SporePrintId MainCollectionId // TODO: make this just parentId, used to be SporePrintId. Ensure handled on ts side
	NotesField
}

// TODO: REALLY FLESH THIS OUT
func createSporeSwabHandler(w http.ResponseWriter, r *http.Request) {
	data := createSporeSwabsRequest{}
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
	ids := NextMainCollectionIds(data.num)
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
	out := make([]interface{}, data.num)
	_, err = newTxn(ctx, func(sessCtx mongo.SessionContext) (any, error) {
		for i, _ := range out {
			temp := SporeSwab{
				MainCollectionIdField:             MainCollectionIdField{ids[i]},
				MainCollectionOptionalParentField: MainCollectionOptionalParentField{&data.Parent},
				ParentTypeField:                   ParentTypeField{utils.Pointer(parentItem.SourceType())}, // TODO: ensure ok
				CreationDateField:                 now.asCreationDate(),
				SpeciesField:                      SpeciesField{*parent.Species},
				SubspeciesOptionalField:           parent.SubspeciesOptionalField,
				NotesField:                        NotesField{data.Notes},
				LastUpdatedField:                  LastUpdatedField{now},
				// Do not check permissions, just pass parent perms to child
				AclField: AclField{parentItem.Permissions()},
			}
			out[i] = temp
			if errr := addToIdMapCollection(sessCtx, &temp); errr != nil {
				return nil, errr
			}
		}
		// Actually add the swabs to their collection
		return mongo.SessionFromContext(sessCtx).Client().Database(dbName).Collection(SporeSwabCollectionName).
			InsertMany(ctx, out)
	})
	if err != nil {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bsOut, err := json.Marshal(out)
	if err != nil {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(bsOut)
	if err != nil {
		handleWriteErr(err, w)
	}
}

type updateSporeSwabRequest struct {
	SaleField
	DisposedField
	NotesUpdateField
	PermsOnRequest
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
	b58Id := Base58Str(r.PathValue("id"))
	id, err := b58Id.toMainCollectionId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	out := data
	ctx, db := Db(r)
	coll := db.Collection(SporeSwabCollectionName)

	// go get current sporeSwab
	existing := SporeSwab{}
	err = coll.FindOne(ctx, bsonFindFilter("_id", id)).Decode(&existing)
	if err != nil {
		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishMainCollItemUpdate(ctx, w, coll, out.modsFor, &existing, out.PermsOnRequest)
}

type importSporeSwabRequest struct {
	CreationDateField
	SpeciesField
	SubspeciesOptionalField
	NotesField
	PermsOnRequest
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

	user, err := GetAuthInfo(r.Context())
	if err != nil {
		http.Error(w, "failed to get auth info: "+err.Error(), http.StatusUnauthorized)
		return
	}
	sp, subsp, err := getSpeciesAndSubspecies(r.Context(), data.Species, data.SubSpecies)
	if err != nil {
		http.Error(w, "failed to get species or subspecies: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var finalPerms *ACL = nil
	if subsp != nil {
		finalPerms = subsp.DefaultAcl.Clone()
	} else {
		finalPerms = sp.DefaultAcl.Clone()
	}
	// Add user to the acl as a writer
	finalPerms.Users[user.Email] = true

	ctx, db := Db(r)
	coll := db.Collection(SporeSwabCollectionName)
	toInsert := SporeSwab{
		MainCollectionIdField:   MainCollectionIdField{id},
		CreationDateField:       data.CreationDateField,
		SpeciesField:            data.SpeciesField,
		SubspeciesOptionalField: data.SubspeciesOptionalField,
		NotesField:              data.NotesField,
		LastUpdatedField:        LastUpdatedFieldForNow(),
		AclField:                AclField{finalPerms},
	}
	finishImportMainCollectionEntry(ctx, coll, &toInsert, data.PermsOnRequest, w)
}
