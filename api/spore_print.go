package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/reeceappling/goUtils/v2/utils"
	"github.com/reeceappling/goUtils/v2/utils/slices"
	"github.com/reeceappling/mushDb/api/env"
	"github.com/reeceappling/mushDb/api/pics"
	"github.com/reeceappling/mushDb/api/request"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	slices2 "slices"
	"strconv"
	"strings"
)

// TODO: harvestFruit from bag/box page? How do I want to track which fruit without writing on them or tagging them? Create a flush collection?
// TODO: createSporePrint from fruit page
// TODO: createSerialSporePrint from sporePrint page

type SporePrint struct {
	MainCollectionIdField `bson:"inline"`
	// Parent is always either fruit, or purchased
	MainCollectionOptionalParentField `bson:"inline"` // won't exist for imports only
	CreationDateField                 `bson:"inline"` // Print or receive date
	SporePrintColorField              `bson:"inline"` // Set later on the print, not on creation, but does get added on import if possible
	SporePrintDensityField            `bson:"inline"` // Set later on the print, not on creation, but does get added on import if possible
	SpeciesField                      `bson:"inline"`
	SubspeciesOptionalField           `bson:"inline"`
	PicsField                         `bson:"inline"`
	SaleField                         `bson:"inline"`
	DisposedField                     `bson:"inline"`
	MostRecentImageField              `bson:"inline"`
	NotesField                        `bson:"inline"`
	LastUpdatedField                  `bson:"inline"`
	AclField                          `bson:"inline"`
}

type SporePrintColor string

const (
	SpColorBlack    SporePrintColor = "Black"
	SpColorTanLight SporePrintColor = "LightTan"
	SpColorClear    SporePrintColor = "Clear"
)

type SporePrintDensity string

const (
	SpDensityHeavy       SporePrintDensity = "Heavy"
	SpDensityAvg         SporePrintDensity = "Average"
	SpDensitySparse      SporePrintDensity = "Sparse"
	spDensityNoneMinimal SporePrintDensity = "None or Minimal"
)

type SporePrintColorField struct {
	Color *SporePrintColor `bson:"color,omitempty" json:"color,omitempty"`
}
type SporePrintDensityField struct {
	Density *SporePrintDensity `bson:"density,omitempty" json:"density,omitempty"`
}

func (sp SporePrint) Innoculatable() error {
	return errors.New("sporePrints not innoculatable")
}

func (sp SporePrint) CanTransferTo(dst geneticSource) error {
	// TODO: allow transfer to plate????
	return errors.New("sporePrints cannot transfer. Only be made into mss or swab")
}
func (sp SporePrint) createSwabInTxn(ctx mongo.SessionContext, swabNotes, xferNotes NotesField) (*SporeSwab, error) {
	ctx, now := request.UnixTimeInTxn(ctx)
	idOut := NextMainCollectionId()
	db := mongo.SessionFromContext(ctx).Client().Database(dbName)
	swab := SporeSwab{
		MainCollectionIdField:             MainCollectionIdField{idOut},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&sp.Id},
		ParentTypeField:                   ParentTypeField{utils.Pointer("sporePrint")},
		CreationDateField:                 CreationDateField{now},
		SpeciesField:                      sp.SpeciesField,
		SubspeciesOptionalField:           sp.SubspeciesOptionalField,
		NotesField:                        swabNotes,
		LastUpdatedField:                  LastUpdatedField{now},
		AclField:                          sp.AclField,
	}
	xfer := Transfer{
		AlternateCollectionIdField: AlternateCollectionIdField{newAlternateCollectionId()},
		From:                       sp.Id,
		To:                         idOut,
		FromType:                   "sporePrint",
		ToType:                     "sporeSwab",
		CreationDateField:          CreationDateField{now},
		Reason:                     xferReasonReady,
		NotesField:                 xferNotes,
		LastUpdatedField:           LastUpdatedField{now},
		AclField:                   sp.AclField,
	}
	err := addToIdMapCollection(ctx, &swab)
	if err != nil {
		return nil, err
	}
	// Update print with new swab id
	// Update xfers out and lastUpdated on parent
	upd, err := NewMods().Push("transfersOut", xfer.Id).withLastUpdated(now).Finalized()
	if err != nil {
		return nil, err
	}
	_, err = db.Collection(SporePrintCollectionName).UpdateByID(ctx, sp.Id, upd)
	if err != nil {
		return nil, err
	}
	_, err = db.Collection(SporeSwabCollectionName).InsertOne(ctx, &swab)
	if err != nil {
		return nil, errors.Join(errors.New("failed to insert new spore print"), err)
	}
	_, err = db.Collection(TransfersCollName).InsertOne(ctx, &xfer)
	if err != nil {
		return nil, errors.Join(errors.New("failed to insert new spore print"), err)
	}
	return &swab, nil
}

