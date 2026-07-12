package api

// TODO: newFromAgarBatch (post-PC) typical

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/reeceappling/goUtils/v2/utils"
	"github.com/reeceappling/mushDb/api/env"
	"github.com/reeceappling/mushDb/api/pics"
	"github.com/reeceappling/mushDb/api/request"
	"github.com/reeceappling/mushDb/api/request/unix"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"slices"
)

// sometimes needed for transfers
// needed for clones

type CondensationCoverageAtSealTimeField struct {
	CondensationCoverageAtSealTime *int `bson:"condensationCoverageAtSealTime,omitempty" json:"condensationCoverageAtSealTime,omitempty"` // TODO: (0-100), HANDLE EVERYWHERE, NEW!
}

func (cc CondensationCoverageAtSealTimeField) condensationCoverage() *int {
	return cc.CondensationCoverageAtSealTime
}

type hasCondensCov interface {
	condensationCoverage() *int
}
type CondensationCoverageAtPourTimeField struct {
	CondensationCoverageAtPourTime *int `bson:"condensationCoverageAtPourTime,omitempty" json:"condensationCoverageAtPourTime,omitempty"` // TODO: (0-100), HANDLE EVERYWHERE, NEW!
}

func (cc CondensationCoverageAtPourTimeField) condensationCoveragePourTime() *int {
	return cc.CondensationCoverageAtPourTime
}

type hasCondensCovPourTime interface {
	condensationCoveragePourTime() *int
}
type PourCoverageField struct {
	PourCoverage *int `bson:"pourCoverage,omitempty" json:"pourCoverage,omitempty"` // PourCoverage (0-100) (or nil for imports)
}

func (f PourCoverageField) pourCoverage() *int {
	return f.PourCoverage
}

type hasPourCoverage interface {
	pourCoverage() *int
}
type WetAtCooledTimeField struct {
	WetAtCooledTime *bool `bson:"wetAtCooledTime,omitempty" json:"wetAtCooledTime,omitempty"` // WetAtCooledTime TODO: (nil==unknown or imported, false==known and not wet, true=known and wet)
}

func (f WetAtCooledTimeField) wetAtCool() *bool {
	return f.WetAtCooledTime
}

type hasWact interface {
	wetAtCool() *bool
}
type AgarOnOutsideAtPourTimeField struct {
	AgarOnOutsideAtPourTime *bool `bson:"agarOnOutsideAtPourTime,omitempty" json:"agarOnOutsideAtPourTime,omitempty"` // Only when mispours happen with plates above this one
}

func (f AgarOnOutsideAtPourTimeField) agarOutside() *bool {
	return f.AgarOnOutsideAtPourTime
}

type hasAgarOutside interface {
	agarOutside() *bool
}

