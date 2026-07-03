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
// transfers sometimes, creating lcSyringes

// TODO: new (PC is created first, so it can be referenced)?

type LiquidCulture struct {
	MainCollectionIdField             `bson:"inline"`
	PcRunField                        `bson:"inline"` // default for purchased
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
	canTransferTo := []string{GrainJarSourceType, PlateSourceType, SlantSourceType, StasisTubeSourceType, LcSourceType, BagSourceType} // TODO: validate this encompasses all...
	if !slices.Contains(canTransferTo, dst.SourceType()) {
		return errors.New("LC cannot transfer to " + dst.SourceType())
	}
	return nil
}

func (l LiquidCulture) GeneticInfoAsParent() (GeneticParentInfo, error) {
	return GeneticParentInfo{
		SpeciesOptionalField:    SpeciesOptionalField{l.Species},
		SubspeciesOptionalField: SubspeciesOptionalField{l.Subspecies},
		KnownFruitableField:     KnownFruitableField{l.KnownFruitable},
		GenerationsFields:       l.GenerationsFields,
	}, nil
}

func (l LiquidCulture) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return l.GenSinceSpore, l.GenSinceFruitOrSpore
}

func (l LiquidCulture) Innoculatable() error {
	return errors.Join(
		l.RequireNoSpecies(),
		l.RequireNoSubspecies(),
		l.RequireNotDisposed(),
		l.RequireUnknownFruitable(),
		l.RequireNoInnoculation())
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
		withSubspecies(parentInfo.Subspecies).
		withPerms(from.Permissions()).
		withLastUpdated(xfer.LastUpdated).
		Finalized()
	if err != nil {
		return ErrFailedToFinalizeMods
	}
	res, err := mongo.SessionFromContext(ctx).Client().Database(dbName).Collection(LCCollectionName).UpdateByID(ctx, l.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer
	}
	return nil
}

func (l LiquidCulture) id() []byte {
	return []byte(l.Id.dbIdStr())
}

