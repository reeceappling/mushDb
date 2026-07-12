package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/reeceappling/goUtils/v2/utils"
	"github.com/reeceappling/mushDb/api/env"
	"github.com/reeceappling/mushDb/api/pics"
	"github.com/reeceappling/mushDb/api/request"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"slices"
)

// needed for
// xfers

// TODO: newFromLC

type LcSyringe struct {
	// TODO: ADD PICTURES!
	MainCollectionIdField `bson:"inline"`
	// Parent is always either purchased (nil), LC, or LcSyringe
	MainCollectionOptionalParentField `bson:"inline"` // won't exist for imported
	MostRecentImageField              `bson:"inline"`
	CreationDateField                 `bson:"inline"` // create or receive date
	SpeciesField                      `bson:"inline"`
	SubspeciesOptionalField           `bson:"inline"`
	SaleField                         `bson:"inline"`
	GenerationsFields                 `bson:"inline"`
	KnownFruitableField               `bson:"inline"`
	ConfirmedCleanField               `bson:"inline"`
	PicsField                         `bson:"inline"`
	TransfersOutField                 `bson:"inline"`
	DisposedField                     `bson:"inline"`
	NotesField                        `bson:"inline"`
	LastUpdatedField                  `bson:"inline"`
	AclField                          `bson:"inline"`
}

func (lcs LcSyringe) Innoculatable() error {
	return errors.New("lcSyringes never innoculatable")
}

func (lcs LcSyringe) CanTransferTo(dst geneticSource) error {
	if !slices.Contains([]string{BagSourceType, GrainJarSourceType, LcSourceType, PlateSourceType, PlugSourceType, SlantSourceType, FruitingChamberSourceType}, dst.SourceType()) {
		return errors.New("lc syringe cannot transfer to " + dst.SourceType())
	}
	return nil
}

//func (sw LcSyringe) setTransferParent(ctx context.Context, xfer Transfer) error {
//	// TODO: can this even happen? // TODO: YES IT CAN! REVAMP!
//	coll := DbFrom(ctx).Collection(sw.CollectionName())
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
		KnownFruitableField:     sw.KnownFruitableField,
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
	coll := DbFrom(ctx).Collection(LcSyringeCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		creationDateIndexModel,
		//newSimpleIndex("parent", "parent", false, true, false),
		newSimpleIndex("species", "species", false, false, false),
		newSimpleIndex("subspecies", "subspecies", false, true, false),
		//saleIndexModel,
		//newSimpleIndex("genSinceSpore", "genSpore", true, true, false),
		//newSimpleIndex("genSinceFruitOrSpore", "genFruitOrSpore", true, true, false),
		//newSimpleIndex("knownFruitable", "knownFruitable", false, true, false),
		//newSimpleIndex("confirmedClean", "confirmedClean", false, true, false),
		//transfersOutIndexModel,
		//newSimpleIndex("disposed", "disposed", false, true, false),
		projectsIndexModel, // TODO: ensure ok and add on many other collections
		//Notes (no index unless tags)
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	return env.IfNotProd(ctx, func() error {
		// If test agar batch does not exist, then create it
		testItem := &LcSyringe{
			MainCollectionIdField:             MainCollectionIdField{Id: exLCS},
			MainCollectionOptionalParentField: MainCollectionOptionalParentField{&exLC},
			MostRecentImageField:              MostRecentImageField{MostRecentImage: &exPics[0]}, // TODO: ok?
			PicsField:                         PicsField{exPics},                                 // TODO: ok?
			CreationDateField:                 CreationDateField{exampleTime},
			SpeciesField:                      SpeciesField{testEntryStringId},
			SubspeciesOptionalField:           SubspeciesOptionalField{&testEntryStringId},
			SaleField:                         SaleField{&exAltId},
			DisposedField:                     DisposedField{&exampleTime},

			NotesField:       NotesField{exampleNotes()},
			LastUpdatedField: LastUpdatedField{exampleTime},
			AclField:         allCanWriteAcl(),
		}
		return addTestMainEntries(ctx, testItem)
	})
}