type Plate struct {
	MainCollectionIdField `bson:"inline"`
	AgarBatchField        `bson:"inline"` // will be empty for preexisting
	// TODO: do we want PC run on here too? and on others like it? (probably not due to data bloat?)
	CreationDateField                   `bson:"inline"`
	CondensationCoverageAtPourTimeField `bson:"inline"` // Percentage of condensation surface area coverage at pour time
	CondensationCoverageAtSealTimeField `bson:"inline"` // Percentage of condensation surface area coverage at seal time
	PourCoverageField                   `bson:"inline"` // Percentage of bottom surface area agar coverage
	WetAtCooledTimeField                `bson:"inline"` // Wet when initially cooled? True, false, or unknown
	AgarOnOutsideAtPourTimeField        `bson:"inline"` // Agar got on the outside of the plate? True, false, or unknown
	SpeciesOptionalField                `bson:"inline"`
	SubspeciesOptionalField             `bson:"inline"`
	InnocField                          `bson:"inline"`
	GenerationsFields                   `bson:"inline"`
	TransfersOutField                   `bson:"inline"`
	ParentTypeField                     `bson:"inline"` // nil == imported, otherwise entryType, what happens with clones?! // TODO: INDEX????
	MainCollectionOptionalParentField   `bson:"inline"`
	PicsField                           `bson:"inline"`
	ContaminationsField                 `bson:"inline"`
	KnownFruitableField                 `bson:"inline"`
	SaleField                           `bson:"inline"` // TODO: how to set not sellable? not ready for sale? ready for sale? returned? sold? resold?
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

func (p Plate) Innoculatable() error {
	return errors.Join(
		p.RequireNoSpecies(),
		p.RequireNoSubspecies(),
		p.RequireNotDisposed(),
		p.RequireUnsold(),
		p.RequireUnknownFruitable(),
		p.RequireNoInnoculation())
}

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
		withSubspecies(parentInfo.Subspecies).
		withKnownFruitable(parentInfo.KnownFruitable).
		withPerms(from.Permissions()).
		withLastUpdated(xfer.LastUpdated).
		Finalized()
	if err != nil {
		return errors.Join(err, ErrFailedToFinalizeMods)
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

func initializePlates(ctx context.Context) error {
	db := DbFrom(ctx)
	coll := db.Collection(PlatesCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		newSimpleIndex("agarBatch", "agarBatch", false, true, false), // Required index for batch deletes
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
		//newSimpleIndex("parent", "parent", false, true, false),
		//newSimpleIndex("parentType", "parentType", false, true, false),
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
	return env.IfNotProd(ctx, func() error {
		if err != nil {
			return err
		}
		testPlateIds := []int{
			idTestPlateBlanketWrite,
			idTestPlateBlanketRead,
			idTestPlateAdminOnly,
			idTestPlateProjectWrite,
			idTestPlateProjectRead,
			idTestPlateUserOutsideProject,
		}
		platesMade := map[Base58Str]string{}
		testPlates := []*Plate{}
		// If test plate does not exist, then create it
		firstPlate := basicTestPlate()
		testId := firstPlate.Id
		platesMade[testId.AsBase58()] = "test plate with id 1"
		testPlates = append(testPlates, &firstPlate)
		emptyPlate := emptyTestPlate()
		emptyTestId := emptyPlate.Id
		platesMade[emptyTestId.AsBase58()] = "test empty plate"

		testPlates = append(testPlates, &emptyPlate)
		for i, permInt := range testPlateIds {
			id := mainCollIdForint(permInt)
			platesMade[id.AsBase58()] = testAclStrings[i]
			var tempCoverage = &i
			var tempBool = utils.Pointer(false)
			switch i {
			case 1:
				tempCoverage = nil
				tempBool = nil
			case 2:
				tempBool = utils.Pointer(true)
			default:
				// Do nothing different
			}
			// Project-user + Project-entry permissions test plates
			newTestPlate := testPlateFor(id, &exAltId, exampleTime, tempCoverage, tempCoverage, tempBool, tempBool,
				&testEntryStringId, &testEntryStringId, &exAltId, &exGenSinceSpore, &exGenSinceFruitSpore, exAlts,
				&exParentType, &testId, exPics, exContams, exBool, &exAltId, &exampleTime, &exPics[0], exampleNotes(),
				exampleTime, testAcls[i])
			testPlates = append(testPlates, &newTestPlate)
		}
		println("Built-in plates entries:")
		for tempId, name := range platesMade {
			println(tempId, name)
		}
		return addTestMainEntries(ctx, testPlates...)
	})
}

func basicTestPlate() Plate {
	testId := mainCollIdForint(idTestPlate) // 0th id, b58==1
	return Plate{
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
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&testId},
		PicsField:                         PicsField{exPics},
		ContaminationsField:               ContaminationsField{exContams},
		KnownFruitableField:               KnownFruitableField{exBool},
		SaleField:                         SaleField{&exAltId},
		DisposedField:                     DisposedField{&exampleTime},
		MostRecentImageField:              MostRecentImageField{&exPics[0]},
		NotesField:                        NotesField{exampleNotes()},
		LastUpdatedField:                  LastUpdatedField{exampleTime},
		AclField:                          allCanWriteAcl(), // TODO: ok?
	}
}
func emptyTestPlate() Plate {
	testEmptyId := mainCollIdForint(idTestPlateEmpty) // 0th id, b58==1
	return Plate{
		MainCollectionIdField:   MainCollectionIdField{testEmptyId},
		AgarBatchField:          AgarBatchField{&exAltId},
		CreationDateField:       CreationDateField{exampleTime},
		SpeciesOptionalField:    SpeciesOptionalField{},
		SubspeciesOptionalField: SubspeciesOptionalField{},
		InnocField:              InnocField{},
		GenerationsFields: GenerationsFields{
			GenSporeField: GenSporeField{},
		},
		TransfersOutField:                 TransfersOutField{},
		ParentTypeField:                   ParentTypeField{},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{},
		PicsField:                         PicsField{},
		ContaminationsField:               ContaminationsField{},
		KnownFruitableField:               KnownFruitableField{},
		SaleField:                         SaleField{},
		DisposedField:                     DisposedField{},
		MostRecentImageField:              MostRecentImageField{},
		NotesField:                        NotesField{},
		LastUpdatedField:                  LastUpdatedField{exampleTime},
		AclField:                          allCanWriteAcl(), // TODO: ok?
	}
}
func testPlateFor(
	id MainCollectionId,
	agarBatchId *AlternateCollectionId,
	creationDate unix.Time,
	condensationCoverageAtSealTime,
	pourCoverage *int,
	wetAtCooledTime,
	agarOnOutsideAtPourTime *bool,
	species,
	subspecies *string,
	innoc *AlternateCollectionId,
	genSpore, genFruit *Generation,
	xfersOut []AlternateCollectionId,
	parentType *string, parent *MainCollectionId,
	picsForItem []PicWithNotes,
	contams []Contamination,
	knownFruitable *bool,
	sale *AlternateCollectionId,
	disposed *unix.Time,
	mostRecentImage *PicWithNotes,
	notes []Note,
	lastUpdated unix.Time,
	acl ACL,
) Plate {
	return Plate{
		MainCollectionIdField:               MainCollectionIdField{id},
		AgarBatchField:                      AgarBatchField{agarBatchId},
		CreationDateField:                   CreationDateField{creationDate},
		CondensationCoverageAtSealTimeField: CondensationCoverageAtSealTimeField{condensationCoverageAtSealTime},
		PourCoverageField:                   PourCoverageField{pourCoverage},
		WetAtCooledTimeField:                WetAtCooledTimeField{wetAtCooledTime},
		AgarOnOutsideAtPourTimeField:        AgarOnOutsideAtPourTimeField{agarOnOutsideAtPourTime},
		SpeciesOptionalField:                SpeciesOptionalField{species},
		SubspeciesOptionalField:             SubspeciesOptionalField{subspecies},
		InnocField:                          InnocField{innoc},
		GenerationsFields: GenerationsFields{
			GenSporeField:        GenSporeField{genSpore},
			GenSinceFruitOrSpore: genFruit,
		},
		TransfersOutField:                 TransfersOutField{xfersOut},
		ParentTypeField:                   ParentTypeField{parentType},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{parent},
		PicsField:                         PicsField{picsForItem},
		ContaminationsField:               ContaminationsField{contams},
		KnownFruitableField:               KnownFruitableField{knownFruitable},
		SaleField:                         SaleField{sale},
		DisposedField:                     DisposedField{disposed},
		MostRecentImageField:              MostRecentImageField{mostRecentImage},
		NotesField:                        NotesField{notes},
		LastUpdatedField:                  LastUpdatedField{lastUpdated},
		AclField:                          AclField{acl},
	}
}

func EmptyTestPlateBinaryId() MainCollectionId {
	return mainCollIdForint(idTestPlateEmpty)
}

type createPlateRequest struct {
	AgarBatch AlternateCollectionId `json:"agarBatch"`
	CondensationCoverageAtPourTimeField
	PourCoverageField
	WetAtCooledTimeField
	AgarOnOutsideAtPourTimeField
	NotesField
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

	ctx, now := request.UnixTime(r.Context()) // TODO: no more r.Context below
	agarBatchField := AgarBatchField{AgarBatch: &data.AgarBatch}
	_, err = agarBatchField.Get(ctx)
	if err != nil && !errors.Is(err, ErrMissingOptionalField) {
		http.Error(w, "agar batch field missing: "+err.Error(), http.StatusBadRequest)
		return
	}
	toInsert := Plate{
		MainCollectionIdField:               MainCollectionIdField{id},
		AgarBatchField:                      agarBatchField,
		CreationDateField:                   CreationDateField{now},
		CondensationCoverageAtPourTimeField: data.CondensationCoverageAtPourTimeField,
		CondensationCoverageAtSealTimeField: CondensationCoverageAtSealTimeField{nil},
		PourCoverageField:                   data.PourCoverageField,
		WetAtCooledTimeField:                data.WetAtCooledTimeField,
		AgarOnOutsideAtPourTimeField:        data.AgarOnOutsideAtPourTimeField,
		NotesField:                          data.NotesField,
		LastUpdatedField:                    LastUpdatedField{now},
		// No Perms here for basic plates
		AclField: allCanWriteAcl(),
	}
	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	finishCreateMainCollectionEntry(ctx, &toInsert, w)
}

type updatePlateRequest struct {
	CondensationCoverageAtSealTimeField
	PourCoverageField
	WetAtCooledTimeField
	AgarOnOutsideAtPourTimeField
	KnownFruitableField
	SaleField
	DisposedField
	NotesUpdateField
	ImagesUpdateField
	ContamsUpdateField
	PermsOnRequest `json:"acl"`
}

func (upr updatePlateRequest) reform() resolvedUpdatePlateRequest {
	return resolvedUpdatePlateRequest{
		CondensationCoverageAtSealTimeField: upr.CondensationCoverageAtSealTimeField,
		PourCoverageField:                   upr.PourCoverageField,
		WetAtCooledTimeField:                upr.WetAtCooledTimeField,
		AgarOnOutsideAtPourTimeField:        upr.AgarOnOutsideAtPourTimeField,
		KnownFruitableField:                 upr.KnownFruitableField,
		SaleField:                           upr.SaleField,
		DisposedField:                       upr.DisposedField,
		NotesUpdateField:                    upr.NotesUpdateField,
		Images:                              imageUpdates(upr.Images),
		Contams:                             contamUpdates(upr.Contams),
		PermsOnRequest:                      upr.PermsOnRequest,
	}
}

func loadMriPics(pics *SplitEntries[picWithNotesForm, PicWithNotes], contams *SplitEntries[contamForm, Contamination], flushes *SplitEntries[picWithNotesForm, PicWithNotes]) []PicWithNotes {
	imagesForUpdateFunc := []PicWithNotes{}
	if pics != nil {
		for _, ex := range pics.Existing {
			if !ex.Disabled {
				imagesForUpdateFunc = append(imagesForUpdateFunc, ex.Data.convert())
			}
		}
		imagesForUpdateFunc = append(imagesForUpdateFunc, pics.New...) // TODO: is this backwards???
	}
	if contams != nil {
		for _, ex := range contams.Existing {
			if !ex.Disabled {
				imagesForUpdateFunc = append(imagesForUpdateFunc, *ex.Data.convert().getPicWithNotes())
			}
		}
		for _, c := range contams.New {
			imagesForUpdateFunc = append(imagesForUpdateFunc, *c.getPicWithNotes())
		}
	}

	if flushes != nil {
		for _, f := range flushes.Existing {
			if !f.Disabled {
				imagesForUpdateFunc = append(imagesForUpdateFunc, f.Data.convert())
			}
		}
		imagesForUpdateFunc = append(imagesForUpdateFunc, flushes.New...) // TODO: is this backwards???
	}
	return imagesForUpdateFunc
}

func (req resolvedUpdatePlateRequest) modsFor(existing *Plate, aclField AclField) (bson.D, error) {
	return NewMods().
		updateCondensationCoverageIfNeeded(req, existing).
		updatePourCoverageIfNeeded(req, existing).
		updateWetAtCooledTimeIfNeeded(req, existing).
		updateAgarOnOutsideAtPourTimeIfNeeded(req, existing).
		updateKnownFruitableIfNeeded(req, existing).
		updateSaleIfNeeded(req.Sale, existing.Sale).
		updateDisposedIfNeeded(req, existing).
		updateNotesIfNeeded(req, existing).
		updatePicsIfNeeded(req.Images, existing.Pics).
		updateMostRecentImageIfNeeded(existing.MostRecentImage, loadMriPics(&req.Images, &req.Contams, nil)).
		updateContamsIfNeeded(req.Contams, existing.Contaminations).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

type resolvedUpdatePlateRequest struct {
	CondensationCoverageAtSealTimeField
	PourCoverageField
	WetAtCooledTimeField
	AgarOnOutsideAtPourTimeField
	KnownFruitableField
	SaleField
	DisposedField
	NotesUpdateField
	Images         SplitEntries[picWithNotesForm, PicWithNotes]
	Contams        SplitEntries[contamForm, Contamination]
	PermsOnRequest `json:"acl"`
}

func updatePlateHandler(w http.ResponseWriter, r *http.Request) {
	data := &updatePlateRequest{}
	b58Id, id, err := mainCollIdFromRequest(r, w)
	if err != nil {
		return
	}
	reqBs, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	println("REQUEST BYTES: ", string(reqBs)) // TODO: del

	newPics, newContams, _, err := fullMultipartWithNoBreaks(w, r, "plate", &data, b58Id)
	if err != nil {
		// Already wrote
		return
	}

	// CHECK THAT ALL NEW PICS EXIST
	// PROCESS ALL NEW PICS AND CONTAMS
	//// TODO: PANICKING WHEN SENDING EMPTY THINGS!!!!
	//for i, picNote := range data.Images.Existing[0].Data.Notes.asEntries() {
	//	println("note", i, picNote.Note)
	//}
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
		} else {
			println("no contam location for", i)
		}
	}
	finalReqBs, err := json.MarshalIndent(out, "", " ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	println("REQUEST BYTES: ", string(finalReqBs)) // TODO: del

	ctx := r.Context()
	client := GetMongoClient(ctx)
	coll := client.Database(dbName).Collection(PlatesCollectionName)
	existing := &Plate{}
	err = coll.FindOne(ctx, BsonFindFilter(IDfld, id)).Decode(existing)
	if err != nil {
		// TODO: an issue here?
		http.Error(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	finishMainCollItemUpdate(ctx, w, out.modsFor, existing, out.PermsOnRequest)
}

func handleUpdateMods[T any, U MainCollectionId | AlternateCollectionId | string](ctx context.Context, w http.ResponseWriter, coll *mongo.Collection, existing T, id U, upd bson.D, err error) {
	if err != nil {
		println("mod creation failure: " + err.Error())
		dbErr(w, "error creating txn:"+err.Error(), http.StatusInternalServerError)
		return
	}
	if len(upd) == 0 {
		dbErr(w, "no changes made", http.StatusBadRequest)
		return
	}
	// write updates to db
	bsonId := BsonFindFilter(IDfld, id)
	err = coll.FindOneAndUpdate(ctx, bsonId, upd).Err()
	if err != nil {
		dbErr(w, "failed to write update to db: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = coll.FindOne(ctx, bsonId).Decode(&existing)
	if err != nil {
		dbErr(w, "failed to write update to db: "+err.Error(), http.StatusInternalServerError)
		return
	}
	bsOut, err := json.Marshal(existing)
	if err != nil {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	bsOut2, err2 := json.MarshalIndent(existing, "", " ") // TODO: delete later
	if err2 != nil {
		dbErr(w, err2.Error(), http.StatusInternalServerError)
		return
	}
	println("Writing update:", string(bsOut2)) // TODO: del
	_, err = w.Write(bsOut)
	handleWriteErr(err, w)
}

func handleUpdateModsInTxn[T any, U MainCollectionId | AlternateCollectionId | string](ctx context.Context, coll *mongo.Collection, existing T, id U, upd bson.D, err error) error {
	if err != nil {
		return err
	}
	if len(upd) == 0 {
		return errors.New("no changes made")
	}
	// write updates to db
	bsonId := BsonFindFilter(IDfld, id)
	err = coll.FindOneAndUpdate(ctx, bsonId, upd).Err()
	if err != nil {
		return err
	}
	err = coll.FindOne(ctx, bsonId).Decode(&existing)
	if err != nil {
		return err
	}
	return nil
}

type importPlateRequest struct {
	CreationDateField
	SpeciesOptionalField
	SubspeciesOptionalField
	KnownFruitableField
	Generation *Generation `json:"generation,omitempty"` // TODO: make required for when innoculated!
	PourCoverageField
	CondensationCoverageAtPourTimeField
	// pic as "img"
	WriteTagToField
}

func importPlateHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctx, now := request.UnixTime(r.Context()) // TODO: no more r.Context below
	data := importPlateRequest{}
	id := NextMainCollectionId()
	b58id := id.AsBase58()
	println("multipart reader if necessary")
	reader, err := multipartReaderForRequest(r.WithContext(ctx), w, &data)
	if err != nil {
		println("failed in multipart reader area") // TODO: del
		// Already written
		return
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
			println("failed in nextPart") // TODO: del
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		fileName := p.FileName()
		defer p.Close()
		if fileName != "img" {
			println("invalid image name") // TODO: del
			http.Error(w, "invalid image name", http.StatusBadRequest)
			return
		}
		// Process file
		fieldBytes, err := multipartToImageBytes(p, w)
		if err != nil {
			// Already wrote
			println("failed in multipartToImageBytes") // TODO: del
			return
		}
		newFileNameWithPrefixPath, errr := pics.SaveFile(ctx, fieldBytes, "plate", string(b58id), "img")
		if errr != nil {
			err = errr
			println("failed to save file: " + err.Error()) // TODO: del
			http.Error(w, "failed to save file: "+err.Error(), http.StatusBadRequest)
			return
		}
		picsSaved = append(picsSaved, newFileNameWithPrefixPath)
		importedPic = utils.Pointer(newPicWithNotes(now, []Note{}, ImageLocation(newFileNameWithPrefixPath)))
	}
	var condensCovSealed *int = nil
	var gen *Generation = nil
	if data.Species != nil {
		if data.Generation == nil {
			println("innoculated must have generation") // TODO: del
			http.Error(w, "innoculated must have generation", http.StatusBadRequest)
			return
		}
		if *data.Generation < 1 {
			println("gen must be positive") // TODO: del
			http.Error(w, "gen must be positive", http.StatusBadRequest)
			return
		}
		gen = data.Generation
		// If innoculated, ensure seal and pour condensation coverage are the same
		condensCovSealed = data.CondensationCoverageAtPourTime
	} else {
		data.KnownFruitable = nil
		data.Subspecies = nil
	}
	pix := []PicWithNotes{}
	if importedPic != nil {
		pix = []PicWithNotes{*importedPic}
	}

	var finalPerms ACL
	if data.Species == nil { // Not innoculated
		finalPerms = allCanWriteAcl().ACL
		data.PourCoverage = nil
	} else {
		finalPerms, err = ImportFinalPerms(ctx, *data.Species, data.Subspecies)
		if err != nil {
			println("failed to get species and/or subspecies: " + err.Error()) // TODO: del
			http.Error(w, "failed to get species and/or subspecies: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	toInsert := Plate{
		MainCollectionIdField:               MainCollectionIdField{id},
		CreationDateField:                   data.CreationDateField,
		CondensationCoverageAtPourTimeField: data.CondensationCoverageAtPourTimeField,
		CondensationCoverageAtSealTimeField: CondensationCoverageAtSealTimeField{condensCovSealed},
		PourCoverageField:                   data.PourCoverageField,
		WetAtCooledTimeField:                WetAtCooledTimeField{nil},
		AgarOnOutsideAtPourTimeField:        AgarOnOutsideAtPourTimeField{nil},
		SpeciesOptionalField:                data.SpeciesOptionalField,
		SubspeciesOptionalField:             data.SubspeciesOptionalField,
		GenerationsFields:                   GenerationsFieldFor(gen),
		PicsField:                           PicsField{pix},
		KnownFruitableField:                 data.KnownFruitableField,
		MostRecentImageField:                MostRecentImageField{importedPic},
		LastUpdatedField:                    LastUpdatedField{now},
		AclField:                            AclField{finalPerms},
	}
	err = writeRfidTagIfNecessary(ctx, data.WriteTagTo, id)
	if err != nil {
		println("failed to write tag: " + err.Error()) // TODO: del
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	println("trying to import the plate...")
	finishImportMainCollectionEntry(ctx, &toInsert, w)
}

// TODO: MOVE!
func ImportFinalPerms(ctx context.Context, spec string, subspec *string) (ACL, error) {
	var finalPerms ACL
	sp, subsp, err := getSpeciesAndSubspecies(ctx, spec, subspec)
	if err != nil {
		return ACL{}, errors.New("failed to get species or subspecies: " + err.Error())
	}
	if subsp != nil {
		finalPerms = subsp.DefaultAcl.Clone()
	} else {
		finalPerms = sp.DefaultAcl.Clone()
	}
	userEmail := GetUserEmail(ctx)
	if finalPerms.Users == nil {
		finalPerms.Users = map[string]bool{}
	}
	finalPerms.Users[userEmail] = true

	return finalPerms, nil
}

func deletePlateHandler(w http.ResponseWriter, r *http.Request) {
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
	item, err := GetMainCollectionItemSpecific[*Plate](ctx, id, &Plate{})
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
		http.Error(w, "Cannot delete innoculated items!", http.StatusConflict) // TODO: type ok?
		return
	}
	if item.TransfersOut != nil && len(item.TransfersOut) > 0 {
		http.Error(w, "Cannot delete items with transfers out", http.StatusConflict) // TODO: type ok?
		return
	}

	// Delete if not found elsewhere!
	DeleteCollectionItem(ctx, item.CollectionName(), id, w)
}
