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

// TODO: sometimes needed for transfers
// TODO: needed for clones

type Slant struct {
	MainCollectionIdField `bson:"inline"`
	AgarBatchField        `bson:"inline"` // TODO: will be empty for preexisting
	// TODO: account for stickType field
	StickType                         *slantStick `bson:"stickType,omitempty" json:"stickType,omitempty"` //If the slant includes a popsicle stick or tongue depressor // TODO: new! use!
	CreationDateField                 `bson:"inline"`
	SpeciesOptionalField              `bson:"inline"`
	SubspeciesOptionalField           `bson:"inline"`
	InnocField                        `bson:"inline"`
	GenerationsFields                 `bson:"inline"`
	TransfersOutField                 `bson:"inline"`
	ParentTypeField                   `bson:"inline"` // nil == mainCollectionType, can also be MSS or clone! // TODO: INDEX????
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

func (s Slant) CanTransferTo(dst geneticSource) error {
	if !slices.Contains([]string{BagSourceType, GrainJarSourceType, LcSourceType, PlateSourceType, PlugSourceType, SlantSourceType, StasisTubeSourceType}, dst.SourceType()) {
		return errors.New("plates cannot transfer to " + dst.SourceType())
	}
	return nil
}

type slantStick string // TODO: rename // TODO: ALLOW TS TO VIEW STICK TYPES!
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

//func (s Slant) setTransferParent(ctx context.Context, xfer Transfer) error {
//	coll := DbFrom(ctx).Collection(s.CollectionName())
//	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
//	if err != nil {
//		return err
//	}
//	res, err := coll.UpdateByID(ctx, s.Id, upd)
//	if err != nil {
//		return err
//	}
//	if res.ModifiedCount == 0 {
//		return ErrNoParentModifiedForTransfer
//	}
//	return nil
//}

func (s Slant) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
	parentInfo, genSpore, genFruitSpore, err := childGensForParent(from)
	if err != nil {
		return err
	}
	upd, err := xfer.
		PicsModsForChild().
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
		return ErrFailedToFinalizeMods
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
		newSimpleIndex("agarBatch", "agarBatch", false, true, false),
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
		// TODO: Contams
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
	return env.IfNotProd(ctx, func() error { // TODO: ensure ok
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
	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
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
		AclField:              allCanWriteAcl(), // TODO: ok?
	}
	_, err = toInsert.AgarBatchField.Get(ctx)
	if err != nil && !errors.Is(err, ErrMissingOptionalField) {
		http.Error(w, "agar batch field missing: "+err.Error(), http.StatusInternalServerError)
		return
	}
	finishCreateMainCollectionEntry(ctx, toInsert, w)
}

// TODO: MOVE
func finishCreateMainCollectionEntry(ctx context.Context, toInsert MainCollectionItem, w http.ResponseWriter) {
	_, err := newTxn(ctx, func(sessCtx mongo.SessionContext) (any, error) {
		return nil, createMainCollectionEntryInTxn(sessCtx, toInsert)
	})
	if err != nil {
		http.Error(w, "failed to create main collection entry in txn:"+err.Error(), http.StatusInternalServerError)
		return
	}

	bsOut, err := json.Marshal(toInsert)
	if err != nil {
		http.Error(w, "failed to marshal result: "+err.Error(), http.StatusInternalServerError)
		return
	}
	bs, err := json.MarshalIndent(toInsert, "", "  ") // TODO; del
	if err != nil {                                   // TODO; del
		println(err.Error()) // TODO; del
	} // TODO; del
	println("imported: ", string(bs)) // TODO; del
	_, err = w.Write(bsOut)
	if err != nil {
		handleWriteErr(err, w)
	}
}

func createMainCollectionEntryInTxn(ctx mongo.SessionContext, toInsert MainCollectionItem) error {
	err := addToIdMapCollection(ctx, toInsert)
	if err != nil {
		return errors.Join(errors.New("failed to insert in map collection"), err)
	}
	_, err = mongo.SessionFromContext(ctx).Client().Database(dbName).Collection(toInsert.CollectionName()).InsertOne(ctx, toInsert)
	if err != nil {
		return errors.Join(errors.New("failed to insert main collection item"), err)
	}
	return nil
}

var ErrTxnWriteFail = errors.New("failed to write in transaction")

// TODO: MOVE
// TODO: used to be: finishCreateAlternateEntry(ctx context.Context, toInsert CollectionItem, w http.ResponseWriter) {
func finishCreateAlternateEntry[T CollectionItem](ctx context.Context, toInsert T, w http.ResponseWriter) {
	coll := DbFrom(ctx).Collection(toInsert.CollectionName())
	_, err := coll.InsertOne(ctx, toInsert)
	if err != nil {
		http.Error(w, "failed to insert one: "+err.Error(), http.StatusInternalServerError)
		return
	}
	bsOut, err := json.Marshal(toInsert)
	if err != nil {
		return
	}
	_, err = w.Write(bsOut)
	if err != nil {
		handleWriteErr(err, w)
	}
}

// TODO: MOVE
func finishImportMainCollectionEntry(ctx context.Context, toInsert MainCollectionItem, w http.ResponseWriter) {
	finishCreateMainCollectionEntry(ctx, toInsert, w)
}

type updateSlantRequest struct { // TODO: overhauled, validate still works
	KnownFruitableField
	SaleField
	DisposedField
	NotesUpdateField
	ImagesUpdateField
	ContamsUpdateField
	PermsOnRequest `json:"acl"`
}

func (upr updateSlantRequest) reform() resolvedUpdateSlantRequest {
	return resolvedUpdateSlantRequest{
		KnownFruitableField: upr.KnownFruitableField,
		SaleField:           upr.SaleField,
		DisposedField:       upr.DisposedField,
		NotesUpdateField:    upr.NotesUpdateField,
		Images:              imageUpdates(upr.Images),
		Contams:             contamUpdates(upr.Contams),
		PermsOnRequest:      upr.PermsOnRequest,
	}
}

type resolvedUpdateSlantRequest resolvedUpdatePlateRequest

func (req resolvedUpdateSlantRequest) modsFor(existing *Slant, aclField AclField) (bson.D, error) {
	return NewMods().
		updateKnownFruitableIfNeeded(req, existing).
		updateSaleIfNeeded(req.Sale, existing.Sale). // TODO: update to a different endpoint if possible
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
		println("failed to standardize main collection id: " + err.Error()) // TODO: del
		http.Error(w, "failed to standardize main collection id: "+err.Error(), http.StatusBadRequest)
		return
	}
	id := *mainCollId
	b58Id := mainCollId.AsBase58()
	ctx, db := Db(r)
	reader, err := multipartReaderForRequest(r.WithContext(ctx), w, &data)
	if err != nil {
		// Already written
		return
	}

	newPics, newContams, _, err := getMultipartImages(ctx, "slant", w, reader, b58Id)
	// TODO: SOME OTHER AREAS NEED TO DO THIS INSTEAD OF fullMultipartWithNoBreaks becaues rfid writer is in-between
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

	coll := db.Collection(SlantsCollectionName)
	// go get current plate
	existing := Slant{}
	err = coll.FindOne(ctx, BsonFindFilter("_id", id)).Decode(&existing)
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
	Generation *Generation // TODO: make required for when innoculated!
	// pic as "img"
	WriteTagToField
}

func importSlantHandler(w http.ResponseWriter, r *http.Request) {
	ctx, now := request.UnixTime(r.Context()) // TODO: no more r.Context below
	data := importSlantRequest{}
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
		newFileNameWithPrefixPath, errSave := pics.SaveFile(r.Context(), fieldBytes, "slant", string(b58id), "img")
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
		finalPerms, err = ImportFinalPerms(r.Context(), *data.Species, data.Subspecies)
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
	finishImportMainCollectionEntry(ctx, &toInsert, w)
}
