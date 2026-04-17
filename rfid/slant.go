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
	"slices"
)

type Slant struct {
	MainCollectionIdField `bson:"inline"`
	AgarBatchField        `bson:"inline"` // TODO: will be empty for preexisting
	// TODO: account for stickType field
	StickType                         *string `bson:"stickType,omitempty" json:"stickType,omitempty"` //If the slant includes a popsicle stick or tongue depressor // TODO: new! use!
	CreationDateField                 `bson:"inline"`
	SpeciesOptionalField              `bson:"inline"`
	SubspeciesOptionalField           `bson:"inline"`
	InnocField                        `bson:"inline"`
	GenerationsFields                 `bson:"inline"`
	TransfersOutField                 `bson:"inline"`
	ParentTypeField                   `bson:"inline"` // nil == mainCollectionType, can also be MSS or clone! // TODO: INDEX????
	MainCollectionOptionalParentField `bson:"inline"` // TODO: binary serverside, b58 clientside? // TODO: can be from any MainCollection, or a fruit (alt) cloning/lcSyringe/sporeSwab
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

type slantStick string // TODO: rename
var (
	slantStickPopsicle        = "popsicle stick"
	slantStickTongueDepressor = "tongue depressor"
	slantStickCardboard       = "cardboard"
	slantStickDowel           = "wooden dowel" // TODO: diff dowel types?
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

//func (s Slant) setTransferParent(ctx context.Context, xfer Transfer) error {
//	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(s.CollectionName())
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
		withSubspecies(parentInfo.SubSpecies).
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

func (s Slant) EntryTypeField() *string {
	return utils.Pointer(SlantSourceType)
}

func (s Slant) id() []byte {
	return []byte(s.Id.dbIdStr())
}

func initializeSlants(ctx context.Context) error {
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(SlantsCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		newSimpleIndex("agarBatch", "agarBatch", false, true, false),
		//newSimpleIndex("stickType", "stickType", false, true, false),
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
		// TODO: Contams
		//newSimpleIndex("knownFruitable", "knownFruitable", false, true, false),
		//saleIndexModel,
		//disposedIndexModel,
		// MostRecentImage
		//Notes (no index) (maybe later with tags?)
		projectsIndexModel,
		lastUpdatedIndexModel,
		// TODO: projectsIndexModel,
	})
	if err != nil {
		return err
	}
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
}

type createSlantRequest struct {
	AgarBatch AlternateCollectionId `json:"agarBatch"`
	StickType *string               `json:"stickType,omitempty"`
	WriteTagToField
}

func createSlantHandler(w http.ResponseWriter, r *http.Request) {
	data := createSlantRequest{}
	id := NextMainCollectionId()
	//id, err := newMainCollectionId(r.Context(), SlantsCollectionName)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
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
	ctx := r.Context()
	toInsert := &Slant{
		MainCollectionIdField: MainCollectionIdField{id},
		AgarBatchField:        AgarBatchField{&data.AgarBatch},
		StickType:             data.StickType,
		CreationDateField:     CreationDateField{now},
		LastUpdatedField:      LastUpdatedField{now},
		AclField:              allCanWriteAcl(),
	}
	_, err = toInsert.AgarBatchField.Get(ctx)
	if err != nil && !errors.Is(err, ErrMissingOptionalField) {
		http.Error(w, "agar batch field missing: "+err.Error(), http.StatusInternalServerError)
		return
	}
	finishCreateMainCollectionEntry(ctx, toInsert, w) // TODO: use in all main creates
}

func finishCreateMainCollectionEntry(ctx context.Context, toInsert MainCollectionItem, w http.ResponseWriter) {
	_, err := newTxn(ctx, func(sessCtx mongo.SessionContext) (any, error) {
		err := addToIdMapCollection(sessCtx, toInsert) // TODO: do this everywhere
		if err != nil {
			return nil, errors.Join(errors.New("failed to insert in map collection"), err)
		}
		_, err = mongo.SessionFromContext(sessCtx).Client().Database(dbName).Collection(toInsert.CollectionName()).InsertOne(ctx, toInsert)
		if err != nil {
			return nil, errors.Join(errors.New("failed to insert main collection item"), err)
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
func finishCreateAlternateEntry(ctx context.Context, coll *mongo.Collection, toInsert CollectionItem, w http.ResponseWriter) {
	_, err := coll.InsertOne(ctx, toInsert)
	if err != nil {
		http.Error(w, "failed to insert one: "+err.Error(), http.StatusInternalServerError)
		return
	}
	bsOut, err := json.Marshal(toInsert)
	if err != nil {
		http.Error(w, "failed to marshal one: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(bsOut)
	if err != nil {
		handleWriteErr(err, w)
	}
}
func finishImportMainCollectionEntry(ctx context.Context, coll *mongo.Collection, toInsert MainCollectionItem, reqPerms PermsOnRequest, w http.ResponseWriter) {
	// TODO: validate that species and subspecies exist???
	genetics, err := toInsert.GeneticInfoAsParent()
	if err != nil {
		http.Error(w, "failed to get genetic info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	sp, subsp, err := genetics.GetSpeciesSubspecies(ctx)
	if err != nil {
		http.Error(w, "failed to get species or subspecies: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Set ACL to default from parent species/subspecies
	// Note: users can always import, but they may not be able to write afterwards if they do not meet the permissions...
	var acl = AclField{ACL: &ACL{}}
	if subsp != nil {
		acl.ACL = subsp.DefaultAcl
	} else {
		acl.ACL = sp.DefaultAcl
	}
	toInsert.SetPerms(acl)
	finishCreateMainCollectionEntry(ctx, toInsert, w)
}

type updateSlantRequest updatePlateRequest

func (upr updateSlantRequest) reform() resolvedUpdateSlantRequest {
	return resolvedUpdateSlantRequest{
		KnownFruitableField: upr.KnownFruitableField,
		SaleField:           upr.SaleField,
		DisposedField:       upr.DisposedField,
		Notes:               upr.Notes,
		Images:              imageUpdates(upr.Images),
		Contams:             contamUpdates(upr.Contams),
		WriteTagToField:     upr.WriteTagToField,
		PermsOnRequest:      upr.PermsOnRequest,
	}
}

type resolvedUpdateSlantRequest resolvedUpdatePlateRequest

func (mods resolvedUpdateSlantRequest) modsFor(existing *Slant, aclField AclField) (bson.D, error) {
	return NewMods().
		updateKnownFruitableIfNeeded(mods.KnownFruitable, existing.KnownFruitable).
		updateSaleIfNeeded(mods.Sale, existing.Sale). // TODO: update to a different endpoint if possible
		updateDisposedIfNeeded(mods.Disposed, existing.Disposed).
		updateNotesIfNeeded(mods.Notes, existing.Notes).
		updatePicsIfNeeded(mods.Images, existing.Pics).
		updateContamsIfNeeded(mods.Contams, existing.Contaminations).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateSlantHandler(w http.ResponseWriter, r *http.Request) {
	data := updateSlantRequest{}
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

	ctx, db := Db(r)
	coll := db.Collection(SlantsCollectionName)
	// go get current plate
	existing := Slant{}
	err = coll.FindOne(ctx, bson.D{{"_id", id}}).Decode(&existing)
	if err != nil {
		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishMainCollItemUpdate(ctx, w, coll, out.modsFor, &existing, out.PermsOnRequest)
}

type importSlantRequest struct {
	CreationDateField         // TODO: was creationTime
	StickType         *string `json:"stickType,omitempty"`
	SpeciesField
	SubspeciesOptionalField
	KnownFruitableField
	Generation *int
	// pic as "img"
	WriteTagToField
	PermsOnRequest
}

func importSlantHandler(w http.ResponseWriter, r *http.Request) {
	data := importSlantRequest{}
	id := NextMainCollectionId()
	//id, err := newMainCollectionId(r.Context(), SlantsCollectionName)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
	b58id := id.asBase58()
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
	//if err = Data.Perms.ValidateUserCanWrite(r.Context()); err != nil {
	//	http.Error(w, "email cannot write to provided perms: "+err.Error(), http.StatusBadRequest)
	//	return // TODO MAKE SURE TO ONLY TAKE SPECIES OVERLAP WITH REQUEST?
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
		newFileNameWithPrefixPath, errSave := pics.SaveFile(r.Context(), fieldBytes, "slant", string(b58id), "img")
		if errSave != nil {
			err = errSave
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

	ctx, db := Db(r)
	coll := db.Collection(SlantsCollectionName)

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

	toInsert := Slant{
		MainCollectionIdField:   MainCollectionIdField{id},
		StickType:               data.StickType,
		CreationDateField:       data.CreationDateField,
		SpeciesOptionalField:    SpeciesOptionalField{&data.Species},
		SubspeciesOptionalField: data.SubspeciesOptionalField,
		GenerationsFields: GenerationsFields{
			GenSporeField:        GenSporeField{gen},
			GenSinceFruitOrSpore: gen,
		},
		PicsField:            PicsField{pix},
		KnownFruitableField:  data.KnownFruitableField,
		MostRecentImageField: MostRecentImageField{importedPic},
		LastUpdatedField:     LastUpdatedField{unixTimeForNow()},
		AclField:             AclField{finalPerms},
	}
	finishImportMainCollectionEntry(ctx, coll, &toInsert, data.PermsOnRequest, w)
}
