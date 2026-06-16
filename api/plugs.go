package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/reeceappling/goUtils/v2/utils"
	"github.com/reeceappling/mushDb/api/request"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"slices"
)

// sometimes needed for transfers

type PlugsJar struct {
	MainCollectionIdField             `bson:"inline"`
	ParentTypeField                   `bson:"inline"` // empty==bought
	MainCollectionOptionalParentField `bson:"inline"` // TODO: empty=bought. From plugs, jar, LC, plate/slant
	CreationDateField                 `bson:"inline"`
	DowelTypes                        []Dowel `bson:"dowelTypes" json:"dowelTypes"`
	GenerationsFields                 `bson:"inline"`
	SpeciesOptionalField              `bson:"inline"`
	TransfersOutField                 `bson:"inline"`
	SubspeciesOptionalField           `bson:"inline"`
	InnocField                        `bson:"inline"`
	//PicsField                           `bson:"inline"` // TODO: pics?
	//ContaminationsField                 `bson:"inline"` // TODO: contams?
	KnownFruitableField `bson:"inline"`
	PcRunOptionalField  `bson:"inline"` // TODO: used to be required, but not found for imports! created before innoculation
	SalesField          `bson:"inline"` // TODO: MULTIPLE!
	DisposedField       `bson:"inline"` // Also changed once all pegs are sold/used?
	NotesField          `bson:"inline"`
	LastUpdatedField    `bson:"inline"`
	AclField            `bson:"inline"`
}

func (pl PlugsJar) CanTransferTo(dst geneticSource) error {
	if !slices.Contains([]string{BagSourceType, GrainJarSourceType, LcSourceType, PlateSourceType, PlugSourceType, SlantSourceType, StasisTubeSourceType}, dst.SourceType()) {
		return errors.New("plugs cannot transfer to " + dst.SourceType())
	}
	return nil
}

func (pl PlugsJar) GeneticInfoAsParent() (GeneticParentInfo, error) {
	return GeneticParentInfo{
		SpeciesOptionalField:    SpeciesOptionalField{pl.Species},
		SubspeciesOptionalField: pl.SubspeciesOptionalField,
		KnownFruitableField:     pl.KnownFruitableField,
		GenerationsFields:       pl.GenerationsFields,
	}, nil
}

func (pl PlugsJar) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
	parentInfo, genSpore, genFruitSpore, err := childGensForParent(from)
	if err != nil {
		return err
	}
	upd, err := xfer.PicsModsForChild().
		withInnoc(xfer).
		withParentType(&xfer.FromType).
		withParent(utils.Pointer(from.DbId())).
		withGens(genSpore, genFruitSpore).
		withSpecies(parentInfo.Species).
		withSubspecies(parentInfo.Subspecies).
		withPerms(from.Permissions()).
		withLastUpdated(xfer.LastUpdated).
		Finalized()
	if err != nil {
		return ErrFailedToFinalizeMods
	}
	res, err := mongo.SessionFromContext(ctx).Client().Database(dbName).Collection(PlugsCollectionName).UpdateByID(ctx, pl.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer
	}
	return nil
}

func (pl PlugsJar) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return pl.GenSinceSpore, pl.GenSinceFruitOrSpore
}

func (pl PlugsJar) Innoculatable() bool {
	return pl.Species == nil &&
		pl.Subspecies == nil &&
		pl.Disposed == nil &&
		pl.KnownFruitable == nil &&
		pl.Innoc == nil
}

type Dowel struct {
	Radius `bson:"inline"`
	Wood   Wood `bson:"wood" json:"wood"`
}

func (d Dowel) validate() error {
	if !slices.Contains(woods, d.Wood) {
		return errors.New("invalid wood type")
	}
	if d.Size <= 0.0 {
		return errors.New("invalid dowel radius magnitude")
	}
	if !slices.Contains(dowelRadiusUnits, d.Units) {
		return errors.New("invalid dowel radius units")
	}
	return nil
}

type Radius struct {
	Size  float64    `bson:"size" json:"size"`
	Units LengthUnit `bson:"units" json:"units"`
}

type Wood string

var woods = []Wood{oak, poplar, bamboo}

const (
	poplar Wood = "Poplar"
	oak    Wood = "Oak"
	bamboo Wood = "Bamboo"
)

type LengthUnit string

