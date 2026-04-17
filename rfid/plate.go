package rfid

// TODO: newFromAgarBatch (post-PC) typical

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

type CondensationCoverageAtSealTimeField struct {
	CondensationCoverageAtSealTime *int `bson:"condensationCoverageAtSealTime,omitempty" json:"condensationCoverageAtSealTime,omitempty"` // TODO: (0-100), HANDLE EVERYWHERE, NEW!
}
type PourCoverageField struct {
	PourCoverage *int `bson:"pourCoverage,omitempty" json:"pourCoverage,omitempty"` // PourCoverage TODO: (0-100) (or nil for imports), HANDLE EVERYWHERE, NEW!
}
type WetAtCooledTimeField struct {
	WetAtCooledTime *bool `bson:"wetAtCooledTime,omitempty" json:"wetAtCooledTime,omitempty"` // WetAtCooledTime TODO: (nil==unknown or imported, false==known and not wet, true=known and wet), HANDLE EVERYWHERE, NEW!
}
type AgarOnOutsideAtPourTimeField struct {
	AgarOnOutsideAtPourTime *bool `bson:"agarOnOutsideAtPourTime,omitempty" json:"agarOnOutsideAtPourTime,omitempty"` // Only when mispours happen with plates above this one // TODO: HANDLE EVERYWHERE, NEW!
}

type Plate struct {
	MainCollectionIdField `bson:"inline"`
	AgarBatchField        `bson:"inline"` // will be empty for preexisting
	// TODO: do we want PC run on here too? and on others like it?
	CreationDateField                   `bson:"inline"`
	CondensationCoverageAtSealTimeField `bson:"inline"` // Percentage of condensation surface area coverage at seal time
	PourCoverageField                   `bson:"inline"` // Percentage of bottom surface area agar coverage
	WetAtCooledTimeField                `bson:"inline"` // Wet when initially cooled? True, false, or unknown
	AgarOnOutsideAtPourTimeField        `bson:"inline"` // Agar got on the outside of the plate? True, false, or unknown
	SpeciesOptionalField                `bson:"inline"`
	SubspeciesOptionalField             `bson:"inline"`
	InnocField                          `bson:"inline"`
	GenerationsFields                   `bson:"inline"`
	TransfersOutField                   `bson:"inline"`
	ParentTypeField                     `bson:"inline"` // nil == mainCollectionType, can also be MSS or clone! // TODO: INDEX????
	MainCollectionOptionalParentField   `bson:"inline"`
	PicsField                           `bson:"inline"`
	ContaminationsField                 `bson:"inline"`
	KnownFruitableField                 `bson:"inline"`
	SaleField                           `bson:"inline"`
	DisposedField                       `bson:"inline"`
	MostRecentImageField                `bson:"inline"`
	NotesField                          `bson:"inline"`
	LastUpdatedField                    `bson:"inline"`
	AclField                            `bson:"inline"`
}

func (p Plate) IdValue() any {
	return p.Id.dbIdStr() // TODO: ensure ok
}

func (p Plate) CanTransferTo(dst geneticSource) error {
	if !slices.Contains([]string{BagSourceType, GrainJarSourceType, LcSourceType, PlateSourceType, PlugSourceType, SlantSourceType, StasisTubeSourceType}, dst.SourceType()) {
		return errors.New("plates cannot transfer to " + dst.SourceType())
	}
	return nil
}

func (p Plate) GeneticInfoAsParent() (GeneticParentInfo, error) {
	return GeneticParentInfo{
		SpeciesOptionalField:    SpeciesOptionalField{p.Species},
		SubspeciesOptionalField: p.SubspeciesOptionalField,
		KnownFruitableField:     p.KnownFruitableField,
		GenerationsFields:       p.GenerationsFields,
	}, nil
}

func (p Plate) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return p.GenSinceSpore, p.GenSinceFruitOrSpore
}