func initializeLCs(ctx context.Context) error {
	db := DbFrom(ctx)
	coll := db.Collection(LCCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		creationDateIndexModel,
		newSimpleIndex("pcRun", "pcRun", false, false, false),
		newSimpleIndex("recipe", "recipe", false, false, false),
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
	return env.IfNotProd(ctx, func() error { // TODO: ensure ok
		// If test LC does not exist, then create it
		testId := mainCollIdForint(idTestLC)
		testItem := &LiquidCulture{
			MainCollectionIdField:   MainCollectionIdField{testId},
			PcRunField:              PcRunField{impPcRun},
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
		testId2 := mainCollIdForint(idTestLC2)
		testItem2 := &LiquidCulture{
			MainCollectionIdField:   MainCollectionIdField{testId2},
			PcRunField:              PcRunField{impPcRun},
			LcRecipeField:           LcRecipeField{exAltId},
			CreationDateField:       CreationDateField{exampleTime},
			SpeciesOptionalField:    SpeciesOptionalField{nil},
			SubspeciesOptionalField: SubspeciesOptionalField{nil},
			InnocField:              InnocField{nil},
			GenerationsFields: GenerationsFields{
				GenSporeField:        GenSporeField{nil},
				GenSinceFruitOrSpore: nil,
			},
			TransfersOutField:                 TransfersOutField{nil},
			ParentTypeField:                   ParentTypeField{nil},
			MainCollectionOptionalParentField: MainCollectionOptionalParentField{Parent: nil},
			PicsField:                         PicsField{nil},
			ConfirmedCleanField:               ConfirmedCleanField{nil},
			ContaminationsField:               ContaminationsField{nil},
			KnownFruitableField:               KnownFruitableField{nil},
			DisposedField:                     DisposedField{nil},
			MostRecentImageField:              MostRecentImageField{nil},
			NotesField:                        NotesField{nil},
			LastUpdatedField:                  LastUpdatedField{exampleTime},
		}
		return addTestMainEntries(ctx, testItem, testItem2)
	})
}

type createLiquidCultureRequest struct {
	LcRecipeField
	PcRunField
	NotesField
	WriteTagToField
}

func createLiquidCultureHandler(w http.ResponseWriter, r *http.Request) {
	data := createLiquidCultureRequest{}
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
	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}

	ctx, now := request.UnixTime(r.Context())

	_, err = data.LcRecipeField.Get(ctx)
	if err != nil {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	toInsert := LiquidCulture{
		MainCollectionIdField: MainCollectionIdField{id},
		LcRecipeField:         data.LcRecipeField,
		PcRunField:            PcRunField{data.PcRun},
		CreationDateField:     CreationDateField{now},
		NotesField:            NotesField{data.Notes},
		LastUpdatedField:      LastUpdatedField{now},
		AclField:              allCanWriteAcl(),
	}

	_, err = toInsert.PcRunField.Get(ctx)
	if err != nil {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	finishCreateMainCollectionEntry(ctx, &toInsert, w)
}

type importLiquidCultureRequest struct {
	CreationDateField
	LcRecipeField
	SpeciesOptionalField // TODO: made optional
	SubspeciesOptionalField
	KnownFruitableField
	Generation *Generation // TODO: make required for when innoculated!
	ConfirmedCleanField
	WriteTagToField
	// image as "img"
}

func importLiquidCultureHandler(w http.ResponseWriter, r *http.Request) {
	data := importLiquidCultureRequest{}
	id := NextMainCollectionId()
	b58id := id.AsBase58()
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
	ctx, now := request.UnixTime(r.Context()) // TODO: ensure r.Context is not used anymore
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
		importedPic = utils.Pointer(newPicWithNotes(now, []Note{}, ImageLocation(newFileNameWithPrefixPath)))
	}
	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
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
		data.ConfirmedClean = nil
	}
	pix := []PicWithNotes{}
	if importedPic != nil {
		pix = []PicWithNotes{*importedPic}
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

	// Validate
	_, err = data.LcRecipeField.Get(ctx)
	if err != nil && errors.Is(err, ErrMissingOptionalField) {
		dbErr(w, "invalid LC recipe: "+err.Error(), http.StatusInternalServerError)
		return
	}
	toInsert := LiquidCulture{
		MainCollectionIdField:   MainCollectionIdField{id},
		PcRunField:              PcRunField{impPcRun},
		LcRecipeField:           LcRecipeField{data.Recipe},
		CreationDateField:       CreationDateField{data.CreationDate},
		SpeciesOptionalField:    SpeciesOptionalField{data.Species},
		SubspeciesOptionalField: data.SubspeciesOptionalField,
		GenerationsFields: GenerationsFields{
			GenSporeField:        GenSporeField{gen},
			GenSinceFruitOrSpore: gen,
		},
		PicsField:            PicsField{pix},
		ConfirmedCleanField:  data.ConfirmedCleanField,
		KnownFruitableField:  data.KnownFruitableField,
		MostRecentImageField: MostRecentImageField{importedPic},
		LastUpdatedField:     LastUpdatedField{now},
		AclField:             AclField{finalPerms},
	}
	finishImportMainCollectionEntry(ctx, &toInsert, w)
}

type updateLiquidCultureRequest struct {
	NotesUpdateField
	KnownFruitableField
	DisposedField
	ConfirmedClean *bool                                                    `json:"confirmedClean,omitempty"`
	Images         SplitEntries[picWithNotesForm, PicWithNotesLessLocation] //"newPic-1"
	Contams        SplitEntries[contamForm, ContaminationLessLocation]      //"newContam-1"
	PermsOnRequest `json:"acl"`
}

func (upr updateLiquidCultureRequest) reform() resolvedUpdateLiquidCultureRequest {
	return resolvedUpdateLiquidCultureRequest{
		ConfirmedClean:      upr.ConfirmedClean,
		KnownFruitableField: upr.KnownFruitableField,
		DisposedField:       upr.DisposedField,
		NotesUpdateField:    upr.NotesUpdateField,
		Images:              imageUpdates(upr.Images),
		Contams:             contamUpdates(upr.Contams),
		PermsOnRequest:      upr.PermsOnRequest,
	}
}

func (req resolvedUpdateLiquidCultureRequest) modsFor(existing *LiquidCulture, aclField AclField) (bson.D, error) {
	return NewMods().
		updateKnownFruitableIfNeeded(req, existing).
		updateConfirmedCleanIfNeeded(req.ConfirmedClean, existing.ConfirmedClean).
		updateDisposedIfNeeded(req, existing).
		updateNotesIfNeeded(req, existing).
		updatePicsIfNeeded(req.Images, existing.Pics).
		updateContamsIfNeeded(req.Contams, existing.Contaminations).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

type resolvedUpdateLiquidCultureRequest struct {
	KnownFruitableField
	Sales []AlternateCollectionId // TODO: maybe do this via a "newSale" endpoint?
	DisposedField
	NotesUpdateField
	ConfirmedClean *bool
	Images         SplitEntries[picWithNotesForm, PicWithNotes]
	Contams        SplitEntries[contamForm, Contamination]
	PermsOnRequest `json:"acl"`
}

// TODO: MOVE ME
func Ternary[T any](val bool, ifTrue, ifFalse T) T {
	if val {
		return ifTrue
	}
	return ifFalse
}

// TODO: MOVE
func TernaryPtr[T any](val *bool, ifTrue, ifFalse, ifNil T) T {
	if val == nil {
		return ifNil
	}
	if *val {
		return ifTrue
	}
	return ifFalse
}

func updateLiquidCultureHandler(w http.ResponseWriter, r *http.Request) {
	data := updateLiquidCultureRequest{}
	idStr, err := UrlDecodeString(r.PathValue("id"))
	if err != nil {
		http.Error(w, "failed to url decode string: "+err.Error(), http.StatusInternalServerError)
		return
	}
	mainCollId, err := StandardizeMainCollectionId(idStr)
	if err != nil {
		env.LogIfDev(r.Context(), "failed to standardize main collection id: "+err.Error())
		http.Error(w, "failed to standardize main collection id: "+err.Error(), http.StatusBadRequest)
		return
	}
	newPics, newContams, _, err := fullMultipartWithNoBreaks(w, r, "lc", &data, mainCollId.AsBase58())
	if err != nil {
		// Already wrotw
		return
	}
	env.LogIfDev(r.Context(), "CONFIRMED CLEAN: "+TernaryPtr(data.ConfirmedClean, "isClean", "isDirty", "empty"))

	// CHECK THAT ALL NEW PICS EXIST
	// PROCESS ALL NEW PICS AND CONTAMS
	req := data.reform()
	for i, _ := range data.Images.New {
		loc, exists := newPics[i]
		if !exists {
			http.Error(w, fmt.Sprintf("error, location for new picture index %d not found (should never happen)", i), http.StatusInternalServerError)
			return
		}
		req.Images.New[i].Location = ImageLocation(loc)
	}
	for i, _ := range data.Contams.New {
		if loc, exists := newContams[i]; exists {
			finalLoc := ImageLocation(loc)
			req.Contams.New[i].Location = &finalLoc
		}
	}
	ctx, db := Db(r)
	coll := db.Collection(LCCollectionName)
	// go get current LC
	existing := LiquidCulture{}
	err = coll.FindOne(ctx, BsonFindFilter("_id", *mainCollId)).Decode(&existing)
	if err != nil {
		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishMainCollItemUpdate(ctx, w, req.modsFor, &existing, req.PermsOnRequest)
}
