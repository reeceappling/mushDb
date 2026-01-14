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
	"slices"
)

const (
	SlantsCollectionName = "slants"
	SlantSourceType      = "slant"
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
	ParentTypeField                   `bson:"inline"` // TODO: NEW! HANDLE! nil == mainCollectionType, can also be MSS or clone! // TODO: INDEX????
	MainCollectionOptionalParentField `bson:"inline"` // TODO: binary serverside, b58 clientside? // TODO: can be from any MainCollection, or a fruit (alt) cloning/lcSyringe/sporeSwab
	PicsField                         `bson:"inline"`
	ContaminationsField               `bson:"inline"`
	KnownFruitableField               `bson:"inline"` // TODO: handle being yes if clone, among other yeses
	SaleField                         `bson:"inline"`
	DisposedField                     `bson:"inline"`
	MostRecentImageField              `bson:"inline"`
	NotesField                        `bson:"inline"`
	LastUpdatedField                  `bson:"inline"`
	AclField                          `bson:"inline"` // TODO: handle EVERYWHERE
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

func (s Slant) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := s
	err := decodeItem(&out, encoded)
	return out, err
}

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

func (s Slant) SourceType() string {
	return SlantSourceType
}

func (s Slant) setTransferParent(ctx mongo.SessionContext, xfer Transfer) error {
	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
	if err != nil {
		return err
	}
	res, err := ctx.Client().Database(dbName).Collection(SlantsCollectionName).UpdateByID(ctx, s.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer
	}
	return nil
}

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
		//updatePermsIfNeeded(xfer.Perms, s.Perms). // TODO: make sure perms are on all setTransferChild
		withLastUpdated(xfer.LastUpdated).
		Finalized()
	if err != nil {
		return ErrFailedToFinalizeMods
	}
	res, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(SlantsCollectionName).UpdateByID(ctx, s.Id, upd)
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

func (s Slant) CollectionName() string {
	return SlantsCollectionName
}

func (s Slant) id() []byte {
	return []byte(s.Id.dbIdStr())
}

