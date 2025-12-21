package rfid

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/reeceappling/goUtils/v2/utils"
	"github.com/reeceappling/mushDb/rfid/pics"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"reflect"
)

const (
	BagSourceType = "bag"
	bagIdPrefix   = "BG"
)

type SubstrateBatchOptionalField struct { // TODO: MOVE
	SubstrateBatch *AlternateCollectionId `bson:"substrateBatch,omitempty" json:"substrateBatch,omitempty"`
}

type Bag struct {
	EntryTypeStructField
	MainCollectionIdField
	SubstrateRecipeField
	SubstrateBatchOptionalField // TODO: NEW! HANDLE
	PcRunOptionalField          // this may not exist for pre-existing bags
	//Size int // TODO: unsure what to do here
	FilterSize string `bson:"filterSize" json:"filterSize"`
	CreationDateField
	GenerationsFields
	SealDate                  *unixTime `bson:"sealDate,omitempty" json:"sealDate,omitempty"` // set on transfer in
	WetnessField                        // Initial wetness (refer to scale on field struct) // TODO: new
	KnownFruitableField                 // set on transfer in, or once fruited
	SpeciesOptionalField                // set on transfer in
	SubspeciesOptionalField             // set on transfer in
	InnocField                          // Set on transfer in. Innoc from LC or grain jar only
	TransfersOutField                   // Set on transfer out
	ParentTypeField                     // (main)lc, plate, or jar only (alt) can come from lcSyringe
	BinaryOptionalParentField           // Set on transfer in
	PicsField                           // Updated independently
	ContaminationsField                 // Updated independently
	MostRecentImageField
	FlushesField // Updated independently
	SaleField
	DisposedField

	NotesField // Updated independently
	LastUpdatedField
	PermsField
}

func (b Bag) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := b
	err := decodeItem(&out, encoded)
	return out, err
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

func (b Bag) SourceType() string {
	return BagSourceType
}

func (b Bag) setTransferParent(ctx mongo.SessionContext, xfer Transfer) error {
	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
	if err != nil {
		return err
	}
	res, err := ctx.Client().Database(dbName).Collection(mainCollectionName).UpdateByID(ctx, b.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer
	}
	return nil
}

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
		updatePermsIfNeeded(xfer.Perms, b.Perms).
		withLastUpdated(xfer.LastUpdated).
		Finalized()
	if err != nil {
		return ErrFailedToFinalizeMods
	}
	res, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(mainCollectionName).UpdateByID(ctx, b.Id, upd)
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

func (b Bag) CollectionName() string {
	return mainCollectionName
}

func (b Bag) id() []byte {
	return b.Id[:]
}

func (b Bag) basicFruit() Fruit {
	return Fruit{
		AlternateCollectionIdField:        AlternateCollectionIdField{AlternateCollectionId(primitive.NewObjectID())},
		SpeciesField:                      SpeciesField{*b.Species},
		SubspeciesOptionalField:           b.SubspeciesOptionalField,
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&b.Id},
		GenSporeField:                     GenSporeField{b.GenSinceSpore.Next()},
		ParentTypeField:                   ParentTypeField{utils.Pointer("bag")},
		LastUpdatedField:                  LastUpdatedField{unixTimeForNow()},
	}
}

func initializeBags(ctx context.Context) error {
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(mainCollectionName)
	// If test agar batch does not exist, then create it
	existingEntry := Bag{}
	testId := mainCollIdForint(idTestBag)
	testItem := Bag{
		EntryTypeStructField:  EntryTypeStructField{*existingEntry.EntryTypeField()},
		MainCollectionIdField: MainCollectionIdField{testId},
		SubstrateRecipeField:  SubstrateRecipeField{exAltId},
		PcRunOptionalField:    PcRunOptionalField{&exAltId},
		FilterSize:            "5nm",
		CreationDateField:     CreationDateField{exampleTime},
		GenerationsFields: GenerationsFields{
			GenSporeField:        GenSporeField{&exGenSinceSpore},
			GenSinceFruitOrSpore: &exGenSinceFruitSpore,
		},
		SealDate:                  &exampleTime,
		KnownFruitableField:       KnownFruitableField{exBool},
		SpeciesOptionalField:      SpeciesOptionalField{&testEntryStringId},
		SubspeciesOptionalField:   SubspeciesOptionalField{&testEntryStringId},
		InnocField:                InnocField{&exAltId},
		TransfersOutField:         TransfersOutField{exAlts},
		ParentTypeField:           ParentTypeField{&exParentType},
		BinaryOptionalParentField: BinaryOptionalParentField{utils.Pointer(exPlate.ToBinaryCollectionId())},
		PicsField:                 PicsField{exPics},
		ContaminationsField:       ContaminationsField{exContams},
		MostRecentImageField:      MostRecentImageField{&exPics[0]},
		FlushesField:              FlushesField{exPics},
		SaleField:                 SaleField{&exAltId},
		DisposedField:             DisposedField{&exampleTime},
		NotesField:                NotesField{exampleNotes()},
		LastUpdatedField:          LastUpdatedField{exampleTime},
	}
	err := coll.FindOne(ctx, bson.D{{"_id", testId}}).Decode(&existingEntry)
	if err == nil {
		if reflect.DeepEqual(existingEntry, testItem) {
			return nil
		}
	}
	return testExistingEntry(ctx, coll, testId, testItem, existingEntry)
}

type createBagRequest struct {
	SubstrateRecipeField // substrate recipe
	PcRunField
	FilterSize string // TODO: validate?
	CreationDateField
	NotesField
	WriteTagToField
}

