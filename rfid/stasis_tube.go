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
	"strconv"
	"strings"
)

type StasisTube struct { // TODO: instructions somewhere?
	MainCollectionIdField             `bson:"inline"`
	PcRunOptionalField                `bson:"inline"` // probably won't exist for pre-existing tubes (imports=="unknown") // TODO: new, also used to not be optional
	CreationDateField                 `bson:"inline"`
	SpeciesOptionalField              `bson:"inline"`
	SubspeciesOptionalField           `bson:"inline"`
	InnocField                        `bson:"inline"`
	GenerationsFields                 `bson:"inline"`
	TransfersOutField                 `bson:"inline"`
	ParentTypeField                   `bson:"inline"` // TODO: must be plate, slant, or purchased
	MainCollectionOptionalParentField `bson:"inline"`
	PicsField                         `bson:"inline"`
	ContaminationsField               `bson:"inline"`
	KnownFruitableField               `bson:"inline"`
	SaleField                         `bson:"inline"`
	DisposedField                     `bson:"inline"`
	MostRecentImageField              `bson:"inline"`
	NotesField                        `bson:"inline"`
	LastUpdatedField                  `bson:"inline"`
	AclField                          `bson:"inline"`
}

func (s StasisTube) CanTransferTo(dst geneticSource) error {
	if !slices.Contains([]string{BagSourceType, GrainJarSourceType, LcSourceType, PlateSourceType, PlugSourceType, SlantSourceType, StasisTubeSourceType}, dst.SourceType()) {
		return errors.New("stasis tubes cannot transfer to " + dst.SourceType())
	}
	return nil
}

func (s StasisTube) GeneticInfoAsParent() (GeneticParentInfo, error) {
	return GeneticParentInfo{}, nil
}

func (s StasisTube) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return s.GenSinceSpore, s.GenSinceFruitOrSpore
}

func (s StasisTube) setTransferParent(ctx context.Context, xfer Transfer) (error, func() error) {
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(s.CollectionName())
	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
	if err != nil {
		return err, nil
	}
	res, err := coll.UpdateByID(ctx, s.Id, upd)
	if err != nil {
		return err, nil
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer, nil
	}
	return nil, func() error {
		return coll.FindOneAndReplace(ctx, bson.D{{"_id", s.Id}}, s).Err()
	}
}

func (s StasisTube) setTransferChild(ctx context.Context, xfer Transfer, from geneticSource) error {
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
		withLastUpdated(xfer.LastUpdated).
		Finalized()
	if err != nil {
		return ErrFailedToFinalizeMods
	}
	res, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(StasisTubeCollectionName).UpdateByID(ctx, s.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("parent not found for transfer update. Should never happen") // TODO: MAKE VAR
	}
	return nil
}

func (s StasisTube) EntryTypeField() *string {
	return utils.Pointer(StasisTubeSourceType)
}

func (s StasisTube) id() []byte {
	return []byte(s.Id.dbIdStr())
}