func initializeSlants(ctx context.Context) error {
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(SlantsCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		newSimpleIndex("agarBatch", "agarBatch", false, true, false),
		newSimpleIndex("stickType", "stickType", false, true, false),
		creationDateIndexModel,
		newSimpleIndex("species", "species", false, true, false),
		newSimpleIndex("subSpecies", "subSpecies", false, true, false),
		newSimpleIndex("innoc", "innoc", false, true, false),
		newSimpleIndex("genSinceSpore", "genSpore", true, true, false),
		newSimpleIndex("genSinceFruitOrSpore", "genFruitOrSpore", true, true, false),
		transfersOutIndexModel,
		newSimpleIndex("parent", "parent", false, true, false),         // TODO: nil is store or outside?
		newSimpleIndex("parentType", "parentType", false, true, false), // TODO: nil is store or outside?

		//Pics (no index)
		// TODO: Contams
		newSimpleIndex("knownFruitable", "knownFruitable", false, true, false),
		saleIndexModel,
		disposedIndexModel,
		// MostRecentImage
		//Notes (no index) (maybe later with tags?)
		lastUpdatedIndexModel,
		// TODO: projectsIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	existingEntry := Slant{}
	testId := mainCollIdForint(idTestSlant)
	testItem := Slant{
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
	err = coll.FindOne(ctx, bson.D{{"_id", testId}}).Decode(&existingEntry)
	if err == nil {
		if reflect.DeepEqual(existingEntry, testItem) {
			return nil
		}
	}
	return testExistingEntry(ctx, coll, testId, testItem, existingEntry)
}

type createSlantRequest struct {
	AgarBatch AlternateCollectionId `json:"agarBatch"`
	StickType *string               `json:"stickType,omitempty"`
	WriteTagToField
}

func createSlantHandler(w http.ResponseWriter, r *http.Request) {
	data := createSlantRequest{}
	id, err := newCollectionId(r.Context(), SlantsCollectionName)
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
		toInsert := Slant{
			MainCollectionIdField: MainCollectionIdField{id},
			AgarBatchField:        AgarBatchField{&data.AgarBatch},
			StickType:             data.StickType,
			CreationDateField:     CreationDateField{now},
			LastUpdatedField:      LastUpdatedField{now},
			AclField:              allCanWriteAcl(),
		}
		_, err = toInsert.AgarBatchField.Get(ctx)
		if err != nil && !errors.Is(err, ErrMissingOptionalField) {
			return DbTxnStdErr(w, "agar batch field missing: "+err.Error(), http.StatusInternalServerError)
		}
		coll := ctx.Client().Database(dbName).Collection(SlantsCollectionName)
		_, err = coll.InsertOne(ctx, toInsert)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		bsOut, err := json.Marshal(toInsert)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		_, err = w.Write(bsOut)
		return nil, err
	})

	if err != nil {
		handleWriteErr(err, w)
	}
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

	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(SlantsCollectionName)
		// go get current plate
		existing := Slant{}

		err = coll.FindOne(ctx, bson.D{{"_id", id}}).Decode(&existing)
		if err != nil {
			return DbTxnStdErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		}
		user, err := GetAuthInfo(ctx)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		if !user.HasPermissionToEdit(existing) {
			return DbTxnStdErr(w, "unauthorized to edit", http.StatusForbidden)
		}
		aclField, err := out.AclFor(ctx, user)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		//if err = minimalPermsBetween(existing.Perms, data.Perms).ValidateUserCanWrite(ctx); err != nil {
		//	return DbTxnStdErr(w, "bad overlapping perms for email:"+err.Error(), http.StatusBadRequest)
		//}
		upd, err := NewMods().
			updateKnownFruitableIfNeeded(out.KnownFruitable, existing.KnownFruitable).
			updateSaleIfNeeded(out.Sale, existing.Sale). // TODO: update to a different endpoint if possible
			updateDisposedIfNeeded(out.Disposed, existing.Disposed).
			updateNotesIfNeeded(out.Notes, existing.Notes).
			updatePicsIfNeeded(out.Images, existing.Pics).
			updateContamsIfNeeded(out.Contams, existing.Contaminations).
			updatePermsIfNeeded(aclField.ACL, existing.ACL).
			updateLastUpdatedIfNeeded().
			Finalized()
		if err != nil {
			return DbTxnStdErr(w, "error creating txn:"+err.Error(), http.StatusInternalServerError)
		}
		if len(upd) == 0 {
			return DbTxnStdErr(w, "no changes made", http.StatusBadRequest)
		}
		// write updates to db
		bsonId := bson.D{{"_id", id}}
		err = coll.FindOneAndUpdate(ctx, bsonId, upd).Err()
		if err != nil {
			return DbTxnStdErr(w, "failed to write update to db: "+err.Error(), http.StatusInternalServerError)
		}
		err = coll.FindOne(ctx, bsonId).Decode(&existing)
		if err != nil {
			return DbTxnStdErr(w, "failed to write update to db: "+err.Error(), http.StatusInternalServerError)
		}
		bsOut, err := json.Marshal(&out)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		return w.Write(bsOut)
	})
	if err != nil {
		HandleHttpWriteError(err)
	}
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
	PermsOnRequest // TODO: handle in typescript and handler!
}

func importSlantHandler(w http.ResponseWriter, r *http.Request) {
	data := importSlantRequest{}
	id, err := newCollectionId(r.Context(), SlantsCollectionName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
		http.Error(w, "unable to read data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	// PARSE INTO CORRECT DATA FORMAT
	err = json.Unmarshal(bs, &data)
	if err != nil {
		http.Error(w, "unable to unmarshal json form data: "+err.Error(), http.StatusBadRequest)
		return
	}
	//if err = data.Perms.ValidateUserCanWrite(r.Context()); err != nil {
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

	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		//finalPerms := data.Perms
		//if data.Perms != nil {
		//	spec, subsp, err := getSpeciesAndSubspecies(ctx, data.Species, data.SubSpecies)
		//	if err != nil {
		//		return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError) // TODO: ok?
		//	}
		//	finalPerms = minimalPermsBetween(data.Perms, spec, subsp) // TODO: maximal with perms if nonWrite
		//}
		perms, err := GetAuthInfo(ctx)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		acl, err := data.AclFor(ctx, perms)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
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
			AclField:             acl,
		}
		coll := ctx.Client().Database(dbName).Collection(SlantsCollectionName)
		_, err = coll.InsertOne(ctx, toInsert)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		bsOut, err := json.Marshal(toInsert)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		return w.Write(bsOut)
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}
