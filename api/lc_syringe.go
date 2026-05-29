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

// TODO: needed for xfers

// TODO: newFromLC

// TODO: new are this, sporeSwab, plugs

// Naming convention "{ParentLCID}-#" // TODO: ?????

//func parseName() // TODO: ???

type LcSyringe struct {
	MainCollectionIdField `bson:"inline"`
	// Parent is always either purchased (nil), LC, or LcSyringe
	MainCollectionOptionalParentField `bson:"inline"` // TODO: likely won't exist for pre-existing
	CreationDateField                 `bson:"inline"` // create or receive date
	SpeciesField                      `bson:"inline"`
	SubspeciesOptionalField           `bson:"inline"`
	SaleField                         `bson:"inline"`
	GenerationsFields                 `bson:"inline"`
	KnownFruitableField               `bson:"inline"`
	ConfirmedCleanField               `bson:"inline"`
	TransfersOutField                 `bson:"inline"`
	DisposedField                     `bson:"inline"`
	NotesField                        `bson:"inline"`
	LastUpdatedField                  `bson:"inline"`
	AclField                          `bson:"inline"`
}

func (lcs LcSyringe) Innoculatable() bool {
	return false
}

func (lcs LcSyringe) CanTransferTo(dst geneticSource) error {
	if !slices.Contains([]string{BagSourceType, GrainJarSourceType, LcSourceType, PlateSourceType, PlugSourceType, SlantSourceType}, dst.SourceType()) {
		return errors.New("lc syringe cannot transfer to " + dst.SourceType())
	}
	return nil
}

//func (sw LcSyringe) setTransferParent(ctx context.Context, xfer Transfer) error {
//	// TODO: can this even happen?
//	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(sw.CollectionName())
//	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
//	if err != nil {
//		return err
//	}
//	res, err := coll.UpdateByID(ctx, sw.Id, upd)
//	if err != nil {
//		return err
//	}
//	if res.ModifiedCount == 0 {
//		return ErrNoParentModifiedForTransfer
//	}
//	return nil
//}

func (sw LcSyringe) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
	// TODO: cannot happen
	panic("does not happen")
}

func (sw LcSyringe) GeneticInfoAsParent() (GeneticParentInfo, error) {
	return GeneticParentInfo{
		SpeciesOptionalField:    sw.SpeciesField.AsOptional(),
		SubspeciesOptionalField: sw.SubspeciesOptionalField,
		GenerationsFields:       GenerationsFieldFor(utils.Pointer(Generation(0))),
	}, nil
}

func (sw LcSyringe) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return utils.Pointer(Generation(0)), utils.Pointer(Generation(0))
}

func (sw LcSyringe) altId() MainCollectionId {
	return sw.Id
}

func (sw LcSyringe) id() []byte {
	return []byte(sw.Id.dbIdStr())
}

//func (sp LcSyringe) knownFruitable() bool {
//	return false
//}

