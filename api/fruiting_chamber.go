package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/reeceappling/goUtils/v2/utils"
	"github.com/reeceappling/mushDb/api/pics"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
)

// needed for creating fruits (unless fruits came from agar or bag)

// TODO: HANDLE MULTIPLE GRAIN INPUTS FOR MONOTUBS (DO MONOTUBS LATER)
type FruitingChamber struct { // TODO: SHOEBOX vs monotub!
	MainCollectionIdField             `bson:"inline"`
	CreationDateField                 `bson:"inline"`
	SubstrateRecipeField              `bson:"inline"`
	SubstrateBatchOptionalField       `bson:"inline"`
	CupsGrain                         float64 `bson:"cupsGrain" json:"cupsGrain"`
	MixedSubstratePerGrain            float64 `bson:"mixedSubstratePerGrain" json:"mixedSubstratePerGrain"` // for a 1:1:0.5 (1 part substrate, 1 part grain, 0.5 part casing) box this will be 1
	CasingPerGrain                    float64 `bson:"casingPerGrain" json:"casingPerGrain"`                 // No casing==0, half casing per grain == 0.5
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

//func (f FruitingChamber) setTransferParent(ctx context.Context, xfer Transfer) error {
//	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(FruitingChamberCollectionName)
//	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
//	if err != nil {
//		return err
//	}
//	res, err := coll.UpdateByID(ctx, f.Id, upd)
//	if err != nil {
//		return err
//	}
//	if res.ModifiedCount == 0 {
//		return ErrNoParentModifiedForTransfer
//	}
//	return nil
//}

// TODO: create box via jar instead? Probably want to do it that way...
func (f FruitingChamber) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
	parentInfo, genSpore, genFruitSpore, err := childGensForParent(from)
	if err != nil {
		return err
	}
	mods := NewMods()
	pictures := []PicWithNotes{}
	if xfer.ToImage != nil {
		pic := newPicWithNotes(xfer.CreationDate, []Note{}, *xfer.ToImage)
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
		withSubspecies(parentInfo.Subspecies).
		withKnownFruitable(parentInfo.KnownFruitable).
		withPerms(from.Permissions()).
		withLastUpdated(xfer.LastUpdated).
		Finalized()
	if err != nil {
		return errors.New("failed to finalize") // TODO: ok?
	}
	res, err := mongo.SessionFromContext(ctx).Client().Database(dbName).Collection(FruitingChamberCollectionName).UpdateByID(ctx, f.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer
	}
	return nil
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
	// TODO: removed: Recipe // substrate recipe (pull from batch)
	SubstrateBatchField
	//ParentJar          MainCollectionId // Parent jar // TODO: do we want this? // TODO: ALLOW USER TO INPUT PARENT AND CHAIN A TRANSFER CREATION AS WELL!
	GrainCups          float64
	MixedSubstrateCups float64
	CasingCups         float64
	NotesField
	WriteTagToField
}

func createFruitingChamberHandler(w http.ResponseWriter, r *http.Request) {
	data := createFruitingChamberRequest{}
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
	// Validation
	//parentJar, err := LookupGrainJar(ctx, data.ParentJar)
	//if err != nil {
	//	http.Error(w, "failed to resolve parent jar"+err.Error(), http.StatusBadRequest)
	//	return
	//}

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
	//innoc := Transfer{
	//	AlternateCollectionIdField: AlternateCollectionIdField{newAlternateCollectionId()},
	//	From:                       parentJar.Id,
	//	To:                         id,
	//	FromType:                   "jar", // TODO: allow different types? lcSyringe? spores?
	//	ToType:                     "fruitingChamber",
	//	CreationDateField:          CreationDateField{now},
	//	Reason:                     xferReasonColonized,
	//	NotesField: NotesField{[]Note{{
	//		RequiredTimeField: RequiredTimeField{now},
	//		Note:              "automated transfer created for new box",
	//	}}},
	//	LastUpdatedField: LastUpdatedField{now},
	//	AclField:         parentJar.AclField,
	//} // TODO: ADD INNOC!
	//newTxn(ctx, func(sessCtx mongo.SessionContext)(any, error){
	//	db := mongo.SessionFromContext(sessCtx).Client().Database(dbName)
	//	db.Collection()
	//})
	//// TODO: handle parent! Add xfer to
	//toInsert := FruitingChamber{
	//	//SpeciesOptionalField:              parentJar.SpeciesOptionalField,
	//	//SubspeciesOptionalField:           parentJar.SubspeciesOptionalField,
	//	//MainCollectionOptionalParentField: MainCollectionOptionalParentField{&data.ParentJar},
	//	//InnocField:                        InnocField{
	//	//	// TODO: CREATE INNOC!
	//	//}, // TODO: THIS!
	//	//GenerationsFields: GenerationsFields{
	//	//	GenSporeField:        GenSporeField{parentJar.GenSinceSpore.Next()},
	//	//	GenSinceFruitOrSpore: parentJar.GenSinceFruitOrSpore.Next(),
	//	//},
	//	//ParentTypeField: ParentTypeField{&innoc.FromType},
	//
	//	//MainCollectionIdField:       MainCollectionIdField{id},
	//	//SubstrateRecipeField:        batch.SubstrateRecipeField,
	//	//SubstrateBatchOptionalField: data.SubstrateBatchField.asOptional(),
	//	//CupsGrain:                   data.GrainCups,
	//	//MixedSubstratePerGrain:      data.MixedSubstrateCups / data.GrainCups,
	//	//CasingPerGrain:              data.CasingCups / data.GrainCups,
	//	//CreationDateField:           CreationDateField{now},
	//	//NotesField:                  NotesField{data.Notes},
	//	//LastUpdatedField:            LastUpdatedField{now},
	//	//AclField:                    parentJar.AclField, // TODO: what if parent does not exist? readonly for creator?
	//}
	toInsert := FruitingChamber{
		MainCollectionIdField:       MainCollectionIdField{id},
		SubstrateRecipeField:        batch.SubstrateRecipeField,
		SubstrateBatchOptionalField: data.SubstrateBatchField.asOptional(),
		CupsGrain:                   data.GrainCups,
		MixedSubstratePerGrain:      data.MixedSubstrateCups / data.GrainCups,
		CasingPerGrain:              data.CasingCups / data.GrainCups,
		CreationDateField:           CreationDateField{now},
		NotesField:                  NotesField{data.Notes},
		LastUpdatedField:            LastUpdatedField{now},
		AclField:                    allCanReadAcl(GetUserEmailPtr(ctx)),
	}
	finishCreateMainCollectionEntry(ctx, &toInsert, w)
}

type importFruitingChamberRequest struct {
	SubstrateRecipeField
	CreationDateField
	SpeciesField   // FCs cannot be imported without species!
	GrainCups      float64
	SubstrateRatio float64 // TODO: used to be optional
	CasingRatio    float64 // TODO: used to be optional
	SubspeciesOptionalField
	Generation *int
	KnownFruitableField
	WriteTagToField
	//PermsOnRequest `json:"acl"` // from spec/subspec
	// image as "img"
}

func importFruitingChamberHandler(w http.ResponseWriter, r *http.Request) {
	data := importFruitingChamberRequest{}
	id := NextMainCollectionId()
	b58id := id.AsBase58()
	r.Body = http.MaxBytesReader(w, r.Body, maxMultipartRequestSize) // TODO: REDO THIS MULTIPART READER? Old comment
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

		importedPic = utils.Pointer(newPicWithNotes(now, []Note{}, ImageLocation(newFileNameWithPrefixPath)))
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
	finalPerms, err := ImportFinalPerms(r.Context(), data.Species, data.Subspecies)
	if err != nil {
		http.Error(w, "failed to get species and/or subspecies: "+err.Error(), http.StatusInternalServerError)
		return
	}

	toInsert := &FruitingChamber{
		MainCollectionIdField:       MainCollectionIdField{id},
		SubstrateRecipeField:        data.SubstrateRecipeField,        // TODO: allow unknown substrate recipe???
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
		AclField:             finalPerms.AsField(),
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
	finishImportMainCollectionEntry(ctx, toInsert, w)

}

type updateFruitingChamberRequest struct {
	NotesUpdateField
	KnownFruitableField
	DisposedField
	SaleField
	ImagesUpdateField  //"newPic-1"
	ContamsUpdateField //"newContam-1"
	FlushesUpdateField //"newFlush-1"
	PermsOnRequest     `json:"acl"`
}

// TODO: MOVE THESE 3!!!!
type ImagesUpdateField struct {
	Images SplitEntries[picWithNotesForm, PicWithNotesLessLocation] `json:"images"` //"newPic-1"
}
type ContamsUpdateField struct {
	Contams SplitEntries[contamForm, ContaminationLessLocation] `json:"contams"` //"newContam-1"
}
type FlushesUpdateField struct {
	Flushes SplitEntries[picWithNotesForm, PicWithNotesLessLocation] `json:"flushes"` //"newFlush-1"
}

func (upr updateFruitingChamberRequest) reform() resolvedUpdateFruitingChamberRequest {
	return resolvedUpdateFruitingChamberRequest{
		KnownFruitableField: upr.KnownFruitableField,
		SaleField:           upr.SaleField,
		DisposedField:       upr.DisposedField,
		NotesUpdateField:    upr.NotesUpdateField,
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
	NotesUpdateField
	Images         SplitEntries[picWithNotesForm, PicWithNotes]
	Contams        SplitEntries[contamForm, Contamination]
	Flushes        SplitEntries[picWithNotesForm, PicWithNotes]
	PermsOnRequest `json:"acl"`
}

func (req resolvedUpdateFruitingChamberRequest) modsFor(existing *FruitingChamber, aclField AclField) (bson.D, error) {
	return NewMods().
		updateKnownFruitableIfNeeded(req, existing).
		updateSaleIfNeeded(req.Sale, existing.Sale).
		updateDisposedIfNeeded(req, existing).
		updateNotesIfNeeded(req, existing).
		updatePicsIfNeeded(req.Images, existing.Pics).
		updateContamsIfNeeded(req.Contams, existing.Contaminations).
		updateFlushesIfNeeded(req.Flushes, existing.Flushes).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateFruitingChamberHandler(w http.ResponseWriter, r *http.Request) {
	data := updateFruitingChamberRequest{}
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

	newPics, newContams, newFlushes, err := fullMultipartWithNoBreaks(w, r, "fruitingChamber", &data, mainCollId.AsBase58())
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
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)

	coll := db.Collection(FruitingChamberCollectionName)
	// go get current FC
	existing := &FruitingChamber{}
	err = coll.FindOne(ctx, BsonFindFilter("_id", *mainCollId)).Decode(existing)
	if err != nil {
		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	// TODO: ensure this is ok. Handle sales elsewhere????
	//if out.Sale != nil && (existing.Sale == nil || *existing.Sale != *out.Sale) {
	//	if err = db.Collection(SalesCollectionName).FindOne(ctx, BsonFindFilter("_id", out.Sale)).Err(); err != nil {
	//		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
	//		return
	//	}
	//}
	finishMainCollItemUpdate(ctx, w, out.modsFor, existing, data.PermsOnRequest)
}
