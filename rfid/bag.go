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
	"reflect"
)

const (
	BagsCollectionName = "fruitingBags" // TODO: USE
	BagSourceType      = "bag"
)

type Bag struct {
	MainCollectionIdField       `bson:"inline"`
	SubstrateRecipeField        `bson:"inline"`
	SubstrateBatchOptionalField `bson:"inline"` // TODO: NEW! HANDLE
	PcRunOptionalField          `bson:"inline"` // this may not exist for pre-existing bags
	//Size string // TODO: unsure what to do here
	FilterSize              string `bson:"filterSize" json:"filterSize"`
	CreationDateField       `bson:"inline"`
	GenerationsFields       `bson:"inline"`
	SealDate                *unixTime       `bson:"sealDate,omitempty" json:"sealDate,omitempty"` // set on transfer in
	WetnessField            `bson:"inline"` // Initial wetness (refer to scale on field struct) // TODO: new
	KnownFruitableField     `bson:"inline"` // set on transfer in, or once fruited
	SpeciesOptionalField    `bson:"inline"` // set on transfer in
	SubspeciesOptionalField `bson:"inline"` // set on transfer in
	InnocField              `bson:"inline"` // Set on transfer in. Innoc from LC or grain jar only
	TransfersOutField       `bson:"inline"` // Set on transfer out
	// TODO: make the next 2 a combo field?
	BinaryOptionalParentField `bson:"inline"` // Set on transfer in
	ParentTypeField           `bson:"inline"` // (main)lc, plate, or jar only (alt) can come from lcSyringe

	PicsField            `bson:"inline"` // Updated independently
	ContaminationsField  `bson:"inline"` // Updated independently
	MostRecentImageField `bson:"inline"`
	FlushesField         `bson:"inline"` // Updated independently
	SaleField            `bson:"inline"`
	DisposedField        `bson:"inline"`

	NotesField       `bson:"inline"` // Updated independently
	LastUpdatedField `bson:"inline"`
	AclField         `bson:"inline"` // TODO: handle EVERYWHERE
}

func (b Bag) CanTransferTo(dst geneticSource) error {
	return errors.New("Bag cannot be transferred (unsure if this is ok)")
	// TODO: make transferrable to plate?
}

func (b Bag) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := Bag{}
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

func (b Bag) setTransferParent(ctx context.Context, xfer Transfer) (error, func() error) {
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(BagsCollectionName)
	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
	if err != nil {
		return err, nil
	}
	res, err := coll.UpdateByID(ctx, b.Id, upd)
	if err != nil {
		return err, nil
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer, nil
	}
	return nil, func() error {
		return coll.FindOneAndReplace(ctx, bson.D{{"_id", b.Id}}, b).Err()
	}
}

// TODO: create bag via jar (or LCSyringe) instead
func (b Bag) setTransferChild(ctx context.Context, xfer Transfer, from geneticSource) error {
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
		//updatePermsIfNeeded(xfer.Perms, b.Perms). // TODO: this!
		withLastUpdated(xfer.LastUpdated).
		Finalized()
	if err != nil {
		return ErrFailedToFinalizeMods
	}
	res, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(BagsCollectionName).UpdateByID(ctx, b.Id, upd)
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
	return BagsCollectionName
}

func (b Bag) id() []byte {
	return []byte(b.Id.dbIdStr())
}

//func (b Bag) basicFruit() Fruit {
//	// TODO: ensure new fruit has a decent mainCollId in map
//	return Fruit{
//		MainCollectionIdField:        MainCollectionIdField{MainCollectionId(primitive.NewObjectID())},
//		SpeciesField:                      SpeciesField{*b.Species},
//		SubspeciesOptionalField:           b.SubspeciesOptionalField,
//		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&b.Email},
//		GenSporeField:                     GenSporeField{b.GenSinceSpore.Next()},
//		ParentTypeField:                   ParentTypeField{utils.Pointer("bag")},
//		LastUpdatedField:                  LastUpdatedField{unixTimeForNow()},
//	}
//}

