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
	FruitingChamberCollectionName = "fruitingChambers" // TODO: USE

	FruitingChamberSourceType = "fruitingChamber"
	fruitingChamberIdPrefix   = "FC"
)

// TODO: HANDLE MULTIPLE GRAIN INPUTS FOR MONOTUBS (DO MONOTUBS LATER)
type FruitingChamber struct { // TODO: SHOEBOX
	MainCollectionIdField
	CreationDateField
	SubstrateRecipeField
	SubstrateBatchOptionalField         // TODO: new! use!
	CupsGrain                   float64 `bson:"cupsGrain" json:"cupsGrain"`                           // TODO: new! use!
	MixedSubstratePerGrain      float64 `bson:"mixedSubstratePerGrain" json:"mixedSubstratePerGrain"` // for a 1:1:0.5 box this will be 1  // TODO: new! use!
	CasingPerGrain              float64 `bson:"casingPerGrain" json:"casingPerGrain"`                 // No casing==0, half casing per grain == 0.5 // TODO: new! use!
	SpeciesOptionalField
	SubspeciesOptionalField
	InnocField
	GenerationsFields
	TransfersOutField
	ParentTypeField                   // can be nil, most (main), or some (alt) like lcSyringe // TODO: NEW! HANDLE! nil == mainCollectionType (or purchased?), can also be MSS or clone! // TODO: INDEX????
	MainCollectionOptionalParentField // TODO: used to be binaryOptional
	PicsField
	ContaminationsField
	FlushesField
	KnownFruitableField
	MostRecentImageField
	SaleField
	DisposedField
	NotesField
	LastUpdatedField
	AclField // TODO: handle EVERYWHERE
}

func (f FruitingChamber) CanTransferTo(dst geneticSource) error {
	return errors.New("fc cannot be transferred (unsure if this is ok)")
}

func (f FruitingChamber) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := f
	err := decodeItem(&out, encoded)
	return out, err
}

func (f FruitingChamber) GeneticInfoAsParent() (GeneticParentInfo, error) {
	return GeneticParentInfo{
		SpeciesOptionalField:    f.SpeciesOptionalField,
		SubspeciesOptionalField: f.SubspeciesOptionalField,
		KnownFruitableField:     f.KnownFruitableField,
		GenerationsFields:       f.GenerationsFields,
	}, nil
}

func (f FruitingChamber) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return f.GenSinceSpore, f.GenSinceFruitOrSpore
}

func (f FruitingChamber) SourceType() string {
	return FruitingChamberSourceType
}

func (f FruitingChamber) setTransferParent(ctx mongo.SessionContext, xfer Transfer) error {
	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
	if err != nil {
		return err
	}
	res, err := ctx.Client().Database(dbName).Collection(FruitingChamberCollectionName).UpdateByID(ctx, f.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer
	}
	return nil
}

// TODO: create box via jar instead
func (f FruitingChamber) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
	parentInfo, genSpore, genFruitSpore, err := childGensForParent(from)
	if err != nil {
		return err
	}
	mods := NewMods()
	pictures := []PicWithNotes{}
	if xfer.ToImage != nil {
		pic := PicWithNotes{
			Time:       xfer.CreationDate,
			Location:   *xfer.ToImage,
			NotesField: NotesField{[]Note{}},
		}
		pictures = []PicWithNotes{pic}
		mods = mods.
			withMostRecentImage(&pic).
			withPics(pictures)
	}
	upd, err := mods.
		withInnoc(xfer).
		withParentType(utils.Pointer(xfer.FromType)).
		withParent(utils.Pointer(from.DbId())). // TODO: will this work for mainCollectionId?
		withGens(genSpore, genFruitSpore).
		withSpecies(parentInfo.Species).
		withSubspecies(parentInfo.SubSpecies).
		withKnownFruitable(parentInfo.KnownFruitable).
		//updatePermsIfNeeded(xfer.Perms, f.Perms).
		withLastUpdated(xfer.LastUpdated).
		Finalized()
	if err != nil {
		return errors.New("failed to finalize") // TODO: ok?
	}
	res, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(FruitingChamberCollectionName).UpdateByID(ctx, f.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer
	}
	return nil
}

func (f FruitingChamber) EntryTypeField() *string {
	return utils.Pointer(FruitingChamberSourceType)
}

func (f FruitingChamber) CollectionName() string {
	return mainCollectionName
}

