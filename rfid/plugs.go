package rfid

import (
	"context"
	"errors"
	"github.com/reeceappling/goUtils/v2/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"slices"
)

// TODO: this whole thing whenever we need to

// TODO: IMPORTS!
// TODO: CREATES!
// TODO: UPDATES!

type PlugsJar struct { // TODO: do this whole file! This should be an alt, not a main, due to multi-sales
	// TODO; any more?
	MainCollectionIdField             `bson:"inline"` // TODO: was alt
	ParentTypeField                   `bson:"inline"` // TODO: get rid of these?           // None, plugs, jars, lcSyringe, plate, slant, etc (both alt and main
	MainCollectionOptionalParentField `bson:"inline"` // TODO: empty=bought, from plugs, jar, LC, plate/slant
	CreationDateField                 `bson:"inline"`
	DowelTypes                        []Dowel `bson:"dowelTypes" json:"dowelTypes"`
	SpeciesOptionalField              `bson:"inline"`
	SubspeciesOptionalField           `bson:"inline"`
	InnocField                        `bson:"inline"`
	PcRunField                        `bson:"inline"` // TODO: created before innoculation
	SalesField                        `bson:"inline"`
	DisposedField                     `bson:"inline"` // Also changed once all pegs are sold/used?
	NotesField                        `bson:"inline"`
	LastUpdatedField                  `bson:"inline"`
	AclField                          `bson:"inline"`

	// TODO: this whole thing whenever we need to
}

func (pl PlugsJar) CanTransferTo(dst geneticSource) error {
	if !slices.Contains([]string{BagSourceType, GrainJarSourceType, LcSourceType, PlateSourceType, PlugSourceType, SlantSourceType, StasisTubeSourceType}, dst.SourceType()) {
		return errors.New("plugs cannot transfer to " + dst.SourceType())
	}
	return nil
}

func (pl PlugsJar) GeneticInfoAsParent() (GeneticParentInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (pl PlugsJar) setTransferChild(ctx context.Context, xfer Transfer, from geneticSource) error {
	// TODO: MUST BE AGAR OR BAG
	//TODO implement me
	panic("implement me")
}

func (pl PlugsJar) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	//TODO implement me
	panic("implement me")
}

func (pl PlugsJar) EntryTypeField() *string {
	return nil
}

type Dowel struct { // TODO: use?
	Radius struct {
		Size  float64
		Units LengthUnit
	}
	Wood Wood
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

func (pl PlugsJar) setTransferParent(ctx context.Context, xfer Transfer) (error, func() error) {
	// TODO: can this even occur?
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(pl.CollectionName())
	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
	if err != nil {
		return err, nil
	}
	res, err := coll.UpdateByID(ctx, pl.Id, upd)
	if err != nil {
		return err, nil
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer, nil
	}
	return nil, func() error {
		return coll.FindOneAndReplace(ctx, bson.D{{"_id", pl.Id}}, pl).Err()
	}
}

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
	// If test agar batch does not exist, then create it
	testItem := &PlugsJar{
		MainCollectionIdField: MainCollectionIdField{exPlugId},
		ParentTypeField: ParentTypeField{
			utils.Pointer("plate"),
		},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&exPlate},
		CreationDateField:                 exampleTime.asCreationDate(),
		DowelTypes:                        nil, // TODO; FIX
		SpeciesOptionalField:              SpeciesOptionalField{&testEntryStringId},
		SubspeciesOptionalField:           SubspeciesOptionalField{&testEntryStringId},
		InnocField:                        InnocField{&exAltId},
		PcRunField:                        PcRunField{exAltId},
		SalesField:                        SalesField{[]AlternateCollectionId{exAltId}},
		DisposedField:                     DisposedField{&exampleTime},
		NotesField:                        NotesField{exampleNotes()},
		LastUpdatedField:                  LastUpdatedField{exampleTime},
		AclField:                          allCanReadAcl(),
	}
	return addTestMainEntries(ctx, testItem)
}