// TODO: createSporePrint should be its own endpoint which accepts a fruit. It can also be called from other spore print pages to do "chaining"
func (sp SporePrint) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
	return errors.New("spore prints cannot be the destination of a transfer")
	//// TODO: can this happen????? should always be from a fruit right?
	//// This is a special case because it will always be 0-gen
	//parentInfo, err := from.GeneticInfoAsParent()
	//if err != nil {
	//	return err
	//}
	//if parentInfo.Species == nil {
	//	return errors.New("parent must have a species")
	//}
	//if from.SourceType() != FruitSourceType {
	//	return errors.New("only fruits are supported as a transfer source type into sporePrints")
	//}
	//upd, err := xfer.
	//	PicsModsForChild().
	//	withInnoc(xfer).
	//	withParent(utils.Pointer(from.DbId())).
	//	withSpecies(parentInfo.Species).
	//	withSubspecies(parentInfo.Subspecies).
	//	withPerms(from.Permissions()).
	//	updateLastUpdatedIfNeeded().
	//	Finalized()
	//res, err := mongo.SessionFromContext(ctx).Client().Database(dbName).Collection(sp.CollectionName()).UpdateByID(ctx, sp.Id, upd)
	//if err != nil {
	//	return err
	//}
	//if res.ModifiedCount == 0 {
	//	return ErrNoParentModifiedForTransfer
	//}
	//return nil
}

func (sp SporePrint) GeneticInfoAsParent() (GeneticParentInfo, error) {
	return GeneticParentInfo{
		SpeciesOptionalField:    sp.SpeciesField.AsOptional(),
		SubspeciesOptionalField: sp.SubspeciesOptionalField,
		GenerationsFields:       GenerationsFieldFor(utils.Pointer(Generation(0))),
	}, nil
}

func (sp SporePrint) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return utils.Pointer(Generation(0)), utils.Pointer(Generation(0))
}

func (sp SporePrint) id() []byte {
	return []byte(sp.Id.dbIdStr())
}

