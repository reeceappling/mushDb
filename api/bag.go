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
	"github.com/reeceappling/mushDb/api/request/unix"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"slices"
)

// needed for creating fruits (unless from box or agar)
// TODO: new (subs batch created first, then PC, so they can be referenced)

type Bag struct {
	MainCollectionIdField             `bson:"inline"`
	SubstrateRecipeField              `bson:"inline"`
	SubstrateBatchOptionalField       `bson:"inline"`
	PcRunField                        `bson:"inline"`
	FilterSize                        string `bson:"filterSize" json:"filterSize"`
	CreationDateField                 `bson:"inline"`
	GenerationsFields                 `bson:"inline"`
	SealDate                          *unix.Time      `bson:"sealDate,omitempty" json:"sealDate,omitempty"` // set on transfer in
	WetnessField                      `bson:"inline"` // Initial wetness (refer to scale on field struct)
	KnownFruitableField               `bson:"inline"` // set on transfer in, or once fruited
	SpeciesOptionalField              `bson:"inline"` // set on transfer in
	SubspeciesOptionalField           `bson:"inline"` // set on transfer in
	InnocField                        `bson:"inline"` // Set on transfer in. Innoc from LC or grain jar only
	TransfersOutField                 `bson:"inline"` // Set on transfer out
	MainCollectionOptionalParentField `bson:"inline"` // Set on transfer in
	ParentTypeField                   `bson:"inline"` // (main)lc, plate, or jar only (alt) can come from lcSyringe
	PicsField                         `bson:"inline"` // Updated independently
	ContaminationsField               `bson:"inline"` // Updated independently
	MostRecentImageField              `bson:"inline"`
	FlushesField                      `bson:"inline"` // Updated independently
	SaleField                         `bson:"inline"`
	DisposedField                     `bson:"inline"`

	NotesField       `bson:"inline"` // Updated independently
	LastUpdatedField `bson:"inline"`
	AclField         `bson:"inline"`
}

func (b Bag) Innoculatable() error {
	return errors.Join(
		b.RequireNoSpecies(),
		b.RequireNoSubspecies(),
		b.RequireNotDisposed(),
		b.RequireUnsold(),
		b.RequireUnknownFruitable(),
		b.RequireNoInnoculation())
}

func (b Bag) CanTransferTo(dst geneticSource) error {
	if !slices.Contains([]string{PlateSourceType, BagSourceType, FruitingChamberSourceType, GrainJarSourceType /*BagSourceType, GrainJarSourceType*/}, dst.SourceType()) {
		return errors.New("Bag cannot transfer to " + dst.SourceType())
	}
	return nil
}

func (b Bag) GeneticInfoAsParent() (GeneticParentInfo, error) {
	return GeneticParentInfo{
		SpeciesOptionalField:    b.SpeciesOptionalField,
		SubspeciesOptionalField: b.SubspeciesOptionalField,
		KnownFruitableField:     b.KnownFruitableField,
		GenerationsFields:       b.GenerationsFields,
	}, nil
}

func (b Bag) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return b.GenSinceSpore, b.GenSinceFruitOrSpore
}

func (b Bag) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
	parentInfo, genSpore, genFruitSpore, err := childGensForParent(from)
	if err != nil {
		return err
	}
	upd, err := xfer.
		PicsModsForChild(b).
		Set("sealDate", xfer.LastUpdated).
		withInnoc(xfer).
		withParentType(utils.Pointer(xfer.FromType)).
		withParent(utils.Pointer(from.DbId())).
		withGens(genSpore, genFruitSpore).
		withSpecies(parentInfo.Species).
		withSubspecies(parentInfo.Subspecies).
		withKnownFruitable(parentInfo.KnownFruitable).
		withPerms(from.Permissions()).
		withLastUpdated(xfer.LastUpdated).
		Finalized()
	if err != nil {
		return errors.Join(err, ErrFailedToFinalizeMods)
	}
	res, err := mongo.SessionFromContext(ctx).Client().Database(dbName).Collection(BagsCollectionName).UpdateByID(ctx, b.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer
	}
	return nil
}

func (b Bag) id() []byte {
	return []byte(b.Id.dbIdStr())
}