//func (f FruitingChamber) basicFruit() Fruit {
//	return Fruit{
//		MainCollectionIdField:        MainCollectionIdField{MainCollectionId(primitive.NewObjectID())},
//		SpeciesField:                      SpeciesField{*f.Species}, // TODO: ensure pointer is not nil
//		SubspeciesOptionalField:           f.SubspeciesOptionalField,
//		GenSporeField:                     GenSporeField{f.GenSinceSpore.Next()},
//		ParentTypeField:                   ParentTypeField{utils.Pointer(FruitingChamberSourceType)},
//		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&f.UserId},
//		LastUpdatedField:                  LastUpdatedField{unixTimeForNow()},
//	}
//}

func initializeFruitingChamber(ctx context.Context) error {
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(FruitingChamberCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		creationDateIndexModel,
		newSimpleIndex("recipe", "recipe", false, false, false), // TODO: this is harvest date
		newSimpleIndex("substrateBatch", "substrateBatch", false, true, false),
		//newSimpleIndex("cupsGrain","cupsGrain", false, false, false),
		//newSimpleIndex("mixedSubstratePerGrain","mixedSubstratePerGrain", false, false, false),
		//newSimpleIndex("casingPerGrain","casingPerGrain", false, false, false),
		newSimpleIndex("species", "species", false, true, false),
		newSimpleIndex("subSpecies", "subSpecies", false, true, false),
		newSimpleIndex("innoc", "innoc", false, true, false),
		newSimpleIndex("genSinceSpore", "genSpore", true, true, false),
		newSimpleIndex("genSinceFruitOrSpore", "genFruitOrSpore", true, true, false),
		transfersOutIndexModel,
		// TODO: prints
		newSimpleIndex("parent", "parent", false, true, false),         // TODO: nil is store or outside?
		newSimpleIndex("parentType", "parentType", false, true, false), // TODO: nil is store or outside?
		//Pics (no index)
		//TODO: Contams
		// Flushes
		newSimpleIndex("knownFruitable", "knownFruitable", false, true, false),
		// MostRecentImage
		saleIndexModel,
		newSimpleIndex("disposed", "disposed", false, true, false),
		//Notes (no index) (maybe later with tags?)
		lastUpdatedIndexModel,
		// TODO: projectsIndexModel,
	})
	if err != nil {
		return err
	}
	// If test FC does not exist, then create it
	existingEntry := FruitingChamber{}
	testId := mainCollIdForint(idTestFC)
	xfer := exAltId
	plateId := mainCollIdForint(idTestPlate)
	testItem := FruitingChamber{

		MainCollectionIdField:       MainCollectionIdField{testId},
		SubstrateRecipeField:        SubstrateRecipeField{exAltId},
		SubstrateBatchOptionalField: SubstrateBatchOptionalField{}, // TODO: ADD ME
		CupsGrain:                   4,                             // quart
		MixedSubstratePerGrain:      1.0,                           // 1:1 grain:subsMixed
		CasingPerGrain:              0.5,                           // Half casing per grain
		CreationDateField:           CreationDateField{exampleTime},
		GenerationsFields: GenerationsFields{
			GenSporeField:        GenSporeField{&exGenSinceSpore},
			GenSinceFruitOrSpore: &exGenSinceFruitSpore,
		},
		KnownFruitableField:               KnownFruitableField{exBool},
		SpeciesOptionalField:              SpeciesOptionalField{&exampleSpecies},
		SubspeciesOptionalField:           SubspeciesOptionalField{exampleSubspecies},
		InnocField:                        InnocField{&xfer},
		TransfersOutField:                 TransfersOutField{exAlts},
		ParentTypeField:                   ParentTypeField{&exParentType},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&plateId},
		PicsField:                         PicsField{exPics},
		ContaminationsField:               ContaminationsField{exContams},
		MostRecentImageField:              MostRecentImageField{&exPics[0]},
		FlushesField:                      FlushesField{exPics},
		SaleField:                         SaleField{utils.Pointer(exAltId)},
		DisposedField:                     DisposedField{}, // TODO: dispose of it in tests?
		NotesField:                        NotesField{exampleNotes()},
		LastUpdatedField:                  LastUpdatedField{exampleTime},
	}
	err = coll.FindOne(ctx, bson.D{{"_id", testId}}).Decode(&existingEntry)
	if err == nil {
		if reflect.DeepEqual(existingEntry, testItem) {
			return nil
		}
	}
	return testExistingEntry(ctx, coll, testId, testItem, existingEntry)
}