func initializeBags(ctx context.Context) error {
	// TODO: INDICES!
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(BagsCollectionName)
	// Indices
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		// TODO: DO INDICES!!!
		newSimpleIndex("recipe", "recipe", false, false, false),
		newSimpleIndex("substrateBatch", "substrateBatch", false, true, false),
		newSimpleIndex("pcRun", "pcRun", false, true, false),
		// TODO: filter size?
		newSimpleIndex("creationDate", "creationDate", true, false, false),
		newSimpleIndex("genSinceSpore", "genSpore", true, true, false),
		newSimpleIndex("genSinceFruitOrSpore", "genFruitOrSpore", true, true, false),
		newSimpleIndex("sealDate", "sealDate", true, true, false), // BAG ONLY
		// TODO: wetness
		// TODO: knownFruitable?
		newSimpleIndex("species", "species", false, false, false),
		newSimpleIndex("subSpecies", "subSpecies", false, true, false),
		newSimpleIndex("innoc", "innoc", false, true, false),
		newSimpleIndex("transfersOut", "transfersOut", false, true, false),
		newSimpleIndex("parent", "parent", false, true, false),
		newSimpleIndex("parentType", "parentType", false, true, false),
		//pics
		//TODO: contams?
		//flushes
		newSimpleIndex("sale", "sale", false, true, false),
		newSimpleIndex("disposed", "disposed", false, true, false),
		//notes
		lastUpdatedIndexModel,
		//TODO: projectsIndexModel,
	})
	// If test agar batch does not exist, then create it
	existingEntry := Bag{}
	testId := mainCollIdForint(idTestBag)
	testItem := Bag{
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
		AclField:                  AclField{&testAcl},
	}
	// TODO: replace the test value
	err = coll.FindOne(ctx, bson.D{{"_id", testId}}).Decode(&existingEntry)
	if err == nil {
		if reflect.DeepEqual(existingEntry, testItem) {
			return nil
		}
	}
	return testExistingEntry(ctx, coll, testId, testItem, existingEntry)
}

type createBagRequest struct {
	SubstrateBatchField // TODO: USE AND VALIDATE
	WetnessField        // TODO: use and validate
	PcRunField
	FilterSize string // TODO: validate?
	CreationDateField
	NotesField
	WriteTagToField
}