type createLCSyringeRequest struct {
	LC MainCollectionId `json:"parent"`
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
	ctx, db := Db(r)
	// Validate inputs and grab parent
	parent := &LiquidCulture{}
	err = db.Collection(LCCollectionName).FindOne(ctx, BsonFindFilter(IDfld, data.LC)).Decode(parent)
	if err != nil {
		dbErr(w, "failed to get parent LC: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if parent.Species == nil {
		dbErr(w, "Parent LC must be innoculated", http.StatusInternalServerError)
		return
	}
	// TODO: CREATE PARENT TRANSFER? maybe not
	ctx, now := request.UnixTime(r.Context()) // TODO: no more r.Context below
	toInsert := LcSyringe{
		MainCollectionIdField:             MainCollectionIdField{Id: id},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&data.LC},
		ConfirmedCleanField:               parent.ConfirmedCleanField,
		KnownFruitableField:               parent.KnownFruitableField,
		CreationDateField:                 CreationDateField{now},
		SpeciesField:                      SpeciesField{Species: *parent.Species},
		SubspeciesOptionalField:           parent.SubspeciesOptionalField,
		GenerationsFields:                 parent.GenerationsFields,
		NotesField:                        NotesField{data.Notes},
		LastUpdatedField:                  LastUpdatedField{now},
		// Do not check permissions, just pass parent perms to child
		AclField: parent.AclField,
	}
	err = writeRfidTagIfNecessary(ctx, data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	finishCreateMainCollectionEntry(ctx, &toInsert, w)
}

type updateSyringeRequest struct {
	SaleField // TODO: validate?
	DisposedField
	ConfirmedClean      *bool `json:"confirmedClean,omitempty"` // TODO: handle in react
	KnownFruitableField       // TODO: handle in react
	ImagesUpdateField
	NotesUpdateField
	PermsOnRequest `json:"acl"`
}

func (upr updateSyringeRequest) baseItem() *LcSyringe {
	return &LcSyringe{}
}

func (upr updateSyringeRequest) reform() resolvedUpdateSyringeRequest {
	return resolvedUpdateSyringeRequest{
		SaleField:           upr.SaleField,
		DisposedField:       upr.DisposedField,
		ConfirmedClean:      upr.ConfirmedClean,
		KnownFruitableField: upr.KnownFruitableField,
		NotesUpdateField:    upr.NotesUpdateField,
		Images:              imageUpdates(upr.Images),
		PermsOnRequest:      upr.PermsOnRequest,
	}
}

type resolvedUpdateSyringeRequest struct {
	SaleField
	DisposedField
	ConfirmedClean *bool `json:"confirmedClean,omitempty"`
	KnownFruitableField
	Images SplitEntries[picWithNotesForm, PicWithNotes]
	NotesUpdateField
	PermsOnRequest `json:"acl"`
}

func (req resolvedUpdateSyringeRequest) modsFor(existing *LcSyringe, aclField AclField) (bson.D, error) {
	mds := NewMods()
	imagesForUpdateFunc := []PicWithNotes{}
	for _, ex := range req.Images.Existing {
		if !ex.Disabled {
			imagesForUpdateFunc = append(imagesForUpdateFunc, ex.Data.convert())
		}
	}
	imagesForUpdateFunc = append(imagesForUpdateFunc, req.Images.New...) // TODO: is this backwards???
	return updatePointerIfNeeded(mds, "confirmedClean", req.ConfirmedClean, existing.ConfirmedClean).
		updateSaleIfNeeded(req.Sale, existing.Sale).
		updateDisposedIfNeeded(req, existing).
		updateKnownFruitableIfNeeded(req, existing).
		updatePicsIfNeeded(req.Images, existing.Pics).
		updateMostRecentImageIfNeeded(existing.MostRecentImage, imagesForUpdateFunc).
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
	data := updateSyringeRequest{}
	b58Id, id, err := mainCollIdFromRequest(r, w)
	if err != nil {
		return
	}
	newPics, _, _, err := fullMultipartWithNoBreaks(w, r, "lcSyringe", &data, b58Id)
	if err != nil {
		// Already wrotw
		return
	}
	out := data.reform()
	for i, _ := range data.Images.New {
		loc, exists := newPics[i]
		if !exists {
			http.Error(w, fmt.Sprintf("error, location for new picture index %d not found (should never happen)", i), http.StatusInternalServerError)
			return
		}
		out.Images.New[i].Location = ImageLocation(loc)
	}
	finalReqBs, err := json.MarshalIndent(out, "", " ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	println("REQUEST BYTES: ", string(finalReqBs)) // TODO: del
	ctx := r.Context()
	client := GetMongoClient(ctx)
	coll := client.Database(dbName).Collection(LcSyringeCollectionName)
	existing := &LcSyringe{}
	err = coll.FindOne(ctx, BsonFindFilter(IDfld, id)).Decode(existing)
	if err != nil {
		// TODO: an issue here?
		http.Error(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishMainCollItemUpdate(ctx, w, out.modsFor, existing, out.PermsOnRequest)

	//req := updateSyringeRequest{}
	//_, id, err := mainCollIdFromRequest(r, w)
	//if err != nil {
	//	return
	//}
	//if err = ReadSimpleStructuredBody(r, w, &req); err != nil { // TODO: use this everywhere
	//	return // Writes already if err
	//}
	//
	//// CHECK THAT ALL NEW PICS EXIST
	//// PROCESS ALL NEW PICS AND CONTAMS
	//out := req.reform()
	//ctx, db := Db(r)
	//coll := db.Collection(LcSyringeCollectionName)
	//// go get current LcSyringe
	//existing := LcSyringe{}
	//err = coll.FindOne(ctx, BsonFindFilter(IDfld, id)).Decode(&existing)
	//if err != nil {
	//	dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
	//	return
	//}
	//
	////_, err = newTxn(ctx, func(sessCtx mongo.SessionContext) (any, error){
	////	if req.ConfirmedClean != nil {
	////		if existing.ConfirmedClean == nil || (*req.ConfirmedClean != *existing.ConfirmedClean){
	////			// TODO: ALSO CHANGE PARENT!!!!!
	////		}
	////	}
	////	return finishMainCollItemUpdateInTxn(sessCtx, w, out.modsFor, &existing, out.GetPermsOnRequest())
	////})
	//
	//finishMainCollItemUpdate(ctx, w, out.modsFor, &existing, out.GetPermsOnRequest())
}

//// TODO: use and move!
//func mainUpdateHandler[T MainCollectionItem](w http.ResponseWriter, r *http.Request, req simpleUpdateHandler[T]) {
//	_, id, err := mainCollIdFromRequest(r, w) // TODO: use this everywhere
//	if err != nil {
//		return // Writes already if err
//	}
//	if err = ReadSimpleStructuredBody(r, w, &req); err != nil { // TODO: use this everywhere
//		return // Writes already if err
//	}
//	// CHECK THAT ALL NEW PICS EXIST
//	// PROCESS ALL NEW PICS AND CONTAMS
//	out := req.reform()
//	existing := req.baseItem()
//	ctx, db := Db(r)
//	coll := db.Collection(existing.CollectionName())
//	// go get current LcSyringe
//
//	err = coll.FindOne(ctx, BsonFindFilter(IDfld, id)).Decode(existing)
//	if err != nil {
//		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
//		return
//	}
//	finishMainCollItemUpdate(ctx, w, out.modsFor, existing, out.GetPermsOnRequest())
//}

//// TODO: MOVE
//type simpleUpdateHandler[T CollectionItem] interface {
//	baseItem() T
//	reform() reformedRequest[T]
//}

// TODO: MOVE, maybe delete
type reformedRequest[T CollectionItem] interface {
	modsFor(existing T, aclField AclField) (bson.D, error)
	GetPermsOnRequest() PermsOnRequest
}

type importLcSyringeRequest struct {
	CreationDateField
	SpeciesField
	SubspeciesOptionalField
	KnownFruitableField
	ConfirmedCleanField
	Generation Generation // TODO: make required!
	NotesField
	// pic as "img"
	WriteTagToField
}

func importLcSyringeHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctx, now := request.UnixTime(r.Context()) // TODO: no more r.Context below
	data, id := importLcSyringeRequest{}, NextMainCollectionId()
	b58id := id.AsBase58()
	reader, err := multipartReaderForRequest(r.WithContext(ctx), w, &data)
	if err != nil {
		println("failed in multipart reader area: " + err.Error()) // TODO: del
		// Already written
		return
	}
	// Try to get pic if exists
	picsSaved := []string{}
	defer func() {
		if err != nil {
			err = pics.DeleteFiles(ctx, picsSaved...)
			if err != nil {
				handleFileDeleteErr(err)
			}
		}
	}()
	// Go to next part, if exists to get image
	var importedPic *PicWithNotes = nil
	p, err := reader.NextPart()
	if err != nil {
		if err != io.EOF {
			println("failed in nextPart: " + err.Error()) // TODO: del
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		fileName := p.FileName()
		defer p.Close()
		if fileName != "img" {
			println("invalid image name: " + fileName) // TODO: del
			http.Error(w, "invalid image name: "+fileName, http.StatusBadRequest)
			return
		}
		// Process file
		fieldBytes, err := multipartToImageBytes(p, w)
		if err != nil {
			// Already wrote
			println("failed in multipartToImageBytes") // TODO: del
			return
		}
		newFileNameWithPrefixPath, errr := pics.SaveFile(ctx, fieldBytes, "lcSyringe", string(b58id), "img")
		if errr != nil {
			err = errr
			println("failed to save file: " + err.Error()) // TODO: del
			http.Error(w, "failed to save file: "+err.Error(), http.StatusBadRequest)
			return
		}
		picsSaved = append(picsSaved, newFileNameWithPrefixPath)
		importedPic = utils.Pointer(newPicWithNotes(now, []Note{}, ImageLocation(newFileNameWithPrefixPath)))
	}
	if data.Generation < 1 {
		http.Error(w, "generation cannot be <=0 for a non-spore import", http.StatusBadRequest)
		return
	}

	finalPerms, err := ImportFinalPerms(r.Context(), data.Species, data.Subspecies)
	if err != nil {
		http.Error(w, "failed to get species and/or subspecies: "+err.Error(), http.StatusInternalServerError)
		return
	}
	pix := []PicWithNotes{}
	if importedPic != nil {
		pix = []PicWithNotes{*importedPic}
	}
	toInsert := LcSyringe{
		MainCollectionIdField: MainCollectionIdField{Id: id},
		MostRecentImageField:  MostRecentImageField{importedPic},
		PicsField:             PicsField{pix},
		CreationDateField:     data.CreationDateField,
		GenerationsFields: GenerationsFields{
			GenSporeField: GenSporeField{
				GenSinceSpore: &data.Generation,
			},
			GenSinceFruitOrSpore: &data.Generation,
		},
		SpeciesField:            data.SpeciesField,
		ConfirmedCleanField:     data.ConfirmedCleanField,
		KnownFruitableField:     data.KnownFruitableField,
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

	//var gen *Generation = nil
	//if data.Species != nil {
	//	if data.Generation == nil {
	//		println("innoculated must have generation") // TODO: del
	//		http.Error(w, "innoculated must have generation", http.StatusBadRequest)
	//		return
	//	}
	//	gen =
	//	if data.Generation < 1 {
	//		println("gen must be positive") // TODO: del
	//		http.Error(w, "gen must be positive", http.StatusBadRequest)
	//		return
	//	}
	//
	//	// If innoculated, ensure seal and pour condensation coverage are the same
	//	condensCovSealed = data.CondensationCoverageAtPourTime
	//} else {
	//	data.KnownFruitable = nil
	//	data.Subspecies = nil
	//}
	//pix := []PicWithNotes{}
	//if importedPic != nil {
	//	pix = []PicWithNotes{*importedPic}
	//}
	//
	//var finalPerms ACL
	//if data.Species == nil { // Not innoculated
	//	finalPerms = allCanWriteAcl().ACL
	//	data.PourCoverage = nil
	//} else {
	//	finalPerms, err = ImportFinalPerms(ctx, *data.Species, data.Subspecies)
	//	if err != nil {
	//		println("failed to get species and/or subspecies: " + err.Error()) // TODO: del
	//		http.Error(w, "failed to get species and/or subspecies: "+err.Error(), http.StatusInternalServerError)
	//		return
	//	}
	//}
	//
	//toInsert := LcSyringe{
	//	MainCollectionIdField:               MainCollectionIdField{id},
	//	CreationDateField:                   data.CreationDateField,
	//	SubspeciesOptionalField:             data.SubspeciesOptionalField,
	//	GenerationsFields:                   GenerationsFieldFor(data.Generation),
	//	PicsField:                           PicsField{pix},
	//	KnownFruitableField:                 data.KnownFruitableField,
	//	MostRecentImageField:                MostRecentImageField{importedPic},
	//	LastUpdatedField:                    LastUpdatedField{now},
	//	AclField:                            AclField{finalPerms},
	//}
	//err = writeRfidTagIfNecessary(ctx, data.WriteTagTo, id)
	//if err != nil {
	//	println("failed to write tag: " + err.Error()) // TODO: del
	//	http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//println("trying to import the plate...")
	//finishImportMainCollectionEntry(ctx, &toInsert, w)
	//
	//
	//
	//
	//data := importLcSyringeRequest{}
	//id := NextMainCollectionId()
	//defer r.Body.Close()
	//// Process text (or object)
	//bs, err := io.ReadAll(r.Body)
	//if err != nil {
	//	http.Error(w, "unable to read Data from form: "+err.Error(), http.StatusBadRequest)
	//	return
	//}
	//// PARSE INTO CORRECT DATA FORMAT
	//err = json.Unmarshal(bs, &data)
	//if err != nil {
	//	http.Error(w, "unable to unmarshal json form Data: "+err.Error(), http.StatusBadRequest)
	//	return
	//}
	//if data.Generation < 1 {
	//	http.Error(w, "generation cannot be <=0 for a non-spore import", http.StatusBadRequest)
	//	return
	//}
	//
	//finalPerms, err := ImportFinalPerms(r.Context(), data.Species, data.Subspecies)
	//if err != nil {
	//	http.Error(w, "failed to get species and/or subspecies: "+err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//
	//ctx, now := request.UnixTime(r.Context())
	//toInsert := LcSyringe{
	//	MainCollectionIdField: MainCollectionIdField{Id: id},
	//	CreationDateField:     data.CreationDateField,
	//	GenerationsFields: GenerationsFields{
	//		GenSporeField: GenSporeField{
	//			GenSinceSpore: &data.Generation,
	//		},
	//		GenSinceFruitOrSpore: &data.Generation,
	//	},
	//	SpeciesField:            data.SpeciesField,
	//	ConfirmedCleanField:     data.ConfirmedCleanField,
	//	KnownFruitableField:     data.KnownFruitableField,
	//	SubspeciesOptionalField: data.SubspeciesOptionalField,
	//	NotesField:              data.NotesField,
	//	LastUpdatedField:        LastUpdatedField{now},
	//	AclField:                AclField{finalPerms},
	//}
	////err = writeRfidTagIfNecessary(ctx, data.WriteTagTo, id) // TODO: this should always only occur right before the true writes
	////if err != nil {
	////	http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
	////	return
	////}
	//finishImportMainCollectionEntry(ctx, &toInsert, w)
}

func deleteLcSyringeHandler(w http.ResponseWriter, r *http.Request) {
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
	item, err := GetMainCollectionItemSpecific[*LcSyringe](ctx, id, &LcSyringe{})
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