func createBagHandler(w http.ResponseWriter, r *http.Request) {
	data := createBagRequest{}
	id, err := generateMainCollectionId(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		// Validate
		_, err = data.PcRunField.Get(ctx)
		if err != nil {
			return DbTxnStdErr(w, "PcRun validation failure: "+err.Error(), http.StatusBadRequest)
		}
		_, err = data.SubstrateRecipeField.Get(ctx)
		if err != nil {
			return DbTxnStdErr(w, "Substrate recipe validation failure: "+err.Error(), http.StatusBadRequest)
		}
		coll := ctx.Client().Database(dbName).Collection(mainCollectionName)
		toInsert := Bag{
			EntryTypeStructField:  EntryTypeStructField{"bag"},
			MainCollectionIdField: MainCollectionIdField{id},
			SubstrateRecipeField:  data.SubstrateRecipeField,
			PcRunOptionalField:    PcRunOptionalField{&data.PcRun},
			FilterSize:            data.FilterSize,
			CreationDateField:     data.CreationDateField,
			NotesField:            data.NotesField,
			LastUpdatedField:      LastUpdatedFieldForNow(),
		}
		_, err := coll.InsertOne(ctx, toInsert)
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

type updateBagRequest struct {
	KnownFruitableField
	SaleField
	DisposedField
	Notes   AllEntries[Note]
	Images  SplitEntries[picWithNotesForm, PicWithNotesLessLocation] //"newPic-1"
	Contams SplitEntries[contamForm, ContaminationLessLocation]      //"newContam-1"
	Flushes SplitEntries[picWithNotesForm, PicWithNotesLessLocation] //"newFlush-1"
	WriteTagToField
	PermsField
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
		PermsField:          upr.PermsField,
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
	PermsField
}

func (out resolvedUpdateBagRequest) modsFor(current Bag) (bson.D, error) {
	return NewMods().
		updateKnownFruitableIfNeeded(out.KnownFruitable, current.KnownFruitable).
		updateSaleIfNeeded(out.Sale, current.Sale).
		updateDisposedIfNeeded(out.Disposed, current.Disposed).
		updateNotesIfNeeded(out.Notes, current.Notes).
		updatePicsIfNeeded(out.Images, current.Pics).
		updateContamsIfNeeded(out.Contams, current.Contaminations).
		updateFlushesIfNeeded(out.Flushes, current.Flushes).
		updatePermsIfNeeded(out.Perms, current.Perms).
		updateLastUpdatedIfNeeded().
		Finalized()
}

const maxMultipartRequestSize = 32<<20 + 1024 // TODO: is this max size ok?

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

	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(mainCollectionName)
		// go get current plate
		existing := Bag{}
		err = coll.FindOne(ctx, bson.D{{"_id", id}}).Decode(&existing)
		if err != nil {
			return DbTxnStdErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		}

		if err = minimalPermsBetween(existing.Perms, data.Perms).ValidateUserCanWrite(ctx); err != nil {
			return DbTxnStdErr(w, "overlapping perms for user err: "+err.Error(), http.StatusUnauthorized)
		}
		upd, err := out.modsFor(existing)
		return handleUpdateMods(ctx, w, coll, existing, id, upd, err) // TODO: DO THIS EVERYWHERE!
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}

type importBagRequest struct {
	CreationDateField              // TODO: MAKE OPTIONAL?
	SubstrateRecipeField           // TODO: MAKE OPTIONAL
	FilterSize              string // TODO: MAKE OPTIONAL
	SpeciesField                   // TODO: Check species perms too
	SubspeciesOptionalField        // TODO: check perms and validate
	Generation              *int
	KnownFruitableField
	WriteTagToField
	// image as "img"
	PermsField // TODO: new
}

func importBagHandler(w http.ResponseWriter, r *http.Request) { // TODO: COPY FRUITING CHAMBER
	data := importBagRequest{}
	id, err := generateMainCollectionId(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
				http.Error(w, "unable to read data from form: "+err.Error(), http.StatusBadRequest)
				return
			}
			// PARSE INTO CORRECT DATA FORMAT
			err = json.Unmarshal(bs, &data)
			if err != nil {
				http.Error(w, "unable to unmarshal json form data: "+err.Error(), http.StatusBadRequest)
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
		err = errors.New("no non-image data found on form request")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	spec, subsp, err := getSpeciesAndSubspecies(r.Context(), data.Species, data.SubSpecies)
	if err != nil {
		http.Error(w, "failed to get spec/subsp: "+err.Error(), http.StatusInternalServerError)
		return
	}
	finalPerms := minimalPermsBetween(data.Perms, spec, subsp) // TODO: subsp ptr ok here?
	if err = finalPerms.ValidateUserCanWrite(r.Context()); err != nil {
		http.Error(w, "user cannot write these perms: "+err.Error(), http.StatusUnauthorized)
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
	out := Bag{
		EntryTypeStructField:    EntryTypeStructField{"bag"},
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
		PermsField:              PermsField{finalPerms}, // TODO: ALSO USE SPEC/SUBS PERMS!
	} // TODO: update sessions (and projects?) with perm updates?
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(mainCollectionName)
		_, err = data.SubstrateRecipeField.Get(ctx)
		if err != nil {
			return DbTxnStdErr(w, "substrate recipe retrieval error: "+err.Error(), http.StatusInternalServerError)
		}
		_, err := coll.InsertOne(ctx, out)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		bs, err := json.Marshal(out)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		return w.Write(bs)
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}
