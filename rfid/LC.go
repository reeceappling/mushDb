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

type LiquidCulture struct {
	MainCollectionIdField             `bson:"inline"`
	PcRunOptionalField                `bson:"inline"` // likely won't exist for pre-existing or purchased
	LcRecipeField                     `bson:"inline"` // always exists (unless purchased)
	CreationDateField                 `bson:"inline"`
	SpeciesOptionalField              `bson:"inline"`
	SubspeciesOptionalField           `bson:"inline"`
	InnocField                        `bson:"inline"`
	GenerationsFields                 `bson:"inline"`
	TransfersOutField                 `bson:"inline"`
	ParentTypeField                   `bson:"inline"`
	MainCollectionOptionalParentField `bson:"inline"` // Must come from (main) LC, plate, slant, (alt) lcSyringe
	PicsField                         `bson:"inline"`
	ConfirmedCleanField               `bson:"inline"`
	ContaminationsField               `bson:"inline"`
	KnownFruitableField               `bson:"inline"`
	DisposedField                     `bson:"inline"`
	MostRecentImageField              `bson:"inline"`
	NotesField                        `bson:"inline"`
	LastUpdatedField                  `bson:"inline"`
	AclField                          `bson:"inline"`
}

func (l LiquidCulture) CanTransferTo(dst geneticSource) error {
	return errors.New("LiquidCulture cannot transfer this way. Must create a new lcSyringe")
}

func (l LiquidCulture) GeneticInfoAsParent() (GeneticParentInfo, error) {
	return GeneticParentInfo{
		SpeciesOptionalField:    SpeciesOptionalField{l.Species},
		SubspeciesOptionalField: SubspeciesOptionalField{l.SubSpecies},
		KnownFruitableField:     KnownFruitableField{l.KnownFruitable},
		GenerationsFields:       l.GenerationsFields,
	}, nil
}

func (l LiquidCulture) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return l.GenSinceSpore, l.GenSinceFruitOrSpore
}

func (l LiquidCulture) setTransferParent(ctx context.Context, xfer Transfer) (error, func() error) {
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(LCCollectionName)
	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
	if err != nil {
		return err, nil
	}
	res, err := coll.UpdateByID(ctx, l.Id, upd)
	if err != nil {
		return err, nil
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer, nil
	}
	return nil, func() error {
		return coll.FindOneAndReplace(ctx, bson.D{{"_id", l.Id}}, l).Err()
	}
}

func (l LiquidCulture) setTransferChild(ctx context.Context, xfer Transfer, from geneticSource) error {
	parentInfo, genSpore, genFruitSpore, err := childGensForParent(from)
	if err != nil {
		return err
	}
	upd, err := xfer.PicsModsForChild().
		withInnoc(xfer).
		withParentType(&xfer.FromType).
		withParent(utils.Pointer(from.DbId())).
		withGens(genSpore, genFruitSpore).
		withSpecies(parentInfo.Species).
		withSubspecies(parentInfo.SubSpecies).
		withLastUpdated(xfer.LastUpdated).
		Finalized()
	if err != nil {
		return ErrFailedToFinalizeMods
	}
	res, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(LCCollectionName).UpdateByID(ctx, l.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer
	}
	return nil
}

func (l LiquidCulture) EntryTypeField() *string {
	return utils.Pointer(LcSourceType)
}

func (l LiquidCulture) id() []byte {
	return []byte(l.Id.dbIdStr())
}

