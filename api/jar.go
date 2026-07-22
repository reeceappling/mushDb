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
	slices2 "slices"
	"strings"
)

// TODO: needed for transfers

// TODO: new (PC is created first, so it can be referenced)

type GrainJar struct {
	MainCollectionIdField   `bson:"inline"`
	SizeCups                int `bson:"sizeCups" json:"sizeCups"` // 1==1cup, 2 == pint, 4==quart, 16==gal
	JarRecipeField          `bson:"inline"`
	GrainBatchOptionalField `bson:"inline"`
	// TODO: multiple grain batches????
	WetnessField                      `bson:"inline"` // 5 is ideal, 0 is ultra-dry, 10 is soaked
	BurstGrainsField                  `bson:"inline"` // 0 is ideal, 1-2 is common, everything above that is oof
	PcRunOptionalField                `bson:"inline"` // Imports default, can be created without a run!
	CreationDateField                 `bson:"inline"`
	SpeciesOptionalField              `bson:"inline"`
	SubspeciesOptionalField           `bson:"inline"`
	InnocField                        `bson:"inline"` // TODO: multiple? What if first innoc does not work?
	GenerationsFields                 `bson:"inline"`
	TransfersOutField                 `bson:"inline"`
	ParentTypeField                   `bson:"inline"` // nil == mainCollectionType, can also be MSS or clone! // TODO: INDEX???? // TODO: multiple?
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

type BurstGrainsField struct { // 0 is none, 10 is most or all
	BurstGrains *int `bson:"burstGrains,omitempty" json:"burstGrains,omitempty"`
}

func (j GrainJar) CanTransferTo(dst geneticSource) error {
	if slices2.Contains([]string{FruitingChamberSourceType, FruitSourceType, LcSyringeSourceType, MssSourceType, SporePrintSourceType, SporeSwabSourceType, WaterJarsSourceType}, dst.SourceType()) {
		return errors.New("jar cannot transfer to " + dst.SourceType())
	}
	return nil
}

func (j GrainJar) GeneticInfoAsParent() (GeneticParentInfo, error) {
	return GeneticParentInfo{
		SpeciesOptionalField:    SpeciesOptionalField{j.Species},
		SubspeciesOptionalField: j.SubspeciesOptionalField,
		KnownFruitableField:     j.KnownFruitableField,
		GenerationsFields:       j.GenerationsFields,
	}, nil
}

func (j GrainJar) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return j.GenSinceSpore, j.GenSinceFruitOrSpore
}

//func (j GrainJar) setTransferParent(ctx context.Context, xfer Transfer) error {
//	coll := DbFrom(ctx).Collection(GrainJarCollectionName)
//	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
//	if err != nil {
//		return err
//	}
//	res, err := coll.UpdateByID(ctx, j.Id, upd)
//	if err != nil {
//		return err
//	}
//	if res.ModifiedCount == 0 {
//		return ErrNoParentModifiedForTransfer
//	}
//	return nil
//}

func (j GrainJar) Innoculatable() error {
	return errors.Join(
		j.RequireNoSpecies(),
		j.RequireNoSubspecies(),
		j.RequireNotDisposed(),
		j.RequireUnsold(),
		j.RequireUnknownFruitable(),
		j.RequireNoInnoculation(),
		j.HasPcRun())
}

func (j GrainJar) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
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
		withKnownFruitable(parentInfo.KnownFruitable).
		withSubspecies(parentInfo.Subspecies).
		withPerms(from.Permissions()).
		withLastUpdated(xfer.LastUpdated).
		Finalized()
	if err != nil {
		return errors.Join(err, ErrFailedToFinalizeMods)
	}
	res, err := mongo.SessionFromContext(ctx).Client().Database(dbName).Collection(GrainJarCollectionName).UpdateByID(ctx, j.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer
	}
	return nil
}

func (j GrainJar) Collection(ctx mongo.SessionContext) *mongo.Collection {
	return ctx.Client().Database(dbName).Collection(GrainJarCollectionName)
}

func LookupGrainJar(ctx context.Context, id MainCollectionId) (j *GrainJar, err error) {
	j = &GrainJar{}
	err = DbFrom(ctx).Collection(GrainJarCollectionName).FindOne(ctx, BsonFindFilter(IDfld, id)).Decode(j)
	return j, err
}