func initializeBags(ctx context.Context) error {
	db := DbFrom(ctx)
	coll := db.Collection(BagsCollectionName)
	// Indices
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		//newSimpleIndex("recipe", "recipe", false, false, false),
		//newSimpleIndex("substrateBatch", "substrateBatch", false, true, false),
		//newSimpleIndex("pcRun", "pcRun", false, true, false),
		//// TODO: filter size?
		//newSimpleIndex("creationDate", "creationDate", true, false, false),
		//newSimpleIndex("genSinceSpore", "genSpore", true, true, false),
		//newSimpleIndex("genSinceFruitOrSpore", "genFruitOrSpore", true, true, false),
		//newSimpleIndex("sealDate", "sealDate", true, true, false), // BAG ONLY
		//// TODO: wetness
		//newSimpleIndex("knownFruitable", "knownFruitable", false, true, false),
		//newSimpleIndex("species", "species", false, false, false),
		//newSimpleIndex("subspecies", "subspecies", false, true, false),
		//newSimpleIndex("innoc", "innoc", false, true, false),
		//newSimpleIndex("transfersOut", "transfersOut", false, true, false),
		//newSimpleIndex("parent", "parent", false, true, false),
		//newSimpleIndex("parentType", "parentType", false, true, false),
		////pics
		////TODO: contams?
		////flushes
		// TODO: substrate recipe???
		// TODO: substrate batch???
		//newSimpleIndex("sale", "sale", false, true, false),
		//newSimpleIndex("disposed", "disposed", false, true, false),
		//notes
		projectsIndexModel,
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// Add test entries if dev
	return env.IfNotProd(ctx, func() error {
		// If test agar batch does not exist, then create it
		testId := mainCollIdForint(idTestBag)
		testItem := &Bag{
			MainCollectionIdField:       MainCollectionIdField{testId},
			SubstrateRecipeField:        SubstrateRecipeField{exAltId},
			SubstrateBatchOptionalField: SubstrateBatchOptionalField{SubstrateBatch: utils.Pointer(altCollIdForint(idWoodPellets))},
			PcRunField:                  PcRunField{exAltId},
			FilterSize:                  filterSizetwoMic,
			CreationDateField:           CreationDateField{exampleTime},
			GenerationsFields: GenerationsFields{
				GenSporeField:        GenSporeField{&exGenSinceSpore},
				GenSinceFruitOrSpore: &exGenSinceFruitSpore,
			},
			SealDate:                          &exampleTime,
			KnownFruitableField:               KnownFruitableField{exBool},
			SpeciesOptionalField:              SpeciesOptionalField{&testEntryStringId},
			SubspeciesOptionalField:           SubspeciesOptionalField{&testEntryStringId},
			InnocField:                        InnocField{&exAltId},
			TransfersOutField:                 TransfersOutField{exAlts},
			ParentTypeField:                   ParentTypeField{&exParentType},
			MainCollectionOptionalParentField: MainCollectionOptionalParentField{&exPlate},
			PicsField:                         PicsField{exPics},
			ContaminationsField:               ContaminationsField{exContams},
			MostRecentImageField:              MostRecentImageField{&exPics[0]},
			FlushesField:                      FlushesField{exPics},
			SaleField:                         SaleField{&exAltId},
			DisposedField:                     DisposedField{&exampleTime},
			NotesField:                        NotesField{exampleNotes()},
			LastUpdatedField:                  LastUpdatedField{exampleTime},
			AclField:                          AclField{testAcl},
		}
		return addTestMainEntries(ctx, testItem)
	})
}

const (
	filterSizetwoMic       = "0.2 micron"
	filterSizeTwentyTwoMic = "0.22 micron"
	filterSizeUnknown      = "unknown"
)

var bagFilterSizes = map[string]string{
	filterSizetwoMic:       "Average large bag with filter", // Avg large bag // TODO; figure out which sizes apply to which bags
	filterSizeTwentyTwoMic: "Average filter patches",        // Avg patches
	//"0.3 micron",
	//"0.45 micron",
	//"0.5 micron",
	//"5 micron": "Airy large bags (do not have)", // TODO: Airy large bag?
	filterSizeUnknown: "monotub filter patches, etc",
}

type createBagRequest struct {
	SubstrateBatchField
	WetnessField
	PcRunField // Bags cannot be created without a PC run.
	FilterSize string
	NotesField
	WriteTagToField
}

