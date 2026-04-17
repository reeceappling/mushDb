package rfid

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/reeceappling/goUtils/v2/utils"
	"github.com/reeceappling/mushDb/rfid/pics"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
)

type Bag struct {
	MainCollectionIdField       `bson:"inline"`
	SubstrateRecipeField        `bson:"inline"`
	SubstrateBatchOptionalField `bson:"inline"`
	PcRunOptionalField          `bson:"inline"` // this may not exist for pre-existing bags
	//Size string // TODO: unsure what to do here
	FilterSize                        string `bson:"filterSize" json:"filterSize"`
	CreationDateField                 `bson:"inline"`
	GenerationsFields                 `bson:"inline"`
	SealDate                          *unixTime       `bson:"sealDate,omitempty" json:"sealDate,omitempty"` // set on transfer in
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

func (b Bag) CanTransferTo(dst geneticSource) error {
	return errors.New("Bag cannot be transferred (unsure if this is ok)")
	// TODO: make transferrable to plate? bag?
}

func (b Bag) GeneticInfoAsParent() (GeneticParentInfo, error) {
	return GeneticParentInfo{
		SpeciesOptionalField:    SpeciesOptionalField{b.Species},
		SubspeciesOptionalField: SubspeciesOptionalField{b.SubSpecies},
		KnownFruitableField:     KnownFruitableField{b.KnownFruitable},
		GenerationsFields:       b.GenerationsFields,
	}, nil
}

func (b Bag) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return b.GenSinceSpore, b.GenSinceFruitOrSpore
}

//func (b Bag) setTransferParent(ctx context.Context, xfer Transfer) error {
//	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(BagsCollectionName)
//	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
//	if err != nil {
//		return err
//	}
//	res, err := coll.UpdateByID(ctx, b.Id, upd)
//	if err != nil {
//		return err
//	}
//	if res.ModifiedCount == 0 {
//		return ErrNoParentModifiedForTransfer
//	}
//	return nil
//}

