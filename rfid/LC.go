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
	LcSourceType = "lc"
	LCIdPrefix   = "LC" // TODO: USE
)

type LiquidCulture struct { // TODO: LIQUID CULTURE SYRINGE???
	EntryTypeStructField
	MainCollectionIdField
	PcRunOptionalField // likely won't exist for pre-existing or purchased
	PressureCookedTouchingWaterOptionalField

	LcRecipeField // likely won't exist for pre-existing or purchased
	CreationDateField
	SpeciesOptionalField
	SubspeciesOptionalField
	InnocField
	GenerationsFields
	TransfersOutField
	ParentTypeField
	BinaryOptionalParentField // TODO: BRAND NEW! // Must come from (main) LC, plate, slant, (alt) lcSyringe
	PicsField
	ConfirmedClean *bool `bson:"confirmedClean,omitempty" json:"confirmedClean,omitempty"` // TODO: change so that we know exactly what confirmed it?
	ContaminationsField
	KnownFruitableField
	DisposedField
	MostRecentImageField
	Sales []AlternateCollectionId `bson:"sales,omitempty" json:"sales,omitempty"`
	NotesField
	LastUpdatedField
	PermsField
}

func (l LiquidCulture) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := l
	err := decodeItem(&out, encoded)
	return out, err
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

func (l LiquidCulture) SourceType() string {
	return LcSourceType
}

func (l LiquidCulture) setTransferParent(ctx mongo.SessionContext, xfer Transfer) error {
	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
	if err != nil {
		return err
	}
	res, err := ctx.Client().Database(dbName).Collection(mainCollectionName).UpdateByID(ctx, l.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer
	}
	return nil
}

func (l LiquidCulture) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
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
	res, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(mainCollectionName).UpdateByID(ctx, l.Id, upd)
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

func (l LiquidCulture) CollectionName() string {
	return mainCollectionName
}

func (l LiquidCulture) id() []byte {
	return l.Id[:]
}