func createBagHandler(w http.ResponseWriter, r *http.Request) {
	data := createBagRequest{}
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
	ctx := r.Context()
	// Validate request
	_, err = data.PcRunField.Get(ctx)
	if err != nil {
		http.Error(w, "PcRun validation failure: "+err.Error(), http.StatusBadRequest)
		return
	}
	batch, err := data.SubstrateBatchField.Get(ctx)
	if err != nil {
		http.Error(w, "Substrate batch validation failure: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err = data.WetnessField.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// validate filter size
	if _, exists := bagFilterSizes[data.FilterSize]; !exists {
		http.Error(w, "filter size validation failure", http.StatusBadRequest)
		return
	}
	// Denying guest edits is done in the upper handlers
	ctx, now := request.UnixTime(ctx)
	toInsert := &Bag{
		MainCollectionIdField:       MainCollectionIdField{id},
		SubstrateRecipeField:        batch.SubstrateRecipeField,
		SubstrateBatchOptionalField: data.SubstrateBatchField.asOptional(),
		WetnessField:                data.WetnessField,
		PcRunField:                  PcRunField{data.PcRun},
		FilterSize:                  data.FilterSize,
		CreationDateField:           CreationDateField{now},
		NotesField:                  data.NotesField,
		LastUpdatedField:            LastUpdatedField{now},
		AclField:                    allCanWriteAcl(),
	}
	err = writeRfidTagIfNecessary(ctx, data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	finishCreateMainCollectionEntry(ctx, toInsert, w)
}

type updateBagRequest struct {
	KnownFruitableField
	//SaleField
	DisposedField
	WetnessField
	NotesUpdateField
	ImagesUpdateField  //"newPic-1"
	ContamsUpdateField //"newContam-1"
	FlushesUpdateField //"newFlush-1"
	PermsOnRequest     `json:"acl"`
}

func (upr updateBagRequest) reform() resolvedUpdateBagRequest {
	return resolvedUpdateBagRequest{
		KnownFruitableField: upr.KnownFruitableField,
		//SaleField:           upr.SaleField,
		WetnessField:     upr.WetnessField,
		DisposedField:    upr.DisposedField,
		NotesUpdateField: upr.NotesUpdateField,
		Images:           imageUpdates(upr.Images),
		Contams:          contamUpdates(upr.Contams),
		Flushes:          imageUpdates(upr.Flushes),
		PermsOnRequest:   upr.PermsOnRequest,
	}
}

type resolvedUpdateBagRequest struct {
	KnownFruitableField
	//SaleField
	DisposedField
	NotesUpdateField
	WetnessField
	Images         SplitEntries[picWithNotesForm, PicWithNotes]
	Contams        SplitEntries[contamForm, Contamination]
	Flushes        SplitEntries[picWithNotesForm, PicWithNotes]
	PermsOnRequest `json:"acl"`
}

func (req resolvedUpdateBagRequest) modsFor(existing *Bag, aclField AclField) (bson.D, error) {
	return NewMods().
		updateKnownFruitableIfNeeded(req, existing).
		//updateSaleIfNeeded(req.Sale, existing.Sale).
		updateWetnessIfNeeded(req.Wetness, existing.Wetness).
		updateDisposedIfNeeded(req, existing).
		updateNotesIfNeeded(req, existing).
		updatePicsIfNeeded(req.Images, existing.Pics).
		updateContamsIfNeeded(req.Contams, existing.Contaminations).
		updateFlushesIfNeeded(req.Flushes, existing.Flushes).
		updateMostRecentImageIfNeeded(existing.MostRecentImage, loadMriPics(&req.Images, &req.Contams, &req.Flushes)).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

//func getBag(ctx context.Context, id MainCollectionId) (*Bag, error) {
//	// go get current plate
//	existing := &Bag{}
//	err := DbFrom(ctx).Collection(BagsCollectionName).FindOne(ctx, BsonFindFilter(IDfld, id)).Decode(existing)
//	return existing, err
//}

func updateBagHandler(w http.ResponseWriter, r *http.Request) {
	data := updateBagRequest{}
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
	newPics, newContams, newFlushes, err := fullMultipartWithNoBreaks(w, r, &data, mainCollId.AsBase58())
	if err != nil {
		// Already wrote
		return
	}

	// CHECK THAT ALL NEW PICS EXIST
	// PROCESS ALL NEW PICS AND CONTAMS
	out := data.reform()
	for i, _ := range data.Images.New {
		loc, exists := newPics[i]
		if !exists {
			http.Error(w, fmt.Sprintf("error, location for new picture index %d not found (should never happen)", i), http.StatusInternalServerError)
			return
		}
		out.Images.New[i].Location = ImageLocation(loc)
	}
	for i, _ := range data.Contams.New {
		if loc, exists := newContams[i]; exists {
			finalLoc := ImageLocation(loc)
			out.Contams.New[i].Location = &finalLoc
		}
	}
	for i, _ := range data.Flushes.New {
		loc, exists := newFlushes[i]
		if !exists {
			http.Error(w, fmt.Sprintf("error, location for new flush index %d not found (should never happen)", i), http.StatusInternalServerError)
			return
		}
		out.Flushes.New[i].Location = ImageLocation(loc)
	}
	ctx := r.Context()
	coll := DbFrom(ctx).Collection(BagsCollectionName)
	existing := &Bag{}
	err = coll.FindOne(ctx, BsonFindFilter(IDfld, *mainCollId)).Decode(existing)
	if err != nil {
		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishMainCollItemUpdate(ctx, w, out.modsFor, existing, data.PermsOnRequest)
}

type importBagRequest struct {
	CreationDateField
	SubstrateRecipeField
	FilterSize string
	SpeciesOptionalField
	SubspeciesOptionalField
	Generation *Generation // required when innoculated
	KnownFruitableField
	WriteTagToField
	// image as "img"
}

func importBagHandler(w http.ResponseWriter, r *http.Request) {
	data := importBagRequest{}
	id := NextMainCollectionId()
	b58id := id.AsBase58()
	r.Body = http.MaxBytesReader(w, r.Body, maxMultipartRequestSize)
	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "unable to open multipart reader: "+err.Error(), http.StatusBadRequest)
		return
	}
	p, err := reader.NextPart()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	picsSaved := []string{}
	defer func() {
		if err != nil {
			err = pics.DeleteFiles(r.Context(), picsSaved...)
			if err != nil {
				handleFileDeleteErr(err)
			}
		}
	}()
	var importedPic *PicWithNotes = nil
	dataProcessed := false
	filesProcessed := 0
	ctx, now := request.UnixTime(r.Context())
	for { // TODO: FIX THIS MULTIPART READER? Unconfirmed that this even needs fixing as of 6/5/26
		fileName := p.FileName()
		defer p.Close()
		if isFile := fileName != ""; isFile {
			if filesProcessed == 1 {
				http.Error(w, "only allowed to create 1 image on import: "+err.Error(), http.StatusBadRequest)
				return
			}
			// Process file
			fieldBytes, err := multipartToImageBytes(p, w)
			if err != nil {
				// Already wrote
				return
			}
			newFileNameWithPrefixPath, errr := pics.SaveFile(ctx, fieldBytes, "bag", string(b58id), "img")
			if errr != nil {
				err = errr
				http.Error(w, "failed to save file: "+err.Error(), http.StatusBadRequest)
				return
			}
			picsSaved = append(picsSaved, newFileNameWithPrefixPath)
			importedPic = &PicWithNotes{
				PicWithNotesLessLocation: newPicWithNotesLessLocation(now, []Note{}),
				Location:                 ImageLocation(newFileNameWithPrefixPath),
			}
			filesProcessed++
		} else {
			// Process text (or object)
			bs, errr := io.ReadAll(p)
			if errr != nil {
				err = errr
				http.Error(w, "unable to read Data from form: "+err.Error(), http.StatusBadRequest)
				return
			}
			// PARSE INTO CORRECT DATA FORMAT
			err = json.Unmarshal(bs, &data)
			if err != nil {
				http.Error(w, "unable to unmarshal json form Data: "+err.Error(), http.StatusBadRequest)
				return
			}
			dataProcessed = true
		}

		// Go to next part or break
		p, err = reader.NextPart()
		if err != nil {
			if err != io.EOF {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			break
		}
	}

	if !dataProcessed {
		err = errors.New("no non-image Data found on form request")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var gen *Generation = nil
	if data.Species != nil {
		if data.Generation == nil {
			http.Error(w, "innoculated must have generation: "+err.Error(), http.StatusBadRequest)
			return
		}
		if *data.Generation < 1 {
			http.Error(w, "gen must be positive", http.StatusBadRequest)
			return
		}
		gen = data.Generation
	} else {
		data.KnownFruitable = nil
		data.Subspecies = nil
	}

	pix := []PicWithNotes{}
	if importedPic != nil {
		pix = []PicWithNotes{*importedPic}
	}
	// Validate
	_, err = data.SubstrateRecipeField.Get(ctx)
	if err != nil {
		dbErr(w, "substrate recipe retrieval error for recipe "+string(data.Substrate.AsBase58())+": "+err.Error(), http.StatusInternalServerError)
		return
	}

	var finalPerms ACL
	innoculated := data.Species != nil
	if !innoculated {
		finalPerms = allCanWriteAcl().ACL
	} else {
		finalPerms, err = ImportFinalPerms(r.Context(), *data.Species, data.Subspecies)
		if err != nil {
			http.Error(w, "failed to get species and/or subspecies: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Write
	toInsert := &Bag{
		MainCollectionIdField:   MainCollectionIdField{id},
		SubstrateRecipeField:    data.SubstrateRecipeField,
		PcRunField:              PcRunField{impPcRun}, // imported id for pc run
		FilterSize:              data.FilterSize,
		CreationDateField:       data.CreationDateField,
		GenerationsFields:       GenerationsFieldFor(gen),
		SealDate:                &data.CreationDate,
		KnownFruitableField:     data.KnownFruitableField,
		SpeciesOptionalField:    SpeciesOptionalField{data.Species},
		SubspeciesOptionalField: data.SubspeciesOptionalField,
		PicsField:               PicsField{pix},
		ContaminationsField:     ContaminationsField{},
		MostRecentImageField:    MostRecentImageField{importedPic},
		FlushesField:            FlushesField{},
		SaleField:               SaleField{},
		DisposedField:           DisposedField{},
		NotesField:              NotesField{},
		LastUpdatedField:        LastUpdatedField{now},
		AclField:                finalPerms.AsField(),
	}
	err = writeRfidTagIfNecessary(ctx, data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	finishImportMainCollectionEntry(ctx, toInsert, w)
}

//func deleteMainCollectionItemHandler[T MainCollectionItem](w http.ResponseWriter, r *http.Request) {
//	idStr := r.PathValue("id")
//	if idStr == "" {
//		http.Error(w, "Empty id for delete request", http.StatusBadRequest)
//		return
//	}
//	id, err := Base58Str(idStr).ToMainCollectionId()
//	if err != nil {
//		http.Error(w, "Invalid ID to delete: "+err.Error(), http.StatusBadRequest)
//		return
//	}
//	// Validate not used in other places...
//	ctx := r.Context()
//	// ensure item does not have any transfers in or out
//	var tempItem T // TODO: ensure works ok
//	item, err := GetMainCollectionItemSpecific[T](ctx, id, tempItem)
//	if err != nil {
//		if errors.Is(err, mongo.ErrNoDocuments) {
//http.Error(w, "Item to be deleted not found! Should never happen!: "+err.Error(), http.StatusNotFound)
//		} else {
//			http.Error(w, "Failed to retrieve item to be deleted: "+err.Error(), http.StatusInternalServerError)
//		}
//		return
//	}
//	if item.Parent != nil { // TODO: FIX THIS!
//		// TODO: what if we want to remove it from the parent as well?
//		http.Error(w, "Cannot delete innoculated items!", http.StatusExpectationFailed)
//		return
//	}
//	if item.TransfersOut != nil && len(item.TransfersOut) > 0 {
//		http.Error(w, "Cannot delete items with transfers out", http.StatusExpectationFailed)
//		return
//	}
//
//	// Delete if not found elsewhere!
//	DeleteCollectionItem(ctx, item.CollectionName(), id, w)
//}

//func deleteBagHandler(w http.ResponseWriter, r *http.Request) {
//	//deleteMainCollectionItemHandler[*Bag](w, r) // TODO; if this works, use it everywhere else!
//	idStr := r.PathValue("id")
//	if idStr == "" {
//		http.Error(w, "Empty id for delete request", http.StatusBadRequest)
//		return
//	}
//	id, err := Base58Str(idStr).ToMainCollectionId()
//	if err != nil {
//		http.Error(w, "Invalid ID to delete: "+err.Error(), http.StatusBadRequest)
//		return
//	}
//	// Validate not used in other places...
//	ctx := r.Context()
//	// ensure item does not have any transfers in or out
//	item, err := GetMainCollectionItemSpecific[*Bag](ctx, id, &Bag{})
//	if err != nil {
//		if errors.Is(err, mongo.ErrNoDocuments) {
//			http.Error(w, "Item to be deleted not found! Should never happen!: "+err.Error(), http.StatusNotFound)
//		} else {
//			http.Error(w, "Failed to retrieve item to be deleted: "+err.Error(), http.StatusInternalServerError)
//		}
//		return
//	}
//	if item.Parent != nil {
//		// TODO: what if we want to remove it from the parent as well?
//		http.Error(w, "Cannot delete innoculated items!", http.StatusExpectationFailed)
//		return
//	}
//	if item.TransfersOut != nil && len(item.TransfersOut) > 0 {
//		http.Error(w, "Cannot delete items with transfers out", http.StatusExpectationFailed)
//		return
//	}
//
//	// Delete if not found elsewhere!
//	DeleteCollectionItem(ctx, item.CollectionName(), id, w)
//}