func initializeSporePrints(ctx context.Context) error {
	// Indices
	coll := DbFrom(ctx).Collection(SporePrintCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		creationDateIndexModel, // This is print date
		//newSimpleIndex("parent", "parent", false, false, false),
		//newSimpleIndex("color", "color", true, true, false),
		//newSimpleIndex("density", "density", true, true, false),
		newSimpleIndex("species", "species", false, false, false),
		newSimpleIndex("subspecies", "subspecies", false, true, false),
		// Pics
		//saleIndexModel,
		//disposedIndexModel,
		// MostRecentImage
		//Notes (no index unless tags)
		projectsIndexModel,
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	return env.IfNotProd(ctx, func() error { // TODO: ensure ok
		// If test agar batch does not exist, then create it
		testItem := &SporePrint{
			MainCollectionIdField:             MainCollectionIdField{exSporePrint},
			MainCollectionOptionalParentField: MainCollectionOptionalParentField{&exFruitId},
			CreationDateField:                 CreationDateField{exampleTime},
			SporePrintColorField:              SporePrintColorField{utils.Pointer(SpColorBlack)},
			SporePrintDensityField:            SporePrintDensityField{utils.Pointer(SpDensityAvg)},
			SpeciesField:                      SpeciesField{testEntryStringId},
			SubspeciesOptionalField:           SubspeciesOptionalField{&testEntryStringId},
			PicsField:                         PicsField{exPics},
			SaleField:                         SaleField{&exAltId},
			DisposedField:                     DisposedField{&exampleTime},
			MostRecentImageField:              MostRecentImageField{utils.Pointer(exPics[0])},
			NotesField:                        NotesField{exampleNotes()},
			LastUpdatedField:                  LastUpdatedField{exampleTime},
		}
		return addTestMainEntries(ctx, testItem)
	})
}

type createSporePrintRequest struct {
	ParentId   MainCollectionId `json:"parent"`
	ParentType string           `json:"parentType"` // TODO: We will have to get parent anyways. No reason to add parentType...
	NotesField
	Pics            []PicWithNotesLessLocation //"newPic-1"
	WriteTagToField                            // TODO: make sure this is on the ts side!
	// USEs PARENT PERMS
}

func (upr createSporePrintRequest) reform() resolvedCreateSporePrintRequest {
	return resolvedCreateSporePrintRequest{
		ParentId:   upr.ParentId,
		ParentType: upr.ParentType,
		NotesField: upr.NotesField,
		PicsField: PicsField{slices.Map(upr.Pics, func(i PicWithNotesLessLocation) PicWithNotes {
			return i.asPicWithNotes(nil)
		})},
	}
}

type resolvedCreateSporePrintRequest struct {
	ParentId   MainCollectionId `json:"parent"`
	ParentType string           `json:"parentType"`
	NotesField
	PicsField
}

func createSporePrintHandler(w http.ResponseWriter, r *http.Request) {
	data := createSporePrintRequest{}
	id := NextMainCollectionId()
	b58Id := id.AsBase58()
	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, maxMultipartRequestSize)
	reader, err := r.MultipartReader() // TODO: do streamlined
	if err != nil {
		http.Error(w, "unable to open multipart reader: "+err.Error(), http.StatusBadRequest)
		return
	}
	p, err := reader.NextPart()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Process text (or object)
	bs, errr := io.ReadAll(p)
	if errr != nil {
		err = errr
		http.Error(w, "failed to read Data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	// PARSE INTO CORRECT DATA FORMAT
	err = json.Unmarshal(bs, &data)
	if err != nil {
		http.Error(w, "failed to unmarshal Data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	// Get any images
	newPics := map[int]string{}
	picsSaved := []string{}
	defer func() {
		if err != nil {
			errDel := pics.DeleteFiles(r.Context(), picsSaved...)
			if errDel != nil {
				handleFileDeleteErr(errDel)
			}
		}
	}()
	for {
		// Go to next part or break
		p, err = reader.NextPart()
		if err != nil {
			if err != io.EOF {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			break
		}
		fileName := p.FileName()
		if fileName == "" {
			http.Error(w, "file name is empty for what should have been an image", http.StatusBadRequest)
			return
		}
		// Process file
		parts := strings.Split(fileName, "-")
		if len(parts) != 2 {
			http.Error(w, "invalid image name: "+fileName, http.StatusBadRequest)
			return
		}
		num, errr := strconv.Atoi(parts[1])
		if errr != nil {
			err = errr
			http.Error(w, "failed to parse image number! "+errr.Error(), http.StatusBadRequest)
			return
		}
		fieldBytes, err := multipartToImageBytes(p, w)
		if err != nil {
			// Already wrote
			return
		}
		switch parts[0] {
		case "newPic":
			newFileNameWithPrefixPath, err := pics.SaveFile(r.Context(), fieldBytes, "sporePrint", string(b58Id), "img")
			if err != nil {
				http.Error(w, "failed to save new picture: "+err.Error(), http.StatusBadRequest)
				return
			}
			picsSaved = append(picsSaved, newFileNameWithPrefixPath)
			newPics[num] = newFileNameWithPrefixPath
		default:
			http.Error(w, "invalid picture name", http.StatusBadRequest)
			return
		}
	}
	// CHECK THAT ALL NEW PICS EXIST
	// PROCESS ALL NEW PICS AND CONTAMS
	out := data.reform()
	for i, _ := range data.Pics {
		loc, exists := newPics[i]
		if !exists {
			http.Error(w, fmt.Sprintf("error, location for new picture index %d not found (should never happen)", i), http.StatusInternalServerError)
			return
		}
		out.Pics[i].Location = ImageLocation(loc)
	}
	ctx := r.Context()
	mcItem, err := GetMainCollectionItemWithId(ctx, data.ParentId)
	if err != nil {
		http.Error(w, "failed to find main collection item via id", http.StatusBadRequest)
		return
	}
	var fr *Fruit
	var printOut *SporePrint
	_, er := newTxn(ctx, func(sessCtx mongo.SessionContext) (any, error) {
		var e error = nil
		switch mcItem.SourceType() {
		case FruitingChamberSourceType, BagSourceType, PlateSourceType, SlantSourceType, PlugSourceType, GrainJarSourceType:
			fr, e = FruitFromSourceInTxn(sessCtx, mcItem)
			if e != nil {
				return nil, e
			}
			break
		case FruitSourceType:
			var ok bool
			fr, ok = mcItem.(*Fruit)
			if !ok {
				return nil, errors.New("fruit is not a Fruit?")
			}
			// TODO: DIRECT! continue!
			break
		default:
			e := errors.New("invalid source type: " + mcItem.SourceType())
			http.Error(w, e.Error(), http.StatusBadRequest)
			return nil, e
		}
		printOut, e = fr.createSporePrintInTxn(sessCtx, PicsField{}, NotesField{}) // TODO: pics and notes?
		return nil, e
		// TODO: WRITE TAG TO!
	})
	if er != nil {
		http.Error(w, er.Error(), http.StatusInternalServerError)
		return
	}
	bsOut, err := json.Marshal(printOut)
	if err != nil {
		http.Error(w, "failed to marshal result: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(bsOut)
	handleWriteErr(err, w)
}

func MainCollItemForEntryType(entryType string) (MainCollectionItem, error) {
	if mainCollItem, exists := map[string]MainCollectionItem{
		"bag":             &Bag{}, // can only go to fruits
		"fruit":           &Fruit{},
		"fruitingChamber": &FruitingChamber{}, // can only go to fruits
		"jar":             &GrainJar{},        // can go anywhere (in theory) except MSS
		"lc":              &LiquidCulture{},   // can go anywhere (in theory) except MSS
		"lcSyringe":       &LcSyringe{},
		"mss":             &MSS{},   // generally only goes to plate
		"plate":           &Plate{}, // can go anywhere (in theory) except MSS
		"plugs":           &PlugsJar{},
		"slant":           &Slant{}, // generally only goes to plate
		"sporePrint":      &SporePrint{},
		"sporeSwab":       &SporeSwab{},
		"stasisTube":      &StasisTube{}, // generally only goes to plate
		"waterJar":        &WaterJar{},
	}[entryType]; exists {
		return mainCollItem, nil
	}
	return nil, errors.New("invalid entry type: " + entryType)
}

type updateSporePrintRequest struct {
	SaleField // TODO: validate?
	DisposedField
	SporePrintColorField
	SporePrintDensityField
	NotesUpdateField
	ImagesUpdateField //"newPic-1"
	PermsOnRequest    `json:"acl"`
}

func (upr updateSporePrintRequest) reform() resolvedUpdateSporePrintRequest {
	return resolvedUpdateSporePrintRequest{
		SporePrintColorField:   upr.SporePrintColorField,
		SporePrintDensityField: upr.SporePrintDensityField,
		SaleField:              upr.SaleField,
		DisposedField:          upr.DisposedField,
		NotesUpdateField:       upr.NotesUpdateField,
		Images:                 imageUpdates(upr.Images),
		PermsOnRequest:         upr.PermsOnRequest,
	}
}

type resolvedUpdateSporePrintRequest struct {
	SporePrintColorField
	SporePrintDensityField
	SaleField
	DisposedField
	NotesUpdateField
	Images         SplitEntries[picWithNotesForm, PicWithNotes]
	PermsOnRequest `json:"acl"`
}

func (req resolvedUpdateSporePrintRequest) modsFor(existing *SporePrint, aclField AclField) (bson.D, error) {
	return NewMods().
		updateSporePrintColorIfNeeded(req.Color, existing.Color).
		updateSporePrintDensityIfNeeded(req.Density, existing.Density).
		updateSaleIfNeeded(req.Sale, existing.Sale).
		updateDisposedIfNeeded(req, existing).
		updateNotesIfNeeded(req, existing).
		updatePicsIfNeeded(req.Images, existing.Pics).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateSporePrintHandler(w http.ResponseWriter, r *http.Request) {
	data := updateSporePrintRequest{}
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
	id := *mainCollId
	b58Id := mainCollId.AsBase58()
	newPics, _, _, err := fullMultipartWithNoBreaks(w, r, "sporePrint", &data, b58Id)
	if err != nil {
		// Already wrote
		return
	}
	// validate spore print color and density inputs
	if data.Color != nil {
		if !slices2.Contains(sporePrintColors, *data.Color) {
			http.Error(w, "invalid spore print color: "+string(*data.Color), http.StatusBadRequest)
			return
		}
	}
	if data.Density != nil {
		if !slices2.Contains(sporePrintDensities, *data.Density) {
			http.Error(w, "invalid spore print density: "+string(*data.Density), http.StatusBadRequest)
			return
		}
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

	ctx, db := Db(r)
	coll := db.Collection(SporePrintCollectionName)
	existing := SporePrint{}
	err = coll.FindOne(ctx, BsonFindFilter("_id", id)).Decode(&existing)
	if err != nil {
		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishMainCollItemUpdate(ctx, w, out.modsFor, &existing, out.PermsOnRequest)
}

type importSporePrintRequest struct {
	CreationDateField
	SporePrintColorField
	SporePrintDensityField
	SpeciesField
	SubspeciesOptionalField
	NotesField
	// pic as "img"
}

func importSporePrintHandler(w http.ResponseWriter, r *http.Request) {
	data := importSporePrintRequest{}
	id := NextMainCollectionId()
	b58id := id.AsBase58()
	r.Body = http.MaxBytesReader(w, r.Body, maxMultipartRequestSize)
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
			errDel := pics.DeleteFiles(r.Context(), picsSaved...)
			if err != nil {
				handleFileDeleteErr(errDel)
			}
		}
	}()
	ctx, now := request.UnixTime(r.Context())
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
		newFileNameWithPrefixPath, errr := pics.SaveFile(r.Context(), fieldBytes, "sporePrint", string(b58id), "img")
		if errr != nil {
			err = errr
			http.Error(w, "failed to save file: "+err.Error(), http.StatusBadRequest)
			return
		}
		picsSaved = append(picsSaved, newFileNameWithPrefixPath)
		importedPic = utils.Pointer(newPicWithNotes(now, []Note{}, ImageLocation(newFileNameWithPrefixPath)))
	}
	pix := []PicWithNotes{}
	if importedPic != nil {
		pix = []PicWithNotes{*importedPic}
	}
	// validate spore print color and density inputs
	if data.Color != nil {
		if !slices2.Contains(sporePrintColors, *data.Color) {
			println("invalid spore print color: " + string(*data.Color)) // TODO: del
			http.Error(w, "invalid spore print color: "+string(*data.Color), http.StatusBadRequest)
			return
		}
	}
	if data.Density != nil {
		if !slices2.Contains(sporePrintDensities, *data.Density) {
			http.Error(w, "invalid spore print density: "+string(*data.Density), http.StatusBadRequest)
			return
		}
	}

	finalPerms, err := ImportFinalPerms(ctx, data.Species, data.Subspecies)
	if err != nil {
		http.Error(w, "failed to get species and/or subspecies: "+err.Error(), http.StatusInternalServerError)
		return
	}

	toInsert := SporePrint{
		MainCollectionIdField:   MainCollectionIdField{id},
		CreationDateField:       data.CreationDateField,
		SporePrintColorField:    data.SporePrintColorField,
		SporePrintDensityField:  data.SporePrintDensityField,
		SpeciesField:            data.SpeciesField,
		SubspeciesOptionalField: data.SubspeciesOptionalField,
		PicsField:               PicsField{pix},
		MostRecentImageField:    MostRecentImageField{importedPic},
		NotesField:              data.NotesField,
		LastUpdatedField:        LastUpdatedField{now},
		AclField:                AclField{finalPerms},
	}
	finishImportMainCollectionEntry(ctx, &toInsert, w)
}