func initializeJars(ctx context.Context) error {
	db := DbFrom(ctx)
	coll := db.Collection(GrainJarCollectionName)
	err := createIndexes(ctx, coll,
		[]mongo.IndexModel{
			creationDateIndexModel,
			//newSimpleIndex("sizeCups", "sizeCups", true, false, false),
			//newSimpleIndex("recipe", "recipe", false, true, false),
			//newSimpleIndex("grainBatch", "grainBatch", false, true, false), // TODO: ????
			//newSimpleIndex("wetness", "wetness", false, true, false),
			//newSimpleIndex("burstGrains", "burstGrains", false, true, false),
			newSimpleIndex("pcRun", "pcRun", false, true, false), // TODO: ????
			newSimpleIndex("species", "species", false, true, false),
			newSimpleIndex("subspecies", "subspecies", false, true, false),
			//newSimpleIndex("innoc", "innoc", false, true, false),
			//newSimpleIndex("genSinceSpore", "genSpore", true, true, false),
			//newSimpleIndex("genSinceFruitOrSpore", "genFruitOrSpore", true, true, false),
			//transfersOutIndexModel,
			//newSimpleIndex("parent", "parent", false, true, false),         // TODO: nil is store or outside?
			//newSimpleIndex("parentType", "parentType", false, true, false), // TODO: nil is store or outside?
			//Pics (no index)
			//TODO: Contams
			//newSimpleIndex("knownFruitable", "knownFruitable", false, true, false),
			//saleIndexModel,
			//newSimpleIndex("disposed", "disposed", false, true, false),
			// MostRecentImage
			//Notes (no index) (maybe later with tags?)
			projectsIndexModel,
			lastUpdatedIndexModel,
		})
	if err != nil {
		// TODO: I dont like the second one here!
		if !strings.Contains(err.Error(), "Identical index already exists:") && !strings.Contains(err.Error(), "Cannot build two identical indexes") {
			println("failed to create indices", err.Error())
			return err
		}
	}
	return env.IfNotProd(ctx, func() error {
		// If test agar batch does not exist, then create it
		testId := mainCollIdForint(idTestJar)
		testItem := &GrainJar{
			MainCollectionIdField:   MainCollectionIdField{testId},
			JarRecipeField:          JarRecipeField{&exAltId},
			PcRunOptionalField:      PcRunOptionalField{&exAltId},
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
			MainCollectionOptionalParentField: MainCollectionOptionalParentField{&exPlate},
			PicsField:                         PicsField{exPics},
			ContaminationsField:               ContaminationsField{exContams},
			KnownFruitableField:               KnownFruitableField{exBool},
			SaleField:                         SaleField{&exAltId},
			DisposedField:                     DisposedField{&exampleTime},
			MostRecentImageField:              MostRecentImageField{&exPics[0]},
			NotesField:                        NotesField{exampleNotes()},
			LastUpdatedField:                  LastUpdatedField{exampleTime},
			SizeCups:                          4,
		}
		return addTestMainEntries(ctx, testItem)
	})
}

//// TODO: RENAME AND MOVE!
//func testExistingEntry[T any](ctx context.Context, coll *mongo.Collection, testId any, testItem T, existingEntry T) error {
//	res, err := coll.InsertOne(ctx, testItem)
//	if err != nil {
//		return err
//	}
//	if res == nil {
//		return errors.New("result should not be nil")
//	}
//	err = coll.FindOne(ctx, BsonFindFilter(IDfld, testId)).Decode(&existingEntry)
//	if err != nil {
//		return errors.New("not found at specified id. " + err.Error())
//	}
//	if !reflect.DeepEqual(existingEntry, testItem) {
//		ee, err := json.Marshal(existingEntry)
//		if err != nil {
//			println("bad existing json")
//		}
//		te, err := json.Marshal(testItem)
//		if err != nil {
//			println("bad test json")
//		}
//		println("-------------------")
//		println(string(te))
//		println(string(ee))
//		return errors.New("entries (as updated) were not equal")
//	}
//	return nil
//}

type createJarRequest struct {
	SizeCups int `json:"sizeCups"`
	GrainBatchField
	WetnessField
	BurstGrainsField
	PcRunOptionalField
	NotesField
	WriteTagToField
}

