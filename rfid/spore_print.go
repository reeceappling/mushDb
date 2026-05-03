package rfid

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/reeceappling/goUtils/v2/utils"
	"github.com/reeceappling/goUtils/v2/utils/slices"
	"github.com/reeceappling/mushDb/rfid/pics"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// TODO: harvestFruit from bag/box page? How do I want to track which fruit without writing on them or tagging them? Create a flush collection?
// TODO: createSporePrint from fruit page
// TODO: createSerialSporePrint from sporePrint page

type SporePrint struct {
	MainCollectionIdField `bson:"inline"`
	// Parent is always either fruit, or purchased
	MainCollectionOptionalParentField `bson:"inline"` // TODO: handle now a pointer // TODO: used to be an altCollId       // TODO: likely won't exist for pre-existing
	CreationDateField                 `bson:"inline"` // Print or receive date
	SporePrintColorField              `bson:"inline"` // Set later on the print, not on creation from transfer // TODO: new! handle in TS!
	SporePrintDensityField            `bson:"inline"` // Set later on the print, not on creation from transfer // TODO: new! rename! handle in TS!
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

// TODO: endpoint to return these?
type SporePrintDensity string

const (
	SpDensityHeavy       SporePrintDensity = "Heavy"
	SpDensityAvg         SporePrintDensity = "Average"
	SpDensitySparse      SporePrintDensity = "Sparse"
	spDensityNoneMinimal SporePrintDensity = "None or Minimal"
)

// TODO: endpoint to return these?

type SporePrintColorField struct {
	Color *SporePrintColor `bson:"color,omitempty" json:"color,omitempty"`
}
type SporePrintDensityField struct {
	Density *SporePrintDensity `bson:"density,omitempty" json:"density,omitempty"`
}

func (sp SporePrint) Innoculatable() bool {
	return false
}

func (sp SporePrint) CanTransferTo(dst geneticSource) error {
	return errors.New("sporePrints cannot transfer. Only be made into mss or swab")
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
	//	withSubspecies(parentInfo.SubSpecies).
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
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(SporePrintCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		//newSimpleIndex("parent", "parent", false, false, false),
		//newSimpleIndex("printDate", "creationDate", true, false, false), // TODO: INDEX CREATION DATES EVERYWHERE!
		//newSimpleIndex("color", "color", true, true, false),
		//newSimpleIndex("density", "density", true, true, false),
		newSimpleIndex("species", "species", false, false, false),
		newSimpleIndex("subspecies", "subspecies", false, true, false),
		// Pics
		// TODO: projectsIndexModel,
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
	// If test agar batch does not exist, then create it
	testItem := &SporePrint{
		MainCollectionIdField:             MainCollectionIdField{exSporePrint},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&exFruitId},
		CreationDateField:                 exampleTime.asCreationDate(),
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
}

type createSporePrintRequest struct {
	FruitId AlternateCollectionId `bson:"fruitId" json:"fruitId"`
	NotesField
	Pics []PicWithNotesLessLocation //"newPic-1" // TODO: maybe do pics on the edit page?
	// TODO: USE PARENT PERMS?
}

func (upr createSporePrintRequest) reform() resolvedCreateSporePrintRequest {
	return resolvedCreateSporePrintRequest{
		FruitId:    upr.FruitId,
		NotesField: upr.NotesField,
		PicsField: PicsField{slices.Map(upr.Pics, func(i PicWithNotesLessLocation) PicWithNotes {
			return i.asPicWithNotes(nil)
		})},
	}
}

type resolvedCreateSporePrintRequest struct {
	FruitId AlternateCollectionId `bson:"fruitId" json:"fruitId"`
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
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	parent := Fruit{}
	err = db.Collection(FruitsCollName).FindOne(ctx, bsonFindFilter("_id", id)).Decode(&parent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	now := unixTimeForNow()
	spid := id
	var mri *PicWithNotes = nil
	if len(out.Pics) > 0 {
		lastPic := out.Pics[len(out.Pics)-1]
		mri = &lastPic
	}
	toInsert := SporePrint{
		MainCollectionIdField:             MainCollectionIdField{spid},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&parent.Id},
		CreationDateField:                 now.asCreationDate(),
		SpeciesField:                      parent.SpeciesField,
		SubspeciesOptionalField:           parent.SubspeciesOptionalField,
		PicsField:                         out.PicsField,
		MostRecentImageField:              MostRecentImageField{mri},
		NotesField:                        NotesField{out.Notes},
		LastUpdatedField:                  LastUpdatedField{now},
		// Do not check permissions, just pass parent perms to child
		AclField: parent.AclField,
	}
	_, err = newTxn(ctx, func(sessCtx mongo.SessionContext) (any, error) {
		err := addToIdMapCollection(sessCtx, &toInsert)
		if err != nil {
			return nil, err
		}
		// Update fruit with new print id
		err = parent.addSporePrint(sessCtx, spid)
		if err != nil {
			return nil, errors.Join(errors.New("failed to add spore print to parent fruit"), err)
		}
		_, err = mongo.SessionFromContext(sessCtx).Client().Database(dbName).Collection(SporePrintCollectionName).InsertOne(ctx, toInsert)
		if err != nil {
			return nil, errors.Join(errors.New("failed to insert new spore print"), err)
		}
		return nil, nil
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bsOut, err := json.Marshal(toInsert)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(bsOut)
	if err != nil {
		handleWriteErr(err, w)
	}
}

type updateSporePrintRequest struct {
	SaleField // TODO: validate?
	DisposedField
	SporePrintColorField   // TODO: validate? // TODO: add to typescript side and validate
	SporePrintDensityField // TODO: validate? // TODO: add to typescript side and validate
	NotesUpdateField
	ImagesUpdateField //"newPic-1"
	PermsOnRequest
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
	SaleField
	DisposedField
	SporePrintColorField
	SporePrintDensityField
	NotesUpdateField
	Images SplitEntries[picWithNotesForm, PicWithNotes]
	PermsOnRequest
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
	// TODO: validate spore print color and density inputs

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
	err = coll.FindOne(ctx, bsonFindFilter("_id", id)).Decode(&existing)
	if err != nil {
		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishMainCollItemUpdate(ctx, w, coll, out.modsFor, &existing, out.PermsOnRequest)
}

type importSporePrintRequest struct {
	CreationDateField
	SporePrintColorField   // TODO: add to typescript side and validate
	SporePrintDensityField // TODO: add to typescript side and validate
	SpeciesField
	SubspeciesOptionalField
	NotesField
	// pic as "img"
	PermsOnRequest
}

func importSporePrintHandler(w http.ResponseWriter, r *http.Request) {
	data := importSporePrintRequest{}
	id := NextMainCollectionId()
	b58id := id.AsBase58()
	r.Body = http.MaxBytesReader(w, r.Body, maxMultipartRequestSize)
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
	// TODO: ???? //if err = Data.Perms.ValidateUserCanWrite(r.Context()); err != nil {
	//	http.Error(w, "email cannot write with these perms: "+err.Error(), http.StatusBadRequest)
	//	return
	//}
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
	now := unixTimeForNow()
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
	// TODO: validate spore print color and density inputs

	ctx, db := Db(r)
	coll := db.Collection(SporePrintCollectionName)

	user, err := GetAuthInfo(r.Context())
	if err != nil {
		http.Error(w, "failed to get auth info: "+err.Error(), http.StatusUnauthorized)
		return
	}
	sp, subsp, err := getSpeciesAndSubspecies(r.Context(), data.Species, data.SubSpecies)
	if err != nil {
		http.Error(w, "failed to get species or subspecies: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var finalPerms *ACL = nil
	if subsp != nil {
		finalPerms = subsp.DefaultAcl.Clone()
	} else {
		finalPerms = sp.DefaultAcl.Clone()
	}
	// Add user to the acl as a writer
	finalPerms.Users[user.Email] = true

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
		LastUpdatedField:        LastUpdatedFieldForNow(),
		AclField:                AclField{finalPerms},
	}
	finishImportMainCollectionEntry(ctx, coll, &toInsert, data.PermsOnRequest, w)
}