func initializeStasisTubes(ctx context.Context) error {
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(StasisTubeCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		newSimpleIndex("pcRun", "pcRun", false, true, false),
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
	testId := mainCollIdForint(idTestStasis)
	testItem := &StasisTube{
		MainCollectionIdField:   MainCollectionIdField{testId},
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

type createStasisTubeRequest struct {
	PcRunField
	WriteTagToField
}

func createStasisTubeHandler(w http.ResponseWriter, r *http.Request) {
	data := createStasisTubeRequest{}
	id, err := newCollectionId(r.Context(), StasisTubeCollectionName)
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
	coll := db.Collection(StasisTubeCollectionName)
	toInsert := StasisTube{
		MainCollectionIdField: MainCollectionIdField{id},
		PcRunOptionalField:    PcRunOptionalField{&data.PcRun},
		CreationDateField:     CreationDateField{now},
		LastUpdatedField:      LastUpdatedField{now},
		AclField:              allCanWriteAcl(),
	}
	// Validate
	if _, err := toInsert.PcRunOptionalField.Get(ctx); err != nil && !errors.Is(err, ErrMissingOptionalField) {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	finishCreateMainCollectionEntry(ctx, coll, &toInsert, w)
}

type updateStasisTubeRequest struct {
	KnownFruitableField
	SaleField
	DisposedField
	Notes   AllEntries[Note]
	Images  SplitEntries[picWithNotesForm, PicWithNotesLessLocation]
	Contams SplitEntries[contamForm, ContaminationLessLocation]
	WriteTagToField
	PermsOnRequest
}

func (upr updateStasisTubeRequest) reform() resolvedUpdateStasisTubeRequest {
	return resolvedUpdateStasisTubeRequest{
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

type resolvedUpdateStasisTubeRequest struct {
	KnownFruitableField
	SaleField // TODO: validate exists
	DisposedField
	Notes   AllEntries[Note]
	Images  SplitEntries[picWithNotesForm, PicWithNotes]
	Contams SplitEntries[contamForm, Contamination]
	WriteTagToField
	PermsOnRequest
}

func (mods resolvedUpdateStasisTubeRequest) modsFor(existing *StasisTube, aclField AclField) (bson.D, error) {
	return NewMods(). // TODO: exactly the same as plate, ok?
				updateKnownFruitableIfNeeded(mods.KnownFruitable, existing.KnownFruitable).
				updateSaleIfNeeded(mods.Sale, existing.Sale).
				updateDisposedIfNeeded(mods.Disposed, existing.Disposed).
				updateNotesIfNeeded(mods.Notes, existing.Notes).
				updatePicsIfNeeded(mods.Images, existing.Pics).
				updateContamsIfNeeded(mods.Contams, existing.Contaminations).
				updatePermsIfNeeded(aclField.ACL, existing.ACL).
				updateLastUpdatedIfNeeded().
				Finalized()
}

func updateStasisTubeHandler(w http.ResponseWriter, r *http.Request) {
	data := updateStasisTubeRequest{}
	b58Id := Base58Str(r.PathValue("id"))
	id, err := b58Id.toMainCollectionId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
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
		http.Error(w, "failed to read Data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	// PARSE INTO CORRECT DATA FORMAT
	err = json.Unmarshal(bs, &data)
	if err != nil {
		http.Error(w, "failed to unmarshal Data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Get any images
	newPics := map[int]string{}
	newContams := map[int]string{}
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
		p, err := reader.NextPart()
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
		//var maxSize
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
			newFileNameWithPrefixPath, err := pics.SaveFile(r.Context(), fieldBytes, "stasisTube", string(b58Id), "img")
			if err != nil {
				http.Error(w, "failed to save new picture: "+err.Error(), http.StatusBadRequest)
				return
			}
			picsSaved = append(picsSaved, newFileNameWithPrefixPath)
			newPics[num] = newFileNameWithPrefixPath
		case "newContam":
			newFileNameWithPrefixPath, err := pics.SaveFile(r.Context(), fieldBytes, "stasisTube", string(b58Id), "contam")
			if err != nil {
				http.Error(w, "failed to save new contamination: "+err.Error(), http.StatusBadRequest)
				return
			}
			picsSaved = append(picsSaved, newFileNameWithPrefixPath)
			newContams[num] = newFileNameWithPrefixPath
		default:
			http.Error(w, "invalid picture name", http.StatusBadRequest)
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
		out.Images.New[i].Location = imageLocation(loc)
	}
	for i, _ := range data.Contams.New {
		if loc, exists := newContams[i]; exists {
			finalLoc := imageLocation(loc)
			out.Contams.New[i].Location = &finalLoc
		}
	}
	ctx, db := Db(r)
	coll := db.Collection(StasisTubeCollectionName)
	// go get current stasisTube
	existing := StasisTube{}
	err = coll.FindOne(ctx, bson.D{{"_id", id}}).Decode(&existing)
	if err != nil {
		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	// Validation
	if out.Sale != nil && (existing.Sale == nil || *existing.Sale != *out.Sale) {
		if err = db.Collection(SalesCollectionName).FindOne(ctx, bson.D{{"_id", out.Sale}}).Err(); err != nil {
			dbErr(w, "failed to find new sale entry: "+err.Error(), http.StatusBadRequest) // TODO: do this everywhere needed
			return
		}
	}
	finishMainCollItemUpdate(ctx, w, coll, out.modsFor, &existing, out.PermsOnRequest)
}

type importStasisTubeRequest struct {
	CreationDateField
	SpeciesField
	SubspeciesOptionalField
	KnownFruitableField
	Generation *int
	// pic as "img"
	WriteTagToField
	PermsOnRequest
}

func importStasisTubeHandler(w http.ResponseWriter, r *http.Request) {
	data := importStasisTubeRequest{}
	id, err := newCollectionId(r.Context(), StasisTubeCollectionName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	b58id := id.asBase58()
	reader, err := multipartReaderForRequest(r, w, &data)
	if err != nil {
		// Already written
		return
	}
	defer r.Body.Close()
	//if err = Data.Perms.ValidateUserCanWrite(r.Context()); err != nil {
	//	http.Error(w, "email cannot write with overlapping perms: "+err.Error(), http.StatusUnauthorized)
	//	return // TODO: ok? check species or no?
	//}
	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Try to get pic if exists
	picsSaved := []string{} // TODO: do streamlined
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
		}
		// Process file
		fieldBytes, err := multipartToImageBytes(p, w)
		if err != nil {
			// Already wrote
			return
		}
		newFileNameWithPrefixPath, errr := pics.SaveFile(r.Context(), fieldBytes, "stasisTube", string(b58id), "img")
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
	var gen *Generation = nil
	if data.Generation != nil {
		gen = (*Generation)(data.Generation)
	}
	pix := []PicWithNotes{}
	if importedPic != nil {
		pix = []PicWithNotes{*importedPic}
	}

	user, err := GetAuthInfo(r.Context()) // TODO: THIS until finalPerms.Users[user.Email] = true is used in many places. Make it its own func
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
	coll := db.Collection(StasisTubeCollectionName)
	toInsert := StasisTube{
		MainCollectionIdField:   MainCollectionIdField{id},
		CreationDateField:       data.CreationDateField,
		SpeciesOptionalField:    data.SpeciesField.AsOptional(),
		SubspeciesOptionalField: data.SubspeciesOptionalField,
		GenerationsFields:       GenerationsFieldFor(gen),
		PicsField:               PicsField{pix},
		KnownFruitableField:     data.KnownFruitableField,
		MostRecentImageField:    MostRecentImageField{importedPic},
		LastUpdatedField:        LastUpdatedField{unixTimeForNow()},
		AclField:                AclField{finalPerms},
	}
	finishImportMainCollectionEntry(ctx, coll, &toInsert, data.PermsOnRequest, w)
}