func initializeSyringes(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(LcSyringeCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		//newSimpleIndex("parent", "parent", false, true, false),
		newSimpleIndex("creationDate", "creationDate", true, false, false), // TODO: INDEX CREATION DATES EVERYWHERE!
		newSimpleIndex("species", "species", false, false, false),
		newSimpleIndex("subspecies", "subspecies", false, true, false),
		//saleIndexModel,
		//newSimpleIndex("genSinceSpore", "genSpore", true, true, false),
		//newSimpleIndex("genSinceFruitOrSpore", "genFruitOrSpore", true, true, false),
		//newSimpleIndex("knownFruitable", "knownFruitable", false, true, false),
		//newSimpleIndex("confirmedClean", "confirmedClean", false, true, false),
		//transfersOutIndexModel,
		//newSimpleIndex("disposed", "disposed", false, true, false),
		//// TODO: Projects?
		//Notes (no index unless tags)
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	testItem := &LcSyringe{
		MainCollectionIdField:             MainCollectionIdField{Id: exLCS},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&exLC},
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

type createLCSyringeRequest struct {
	LC MainCollectionId `json:"lc"`
	NotesField
	WriteTagToField
}

func createSyringeHandler(w http.ResponseWriter, r *http.Request) {
	data := createLCSyringeRequest{}
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id := NextMainCollectionId()
	err = json.Unmarshal(bs, &data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx, db := Db(r)
	// Validate inputs and grab parent
	parent := &LiquidCulture{}
	err = db.Collection(LCCollectionName).FindOne(ctx, bsonFindFilter("_id", id)).Decode(&parent)
	if err != nil {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if parent.Species == nil {
		dbErr(w, "Parent LC must be innoculated", http.StatusInternalServerError)
		return
	}
	now := unixTimeForNow()
	toInsert := LcSyringe{
		MainCollectionIdField:             MainCollectionIdField{Id: id},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&parent.Id},
		ConfirmedCleanField:               parent.ConfirmedCleanField, // TODO: is this ok?
		KnownFruitableField:               parent.KnownFruitableField,
		CreationDateField:                 now.asCreationDate(),
		SpeciesField:                      SpeciesField{Species: *parent.Species},
		SubspeciesOptionalField:           parent.SubspeciesOptionalField,
		GenerationsFields:                 parent.GenerationsFields,
		NotesField:                        NotesField{data.Notes},
		LastUpdatedField:                  LastUpdatedField{now},
		// Do not check permissions, just pass parent perms to child
		AclField: parent.AclField,
	}
	finishCreateMainCollectionEntry(ctx, &toInsert, w)
}

type updateSyringeRequest struct {
	SaleField // TODO: validate?
	DisposedField
	ConfirmedClean      *bool `json:"confirmedClean,omitempty"` // TODO: handle in react
	KnownFruitableField       // TODO: handle in react
	NotesUpdateField
	PermsOnRequest
}

func (upr updateSyringeRequest) baseItem() *LcSyringe {
	return &LcSyringe{}
}

func (upr updateSyringeRequest) reform() reformedRequest[*LcSyringe] {
	return resolvedUpdateSyringeRequest{
		SaleField:           upr.SaleField,
		DisposedField:       upr.DisposedField,
		ConfirmedClean:      upr.ConfirmedClean,
		KnownFruitableField: upr.KnownFruitableField,
		NotesUpdateField:    upr.NotesUpdateField,
		PermsOnRequest:      upr.PermsOnRequest,
	}
}

type resolvedUpdateSyringeRequest struct {
	SaleField
	DisposedField
	ConfirmedClean *bool `json:"confirmedClean,omitempty"`
	KnownFruitableField
	NotesUpdateField
	PermsOnRequest
}

func (req resolvedUpdateSyringeRequest) modsFor(existing *LcSyringe, aclField AclField) (bson.D, error) {
	mds := NewMods()
	updatePointerIfNeeded(mds, "confirmedClean", req.ConfirmedClean, existing.ConfirmedClean)
	return mds.updateSaleIfNeeded(req.Sale, existing.Sale).
		updateDisposedIfNeeded(req, existing).
		updateKnownFruitableIfNeeded(req, existing).
		updateNotesIfNeeded(req, existing).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

// TODO: MOVE
func mainCollIdFromRequest(r *http.Request, w http.ResponseWriter) (b58id Base58Str, id MainCollectionId, err error) {
	var idStr string
	idStr, err = UrlDecodeString(r.PathValue("id"))
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
	b58id, id = mainCollId.AsBase58(), *mainCollId
	return
}
func altCollIdFromRequest(r *http.Request, w http.ResponseWriter) (b58id Base58Str, id AlternateCollectionId, err error) {
	var idStr string
	idStr, err = UrlDecodeString(r.PathValue("id"))
	if err != nil {
		http.Error(w, "failed to url decode altCollId string: "+err.Error(), http.StatusInternalServerError)
		return
	}
	altCollId, err := StandardizeAltCollectionId(idStr)
	if err != nil {
		http.Error(w, "failed to standardize alt collection id: "+err.Error(), http.StatusBadRequest)
		return
	}
	b58id, id = altCollId.AsBase58(), *altCollId
	return
}

func updateSyringeHandler(w http.ResponseWriter, r *http.Request) {
	//mainUpdateHandler[*LcSyringe](w, r, updateSyringeRequest{})
	req := updateSyringeRequest{}
	_, id, err := mainCollIdFromRequest(r, w) // TODO: use this everywhere
	if err != nil {
		return
	}
	if err = ReadSimpleStructuredBody(r, w, &req); err != nil { // TODO: use this everywhere
		return // Writes already if err
	}

	// CHECK THAT ALL NEW PICS EXIST
	// PROCESS ALL NEW PICS AND CONTAMS
	out := req.reform()
	ctx, db := Db(r)
	coll := db.Collection(LcSyringeCollectionName)
	// go get current LcSyringe
	existing := LcSyringe{}
	err = coll.FindOne(ctx, bsonFindFilter("_id", id)).Decode(&existing)
	if err != nil {
		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishMainCollItemUpdate(ctx, w, coll, out.modsFor, &existing, out.GetPermsOnRequest())
}

// TODO: use and move!
func mainUpdateHandler[T MainCollectionItem](w http.ResponseWriter, r *http.Request, req simpleUpdateHandler[T]) {
	_, id, err := mainCollIdFromRequest(r, w) // TODO: use this everywhere
	if err != nil {
		return // Writes already if err
	}
	if err = ReadSimpleStructuredBody(r, w, &req); err != nil { // TODO: use this everywhere
		return // Writes already if err
	}
	// CHECK THAT ALL NEW PICS EXIST
	// PROCESS ALL NEW PICS AND CONTAMS
	out := req.reform()
	existing := req.baseItem()
	ctx, db := Db(r)
	coll := db.Collection(existing.CollectionName())
	// go get current LcSyringe

	err = coll.FindOne(ctx, bsonFindFilter("_id", id)).Decode(existing)
	if err != nil {
		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishMainCollItemUpdate(ctx, w, coll, out.modsFor, existing, out.GetPermsOnRequest())
}

// TODO: MOVE
type simpleUpdateHandler[T CollectionItem] interface {
	baseItem() T
	reform() reformedRequest[T]
}

// TODO: MOVE
type reformedRequest[T CollectionItem] interface {
	modsFor(existing T, aclField AclField) (bson.D, error)
	GetPermsOnRequest() PermsOnRequest
}

type importLcSyringeRequest struct {
	CreationDateField
	SpeciesField
	SubspeciesOptionalField
	KnownFruitableField
	NotesField
	// pic as "img"
	PermsOnRequest
}

// TODO: USE!!!
func importLcSyringeHandler(w http.ResponseWriter, r *http.Request) {
	data := importLcSyringeRequest{}
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
	coll := db.Collection(LcSyringeCollectionName)
	toInsert := LcSyringe{
		MainCollectionIdField:   MainCollectionIdField{Id: id},
		CreationDateField:       data.CreationDateField,
		SpeciesField:            data.SpeciesField,
		KnownFruitableField:     data.KnownFruitableField,
		SubspeciesOptionalField: data.SubspeciesOptionalField,
		NotesField:              data.NotesField,
		LastUpdatedField:        LastUpdatedFieldForNow(),
		AclField:                AclField{finalPerms},
	}
	finishImportMainCollectionEntry(ctx, coll, &toInsert, data.PermsOnRequest, w)
}