//func (p Plate) setTransferParent(ctx context.Context, xfer Transfer) error {
//	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(PlatesCollectionName)
//	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
//	// TODO: if transfer has a fromPic on it, can we add it to the parent?
//	if err != nil {
//		return err
//	}
//	res, err := coll.UpdateByID(ctx, p.Id, upd)
//	if err != nil {
//		return err
//	}
//	if res.ModifiedCount == 0 {
//		return ErrNoParentModifiedForTransfer
//	}
//	return nil
//}

func (p Plate) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
	parentInfo, genSpore, genFruitSpore, err := childGensForParent(from)
	if err != nil {
		return err
	}
	// TODO: if xfer has a pic on it for the to, can we add it to the child?
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
	res, err := mongo.SessionFromContext(ctx).Client().Database(dbName).Collection(PlatesCollectionName).UpdateByID(ctx, p.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer
	}
	return nil
}

func (p Plate) EntryTypeField() *string {
	return utils.Pointer(PlateSourceType)
}

func initializePlates(ctx context.Context) error {
	println("1")
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	println("2")
	coll := db.Collection(PlatesCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		newSimpleIndex("agarBatch", "agarBatch", false, true, false),
		creationDateIndexModel,
		//newSimpleIndex("condensationCoverageAtSealTime", "condensationCoverageAtSealTime", true, true, false),
		//newSimpleIndex("pourCoverage", "pourCoverage", false, true, false),
		//newSimpleIndex("wetAtCooledTime", "wetAtCooledTime", true, true, false),
		//newSimpleIndex("agarOnOutsideAtPourTime", "agarOnOutsideAtPourTime", true, true, false),
		newSimpleIndex("species", "species", false, true, false),
		newSimpleIndex("subspecies", "subspecies", false, true, false),
		//newSimpleIndex("innoc", "innoc", false, true, false),
		//newSimpleIndex("genSinceSpore", "genSpore", true, true, false),
		//newSimpleIndex("genSinceFruitOrSpore", "genFruitOrSpore", true, true, false),
		//transfersOutIndexModel,
		//newSimpleIndex("parent", "parent", false, true, false),         // TODO: nil is store or outside?
		//newSimpleIndex("parentType", "parentType", false, true, false), // TODO: nil is store or outside?
		//
		////Pics (no index)
		//// TODO: Contams
		//newSimpleIndex("knownFruitable", "knownFruitable", false, true, false),
		//saleIndexModel,
		//disposedIndexModel,
		//// MostRecentImage
		////Notes (no index) (maybe later with tags?)
		lastUpdatedIndexModel,
		projectsIndexModel,
	})
	if err != nil {
		return err
	}
	// If test plate does not exist, then create it
	testId := mainCollIdForint(idTestPlate) // 0th id, b58==1
	testItem := &Plate{
		MainCollectionIdField:   MainCollectionIdField{testId},
		AgarBatchField:          AgarBatchField{&exAltId},
		CreationDateField:       CreationDateField{exampleTime},
		SpeciesOptionalField:    SpeciesOptionalField{&testEntryStringId},
		SubspeciesOptionalField: SubspeciesOptionalField{&testEntryStringId},
		InnocField:              InnocField{&exAltId}, // TODO: multiple?
		GenerationsFields: GenerationsFields{
			GenSporeField:        GenSporeField{&exGenSinceSpore},
			GenSinceFruitOrSpore: &exGenSinceFruitSpore,
		},
		TransfersOutField:                 TransfersOutField{exAlts},
		ParentTypeField:                   ParentTypeField{&exParentType},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&testId}, // TODO: ok? // TODO: multiple?
		PicsField:                         PicsField{exPics},
		ContaminationsField:               ContaminationsField{exContams},
		KnownFruitableField:               KnownFruitableField{exBool},
		SaleField:                         SaleField{&exAltId},
		DisposedField:                     DisposedField{&exampleTime},
		MostRecentImageField:              MostRecentImageField{&exPics[0]},
		NotesField:                        NotesField{exampleNotes()},
		LastUpdatedField:                  LastUpdatedField{exampleTime},
	}
	err = addTestMainEntries(ctx, testItem)
	if err != nil {
		return err
	}
	// TODO: del after
	testPlateIds := []int{idTestPlateBlanketRead, idTestPlateUserWrite,
		idTestPlateUserRead,
		idTestPlateProjectWrite,
		idTestPlateProjectRead,
		idTestPlateUserWriteProjRead,
		idTestPlateUserOutsideProject,
	}
	testPlates := make([]*Plate, len(testPlateIds))
	platesMade := map[Base58Str]string{}
	for i, permInt := range testPlateIds {
		id := mainCollIdForint(permInt)
		platesMade[id.asBase58()] = testAclStrings[i]
		var tempCoverage = &i
		var tempBool *bool = utils.Pointer(false)
		switch i {
		case 1:
			tempCoverage = nil
			tempBool = nil
		case 2:
			tempBool = utils.Pointer(true)
		default:
			// Do nothing different
		}
		testPlates[i] = &Plate{
			MainCollectionIdField:               MainCollectionIdField{id},
			AgarBatchField:                      AgarBatchField{&exAltId},
			CreationDateField:                   CreationDateField{exampleTime},
			CondensationCoverageAtSealTimeField: CondensationCoverageAtSealTimeField{CondensationCoverageAtSealTime: tempCoverage},
			PourCoverageField:                   PourCoverageField{PourCoverage: tempCoverage},
			WetAtCooledTimeField:                WetAtCooledTimeField{WetAtCooledTime: tempBool},
			AgarOnOutsideAtPourTimeField:        AgarOnOutsideAtPourTimeField{AgarOnOutsideAtPourTime: tempBool},
			SpeciesOptionalField:                SpeciesOptionalField{&testEntryStringId},
			SubspeciesOptionalField:             SubspeciesOptionalField{&testEntryStringId},
			InnocField:                          InnocField{&exAltId}, // TODO: multiple?
			GenerationsFields: GenerationsFields{
				GenSporeField:        GenSporeField{&exGenSinceSpore},
				GenSinceFruitOrSpore: &exGenSinceFruitSpore,
			},
			TransfersOutField:                 TransfersOutField{exAlts},
			ParentTypeField:                   ParentTypeField{&exParentType},
			MainCollectionOptionalParentField: MainCollectionOptionalParentField{&testId}, // TODO: ok? // TODO: multiple?
			PicsField:                         PicsField{exPics},
			ContaminationsField:               ContaminationsField{exContams},
			KnownFruitableField:               KnownFruitableField{exBool},
			SaleField:                         SaleField{&exAltId},
			DisposedField:                     DisposedField{&exampleTime},
			MostRecentImageField:              MostRecentImageField{&exPics[0]},
			NotesField:                        NotesField{exampleNotes()},
			LastUpdatedField:                  LastUpdatedField{exampleTime},
			AclField:                          AclField{ACL: testAcls[i]},
		}
	}
	err = addTestMainEntries(ctx, testPlates...)
	if err != nil {
		return err
	}
	println("test plate urls: ") // TODO: del
	for id, str := range platesMade {
		println(id, str)
	}
	return nil
}

