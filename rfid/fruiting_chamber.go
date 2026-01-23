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

// TODO: HANDLE MULTIPLE GRAIN INPUTS FOR MONOTUBS (DO MONOTUBS LATER)
type FruitingChamber struct { // TODO: SHOEBOX
	MainCollectionIdField             `bson:"inline"`
	CreationDateField                 `bson:"inline"`
	SubstrateRecipeField              `bson:"inline"`
	SubstrateBatchOptionalField       `bson:"inline"`
	CupsGrain                         float64 `bson:"cupsGrain" json:"cupsGrain"`                           // TODO: new! use!
	MixedSubstratePerGrain            float64 `bson:"mixedSubstratePerGrain" json:"mixedSubstratePerGrain"` // for a 1:1:0.5 box this will be 1  // TODO: new! use!
	CasingPerGrain                    float64 `bson:"casingPerGrain" json:"casingPerGrain"`                 // No casing==0, half casing per grain == 0.5 // TODO: new! use!
	SpeciesOptionalField              `bson:"inline"`
	SubspeciesOptionalField           `bson:"inline"`
	InnocField                        `bson:"inline"`
	GenerationsFields                 `bson:"inline"`
	TransfersOutField                 `bson:"inline"`
	ParentTypeField                   `bson:"inline"` // can be nil, most (main), or some (alt) like lcSyringe // nil == mainCollectionType (or purchased?), can also be MSS or clone! // TODO: INDEX????
	MainCollectionOptionalParentField `bson:"inline"`
	PicsField                         `bson:"inline"`
	ContaminationsField               `bson:"inline"`
	FlushesField                      `bson:"inline"`
	KnownFruitableField               `bson:"inline"`
	MostRecentImageField              `bson:"inline"`
	SaleField                         `bson:"inline"`
	DisposedField                     `bson:"inline"`
	NotesField                        `bson:"inline"`
	LastUpdatedField                  `bson:"inline"`
	AclField                          `bson:"inline"`
}

