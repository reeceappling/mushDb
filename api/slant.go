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

// sometimes needed for: transfers, clones

type Slant struct {
	MainCollectionIdField `bson:"inline"`
	AgarBatchField        `bson:"inline"` // will be empty for preexisting
	// TODO: account for stickType field everywhere
	StickType                         *slantStick `bson:"stickType,omitempty" json:"stickType,omitempty"` //If the slant includes a popsicle stick or tongue depressor // TODO: new! use!
	CreationDateField                 `bson:"inline"`
	SpeciesOptionalField              `bson:"inline"`
	SubspeciesOptionalField           `bson:"inline"`
	InnocField                        `bson:"inline"`
	GenerationsFields                 `bson:"inline"`
	TransfersOutField                 `bson:"inline"`
	ParentTypeField                   `bson:"inline"` // nil == mainCollectionType, can also be MSS or clone!
	MainCollectionOptionalParentField `bson:"inline"`
	PicsField                         `bson:"inline"`
	ContaminationsField               `bson:"inline"`
	KnownFruitableField               `bson:"inline"` // TODO: handle being yes if clone, among other yeses
	SaleField                         `bson:"inline"`
	DisposedField                     `bson:"inline"`
	MostRecentImageField              `bson:"inline"`
	NotesField                        `bson:"inline"`
	LastUpdatedField                  `bson:"inline"`
	AclField                          `bson:"inline"`
}

//func (s Slant) Blank() CollectionItem {
//	return &Slant{}
//}

func (s Slant) CanTransferTo(dst geneticSource) error {
	if !slices.Contains([]string{BagSourceType, GrainJarSourceType, LcSourceType, PlateSourceType, PlugSourceType, SlantSourceType, StasisTubeSourceType}, dst.SourceType()) {
		return errors.New("plates cannot transfer to " + dst.SourceType())
	}
	return nil
}

type slantStick string // TODO: ALLOW TS TO VIEW STICK TYPES!!!!
var (
	slantStickPopsicle        slantStick = "popsicle stick"
	slantStickTongueDepressor slantStick = "tongue depressor"
	slantStickCardboard       slantStick = "cardboard"
	slantStickDowel           slantStick = "wooden dowel" // TODO: diff dowel types?
)

func (s Slant) GeneticInfoAsParent() (GeneticParentInfo, error) {
	return GeneticParentInfo{
		SpeciesOptionalField:    SpeciesOptionalField{s.Species},
		SubspeciesOptionalField: s.SubspeciesOptionalField,
		KnownFruitableField:     s.KnownFruitableField,
		GenerationsFields:       s.GenerationsFields,
	}, nil
}

func (s Slant) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return s.GenSinceSpore, s.GenSinceFruitOrSpore
}

func (s Slant) Innoculatable() error {
	return errors.Join(
		s.RequireNoSpecies(),
		s.RequireNoSubspecies(),
		s.RequireNotDisposed(),
		s.RequireUnsold(),
		s.RequireUnknownFruitable(),
		s.RequireNoInnoculation())
}

func (s Slant) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
	parentInfo, genSpore, genFruitSpore, err := childGensForParent(from)
	if err != nil {
		return err
	}
	upd, err := xfer.PicsModsForChild(s).
		withInnoc(xfer).
		withParentType(&xfer.FromType).
		withParent(utils.Pointer(from.DbId())).
		withGens(genSpore, genFruitSpore).
		withSpecies(parentInfo.Species).
		withSubspecies(parentInfo.Subspecies).
		withKnownFruitable(parentInfo.KnownFruitable).
		withPerms(from.Permissions()).
		withLastUpdated(xfer.LastUpdated).
		Finalized()
	if err != nil {
		return errors.Join(err, ErrFailedToFinalizeMods)
	}
	res, err := mongo.SessionFromContext(ctx).Client().Database(dbName).Collection(SlantsCollectionName).UpdateByID(ctx, s.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer
	}
	return nil
}

func (s Slant) id() []byte {
	return []byte(s.Id.dbIdStr())
}