type createPlateRequest struct {
	AgarBatch AlternateCollectionId `json:"agarBatch"`
	CondensationCoverageAtSealTimeField
	PourCoverageField
	WetAtCooledTimeField
	AgarOnOutsideAtPourTimeField
	WriteTagToField
}

func createPlateHandler(w http.ResponseWriter, r *http.Request) {
	data := createPlateRequest{}
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

	now := unixTimeForNow()
	ctx := r.Context()
	toInsert := Plate{
		MainCollectionIdField:               MainCollectionIdField{id},
		AgarBatchField:                      AgarBatchField{AgarBatch: &data.AgarBatch},
		CreationDateField:                   CreationDateField{now},
		CondensationCoverageAtSealTimeField: data.CondensationCoverageAtSealTimeField, // TODO: handle on typescript side
		PourCoverageField:                   data.PourCoverageField,                   // TODO: handle on typescript side
		WetAtCooledTimeField:                data.WetAtCooledTimeField,                // TODO: handle on typescript side
		AgarOnOutsideAtPourTimeField:        data.AgarOnOutsideAtPourTimeField,        // TODO: handle on typescript side
		LastUpdatedField:                    LastUpdatedField{now},
		// No Perms here for basic plates
		AclField: allCanWriteAcl(),
	}
	_, err = toInsert.AgarBatchField.Get(ctx)
	if err != nil && !errors.Is(err, ErrMissingOptionalField) {
		http.Error(w, "agar batch field missing: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishCreateMainCollectionEntry(ctx, &toInsert, w) // TODO: use in all main creates
}

type updatePlateRequest struct {
	KnownFruitableField
	SaleField
	DisposedField
	Notes   AllEntries[Note]                                         `json:"notes"`
	Images  SplitEntries[picWithNotesForm, PicWithNotesLessLocation] `json:"images"`
	Contams SplitEntries[contamForm, ContaminationLessLocation]      `json:"contams"`
	WriteTagToField
	PermsOnRequest
}

func (upr updatePlateRequest) reform() resolvedUpdatePlateRequest {
	return resolvedUpdatePlateRequest{
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

func (mods resolvedUpdatePlateRequest) modsFor(existing *Plate, aclField AclField) (bson.D, error) {
	return NewMods().
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

type resolvedUpdatePlateRequest struct {
	KnownFruitableField
	SaleField
	DisposedField
	Notes   AllEntries[Note]
	Images  SplitEntries[picWithNotesForm, PicWithNotes]
	Contams SplitEntries[contamForm, Contamination]
	WriteTagToField
	PermsOnRequest
}

func updatePlateHandler(w http.ResponseWriter, r *http.Request) {
	println("RECEIVED UPDATE PLATE REQUEST") // TODO: THIS
	data := &updatePlateRequest{}
	b58Id := Base58Str(r.PathValue("id"))
	id, err := b58Id.toMainCollectionId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	println("CREATED READER") // TODO: THIS
	reader, err := multipartReaderForRequest(r, w, data)
	if err != nil {
		// Already written
		return
	}
	reqBs, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	println(string(reqBs)) // TODO: del
	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	println("GETTING IMAGES") // TODO: THIS
	newPics, newContams, _, err := getMultipartImages(r.Context(), "plate", w, reader, b58Id)
	if err != nil {
		// Already wrote
		return
	}

	// CHECK THAT ALL NEW PICS EXIST
	// PROCESS ALL NEW PICS AND CONTAMS
	println("REFORMING") // TODO: THIS
	for i, picNote := range data.Images.Existing[0].Data.Notes.asEntries() {
		println("note", i, picNote.Note)
	}
	out := data.reform()
	for i, _ := range data.Images.New {
		loc, exists := newPics[i]
		if !exists {
			println("FAILED TO GET LOCATION") // TODO: THIS
			http.Error(w, fmt.Sprintf("error, location for new picture index %d not found (should never happen)", i), http.StatusInternalServerError)
			return
		}
		out.Images.New[i].Location = imageLocation(loc)
	}
	println("PARSING CONTAMS") // TODO: THIS
	for i, _ := range data.Contams.New {
		if loc, exists := newContams[i]; exists {
			finalLoc := imageLocation(loc)
			out.Contams.New[i].Location = &finalLoc
		} else {
			println("no contam location for", i)
		}
	}
	ctx := r.Context()
	client := ctx.Value(mongoClientContextKey).(*mongo.Client)
	coll := client.Database(dbName).Collection(PlatesCollectionName)
	existing := &Plate{}
	err = coll.FindOne(ctx, bson.D{{"_id", id}}).Decode(existing)
	if err != nil {
		// TODO: an issue here?
		http.Error(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	for i, note := range out.Notes.asEntries() { // TODO: del
		println("Note", i, note.Note)
	}
	finishMainCollItemUpdate(ctx, w, coll, out.modsFor, existing, out.PermsOnRequest)
}

func handleUpdateMods[T any, U MainCollectionId | AlternateCollectionId | string](ctx context.Context, w http.ResponseWriter, coll *mongo.Collection, existing T, id U, upd bson.D, err error) {
	if err != nil {
		println("mod creation failure: " + err.Error())
		dbErr(w, "error creating txn:"+err.Error(), http.StatusInternalServerError)
		return
	}
	if len(upd) == 0 {
		println("no changes made") // TODO: del
		dbErr(w, "no changes made", http.StatusBadRequest)
		return
	}
	// write updates to db
	println("updating") // TODO: del
	bsonId := bson.D{{"_id", id}}
	err = coll.FindOneAndUpdate(ctx, bsonId, upd).Err()
	if err != nil {
		println("failed to update") // TODO: del
		dbErr(w, "failed to write update to db: "+err.Error(), http.StatusInternalServerError)
		return
	}
	println("finding") // TODO: del
	err = coll.FindOne(ctx, bsonId).Decode(&existing)
	if err != nil {
		println("failed to find") // TODO: del
		dbErr(w, "failed to write update to db: "+err.Error(), http.StatusInternalServerError)
		return
	}
	println("marshalling") // TODO: del
	bsOut, err := json.Marshal(existing)
	if err != nil {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	println("Writing update:", string(bsOut))
	_, err = w.Write(bsOut)
	println("Wrote update:", string(bsOut))
	handleWriteErr(err, w)
}

type importPlateRequest struct {
	CreationDateField
	SpeciesField
	SubspeciesOptionalField
	KnownFruitableField
	Generation *int `json:"generation,omitempty"`
	PourCoverageField
	// pic as "img"
	WriteTagToField
	PermsOnRequest
}

func importPlateHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data := importPlateRequest{}
	id := NextMainCollectionId()
	//id, err := newMainCollectionId(r.Context(), PlatesCollectionName)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
	b58id := id.asBase58()
	println("multipart reader if necessary")
	reader, err := multipartReaderForRequest(r, w, &data)
	if err != nil {
		// Already written
		return
	}
	//if err = Data.Perms.ValidateUserCanWrite(r.Context()); err != nil {
	//	http.Error(w, "email cannot write perms: "+err.Error(), http.StatusBadRequest)
	//	return
	//}
	println("writing tag if necessary")
	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Try to get pic if exists
	picsSaved := []string{} // TODO: do this streamlined with the multipart function
	defer func() {
		if err != nil {
			err = pics.DeleteFiles(r.Context(), picsSaved...)
			if err != nil {
				handleFileDeleteErr(err)
			}
		}
	}()
	// Go to next part, if exists to get image
	println("reading parts")
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
		newFileNameWithPrefixPath, errr := pics.SaveFile(r.Context(), fieldBytes, "plate", string(b58id), "img")
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
	println("getting auth info")
	user, err := GetAuthInfo(r.Context())
	if err != nil {
		http.Error(w, "failed to get auth info: "+err.Error(), http.StatusUnauthorized)
		return
	}
	println("getting spsub")
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

	ctx, db := Db(r)
	coll := db.Collection(PlatesCollectionName)

	toInsert := Plate{
		MainCollectionIdField:               MainCollectionIdField{id},
		CreationDateField:                   data.CreationDateField,
		CondensationCoverageAtSealTimeField: CondensationCoverageAtSealTimeField{nil},
		PourCoverageField:                   data.PourCoverageField,
		WetAtCooledTimeField:                WetAtCooledTimeField{nil},
		AgarOnOutsideAtPourTimeField:        AgarOnOutsideAtPourTimeField{nil},
		SpeciesOptionalField:                data.SpeciesField.AsOptional(),
		SubspeciesOptionalField:             data.SubspeciesOptionalField,
		GenerationsFields:                   GenerationsFieldFor(gen),
		PicsField:                           PicsField{pix},
		KnownFruitableField:                 data.KnownFruitableField,
		MostRecentImageField:                MostRecentImageField{importedPic},
		LastUpdatedField:                    LastUpdatedField{unixTimeForNow()},
		AclField:                            AclField{finalPerms},
	}
	println("inserting plate")
	finishImportMainCollectionEntry(ctx, coll, &toInsert, data.PermsOnRequest, w)
}