var filterSizeUnits = []LengthUnit{um}          // TODO: use? can be 0.2 micron (not pc-able), 0.5 micron (pc-able), or 5 micron (pc-able)
var dowelRadiusUnits = []LengthUnit{mm, cm, in} // TODO: use?
// var dowelLengthUnits = []LengthUnit{mm, cm, in, ft} // TODO: use?
// TODO: use?
var lengthUnits = []LengthUnit{um, mm, cm, in, ft, meter} // TODO: use?
const (
	um    = "um"
	mm    = "mm"
	cm    = "cm"
	in    = "in"
	ft    = "ft"
	meter = "m"
)

//func (pl PlugsJar) setTransferParent(ctx context.Context, xfer Transfer) error {
//	// TODO: can this even occur?
//	coll := DbFrom(ctx).Collection(pl.CollectionName())
//	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
//	if err != nil {
//		return err
//	}
//	res, err := coll.UpdateByID(ctx, pl.Id, upd)
//	if err != nil {
//		return err
//	}
//	if res.ModifiedCount == 0 {
//		return ErrNoParentModifiedForTransfer
//	}
//	return nil
//}

func initializePlugs(ctx context.Context) error {
	// Indices
	coll := DbFrom(ctx).Collection(PlugsCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		creationDateIndexModel,
		//newSimpleIndex("parentType", "parentType", false, true, false), // TODO: nil is store or outside?
		//newSimpleIndex("parent", "parent", false, true, false),         // TODO: nil is store or outside?
		// TODO: combo dowel types index?
		//newSimpleIndex("dowelTypes", "dowelTypes.wood", false, false, false),
		//newSimpleIndex("dowelSizes", "dowelTypes.radius", false, false, false),
		newSimpleIndex("species", "species", false, true, false),
		newSimpleIndex("subspecies", "subspecies", false, true, false),
		//newSimpleIndex("innoc", "innoc", false, true, false),
		newSimpleIndex("pcRun", "pcRun", false, false, false),
		// TODO: ensure sales index is for each item!
		//newSimpleIndex("sales", "sales", false, true, false),
		//disposedIndexModel,
		//Notes (no index unless tags)
		projectsIndexModel,
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// Create test plugs
	testItem := &PlugsJar{
		MainCollectionIdField: MainCollectionIdField{exPlugId},
		ParentTypeField: ParentTypeField{
			utils.Pointer("plate"),
		},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&exPlate},
		CreationDateField:                 CreationDateField{exampleTime},
		DowelTypes: []Dowel{
			{
				Radius: Radius{
					Size:  2,
					Units: "cm",
				},
				Wood: "oak",
			},
		},
		SpeciesOptionalField:    SpeciesOptionalField{&testEntryStringId},
		SubspeciesOptionalField: SubspeciesOptionalField{&testEntryStringId},
		InnocField:              InnocField{&exAltId},
		PcRunOptionalField:      PcRunOptionalField{&exAltId},
		SalesField:              SalesField{[]AlternateCollectionId{exAltId}},
		DisposedField:           DisposedField{&exampleTime},
		NotesField:              NotesField{exampleNotes()},
		LastUpdatedField:        LastUpdatedField{exampleTime},
		AclField:                allCanReadAcl(nil),
	}
	return addTestMainEntries(ctx, testItem)
}

type createPlugsRequest struct {
	DowelTypes         []Dowel `json:"dowelTypes"`
	PcRunOptionalField         // OPTIONAL! // TODO: Can be created before pc!?
	NotesField
	WriteTagToField
}

func createPlugsHandler(w http.ResponseWriter, r *http.Request) { // TODO: fully test!
	ctx := r.Context()
	data := createPlugsRequest{}
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
	for i, d := range data.DowelTypes {
		if err = d.validate(); err != nil {
			errTxt := fmt.Sprintf("failed to validate dowel type for entry #%d", i)
			http.Error(w, errTxt+": "+err.Error(), http.StatusBadRequest)
			return
		}
	}
	if _, err = data.PcRunOptionalField.Get(ctx); err != nil {
		if !errors.Is(err, ErrMissingOptionalField) {
			http.Error(w, "invalid pc run field: "+err.Error(), http.StatusBadRequest)
			return
		}
	}
	// TODO: validate dowels
	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}

	ctx, now := request.UnixTime(r.Context()) // TODO: no more r.Context below
	toInsert := PlugsJar{
		MainCollectionIdField: MainCollectionIdField{id},
		CreationDateField:     CreationDateField{now},
		DowelTypes:            data.DowelTypes,
		PcRunOptionalField:    PcRunOptionalField{data.PcRun},
		NotesField:            NotesField{data.Notes},
		LastUpdatedField:      LastUpdatedField{now},
		// No Perms here for basic plugs
		AclField: allCanWriteAcl(),
	}

	finishCreateMainCollectionEntry(ctx, &toInsert, w)
}