type createFruitingChamberRequest struct {
	Recipe              AlternateCollectionId // substrate recipe // TODO: do not use this. Pull from batch
	SubstrateBatchField                       // TODO: USE ME
	ParentJar           MainCollectionId      // TODO: USE ME
	MixedSubstrateCups  float64
	CasingCups          float64
	NotesField
	WriteTagToField
}

func createFruitingChamberHandler(w http.ResponseWriter, r *http.Request) {
	data := createFruitingChamberRequest{}
	id, err := newMainCollectionId(r.Context())
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

	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		parentJar, err := LookupGrainJar(ctx, data.ParentJar)
		if err != nil {
			return DbTxnStdErr(w, "failed to resolve parent jar"+err.Error(), http.StatusBadRequest)
		}

		// TODO: figure out cupSize of grain
		coll := ctx.Client().Database(dbName).Collection(FruitingChamberCollectionName)
		now := unixTimeForNow()
		_, err = SubstrateRecipeField{data.Recipe}.Get(ctx)
		if err != nil {
			return DbTxnStdErr(w, "invalid substrate recipe: "+err.Error(), http.StatusBadRequest)
		}
		err = writeRfidTagIfNecessary(ctx, data.WriteTagTo, id)
		if err != nil {
			return DbTxnStdErr(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		}
		toInsert := FruitingChamber{
			MainCollectionIdField:  MainCollectionIdField{id},
			SubstrateRecipeField:   SubstrateRecipeField{data.Recipe},
			CupsGrain:              float64(parentJar.SizeCups),
			MixedSubstratePerGrain: data.MixedSubstrateCups / float64(parentJar.SizeCups), // TODO: ensure ok
			CasingPerGrain:         data.CasingCups / float64(parentJar.SizeCups),         // TODO: ensure ok
			CreationDateField:      CreationDateField{now},
			NotesField:             NotesField{data.Notes},
			LastUpdatedField:       LastUpdatedField{now},
			AclField:               parentJar.AclField,
		}
		err = addToIdMapCollectionInTxn(ctx, id.ToBinaryCollectionId(), toInsert.SourceType())
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
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

type importFruitingChamberRequest struct {
	SubstrateRecipeField
	CreationDateField
	SpeciesField
	GrainCups      float64  // TODO: USE
	SubstrateRatio *float64 // TODO: USE
	CasingRatio    *float64 // TODO: USE
	SubspeciesOptionalField
	Generation *int
	KnownFruitableField
	WriteTagToField
	PermsOnRequest // TODO: handle in typescript and handler!
	// image as "img"
}

func importFruitingChamberHandler(w http.ResponseWriter, r *http.Request) {
	data := importFruitingChamberRequest{}
	id, err := newMainCollectionId(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	b58id := id.asBase58()
	r.Body = http.MaxBytesReader(w, r.Body, maxMultipartRequestSize) // TODO: REDO THIS MULTIPART READER
	defer r.Body.Close()
	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "unable to open multipart reader: "+err.Error(), http.StatusBadRequest)
		return
	}
	p1, err := reader.NextPart()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer p1.Close()
	// Process text (or object)
	bs, errr := io.ReadAll(p1)
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
	// Try to get pic if exists
	picsSaved := []string{}
	defer func() {
		if err != nil {
			err = pics.DeleteFiles(r.Context(), picsSaved...)
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		fileName := p.FileName()
		defer p.Close()
		if fileName != "img" {
			http.Error(w, "invalid image name", http.StatusBadRequest)
			return
		}
		// Process file
		fieldBytes, errr := multipartToImageBytes(p, w)
		if errr != nil {
			// Already wrote
			return
		}
		newFileNameWithPrefixPath, errr := pics.SaveFile(r.Context(), fieldBytes, "fruitingChamber", string(b58id), "img")
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
	}
	var gen *Generation = nil
	if data.Generation != nil {
		gen = (*Generation)(data.Generation)
	}
	pix := []PicWithNotes{}
	if importedPic != nil {
		pix = []PicWithNotes{*importedPic}
	}
	//sp, subsp, err := getSpeciesAndSubspecies(r.Context(), data.Species, data.SubSpecies)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError) // TODO: normalize
	//	return
	//}
	//finalPerms := minimalPermsBetween(data.Perms, sp, subsp)
	//if err = finalPerms.ValidateUserCanWrite(r.Context()); err != nil { // TODO: maybe dont do this?
	//	http.Error(w, "user cannot write with the provided perms: "+err.Error(), http.StatusBadRequest)
	//	return
	//}

	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		perms, err := GetAuthInfo(ctx)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}

		out := FruitingChamber{
			MainCollectionIdField:       MainCollectionIdField{id},
			SubstrateRecipeField:        data.SubstrateRecipeField,
			SubstrateBatchOptionalField: SubstrateBatchOptionalField{nil}, // Unknown for imports
			CreationDateField:           CreationDateField{data.CreationDate},
			CupsGrain:                   data.GrainCups,
			MixedSubstratePerGrain:      utils.Default(data.SubstrateRatio, 1.0),
			CasingPerGrain:              utils.Default(data.CasingRatio, 0.5),
			SpeciesOptionalField:        SpeciesOptionalField{&data.Species},
			SubspeciesOptionalField:     data.SubspeciesOptionalField,
			GenerationsFields: GenerationsFields{
				GenSporeField:        GenSporeField{gen},
				GenSinceFruitOrSpore: gen,
			},
			PicsField:            PicsField{pix},
			KnownFruitableField:  data.KnownFruitableField,
			MostRecentImageField: MostRecentImageField{importedPic},
			LastUpdatedField:     LastUpdatedField{unixTimeForNow()},
			AclField:             data.AclFor(ctx, perms),
		}
		_, err = data.SubstrateRecipeField.Get(ctx)
		if err != nil {
			return DbTxnStdErr(w, "bad substrate recipe: "+err.Error(), http.StatusInternalServerError) // TODO: normalize
		}
		err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
		if err != nil {
			return DbTxnStdErr(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		}
		coll := ctx.Client().Database(dbName).Collection(mainCollectionName)
		_, err := coll.InsertOne(ctx, out)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		bsOut, err := json.Marshal(out)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		return w.Write(bsOut)
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}

type updateFruitingChamberRequest struct {
	Notes AllEntries[Note]
	KnownFruitableField
	DisposedField
	SaleField
	Images  SplitEntries[picWithNotesForm, PicWithNotesLessLocation] //"newPic-1"
	Contams SplitEntries[contamForm, ContaminationLessLocation]      //"newContam-1"
	Flushes SplitEntries[picWithNotesForm, PicWithNotesLessLocation] //"newFlush-1"
	WriteTagToField
	PermsOnRequest // TODO: handle in typescript and handler!
}

func (upr updateFruitingChamberRequest) reform() resolvedUpdateFruitingChamberRequest {
	return resolvedUpdateFruitingChamberRequest{
		KnownFruitableField: upr.KnownFruitableField,
		SaleField:           upr.SaleField,
		DisposedField:       upr.DisposedField,
		Notes:               upr.Notes,
		Images:              imageUpdates(upr.Images),
		Contams:             contamUpdates(upr.Contams),
		Flushes:             imageUpdates(upr.Flushes),
		PermsOnRequest:      req.PermsOnRequest,
	}
}

type resolvedUpdateFruitingChamberRequest struct {
	KnownFruitableField
	SaleField
	DisposedField
	Notes          AllEntries[Note]
	Images         SplitEntries[picWithNotesForm, PicWithNotes]
	Contams        SplitEntries[contamForm, Contamination]
	Flushes        SplitEntries[picWithNotesForm, PicWithNotes]
	PermsOnRequest // TODO: handle in typescript and handler!
}

func updateFruitingChamberHandler(w http.ResponseWriter, r *http.Request) {
	data := updateFruitingChamberRequest{}
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
	newPics, newContams, newFlushes, err := getMultipartImages(r.Context(), "fruitingChamber", w, reader, b58Id)
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
		db := ctx.Client().Database(dbName)
		coll := db.Collection(mainCollectionName)
		// go get current plate
		existing := FruitingChamber{}
		err = coll.FindOne(ctx, bson.D{{"_id", id}}).Decode(&existing)
		if err != nil {
			return DbTxnStdErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		}
		if out.Sale != nil && (existing.Sale == nil || *existing.Sale != *out.Sale) {
			if err = db.Collection(salesCollectionName).FindOne(ctx, bson.D{{"_id", out.Sale}}).Err(); err != nil {
				return DbTxnStdErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
			}
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
		//if err = minimalPermsBetween(existing.Perms, data.Perms).ValidateUserCanWrite(ctx); err != nil {
		//	return DbTxnStdErr(w, "overlapping perms invalid for user: "+err.Error(), http.StatusBadRequest)
		//}
		upd, err := NewMods().
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
		return handleUpdateMods(ctx, w, coll, existing, id, upd, err)
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}
