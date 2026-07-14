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
	"strconv"
	"strings"
)

// TODO: newFromPcRunWaterContainer? (allow either creation from water jar, or direct through pc
// TODO: fromPcRun (when put in pc while already containing water)

type StasisTube struct { // TODO: instructions somewhere?
	MainCollectionIdField             `bson:"inline"`
	PcRunField                        `bson:"inline"` // All tubes must go through the PC. Created with PC. (imports==default)
	WaterJarOptionalField             `bson:"inline"` // Only populated if the tubes are not PC'd with water inside // TODO: HANDLE THIS EVERYWHERE! NOT YET DONE IN TS
	CreationDateField                 `bson:"inline"`
	SpeciesOptionalField              `bson:"inline"`
	SubspeciesOptionalField           `bson:"inline"`
	InnocField                        `bson:"inline"`
	GenerationsFields                 `bson:"inline"`
	TransfersOutField                 `bson:"inline"`
	ParentTypeField                   `bson:"inline"` // must be plate, slant, or empty (purchased/other)
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
	return GeneticParentInfo{
		SpeciesOptionalField:    s.SpeciesOptionalField,
		SubspeciesOptionalField: s.SubspeciesOptionalField,
		KnownFruitableField:     s.KnownFruitableField,
		GenerationsFields:       s.GenerationsFields,
	}, nil
}

func (s StasisTube) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return s.GenSinceSpore, s.GenSinceFruitOrSpore
}

func (s StasisTube) Innoculatable() error {
	return errors.Join(
		s.RequireNoSpecies(),
		s.RequireNoSubspecies(),
		s.RequireNotDisposed(),
		s.RequireUnsold(),
		s.RequireUnknownFruitable(),
		s.RequireNoInnoculation())
}

//func (s StasisTube) setTransferParent(ctx context.Context, xfer Transfer) error {
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

func (s StasisTube) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
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
		return errors.Join(err, ErrFailedToFinalizeMods)
	}
	res, err := mongo.SessionFromContext(ctx).Client().Database(dbName).Collection(StasisTubeCollectionName).UpdateByID(ctx, s.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("parent not found for transfer update. Should never happen") // TODO: MAKE VAR
	}
	return nil
}

func (s StasisTube) id() []byte {
	return []byte(s.Id.dbIdStr())
}

func initializeStasisTubes(ctx context.Context) error {
	db := DbFrom(ctx)
	coll := db.Collection(StasisTubeCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		creationDateIndexModel,
		newSimpleIndex("pcRun", "pcRun", false, true, false),
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
		// TODO: waterJar?
		projectsIndexModel,
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	return env.IfNotProd(ctx, func() error {
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
			AclField:                          allCanWriteAcl(),
		}
		return addTestMainEntries(ctx, testItem)
	})
}

type createStasisTubeRequest struct {
	//WaterJarOptionalField // TODO: Probably don't do this... sts should always be pc'd with water in... // TODO: ALLOW THIS ONLY IF ADDING AFTER PC RUN
	PcRunField
	NotesField
	WriteTagToField
}