type importPlugsRequest struct {
	DowelTypes []Dowel     `json:"dowelTypes"`
	Generation *Generation `json:"generation"` // TODO: what to do about nil?
	SpeciesOptionalField
	SubspeciesOptionalField
	KnownFruitableField
	NotesField
	WriteTagToField
	// TODO: perms should follow species/subspec if exists, otherwise all can write
}

func importPlugsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := importPlugsRequest{}
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
	// TODO: perms from spec/subspec+user if innoculated, otherwise allCanWrite
	toInsert := PlugsJar{
		MainCollectionIdField: MainCollectionIdField{id},
		//ParentTypeField:                   ParentTypeField{},
		//MainCollectionOptionalParentField: MainCollectionOptionalParentField{},
		CreationDateField: CreationDateField{now},
		DowelTypes:        data.DowelTypes,
		GenerationsFields: GenerationsFields{
			GenSporeField:        GenSporeField{data.Generation},
			GenSinceFruitOrSpore: data.Generation,
		},
		SpeciesOptionalField: SpeciesOptionalField{data.Species},
		//TransfersOutField:       TransfersOutField{},
		SubspeciesOptionalField: SubspeciesOptionalField{data.Subspecies},
		//InnocField:              InnocField{},
		KnownFruitableField: data.KnownFruitableField,
		//PcRunOptionalField:      PcRunOptionalField{nil},
		//SalesField:              SalesField{},
		//DisposedField:           DisposedField{},
		NotesField:       NotesField{data.Notes},
		LastUpdatedField: LastUpdatedField{now},
		// No Perms here for basic plugs
		AclField: allCanWriteAcl(), // TODO: add user and use species/subspecies perms
	}
	for i, d := range data.DowelTypes {
		if err = d.validate(); err != nil {
			errTxt := fmt.Sprintf("failed to validate dowel type for entry #%d", i)
			http.Error(w, errTxt+": "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	var finalPerms ACL
	if data.Species == nil {
		if data.Generation != nil {
			http.Error(w, "generation without species: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if data.Subspecies != nil {
			http.Error(w, "subspecies without species: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if data.KnownFruitable != nil {
			http.Error(w, "knownFruitable without species: "+err.Error(), http.StatusInternalServerError)
			return
		}
		finalPerms = allCanWriteAcl().ACL
	} else {
		finalPerms, err = ImportFinalPerms(r.Context(), *data.Species, data.Subspecies)
		if err != nil {
			http.Error(w, "failed to get species and/or subspecies: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	toInsert.ACL = finalPerms
	if err = data.Generation.validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	finishCreateMainCollectionEntry(ctx, &toInsert, w)
}

// TODO: new sale?

type updatePlugsRequest struct {
	PcRunOptionalField // Can only be set once!
	KnownFruitableField
	NotesUpdateField
	DisposedField
	PermsOnRequest `json:"acl"`
}

func (req updatePlugsRequest) modsFor(existing *PlugsJar, acl AclField) (bson.D, error) {
	mods := NewMods()
	return mods.
		updatePcRunIfNeeded(req, existing).
		updateNotesIfNeeded(req, existing).
		updateDisposedIfNeeded(req, existing).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updatePlugsHandler(w http.ResponseWriter, r *http.Request) {
	req := &updatePlugsRequest{}
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
	ctx := r.Context()
	existing := &PlugsJar{}
	db := DbFrom(ctx)
	err = db.Collection(PlugsCollectionName).FindOne(ctx, BsonFindFilter("_id", *mainCollId)).Decode(existing)
	if err != nil {
		http.Error(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.PcRun != nil {
		_, err = req.PcRunOptionalField.Get(ctx)
		if err != nil {
			http.Error(w, "invalid pc run: "+err.Error(), http.StatusBadRequest)
			return
		}
	}
	finishMainCollItemUpdate(ctx, w, req.modsFor, existing, req.PermsOnRequest)
}