func initializeLCs(ctx context.Context) error {
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(LCCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		newSimpleIndex("pcRun", "pcRun", false, true, false),
		newSimpleIndex("recipe", "recipe", false, false, false),
		creationDateIndexModel,
		newSimpleIndex("species", "species", false, true, false),
		newSimpleIndex("subspecies", "subspecies", false, true, false),
		//newSimpleIndex("innoc", "innoc", false, true, false),
		//newSimpleIndex("genSinceSpore", "genSpore", true, true, false),
		//newSimpleIndex("genSinceFruitOrSpore", "genFruitOrSpore", true, true, false),
		//transfersOutIndexModel,
		//newSimpleIndex("parent", "parent", false, true, false),         // TODO: nil is store or outside?
		//newSimpleIndex("parentType", "parentType", false, true, false), // TODO: nil is store or outside?
		//Pics (no index)
		//newSimpleIndex("confirmedClean", "confirmedClean", false, true, false),
		// TODO: Contams
		// Flushes
		//newSimpleIndex("knownFruitable", "knownFruitable", false, true, false),
		//newSimpleIndex("disposed", "disposed", false, true, false),
		// MostRecentImage
		//Notes (no index) (maybe later with tags?)
		projectsIndexModel,
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	testId := mainCollIdForint(idTestLC)
	testItem := &LiquidCulture{
		MainCollectionIdField:   MainCollectionIdField{testId},
		PcRunOptionalField:      PcRunOptionalField{&exAltId},
		LcRecipeField:           LcRecipeField{exAltId},
		CreationDateField:       CreationDateField{exampleTime},
		SpeciesOptionalField:    SpeciesOptionalField{&exampleSpecies},
		SubspeciesOptionalField: SubspeciesOptionalField{exampleSubspecies},
		InnocField:              InnocField{&exAltId},
		GenerationsFields: GenerationsFields{
			GenSporeField:        GenSporeField{&exGenSinceSpore},
			GenSinceFruitOrSpore: &exGenSinceFruitSpore,
		},
		TransfersOutField:                 TransfersOutField{exAlts},
		ParentTypeField:                   ParentTypeField{&exParentType},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{Parent: &exPlate},
		PicsField:                         PicsField{exPics},
		ConfirmedCleanField:               ConfirmedCleanField{exBool},
		ContaminationsField:               ContaminationsField{exContams},
		KnownFruitableField:               KnownFruitableField{exBool},
		DisposedField:                     DisposedField{&exampleTime},
		MostRecentImageField:              MostRecentImageField{&exPics[0]},
		NotesField:                        NotesField{exampleNotes()},
		LastUpdatedField:                  LastUpdatedField{exampleTime},
	}
	return addTestMainEntries(ctx, testItem)
}

type createLiquidCultureRequest struct {
	LcRecipeField
	CreationDateField
	PcRunField
	NotesField
	WriteTagToField
}

func createLiquidCultureHandler(w http.ResponseWriter, r *http.Request) {
	data := createLiquidCultureRequest{}
	id, err := newCollectionId(r.Context(), LCCollectionName)
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
	now := unixTimeForNow()
	ctx, db := Db(r)
	coll := db.Collection(LCCollectionName)

	_, err = data.LcRecipeField.Get(ctx)
	if err != nil {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	toInsert := LiquidCulture{
		MainCollectionIdField: MainCollectionIdField{id},
		LcRecipeField:         data.LcRecipeField,
		PcRunOptionalField:    PcRunOptionalField{&data.PcRun},
		CreationDateField:     CreationDateField{data.CreationDate},
		NotesField:            NotesField{data.Notes},
		LastUpdatedField:      LastUpdatedField{now},
		AclField:              allCanWriteAcl(),
	}

	_, err = toInsert.PcRunOptionalField.Get(ctx)
	if err != nil && !errors.Is(err, ErrMissingOptionalField) {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	finishCreateMainCollectionEntry(ctx, coll, &toInsert, w)
}

type importLiquidCultureRequest struct {
	CreationDateField
	LcRecipeField
	SpeciesField
	SubspeciesOptionalField
	KnownFruitableField
	Generation *int
	ConfirmedCleanField
	WriteTagToField
	PermsOnRequest
	// image as "img"
}

func importLiquidCultureHandler(w http.ResponseWriter, r *http.Request) {
	data := importLiquidCultureRequest{}
	id, err := newCollectionId(r.Context(), LCCollectionName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	b58id := id.asBase58()
	r.Body = http.MaxBytesReader(w, r.Body, maxMultipartRequestSize) // TODO: do the multipart reader differently
	defer r.Body.Close()
	reader, err := r.MultipartReader() // TODO: do streamlined
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
	//authinfo, err := GetAuthInfo(r.Context())
	//if err != nil {
	//	http.Error(w, "failed to get auth info: "+err.Error(), http.StatusUnauthorized)
	//	return
	//}
	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
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
		fieldBytes, err := multipartToImageBytes(p, w)
		if err != nil {
			// Already wrote
			return
		}

		newFileNameWithPrefixPath, errr := pics.SaveFile(r.Context(), fieldBytes, "lc", string(b58id), "img")
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
	user, err := GetAuthInfo(r.Context())
	if err != nil {
		http.Error(w, "failed to get auth info: "+err.Error(), http.StatusUnauthorized)
		return
	}
	sp, subsp, err := getSpeciesAndSubspecies(r.Context(), data.Species, data.SubSpecies)
	if err != nil {
		http.Error(w, "failed to get species or subspecies: "+err.Error(), http.StatusInternalServerError) // TODO: normalize
		return
	}
	var finalPerms *ACL = nil
	if subsp != nil {
		finalPerms = subsp.DefaultAcl.Clone()
	} else {
		sp.DefaultAcl.Clone()
	}
	// Add user to the acl as a writer
	finalPerms.Users[user.Email] = true

	ctx, db := Db(r)
	coll := db.Collection(LCCollectionName)
	// Validate
	_, err = data.LcRecipeField.Get(ctx)
	if err != nil && errors.Is(err, ErrMissingOptionalField) {
		dbErr(w, "invalid LC recipe: "+err.Error(), http.StatusInternalServerError)
		return
	}
	toInsert := LiquidCulture{
		MainCollectionIdField: MainCollectionIdField{id},
		//PcRunOptionalField:      PcRunOptionalField{},       // No pc runs on imports
		LcRecipeField:           LcRecipeField{data.Recipe},
		CreationDateField:       CreationDateField{data.CreationDate},
		SpeciesOptionalField:    SpeciesOptionalField{&data.Species},
		SubspeciesOptionalField: data.SubspeciesOptionalField,
		GenerationsFields: GenerationsFields{
			GenSporeField:        GenSporeField{gen},
			GenSinceFruitOrSpore: gen,
		},
		PicsField:            PicsField{pix},
		ConfirmedCleanField:  data.ConfirmedCleanField,
		KnownFruitableField:  data.KnownFruitableField,
		MostRecentImageField: MostRecentImageField{importedPic},
		LastUpdatedField:     LastUpdatedField{unixTimeForNow()},
		AclField:             AclField{finalPerms},
	}
	finishImportMainCollectionEntry(ctx, coll, &toInsert, data.PermsOnRequest, w)
}

type updateLiquidCultureRequest struct {
	Notes AllEntries[Note]
	KnownFruitableField
	DisposedField
	ConfirmedClean *bool
	Images         SplitEntries[picWithNotesForm, PicWithNotesLessLocation] //"newPic-1"
	Contams        SplitEntries[contamForm, ContaminationLessLocation]      //"newContam-1"
	WriteTagToField
	PermsOnRequest
}

func (upr updateLiquidCultureRequest) reform() resolvedUpdateLiquidCultureRequest {
	return resolvedUpdateLiquidCultureRequest{
		ConfirmedClean:      upr.ConfirmedClean,
		KnownFruitableField: upr.KnownFruitableField,
		DisposedField:       upr.DisposedField,
		Notes:               upr.Notes,
		Images:              imageUpdates(upr.Images),
		Contams:             contamUpdates(upr.Contams),
		PermsOnRequest:      upr.PermsOnRequest,
	}
}

func (mods resolvedUpdateLiquidCultureRequest) modsFor(existing *LiquidCulture, aclField AclField) (bson.D, error) {
	return NewMods().
		updateKnownFruitableIfNeeded(mods.KnownFruitable, existing.KnownFruitable).
		updateConfirmedCleanIfNeeded(mods.ConfirmedClean, existing.ConfirmedClean).
		updateDisposedIfNeeded(mods.Disposed, existing.Disposed).
		updateNotesIfNeeded(mods.Notes, existing.Notes).
		updatePicsIfNeeded(mods.Images, existing.Pics).
		updateContamsIfNeeded(mods.Contams, existing.Contaminations).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

type resolvedUpdateLiquidCultureRequest struct {
	KnownFruitableField
	Sales []AlternateCollectionId // TODO: maybe do this via a "newSale" endpoint?
	DisposedField
	Notes          AllEntries[Note]
	ConfirmedClean *bool
	Images         SplitEntries[picWithNotesForm, PicWithNotes]
	Contams        SplitEntries[contamForm, Contamination]
	PermsOnRequest
}

func updateLiquidCultureHandler(w http.ResponseWriter, r *http.Request) {
	data := updateLiquidCultureRequest{}
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
	newPics, newContams, _, err := getMultipartImages(r.Context(), "lc", w, reader, b58Id)
	if err != nil {
		// Already wrotw
		return
	}

	// CHECK THAT ALL NEW PICS EXIST
	// PROCESS ALL NEW PICS AND CONTAMS
	req := data.reform()
	for i, _ := range data.Images.New {
		loc, exists := newPics[i]
		if !exists {
			http.Error(w, fmt.Sprintf("error, location for new picture index %d not found (should never happen)", i), http.StatusInternalServerError)
			return
		}
		req.Images.New[i].Location = imageLocation(loc)
	}
	for i, _ := range data.Contams.New {
		if loc, exists := newContams[i]; exists {
			finalLoc := imageLocation(loc)
			req.Contams.New[i].Location = &finalLoc
		}
	}
	ctx, db := Db(r)
	coll := db.Collection(LCCollectionName)
	// go get current LC
	existing := LiquidCulture{}
	err = coll.FindOne(ctx, bson.D{{"_id", id}}).Decode(&existing)
	if err != nil {
		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishMainCollItemUpdate(ctx, w, coll, req.modsFor, &existing, req.PermsOnRequest)
}