func initializeLCs(ctx context.Context) error {
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(mainCollectionName)
	// If test agar batch does not exist, then create it
	existingEntry := LiquidCulture{}
	testId := mainCollIdForint(idTestLC)
	testItem := LiquidCulture{
		EntryTypeStructField:    EntryTypeStructField{*existingEntry.EntryTypeField()},
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
		TransfersOutField:         TransfersOutField{exAlts},
		ParentTypeField:           ParentTypeField{&exParentType},
		BinaryOptionalParentField: BinaryOptionalParentField{Parent: utils.Pointer(exPlate.ToBinaryCollectionId())},
		PicsField:                 PicsField{exPics},
		ConfirmedClean:            exBool,
		ContaminationsField:       ContaminationsField{exContams},
		KnownFruitableField:       KnownFruitableField{exBool},
		DisposedField:             DisposedField{&exampleTime},
		MostRecentImageField:      MostRecentImageField{&exPics[0]},
		Sales:                     exAlts,
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

type createLiquidCultureRequest struct {
	LcRecipeField
	CreationDateField
	PcRunField
	NotesField
	WriteTagToField
}

func createLiquidCultureHandler(w http.ResponseWriter, r *http.Request) {
	data := createLiquidCultureRequest{}
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
	now := unixTimeForNow()
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		_, err = data.LcRecipeField.Get(ctx)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		toInsert := LiquidCulture{
			EntryTypeStructField:  EntryTypeStructField{"lc"},
			MainCollectionIdField: MainCollectionIdField{id},
			LcRecipeField:         data.LcRecipeField,
			PcRunOptionalField:    PcRunOptionalField{&data.PcRun},
			CreationDateField:     CreationDateField{data.CreationDate},
			NotesField:            NotesField{data.Notes},
			LastUpdatedField:      LastUpdatedField{now},
			// No perms because initial is for everyone
		}

		_, err = toInsert.PcRunOptionalField.Get(ctx)
		if err != nil && !errors.Is(err, ErrMissingOptionalField) {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		coll := ctx.Client().Database(dbName).Collection(mainCollectionName)
		_, err = coll.InsertOne(ctx, toInsert)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		bs, err = json.Marshal(toInsert)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		return w.Write(bs)
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}

type importLiquidCultureRequest struct {
	CreationDateField
	LcRecipeField
	SpeciesField
	SubspeciesOptionalField
	KnownFruitableField
	Generation     *int
	ConfirmedClean *bool
	WriteTagToField
	PermsField
	// image as "img"
}

func importLiquidCultureHandler(w http.ResponseWriter, r *http.Request) {
	data := importLiquidCultureRequest{}
	id, err := generateMainCollectionId(r.Context())
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
		http.Error(w, "unable to read data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	// PARSE INTO CORRECT DATA FORMAT
	err = json.Unmarshal(bs, &data)
	if err != nil {
		http.Error(w, "unable to unmarshal json form data: "+err.Error(), http.StatusBadRequest)
		return
	}
	authinfo, err := GetAuthInfo(r.Context())
	if err != nil {
		http.Error(w, "failed to get auth info: "+err.Error(), http.StatusUnauthorized)
		return
	}
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
	spec, subsp, err := getSpeciesAndSubspecies(r.Context(), data.Species, data.SubSpecies)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get species and subspecies: %s", err), http.StatusInternalServerError)
		return
	}
	finalPerms := minimalPermsBetween(data.Perms, spec, subsp)
	finalPerms.Users = finalPerms.Users.WithAuthor(authinfo.Id) // Add user to perms if not already there
	out := LiquidCulture{
		EntryTypeStructField:  EntryTypeStructField{"lc"},
		MainCollectionIdField: MainCollectionIdField{id},
		//PcRunOptionalField:      PcRunOptionalField{},       // No pc runs on imports
		LcRecipeField:           LcRecipeField{data.Recipe}, // TODO: optional for imports?
		CreationDateField:       CreationDateField{data.CreationDate},
		SpeciesOptionalField:    SpeciesOptionalField{&data.Species},
		SubspeciesOptionalField: data.SubspeciesOptionalField,
		GenerationsFields: GenerationsFields{
			GenSporeField:        GenSporeField{gen},
			GenSinceFruitOrSpore: gen,
		},
		PicsField:            PicsField{pix},
		ConfirmedClean:       data.ConfirmedClean,
		KnownFruitableField:  data.KnownFruitableField,
		MostRecentImageField: MostRecentImageField{importedPic},
		LastUpdatedField:     LastUpdatedField{unixTimeForNow()},
		PermsField:           PermsField{finalPerms},
	}

	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		_, err = out.LcRecipeField.Get(ctx)
		if err != nil && errors.Is(err, ErrMissingOptionalField) {
			return DbTxnStdErr(w, "invalid LC recipe: "+err.Error(), http.StatusInternalServerError)
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

type updateLiquidCultureRequest struct {
	Notes AllEntries[Note] // TODO: change into an anonymous struct???
	KnownFruitableField
	DisposedField
	Sales          []AlternateCollectionId `json:"sale,omitempty"` // TODO: validate
	ConfirmedClean *bool
	Images         SplitEntries[picWithNotesForm, PicWithNotesLessLocation] //"newPic-1"
	Contams        SplitEntries[contamForm, ContaminationLessLocation]      //"newContam-1"
	WriteTagToField
	PermsField
}

func (upr updateLiquidCultureRequest) reform() resolvedUpdateLiquidCultureRequest {
	return resolvedUpdateLiquidCultureRequest{
		ConfirmedClean:      upr.ConfirmedClean,
		KnownFruitableField: upr.KnownFruitableField,
		Sales:               upr.Sales,
		DisposedField:       upr.DisposedField,
		Notes:               upr.Notes,
		Images:              imageUpdates(upr.Images),
		Contams:             contamUpdates(upr.Contams),
		PermsField:          upr.PermsField,
	}
}

func (mods resolvedUpdateLiquidCultureRequest) modsFor(existing LiquidCulture) (bson.D, error) {
	return NewMods().
		updateKnownFruitableIfNeeded(mods.KnownFruitable, existing.KnownFruitable).
		updateConfirmedCleanIfNeeded(mods.ConfirmedClean, existing.ConfirmedClean).
		updateSalesIfNeeded(mods.Sales, existing.Sales). // TODO: maybe do this via a "newSale" endpoint?
		updateDisposedIfNeeded(mods.Disposed, existing.Disposed).
		updateNotesIfNeeded(mods.Notes, existing.Notes).
		updatePicsIfNeeded(mods.Images, existing.Pics).
		updateContamsIfNeeded(mods.Contams, existing.Contaminations).
		updatePermsIfNeeded(mods.Perms, existing.Perms).
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
	PermsField
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
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(mainCollectionName)
		// go get current plate
		existing := LiquidCulture{}
		err := coll.FindOne(ctx, bson.D{{"_id", id}}).Decode(&existing)
		if err != nil {
			return DbTxnStdErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		}
		restrictivePerms := minimalPermsBetween(data.Perms, existing.Perms)
		if err = restrictivePerms.ValidateUserCanWrite(ctx); err != nil {
			return DbTxnStdErr(w, "does not meet either old or new permissions: "+err.Error(), http.StatusBadRequest)
		}
		upd, err := out.modsFor(existing)
		return handleUpdateMods(ctx, w, coll, existing, id, upd, err) // TODO: use this on all updates!!!!!!!
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}