// TODO: create bag via substrate batch???
func (b Bag) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
	parentInfo, genSpore, genFruitSpore, err := childGensForParent(from)
	if err != nil {
		return err
	}
	upd, err := xfer.
		PicsModsForChild().
		Set("sealDate", xfer.LastUpdated).
		withInnoc(xfer).
		withParentType(utils.Pointer(xfer.FromType)).
		withParent(utils.Pointer(from.DbId())).
		withGens(genSpore, genFruitSpore).
		withSpecies(parentInfo.Species).
		withSubspecies(parentInfo.SubSpecies).
		withKnownFruitable(parentInfo.KnownFruitable).
		withPerms(from.Permissions()).
		withLastUpdated(xfer.LastUpdated).
		Finalized()
	if err != nil {
		return ErrFailedToFinalizeMods
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

func (b Bag) EntryTypeField() *string {
	return utils.Pointer(BagSourceType)
}

func (b Bag) id() []byte {
	return []byte(b.Id.dbIdStr())
}

func initializeBags(ctx context.Context) error {
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
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
	// If test agar batch does not exist, then create it
	testId := mainCollIdForint(idTestBag)
	testItem := &Bag{
		MainCollectionIdField:       MainCollectionIdField{testId},
		SubstrateRecipeField:        SubstrateRecipeField{exAltId},
		SubstrateBatchOptionalField: SubstrateBatchOptionalField{SubstrateBatch: utils.Pointer(altCollIdForint(idWoodPellets))},
		PcRunOptionalField:          PcRunOptionalField{&exAltId},
		FilterSize:                  "5nm",
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
		AclField:                          AclField{&testAcl},
	}
	return addTestMainEntries(ctx, testItem)
}

type createBagRequest struct {
	SubstrateBatchField
	WetnessField
	PcRunField
	FilterSize string // TODO: ???? Also handle on ts side
	CreationDateField
	NotesField
	WriteTagToField
}

func createBagHandler(w http.ResponseWriter, r *http.Request) {
	data := createBagRequest{}
	id := NextMainCollectionId()
	//id, err := newMainCollectionId(r.Context(), BagsCollectionName)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
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
	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	ctx := r.Context()
	// Validate request // TODO: move validation within a session?
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
	// TODO: validate filter size
	// Denying guest edits is done in the upper handlers
	toInsert := &Bag{
		MainCollectionIdField:       MainCollectionIdField{id},
		SubstrateRecipeField:        batch.SubstrateRecipeField,
		SubstrateBatchOptionalField: data.SubstrateBatchField.asOptional(),
		WetnessField:                data.WetnessField,
		PcRunOptionalField:          PcRunOptionalField{&data.PcRun},
		FilterSize:                  data.FilterSize,
		CreationDateField:           data.CreationDateField,
		NotesField:                  data.NotesField,
		LastUpdatedField:            LastUpdatedFieldForNow(),
		AclField:                    allCanWriteAcl(),
	}
	finishCreateMainCollectionEntry(ctx, toInsert, w)
}

type updateBagRequest struct {
	KnownFruitableField
	SaleField
	DisposedField
	Notes   AllEntries[Note]
	Images  SplitEntries[picWithNotesForm, PicWithNotesLessLocation] //"newPic-1"
	Contams SplitEntries[contamForm, ContaminationLessLocation]      //"newContam-1"
	Flushes SplitEntries[picWithNotesForm, PicWithNotesLessLocation] //"newFlush-1"
	WriteTagToField
	PermsOnRequest
}

func (upr updateBagRequest) reform() resolvedUpdateBagRequest {
	return resolvedUpdateBagRequest{
		KnownFruitableField: upr.KnownFruitableField,
		SaleField:           upr.SaleField,
		DisposedField:       upr.DisposedField,
		Notes:               upr.Notes,
		Images:              imageUpdates(upr.Images),
		Contams:             contamUpdates(upr.Contams),
		Flushes:             imageUpdates(upr.Flushes),
		PermsOnRequest:      upr.PermsOnRequest,
	}
}

type resolvedUpdateBagRequest struct {
	KnownFruitableField
	SaleField
	DisposedField
	Notes   AllEntries[Note]
	Images  SplitEntries[picWithNotesForm, PicWithNotes]
	Contams SplitEntries[contamForm, Contamination]
	Flushes SplitEntries[picWithNotesForm, PicWithNotes]
	PermsOnRequest
}

func (out resolvedUpdateBagRequest) modsFor(existing *Bag, aclField AclField) (bson.D, error) {
	return NewMods().
		updateKnownFruitableIfNeeded(out.KnownFruitable, existing.KnownFruitable).
		updateSaleIfNeeded(out.Sale, existing.Sale).
		updateDisposedIfNeeded(out.Disposed, existing.Disposed).
		updateNotesIfNeeded(out.Notes, existing.Notes).
		updatePicsIfNeeded(out.Images, existing.Pics).
		updateContamsIfNeeded(out.Contams, existing.Contaminations).
		updateFlushesIfNeeded(out.Flushes, existing.Flushes).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

const maxMultipartRequestSize = 32<<25 + 1024 //32<<20 + 1024 // TODO: is this max size ok?

func getBag(ctx context.Context, id MainCollectionId) (*Bag, error) {
	// go get current plate
	existing := &Bag{}
	err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(BagsCollectionName).FindOne(ctx, bson.D{{"_id", id}}).Decode(existing)
	return existing, err
}

func updateBagHandler(w http.ResponseWriter, r *http.Request) {
	data := updateBagRequest{}
	b58Id := Base58Str(r.PathValue("id"))
	id, err := b58Id.toMainCollectionId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	reader, err := multipartReaderForRequest(r, w, &data)
	if err != nil {
		// Already written
		return
	}
	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	newPics, newContams, newFlushes, err := getMultipartImages(r.Context(), "bag", w, reader, b58Id)
	if err != nil {
		// Already wrotw
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
		out.Images.New[i].Location = imageLocation(loc)
	}
	for i, _ := range data.Contams.New {
		if loc, exists := newContams[i]; exists {
			finalLoc := imageLocation(loc)
			out.Contams.New[i].Location = &finalLoc
		}
	}
	for i, _ := range data.Flushes.New {
		loc, exists := newFlushes[i]
		if !exists {
			http.Error(w, fmt.Sprintf("error, location for new flush index %d not found (should never happen)", i), http.StatusInternalServerError)
			return
		}
		out.Flushes.New[i].Location = imageLocation(loc)
	}
	ctx := r.Context()
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(BagsCollectionName)
	existing := &Bag{}
	err = coll.FindOne(ctx, bson.D{{"_id", id}}).Decode(existing)
	if err != nil {
		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishMainCollItemUpdate(ctx, w, coll, out.modsFor, existing, data.PermsOnRequest)
}

type importBagRequest struct {
	CreationDateField
	SubstrateRecipeField
	FilterSize string
	SpeciesField
	SubspeciesOptionalField
	Generation *int
	KnownFruitableField
	WriteTagToField
	// image as "img"
	PermsOnRequest
}

func importBagHandler(w http.ResponseWriter, r *http.Request) {
	data := importBagRequest{}
	//id, err := newMainCollectionId(r.Context(), BagsCollectionName)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
	id := NextMainCollectionId()
	b58id := id.asBase58()
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
	for { // TODO: FIX THIS MULTIPART READER
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
			newFileNameWithPrefixPath, errr := pics.SaveFile(r.Context(), fieldBytes, "bag", string(b58id), "img")
			if errr != nil {
				err = errr
				http.Error(w, "failed to save file: "+err.Error(), http.StatusBadRequest)
				return
			}
			picsSaved = append(picsSaved, newFileNameWithPrefixPath)
			now := unixTimeForNow()
			importedPic = &PicWithNotes{
				Time:       now,
				Location:   imageLocation(newFileNameWithPrefixPath),
				NotesField: NotesField{[]Note{}},
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
	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var gen *Generation = nil
	if data.Generation != nil {
		gen = (*Generation)(data.Generation)
	}
	pix := []PicWithNotes{}
	if importedPic != nil {
		pix = []PicWithNotes{*importedPic}
	}
	ctx := r.Context()
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(BagsCollectionName)
	// Validate
	_, err = data.SubstrateRecipeField.Get(ctx)
	if err != nil {
		dbErr(w, "substrate recipe retrieval error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Write
	toInsert := &Bag{
		MainCollectionIdField:   MainCollectionIdField{id},
		SubstrateRecipeField:    data.SubstrateRecipeField,
		PcRunOptionalField:      PcRunOptionalField{},
		FilterSize:              data.FilterSize,
		CreationDateField:       data.CreationDateField,
		GenerationsFields:       GenerationsFieldFor(gen),
		SealDate:                &data.CreationDate,
		KnownFruitableField:     data.KnownFruitableField,
		SpeciesOptionalField:    SpeciesOptionalField{&data.Species},
		SubspeciesOptionalField: data.SubspeciesOptionalField,
		PicsField:               PicsField{pix},
		ContaminationsField:     ContaminationsField{},
		MostRecentImageField:    MostRecentImageField{importedPic},
		FlushesField:            FlushesField{},
		SaleField:               SaleField{},
		DisposedField:           DisposedField{},
		NotesField:              NotesField{},
		LastUpdatedField:        LastUpdatedField{unixTimeForNow()},
	}
	finishImportMainCollectionEntry(ctx, coll, toInsert, data.PermsOnRequest, w)
}