func createBagHandler(w http.ResponseWriter, r *http.Request) {
	data := createBagRequest{}
	id, err := newCollectionId(r.Context(), BagsCollectionName)
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
	// Denying guest edits is done in the upper handlers
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		// Validate
		_, err = data.PcRunField.Get(ctx)
		if err != nil {
			return DbTxnStdErr(w, "PcRun validation failure: "+err.Error(), http.StatusBadRequest)
		}

		batch, err := data.SubstrateBatchField.Get(ctx) // TODO: get and validate substrate batch field
		if err != nil {
			return DbTxnStdErr(w, "Substrate batch validation failure: "+err.Error(), http.StatusBadRequest)
		}
		// validate wetness
		if data.Wetness != nil {
			if *data.Wetness < 0 || *data.Wetness > 10 {
				return DbTxnStdErr(w, "Invalid wetness, must either be nonexistent or 0-10: "+err.Error(), http.StatusBadRequest)
			}
		}

		coll := ctx.Client().Database(dbName).Collection(BagsCollectionName)
		toInsert := Bag{
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
		_, err = coll.InsertOne(ctx, toInsert)
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

func innoculateBagHandler(w http.ResponseWriter, r *http.Request) {
	b58Id := Base58Str(r.PathValue("id"))
	bagId, err := b58Id.toMainCollectionId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	b, err := getBag(r.Context(), bagId)
	if err != nil {
		http.Error(w, "failed to find bag: "+err.Error(), http.StatusBadRequest)
		return
	}
	if b.Innoc != nil {
		http.Error(w, "bag already innoculated before", http.StatusBadRequest)
		return
	}
	// TODO: create transfer
	// TODO: calculate all new things
	// TODO: save new things
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	req := inncoulateBagRequest{}
	err = json.Unmarshal(bs, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	switch req.sourceType {
	case lcSyringeSourceType:
		// TODO: handle lcSyringe innoc
	case GrainJarSourceType:
		// TODO: handle grainJarInnoc
	default:
		http.Error(w, "invalid source type", http.StatusBadRequest)
		return
	}
}

type inncoulateBagRequest struct {
	sourceType string // TODO: jar or lcSyringe
	parent     MainCollectionId
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
	PermsOnRequest // TODO: handle in typescript and handler!
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
	Notes          AllEntries[Note]
	Images         SplitEntries[picWithNotesForm, PicWithNotes]
	Contams        SplitEntries[contamForm, Contamination]
	Flushes        SplitEntries[picWithNotesForm, PicWithNotes]
	PermsOnRequest // TODO: handle in typescript and handler!
}

func (out resolvedUpdateBagRequest) modsFor(current Bag, aclField AclField) (bson.D, error) {
	return NewMods().
		updateKnownFruitableIfNeeded(out.KnownFruitable, current.KnownFruitable).
		updateSaleIfNeeded(out.Sale, current.Sale).
		updateDisposedIfNeeded(out.Disposed, current.Disposed).
		updateNotesIfNeeded(out.Notes, current.Notes).
		updatePicsIfNeeded(out.Images, current.Pics).
		updateContamsIfNeeded(out.Contams, current.Contaminations).
		updateFlushesIfNeeded(out.Flushes, current.Flushes).
		updatePermsIfNeeded(aclField.ACL, current.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

const maxMultipartRequestSize = 32<<20 + 1024 // TODO: is this max size ok?

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
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {

		coll := ctx.Client().Database(dbName).Collection(BagsCollectionName)
		// go get current plate
		existing := Bag{}
		err = coll.FindOne(ctx, bson.D{{"_id", id}}).Decode(&existing)
		if err != nil {
			return DbTxnStdErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		}
		user, err := GetAuthInfo(ctx)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		if !user.HasPermissionToEdit(existing) {
			return DbTxnStdErr(w, "unauthorized to edit", http.StatusForbidden)
		}
		aclField, err := data.AclFor(ctx, user)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		upd, err := out.modsFor(existing, aclField)                   // TODO: ACL?
		return handleUpdateMods(ctx, w, coll, existing, id, upd, err) // TODO: DO THIS EVERYWHERE!
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}

type importBagRequest struct {
	CreationDateField
	SubstrateRecipeField
	FilterSize              string
	SpeciesField            // TODO: Check species perms too
	SubspeciesOptionalField // TODO: check perms and validate
	Generation              *int
	KnownFruitableField
	WriteTagToField
	// image as "img"
	PermsOnRequest // TODO: handle in typescript and handler!
}

func importBagHandler(w http.ResponseWriter, r *http.Request) { // TODO: COPY FRUITING CHAMBER
	data := importBagRequest{}
	id, err := newCollectionId(r.Context(), BagsCollectionName)
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
	//spec, subsp, err := getSpeciesAndSubspecies(r.Context(), data.Species, data.SubSpecies)
	//if err != nil {
	//	http.Error(w, "failed to get spec/subsp: "+err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//finalPerms := minimalPermsBetween(data.Perms, spec, subsp) // TODO: subsp ptr ok here?
	//if err = finalPerms.ValidateUserCanWrite(r.Context()); err != nil {
	//	http.Error(w, "email cannot write these perms: "+err.Error(), http.StatusUnauthorized)
	//	return
	//}
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
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		perms, err := GetAuthInfo(ctx)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		acl, err := data.AclFor(ctx, perms)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		out := Bag{
			MainCollectionIdField: MainCollectionIdField{id},
			SubstrateRecipeField:  data.SubstrateRecipeField,
			//SubstrateBatchOptionalField: nil,
			//WetnessField,
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
			AclField:                acl,
		}
		// TODO: for non-all acls, update each email and project
		coll := ctx.Client().Database(dbName).Collection(BagsCollectionName)
		_, err = data.SubstrateRecipeField.Get(ctx)
		if err != nil {
			return DbTxnStdErr(w, "substrate recipe retrieval error: "+err.Error(), http.StatusInternalServerError)
		}
		_, err = coll.InsertOne(ctx, out)
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