func (f FruitingChamber) CanTransferTo(dst geneticSource) error {
	return errors.New("fc cannot be transferred (unsure if this is ok)")
	// TODO: make transferrable to plate? box? bag?
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

func (f FruitingChamber) setTransferParent(ctx context.Context, xfer Transfer) (error, func() error) {
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(FruitingChamberCollectionName)
	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
	if err != nil {
		return err, nil
	}
	res, err := coll.UpdateByID(ctx, f.Id, upd)
	if err != nil {
		return err, nil
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer, nil
	}
	return nil, func() error {
		return coll.FindOneAndReplace(ctx, bson.D{{"_id", f.Id}}, f).Err()
	}
}

// TODO: create box via jar instead
func (f FruitingChamber) setTransferChild(ctx context.Context, xfer Transfer, from geneticSource) error {
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

//func (f FruitingChamber) basicFruit() Fruit {
//	return Fruit{
//		MainCollectionIdField:        MainCollectionIdField{MainCollectionId(primitive.NewObjectID())},
//		SpeciesField:                      SpeciesField{*f.Species}, // TODO: ensure pointer is not nil
//		SubspeciesOptionalField:           f.SubspeciesOptionalField,
//		GenSporeField:                     GenSporeField{f.GenSinceSpore.Next()},
//		ParentTypeField:                   ParentTypeField{utils.Pointer(FruitingChamberSourceType)},
//		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&f.Email},
//		LastUpdatedField:                  LastUpdatedField{unixTimeForNow()},
//	}
//}

func initializeFruitingChamber(ctx context.Context) error {
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(FruitingChamberCollectionName)
	err := createIndexes(ctx, coll,
		[]mongo.IndexModel{
			creationDateIndexModel,
			//newSimpleIndex("recipe", "recipe", false, false, false), // TODO: this is harvest date
			// TODO: newSimpleIndex("substrateBatch", "substrateBatch", false, true, false),
			//newSimpleIndex("cupsGrain","cupsGrain", false, false, false),
			//newSimpleIndex("mixedSubstratePerGrain","mixedSubstratePerGrain", false, false, false),
			//newSimpleIndex("casingPerGrain","casingPerGrain", false, false, false),
			newSimpleIndex("species", "species", false, true, false),
			newSimpleIndex("subspecies", "subspecies", false, true, false),
			//newSimpleIndex("innoc", "innoc", false, true, false),
			//newSimpleIndex("genSinceSpore", "genSpore", true, true, false),
			//newSimpleIndex("genSinceFruitOrSpore", "genFruitOrSpore", true, true, false),
			//transfersOutIndexModel,
			//newSimpleIndex("parent", "parent", false, true, false),         // TODO: nil is store or outside?
			//newSimpleIndex("parentType", "parentType", false, true, false), // TODO: nil is store or outside?
			//Pics (no index)
			//TODO: Contams
			// Flushes
			//newSimpleIndex("knownFruitable", "knownFruitable", false, true, false),
			// MostRecentImage
			//saleIndexModel,
			//newSimpleIndex("disposed", "disposed", false, true, false),
			//Notes (no index) (maybe later with tags?)
			lastUpdatedIndexModel,
			projectsIndexModel,
		})
	if err != nil {
		return err
	}
	// If test FC does not exist, then create it
	testId := mainCollIdForint(idTestFC)
	xfer := exAltId
	plateId := mainCollIdForint(idTestPlate)
	testItem := &FruitingChamber{

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
	return addTestMainEntries(ctx, testItem)
}

type createFruitingChamberRequest struct {
	// TODO: removed: Recipe // substrate recipe // TODO: do not use this. Pull from batch
	SubstrateBatchField
	ParentJar          MainCollectionId
	MixedSubstrateCups float64
	CasingCups         float64
	NotesField
	WriteTagToField
}

func createFruitingChamberHandler(w http.ResponseWriter, r *http.Request) {
	data := createFruitingChamberRequest{}
	id, err := newCollectionId(r.Context(), FruitingChamberCollectionName)
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
	ctx := r.Context()
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(FruitingChamberCollectionName)
	// Validation

	parentJar, err := LookupGrainJar(ctx, data.ParentJar)
	if err != nil {
		http.Error(w, "failed to resolve parent jar"+err.Error(), http.StatusBadRequest)
		return
	}

	now := unixTimeForNow()
	batch, err := data.SubstrateBatchField.Get(ctx)
	if err != nil {
		http.Error(w, "invalid substrate batch: "+err.Error(), http.StatusBadRequest)
		return
	}
	err = writeRfidTagIfNecessary(ctx, data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	toInsert := FruitingChamber{
		MainCollectionIdField:       MainCollectionIdField{id},
		SubstrateRecipeField:        batch.SubstrateRecipeField,
		SubstrateBatchOptionalField: data.SubstrateBatchField.asOptional(),
		CupsGrain:                   float64(parentJar.SizeCups),
		MixedSubstratePerGrain:      data.MixedSubstrateCups / float64(parentJar.SizeCups),
		CasingPerGrain:              data.CasingCups / float64(parentJar.SizeCups),
		CreationDateField:           CreationDateField{now},
		NotesField:                  NotesField{data.Notes},
		LastUpdatedField:            LastUpdatedField{now},
		AclField:                    parentJar.AclField,
	}
	finishCreateMainCollectionEntry(ctx, coll, &toInsert, w)
}

type importFruitingChamberRequest struct {
	SubstrateRecipeField
	CreationDateField
	SpeciesField
	GrainCups      float64
	SubstrateRatio float64 // TODO: used to be optional
	CasingRatio    float64 // TODO: used to be optional
	SubspeciesOptionalField
	Generation *int
	KnownFruitableField
	WriteTagToField
	PermsOnRequest
	// image as "img"
}

func importFruitingChamberHandler(w http.ResponseWriter, r *http.Request) {
	data := importFruitingChamberRequest{}
	id, err := newCollectionId(r.Context(), FruitingChamberCollectionName)
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
		http.Error(w, "unable to read Data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	// PARSE INTO CORRECT DATA FORMAT
	err = json.Unmarshal(bs, &data)
	if err != nil {
		http.Error(w, "unable to unmarshal json form Data: "+err.Error(), http.StatusBadRequest)
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
	ctx := r.Context()
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(FruitingChamberCollectionName)
	perms, err := GetAuthInfo(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	acl, err := data.AclForUser(ctx, perms)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	toInsert := &FruitingChamber{
		MainCollectionIdField:       MainCollectionIdField{id},
		SubstrateRecipeField:        data.SubstrateRecipeField,
		SubstrateBatchOptionalField: SubstrateBatchOptionalField{nil}, // Unknown for imports
		CreationDateField:           CreationDateField{data.CreationDate},
		CupsGrain:                   data.GrainCups,
		MixedSubstratePerGrain:      data.SubstrateRatio,
		CasingPerGrain:              data.CasingRatio,
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
		AclField:             acl,
	}
	_, err = data.SubstrateRecipeField.Get(ctx)
	if err != nil {
		http.Error(w, "bad substrate recipe: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	finishImportMainCollectionEntry(ctx, coll, toInsert, data.PermsOnRequest, w)

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
	PermsOnRequest
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
		PermsOnRequest:      upr.PermsOnRequest,
	}
}

type resolvedUpdateFruitingChamberRequest struct {
	KnownFruitableField
	SaleField
	DisposedField
	Notes   AllEntries[Note]
	Images  SplitEntries[picWithNotesForm, PicWithNotes]
	Contams SplitEntries[contamForm, Contamination]
	Flushes SplitEntries[picWithNotesForm, PicWithNotes]
	PermsOnRequest
}

func (out resolvedUpdateFruitingChamberRequest) modsFor(existing *FruitingChamber, aclField AclField) (bson.D, error) {
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

	ctx := r.Context()
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)

	coll := db.Collection(FruitingChamberCollectionName)
	// go get current FC
	existing := &FruitingChamber{}
	err = coll.FindOne(ctx, bson.D{{"_id", id}}).Decode(existing)
	if err != nil {
		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	// TODO: ensure this is ok
	if out.Sale != nil && (existing.Sale == nil || *existing.Sale != *out.Sale) {
		if err = db.Collection(SalesCollectionName).FindOne(ctx, bson.D{{"_id", out.Sale}}).Err(); err != nil {
			dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
			return
		}
	}
	finishMainCollItemUpdate(ctx, w, coll, out.modsFor, existing, data.PermsOnRequest)
}