func createJarHandler(w http.ResponseWriter, r *http.Request) {
	data := createJarRequest{}
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

	ctx, _ := Db(r)
	if data.SizeCups < 1 {
		http.Error(w, fmt.Sprintf("size must be >0: was %d", data.SizeCups), http.StatusBadRequest)
		return
	}
	batch, err := data.GrainBatchField.Get(ctx)
	if err != nil {
		http.Error(w, "failed to find grain batch: "+err.Error(), http.StatusBadRequest)
		return
	}
	_, err = data.PcRunOptionalField.Get(ctx)
	if err != nil {
		if !errors.Is(err, ErrMissingOptionalField) {
			dbErr(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	ctx, now := request.UnixTime(r.Context())
	toInsert := GrainJar{
		MainCollectionIdField:   MainCollectionIdField{id},
		JarRecipeField:          JarRecipeField{&batch.Recipe},
		GrainBatchOptionalField: data.GrainBatchField.asOptional(),
		SizeCups:                data.SizeCups,
		PcRunOptionalField:      data.PcRunOptionalField, // can be nil, which is ok, means not PC'd yet.
		BurstGrainsField:        data.BurstGrainsField,   //      data.BurstGrainsField, // can be updated later if not initially set
		WetnessField:            data.WetnessField,       //           data.WetnessField, // can be updated later if not initially set
		CreationDateField:       CreationDateField{now},
		NotesField:              NotesField{data.Notes},
		LastUpdatedField:        LastUpdatedField{now},
		AclField:                allCanWriteAcl(),
	}

	err = writeRfidTagIfNecessary(ctx, data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	finishCreateMainCollectionEntry(ctx, &toInsert, w)
}

type importJarRequest struct {
	SizeCups int                    `json:"sizeCups"`
	Recipe   *AlternateCollectionId `json:"recipe,omitempty"` // Jar Recipe // TODO: optional for imports?
	CreationDateField
	SpeciesOptionalField // Only empty when non-innoc'd
	SubspeciesOptionalField
	Generation *Generation // required when innoculated
	WetnessField
	BurstGrainsField
	KnownFruitableField
	WriteTagToField
	// image as "img"
}

func importJarHandler(w http.ResponseWriter, r *http.Request) {
	data := importJarRequest{}
	id := NextMainCollectionId()
	b58id := id.AsBase58()
	r.Body = http.MaxBytesReader(w, r.Body, maxMultipartRequestSize) // TODO: do multipart streamlined way
	defer r.Body.Close()
	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "unable to open multipart reader: "+err.Error(), http.StatusBadRequest)
		return
	}
	p1, err := reader.NextPart()
	if err != nil {
		http.Error(w, "failed to get reader next part: "+err.Error(), http.StatusInternalServerError)
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
			err = pics.DeleteFiles(r.Context(), picsSaved...)
			if err != nil {
				handleFileDeleteErr(err)
			}
		}
	}()
	// Go to next part, if exists to get image
	var importedPic *PicWithNotes = nil
	ctx, now := request.UnixTime(r.Context())
	p, err := reader.NextPart()
	if err != nil {
		if err != io.EOF {
			http.Error(w, "EOF on next part: "+err.Error(), http.StatusInternalServerError)
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
		newFileNameWithPrefixPath, errr := pics.SaveFile(r.Context(), fieldBytes, "bag", string(b58id), "img")
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
	toInsert := GrainJar{
		MainCollectionIdField:   MainCollectionIdField{id},
		SizeCups:                data.SizeCups,
		JarRecipeField:          JarRecipeField{data.Recipe},
		GrainBatchOptionalField: GrainBatchOptionalField{nil}, // Batch not provided for import
		WetnessField:            data.WetnessField,
		BurstGrainsField:        data.BurstGrainsField,
		PcRunOptionalField:      PcRunOptionalField{&impPcRun}, // Default run on import
		CreationDateField:       CreationDateField{data.CreationDate},
		SpeciesOptionalField:    SpeciesOptionalField{data.Species},
		SubspeciesOptionalField: data.SubspeciesOptionalField,
		GenerationsFields: GenerationsFields{
			GenSporeField:        GenSporeField{gen},
			GenSinceFruitOrSpore: gen,
		},
		PicsField:            PicsField{pix},
		KnownFruitableField:  data.KnownFruitableField,
		MostRecentImageField: MostRecentImageField{importedPic},
		NotesField:           NotesField{}, // TODO: fix?
		LastUpdatedField:     LastUpdatedField{now},
		AclField:             AclField{finalPerms},
	}

	_, err = toInsert.PcRunOptionalField.Get(ctx)
	if err != nil && !errors.Is(err, ErrMissingOptionalField) {
		dbErr(w, "invalid pc run: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = toInsert.JarRecipeField.Get(ctx)
	if err != nil {
		if !errors.Is(err, ErrMissingOptionalField) {
			dbErr(w, "invalid jar recipe: "+err.Error(), http.StatusInternalServerError)
			return
		}

	}
	err = writeRfidTagIfNecessary(ctx, data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	finishImportMainCollectionEntry(ctx, &toInsert, w)
}

type updateJarRequest struct {
	NotesUpdateField
	KnownFruitableField
	DisposedField
	SaleField
	ImagesUpdateField  //"newPic-1"
	ContamsUpdateField //"newContam-1"
	WetnessField
	BurstGrainsField
	PermsOnRequest `json:"acl"`
}

func (upr updateJarRequest) reform() resolvedUpdateJarRequest {
	return resolvedUpdateJarRequest{
		KnownFruitableField: upr.KnownFruitableField,
		SaleField:           upr.SaleField,
		DisposedField:       upr.DisposedField,
		WetnessField:        upr.WetnessField,
		BurstGrainsField:    upr.BurstGrainsField,
		NotesUpdateField:    upr.NotesUpdateField,
		Images:              imageUpdates(upr.Images),
		Contams:             contamUpdates(upr.Contams),
		PermsOnRequest:      upr.PermsOnRequest,
	}
}

type resolvedUpdateJarRequest struct {
	KnownFruitableField
	SaleField
	DisposedField
	WetnessField
	BurstGrainsField
	NotesUpdateField
	Images  SplitEntries[picWithNotesForm, PicWithNotes]
	Contams SplitEntries[contamForm, Contamination]
	PermsOnRequest
}

func (req resolvedUpdateJarRequest) modsFor(existing *GrainJar, aclField AclField) (bson.D, error) {
	return NewMods().
		updateKnownFruitableIfNeeded(req, existing).
		updateSaleIfNeeded(req.Sale, existing.Sale).
		updateDisposedIfNeeded(req, existing).
		updateNotesIfNeeded(req, existing).
		updatePicsIfNeeded(req.Images, existing.Pics).
		updateMostRecentImageIfNeeded(existing.MostRecentImage, loadMriPics(&req.Images, &req.Contams, nil)).
		updateWetnessIfNeeded(req.Wetness, existing.Wetness).
		updateBurstGrainsIfNeeded(req.BurstGrains, existing.BurstGrains).
		updateContamsIfNeeded(req.Contams, existing.Contaminations).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateJarHandler(w http.ResponseWriter, r *http.Request) {
	data := updateJarRequest{}
	defer r.Body.Close()
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

	newPics, newContams, _, err := fullMultipartWithNoBreaks(w, r, &data, mainCollId.AsBase58())
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
	ctx := r.Context()
	// go get current
	existing := &GrainJar{}
	err = DbFrom(ctx).
		Collection(GrainJarCollectionName).FindOne(ctx, BsonFindFilter(IDfld, *mainCollId)).Decode(existing)
	if err != nil {
		dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishMainCollItemUpdate(ctx, w, out.modsFor, existing, out.PermsOnRequest)
}

func Db(r *http.Request) (context.Context, *mongo.Database) {
	ctx := r.Context()
	return ctx, DbFrom(ctx)
}

func deleteJarHandler(w http.ResponseWriter, r *http.Request) {
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
	// ensure item does not have any transfers in or out
	item, err := GetMainCollectionItemSpecific[*GrainJar](ctx, id, &GrainJar{})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			http.Error(w, "Item to be deleted not found: "+err.Error(), http.StatusNotFound) // TODO: ok?
		} else {
			http.Error(w, "Failed to retrieve item to be deleted: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if item.Parent != nil {
		// TODO: what if we want to remove it from the parent as well?
		http.Error(w, "Cannot delete innoculated items!", http.StatusConflict)
		return
	}
	if item.TransfersOut != nil && len(item.TransfersOut) > 0 {
		http.Error(w, "Cannot delete items with transfers out", http.StatusConflict)
		return
	}

	// Delete if not found elsewhere!
	DeleteCollectionItem(ctx, item.CollectionName(), id, w)
}
