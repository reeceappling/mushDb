package rfid

import (
	"context"
	"errors"
	"github.com/reeceappling/goUtils/v2/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"slices"
)

// sometimes needed for transfers

type PlugsJar struct {
	// TODO; any more?
	MainCollectionIdField             `bson:"inline"`
	ParentTypeField                   `bson:"inline"` // empty==bought
	MainCollectionOptionalParentField `bson:"inline"` // TODO: empty=bought, from plugs, jar, LC, plate/slant
	CreationDateField                 `bson:"inline"`
	DowelTypes                        []Dowel         `bson:"dowelTypes" json:"dowelTypes"`
	GenerationsFields                 `bson:"inline"` // TODO: NEW! fix!
	SpeciesOptionalField              `bson:"inline"`
	TransfersOutField                 `bson:"inline"` // TODO: NEW!!!!!
	SubspeciesOptionalField           `bson:"inline"`
	InnocField                        `bson:"inline"`
	KnownFruitableField               `bson:"inline"` // TODO: NEW! FIX!
	PcRunOptionalField                `bson:"inline"` // TODO: used to be required, but not found for imports! created before innoculation
	SalesField                        `bson:"inline"` // TODO: MULTIPLE!
	DisposedField                     `bson:"inline"` // Also changed once all pegs are sold/used?
	NotesField                        `bson:"inline"`
	LastUpdatedField                  `bson:"inline"`
	AclField                          `bson:"inline"`
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
		withSubspecies(parentInfo.SubSpecies).
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

type Dowel struct {
	Radius `bson:"inline"`
	Wood   Wood `bson:"wood" json:"wood"`
}
type Radius struct {
	Size  float64    `bson:"size" json:"size"`
	Units LengthUnit `bson:"units" json:"units"`
}

type Wood string

var woods = []Wood{oak, poplar, bamboo} // TODO: use?

const (
	poplar Wood = "Poplar"
	oak    Wood = "Oak"
	bamboo Wood = "Bamboo"
)

type LengthUnit string

var filterSizeUnits = []LengthUnit{um}              // TODO: use?
var dowelRadiusUnits = []LengthUnit{mm, cm, in}     // TODO: use?
var dowelLengthUnits = []LengthUnit{mm, cm, in, ft} // TODO: use?
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
//	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(pl.CollectionName())
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
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(PlugsCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		// TODO: which indices are needed?
		//newSimpleIndex("parentType", "parentType", false, true, false), // TODO: nil is store or outside?
		//newSimpleIndex("parent", "parent", false, true, false),         // TODO: nil is store or outside?
		creationDateIndexModel,
		// TODO: DOWEL TYPES
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
		CreationDateField:                 exampleTime.asCreationDate(),
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
		AclField:                allCanReadAcl(),
	}
	return addTestMainEntries(ctx, testItem)
}

// TODO: create new plugs request????
type createPlugsRequest struct { // TODO: USE THIS!
	DowelTypes         []Dowel `bson:"dowelTypes" json:"dowelTypes"`
	PcRunOptionalField         // TODO: OPTIONAL!
	NotesField
	WriteTagToField
}

func createPlugsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: DO THIS WHOLE THING!
}

// TODO: import plugs request!
type importPlugsRequest struct { // TODO: USE THIS!
	DowelTypes []Dowel `bson:"dowelTypes" json:"dowelTypes"` // TODO: ok?
	Generation
	SpeciesOptionalField    `bson:"inline"` // TODO: ok?
	SubspeciesOptionalField `bson:"inline"` // TODO: ok?
	KnownFruitableField     `bson:"inline"` // TODO: ok?
	NotesField              `bson:"inline"` // TODO: ok?
	WriteTagToField
}

func importPlugsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: DO THIS WHOLE THING!
}

// TODO: new sale?

type updatePlugsRequest struct {
	PcRunOptionalField // Can only be set once!
	NotesUpdateField
	PermsOnRequest
	WriteTagToField
	DisposedField
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
	id := *mainCollId
	err = writeRfidTagIfNecessary(r.Context(), req.WriteTagTo, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	existing := &PlugsJar{}
	client := ctx.Value(mongoClientContextKey).(*mongo.Client)
	coll := client.Database(dbName).Collection(PlugsCollectionName)
	err = coll.FindOne(ctx, bsonFindFilter("_id", id)).Decode(existing)
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
	finishMainCollItemUpdate(ctx, w, coll, req.modsFor, existing, req.PermsOnRequest)
}