func createStasisTubeHandler(w http.ResponseWriter, r *http.Request) {
	data := createStasisTubeRequest{}
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
	toInsert := StasisTube{
		MainCollectionIdField: MainCollectionIdField{id},
		PcRunField:            PcRunField{data.PcRun},
		CreationDateField:     CreationDateField{now},
		NotesField:            data.NotesField,
		LastUpdatedField:      LastUpdatedField{now},
		AclField:              allCanWriteAcl(), // Because initial stasis tubes are empty
	}
	// Validate
	if _, err := toInsert.PcRunField.Get(ctx); err != nil && !errors.Is(err, ErrMissingOptionalField) {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = writeRfidTagIfNecessary(ctx, data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	finishCreateMainCollectionEntry(ctx, &toInsert, w)
}

type updateStasisTubeRequest struct {
	KnownFruitableField
	SaleField
	DisposedField
	NotesUpdateField
	ImagesUpdateField
	ContamsUpdateField
	PermsOnRequest `json:"acl"`
}

func (upr updateStasisTubeRequest) reform() resolvedUpdateStasisTubeRequest {
	return resolvedUpdateStasisTubeRequest{
		KnownFruitableField: upr.KnownFruitableField,
		SaleField:           upr.SaleField,
		DisposedField:       upr.DisposedField,
		NotesUpdateField:    upr.NotesUpdateField,
		Images:              imageUpdates(upr.Images),
		Contams:             contamUpdates(upr.Contams),
		PermsOnRequest:      upr.PermsOnRequest,
	}
}

type resolvedUpdateStasisTubeRequest struct {
	KnownFruitableField
	SaleField // TODO: validate exists?
	DisposedField
	NotesUpdateField
	Images  SplitEntries[picWithNotesForm, PicWithNotes]
	Contams SplitEntries[contamForm, Contamination]
	PermsOnRequest
}

func (req resolvedUpdateStasisTubeRequest) modsFor(existing *StasisTube, aclField AclField) (bson.D, error) {
	return NewMods().
		updateKnownFruitableIfNeeded(req, existing).
		updateSaleIfNeeded(req.Sale, existing.Sale).
		updateDisposedIfNeeded(req, existing).
		updateNotesIfNeeded(req, existing).
		updatePicsIfNeeded(req.Images, existing.Pics).
		updateContamsIfNeeded(req.Contams, existing.Contaminations).
		updateMostRecentImageIfNeeded(existing.MostRecentImage, loadMriPics(&req.Images, &req.Contams, nil)).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateStasisTubeHandler(w http.ResponseWriter, r *http.Request) {
	data := updateStasisTubeRequest{}
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
			newFileNameWithPrefixPath, err := pics.SaveFile(r.Context(), fieldBytes, "stasisTube", string(b58Id), "contam") // TODO: contam even needed here???
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
		out.Images.New[i].Location = ImageLocation(loc)
	}
	for i, _ := range data.Contams.New {
		if loc, exists := newContams[i]; exists {
			finalLoc := ImageLocation(loc)
			out.Contams.New[i].Location = &finalLoc
		}
	}
	ctx, db := Db(r)
	coll := db.Collection(StasisTubeCollectionName)
	// go get current stasisTube
	existing := StasisTube{}
	err = coll.FindOne(ctx, BsonFindFilter(IDfld, id)).Decode(&existing)
	if err != nil {
		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	// Validation
	if out.Sale != nil && (existing.Sale == nil || *existing.Sale != *out.Sale) {
		if err = db.Collection(SalesCollectionName).FindOne(ctx, BsonFindFilter(IDfld, out.Sale)).Err(); err != nil {
			dbErr(w, "failed to find new sale entry: "+err.Error(), http.StatusBadRequest) // TODO: do this everywhere needed? or get rid of the sale...
			return
		}
	}
	finishMainCollItemUpdate(ctx, w, out.modsFor, &existing, out.PermsOnRequest)
}

type importStasisTubeRequest struct {
	CreationDateField
	SpeciesOptionalField
	// Optional
	SubspeciesOptionalField
	KnownFruitableField
	Generation *Generation // required for when innoculated!
	// pic as "img"
	NotesField
	WriteTagToField
}

func importStasisTubeHandler(w http.ResponseWriter, r *http.Request) {
	data := importStasisTubeRequest{}
	ctx, now := request.UnixTime(r.Context())
	id := NextMainCollectionId()
	b58id := id.AsBase58()
	reader, err := multipartReaderForRequest(r.WithContext(ctx), w, &data)
	if err != nil {
		// Already written
		return
	}
	defer r.Body.Close()
	err = writeRfidTagIfNecessary(ctx, data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Try to get pic if exists
	picsSaved := []string{} // TODO: do streamlined
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
		newFileNameWithPrefixPath, errr := pics.SaveFile(r.Context(), fieldBytes, "stasisTube", string(b58id), "img")
		if errr != nil {
			err = errr
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

	toInsert := StasisTube{
		MainCollectionIdField:   MainCollectionIdField{id},
		CreationDateField:       data.CreationDateField,
		SpeciesOptionalField:    data.SpeciesOptionalField,
		SubspeciesOptionalField: data.SubspeciesOptionalField,
		PcRunField:              PcRunField{impPcRun}, // import pc run!
		GenerationsFields:       GenerationsFieldFor(gen),
		PicsField:               PicsField{pix},
		KnownFruitableField:     data.KnownFruitableField,
		MostRecentImageField:    MostRecentImageField{importedPic},
		NotesField:              data.NotesField,
		LastUpdatedField:        LastUpdatedField{now},
		AclField:                AclField{finalPerms},
	}
	err = writeRfidTagIfNecessary(ctx, data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	finishImportMainCollectionEntry(ctx, &toInsert, w)
}

func deleteStasisTubeHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Empty id for delete request", http.StatusBadRequest)
		return
	}
	id, err := Base58Str(idStr).ToMainCollectionId()
	if err != nil {
		http.Error(w, "Invalid ID to delete: "+err.Error(), http.StatusBadRequest)
		return
	}
	// Validate not used in other places...
	ctx := r.Context()
	db := DbFrom(ctx)
	// ensure item does not have any transfers in or out
	item, err := GetMainCollectionItemSpecific[*StasisTube](ctx, id, &StasisTube{})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			http.Error(w, "Item to be deleted not found: "+err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve item to be deleted: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if item.Innoc != nil {
		http.Error(w, "Cannot delete innoculated items!", http.StatusConflict)
		return
	}
	if item.TransfersOut != nil && len(item.TransfersOut) > 0 {
		http.Error(w, "Cannot delete items with transfers out", http.StatusConflict)
		return
	}

	// Delete if not found elsewhere!
	deleteResult, err := db.Collection(StasisTubeCollectionName).DeleteOne(ctx, bson.M{IDfld: id})
	if err != nil {
		http.Error(w, "failed to delete: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if deleteResult.DeletedCount == 0 {
		http.Error(w, "failed to delete id "+idStr+" from stasis tubes. Id not found", http.StatusNotFound)
		return
	}
	_, err = w.Write([]byte(idStr))
	handleWriteErr(err, w)
}