func initializeSlants(ctx context.Context) error {
	db := DbFrom(ctx)
	coll := db.Collection(SlantsCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		creationDateIndexModel,
		newSimpleIndex("agarBatch", "agarBatch", false, true, false), // Required index for batch deletes
		//newSimpleIndex("stickType", "stickType", false, true, false),
		newSimpleIndex("species", "species", false, true, false),
		newSimpleIndex("subspecies", "subspecies", false, true, false),
		//newSimpleIndex("innoc", "innoc", false, true, false),
		//newSimpleIndex("genSinceSpore", "genSpore", true, true, false),
		//newSimpleIndex("genSinceFruitOrSpore", "genFruitOrSpore", true, true, false),
		//transfersOutIndexModel,
		//newSimpleIndex("parent", "parent", false, true, false),         // TODO: nil is store or outside?
		//newSimpleIndex("parentType", "parentType", false, true, false), // TODO: nil is store or outside?

		//Pics (no index)
		// Contams?
		//newSimpleIndex("knownFruitable", "knownFruitable", false, true, false),
		//saleIndexModel,
		//disposedIndexModel,
		// MostRecentImage
		//Notes (no index) (maybe later with tags?)
		projectsIndexModel,
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	return env.IfNotProd(ctx, func() error {
		// If test agar batch does not exist, then create it
		testId := mainCollIdForint(idTestSlant)
		testItem := &Slant{
			MainCollectionIdField:   MainCollectionIdField{testId},
			AgarBatchField:          AgarBatchField{&exAltId},
			CreationDateField:       CreationDateField{exampleTime},
			SpeciesOptionalField:    SpeciesOptionalField{&testEntryStringId},
			SubspeciesOptionalField: SubspeciesOptionalField{&testEntryStringId},
			InnocField:              InnocField{&exAltId},
			GenerationsFields: GenerationsFields{
				GenSporeField:        GenSporeField{&exGenSinceSpore},
				GenSinceFruitOrSpore: &exGenSinceFruitSpore,
			},
			TransfersOutField:                 TransfersOutField{exAlts},
			ParentTypeField:                   ParentTypeField{&exParentType},
			MainCollectionOptionalParentField: MainCollectionOptionalParentField{&exPlate},
			PicsField:                         PicsField{exPics},
			ContaminationsField:               ContaminationsField{exContams},
			KnownFruitableField:               KnownFruitableField{exBool},
			SaleField:                         SaleField{&exAltId},
			DisposedField:                     DisposedField{&exampleTime},
			MostRecentImageField:              MostRecentImageField{&exPics[0]},
			NotesField:                        NotesField{exampleNotes()},
			LastUpdatedField:                  LastUpdatedField{exampleTime},
		}
		return addTestMainEntries(ctx, testItem)
	})
}

type createSlantRequest struct {
	AgarBatch AlternateCollectionId `json:"agarBatch"`
	StickType *slantStick           `json:"stickType,omitempty"`
	NotesField
	WriteTagToField
}

func createSlantHandler(w http.ResponseWriter, r *http.Request) {
	data := createSlantRequest{}
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

	ctx, now := request.UnixTime(r.Context())
	toInsert := &Slant{
		MainCollectionIdField: MainCollectionIdField{id},
		AgarBatchField:        AgarBatchField{&data.AgarBatch},
		StickType:             data.StickType,
		CreationDateField:     CreationDateField{now},
		NotesField:            data.NotesField,
		LastUpdatedField:      LastUpdatedField{now},
		AclField:              allCanWriteAcl(),
	}
	_, err = toInsert.AgarBatchField.Get(ctx)
	if err != nil && !errors.Is(err, ErrMissingOptionalField) {
		http.Error(w, "agar batch field missing: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = writeRfidTagIfNecessary(ctx, data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	finishCreateMainCollectionEntry(ctx, toInsert, w)
}

var ErrTxnWriteFail = errors.New("failed to write in transaction")

type updateSlantRequest struct { // TODO: overhauled, validate still works
	KnownFruitableField
	DisposedField
	NotesUpdateField
	ImagesUpdateField
	ContamsUpdateField
	PermsOnRequest `json:"acl"`
}

func (upr updateSlantRequest) reform() resolvedUpdateSlantRequest {
	return resolvedUpdateSlantRequest{
		KnownFruitableField: upr.KnownFruitableField,
		//SaleField:           upr.SaleField,
		DisposedField:    upr.DisposedField,
		NotesUpdateField: upr.NotesUpdateField,
		Images:           imageUpdates(upr.Images),
		Contams:          contamUpdates(upr.Contams),
		PermsOnRequest:   upr.PermsOnRequest,
	}
}

type resolvedUpdateSlantRequest resolvedUpdatePlateRequest

func (req resolvedUpdateSlantRequest) modsFor(existing *Slant, aclField AclField) (bson.D, error) {
	return NewMods().
		updateKnownFruitableIfNeeded(req, existing).
		updateDisposedIfNeeded(req, existing).
		updateNotesIfNeeded(req, existing).
		updatePicsIfNeeded(req.Images, existing.Pics).
		updateContamsIfNeeded(req.Contams, existing.Contaminations).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateSlantHandler(w http.ResponseWriter, r *http.Request) {
	data := updateSlantRequest{}
	idStr, err := UrlDecodeString(r.PathValue("id"))
	if err != nil {
		http.Error(w, "failed to url decode string: "+err.Error(), http.StatusInternalServerError)
		return
	}
	mainCollId, err := StandardizeMainCollectionId(idStr)
	if err != nil {
		http.Error(w, "failed to standardize main collection id: "+err.Error(), http.StatusBadRequest)
		return
	}
	id := *mainCollId
	b58Id := mainCollId.AsBase58()
	ctx, db := Db(r)
	reader, err := multipartReaderForRequest(r.WithContext(ctx), w, &data) // TODO: consider swapping for multipartReaderInitialize
	if err != nil {
		// Already written
		return
	}

	newPics, newContams, _, err := getMultipartImages(ctx, "slant", w, reader, b58Id)
	// TODO: SOME OTHER AREAS NEED TO DO THIS INSTEAD OF fullMultipartWithNoBreaks becaues rfid writer is in-between
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
		out.Images.New[i].Location = ImageLocation(loc)
	}
	for i, _ := range data.Contams.New {
		if loc, exists := newContams[i]; exists {
			finalLoc := ImageLocation(loc)
			out.Contams.New[i].Location = &finalLoc
		}
	}

	coll := db.Collection(SlantsCollectionName)
	// go get current plate
	existing := Slant{}
	err = coll.FindOne(ctx, BsonFindFilter(IDfld, id)).Decode(&existing)
	if err != nil {
		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishMainCollItemUpdate(ctx, w, out.modsFor, &existing, out.PermsOnRequest)
}

type importSlantRequest struct {
	CreationDateField
	StickType *slantStick `json:"stickType,omitempty"`
	SpeciesOptionalField
	SubspeciesOptionalField
	KnownFruitableField
	Generation *Generation // required when innoculated!
	// pic as "img"
	WriteTagToField
}

func importSlantHandler(w http.ResponseWriter, r *http.Request) {
	ctx, now := request.UnixTime(r.Context())
	data := importSlantRequest{}
	id := NextMainCollectionId()
	b58id := id.AsBase58()
	reader, err := multipartReaderInitialize(r.Context(), w, r, &data)
	defer r.Body.Close()
	if err != nil {
		return // Already wrote
	}
	// Try to get pic if exists
	picsSaved := []string{}
	defer func() {
		if err != nil {
			err = pics.DeleteFiles(ctx, picsSaved...)
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
		newFileNameWithPrefixPath, errSave := pics.SaveFile(ctx, fieldBytes, "slant", string(b58id), "img")
		if errSave != nil {
			err = errSave
			http.Error(w, "failed to save file: "+err.Error(), http.StatusBadRequest)
			return
		}
		picsSaved = append(picsSaved, newFileNameWithPrefixPath)
		importedPic = utils.Pointer(newPicWithNotes(now, []Note{}, ImageLocation(newFileNameWithPrefixPath)))
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
		finalPerms, err = ImportFinalPerms(ctx, *data.Species, data.Subspecies)
		if err != nil {
			http.Error(w, "failed to get species and/or subspecies: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	toInsert := Slant{
		MainCollectionIdField:   MainCollectionIdField{id},
		StickType:               data.StickType,
		CreationDateField:       data.CreationDateField,
		SpeciesOptionalField:    SpeciesOptionalField{data.Species},
		SubspeciesOptionalField: data.SubspeciesOptionalField,
		GenerationsFields: GenerationsFields{
			GenSporeField:        GenSporeField{gen},
			GenSinceFruitOrSpore: gen,
		},
		PicsField:            PicsField{pix},
		KnownFruitableField:  data.KnownFruitableField,
		MostRecentImageField: MostRecentImageField{importedPic},
		LastUpdatedField:     LastUpdatedField{now},
		AclField:             AclField{finalPerms},
	}
	err = writeRfidTagIfNecessary(ctx, data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	finishImportMainCollectionEntry(ctx, &toInsert, w)
}

//func deleteSlantHandler(w http.ResponseWriter, r *http.Request) {
//	idStr := r.PathValue("id")
//	if idStr == "" {
//		http.Error(w, "Empty id for delete request", http.StatusBadRequest)
//		return
//	}
//	id, err := Base58Str(idStr).ToMainCollectionId()
//	if err != nil {
//		http.Error(w, "Invalid ID to delete: "+err.Error(), http.StatusBadRequest)
//		return
//	}
//	// Validate not used in other places...
//	ctx := r.Context()
//	// ensure item does not have any transfers in or out
//	item, err := GetMainCollectionItemSpecific[*Slant](ctx, id, &Slant{})
//	if err != nil {
//		if errors.Is(err, mongo.ErrNoDocuments) {
//			http.Error(w, "Item to be deleted not found! Should never happen!: "+err.Error(), http.StatusNotFound)
//		} else {
//			http.Error(w, "Failed to retrieve item to be deleted: "+err.Error(), http.StatusInternalServerError)
//		}
//		return
//	}
//	if item.Parent != nil {
//		// TODO: what if we want to remove it from the parent as well?
//		http.Error(w, "Cannot delete innoculated items!", http.StatusExpectationFailed)
//		return
//	}
//	if item.TransfersOut != nil && len(item.TransfersOut) > 0 {
//		http.Error(w, "Cannot delete items with transfers out", http.StatusExpectationFailed)
//		return
//	}
//
//	// Delete if not found elsewhere!
//	DeleteCollectionItem(ctx, item.CollectionName(), id, w)
//}
