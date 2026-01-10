package rfid

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
	"slices"
)

// TODO: this whole thing whenever we need to

// TODO: IMPORTS!
// TODO: CREATES!
// TODO: UPDATES!

const (
	PlugsCollectionName = "plugs" // TODO: USE
	PlugSourceType      = "plug"
	plugsCollectionName = "plugs"
	plugIdPrefix        = "pl"
)

type PlugsJar struct { // TODO: do this whole file! This should be an alt, not a main, due to multi-sales
	// TODO; any more?
	MainCollectionIdField             // TODO: was alt
	ParentTypeField                   // TODO: get rid of these?           // None, plugs, jars, lcSyringe, plate, slant, etc (both alt and main
	MainCollectionOptionalParentField // TODO: empty=bought, from plugs, jar, LC, plate/slant
	CreationDateField
	DowelTypes []Dowel `bson:"dowelTypes" json:"dowelTypes"`
	SpeciesOptionalField
	SubspeciesOptionalField
	InnocField
	PcRunField // TODO: created before innoculation
	SalesField
	DisposedField // Also changed once all pegs are sold/used?
	NotesField
	LastUpdatedField
	AclField // TODO: handle EVERYWHERE

	// TODO: this whole thing whenever we need to
}

func (pl PlugsJar) CanTransferTo(dst geneticSource) error {
	if !slices.Contains([]string{BagSourceType, GrainJarSourceType, LcSourceType, PlateSourceType, PlugSourceType, SlantSourceType, StasisTubeSourceType}, dst.SourceType()) {
		return errors.New("plugs cannot transfer to " + dst.SourceType())
	}
	return nil
}

func (pl PlugsJar) SourceType() string {
	return PlugSourceType
}

func (pl PlugsJar) GeneticInfoAsParent() (GeneticParentInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (pl PlugsJar) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
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

func (pl PlugsJar) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := pl
	err := decodeItem(&out, encoded) // TODO: likely wont work
	return out, err
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

func (pl PlugsJar) CollectionName() string {
	return PlugsCollectionName
}

func (pl PlugsJar) setTransferParent(ctx mongo.SessionContext, xfer Transfer) error {
	// TODO: can this even occur?
	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
	if err != nil {
		return err
	}
	res, err := ctx.Client().Database(dbName).Collection(pl.CollectionName()).UpdateByID(ctx, pl.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer
	}
	return nil
}

func initializePlugs(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(plugsCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		// TODO: which indices are needed?
		newSimpleIndex("parentType", "parentType", false, true, false), // TODO: nil is store or outside?
		newSimpleIndex("parent", "parent", false, true, false),         // TODO: nil is store or outside?
		creationDateIndexModel,
		// TODO: DOWEL TYPES
		newSimpleIndex("dowelTypes", "dowelTypes.wood", false, false, false),
		newSimpleIndex("dowelSizes", "dowelTypes.radius", false, false, false),
		newSimpleIndex("species", "species", false, true, false),
		newSimpleIndex("subSpecies", "subSpecies", false, true, false),
		newSimpleIndex("innoc", "innoc", false, true, false),
		newSimpleIndex("pcRun", "pcRun", false, false, false),
		// TODO: ensure sales index is for each item!
		newSimpleIndex("sales", "sales", false, true, false),
		disposedIndexModel,
		// TODO: PROJECTS
		//Notes (no index unless tags)
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	existingEntry := PlugsJar{}
	testItem := PlugsJar{
		MainCollectionIdField: MainCollectionIdField{exPlugId},
		ParentTypeField:       ParentTypeField{
			// TODO; FIX!
		},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&exPlate},
		CreationDateField:                 exampleTime.asCreationDate(),
		DowelTypes:                        nil, // TODO; FIX
		SpeciesOptionalField:              SpeciesOptionalField{&testEntryStringId},
		SubspeciesOptionalField:           SubspeciesOptionalField{&testEntryStringId},
		InnocField:                        InnocField{}, // TODO; FIX
		PcRunField:                        PcRunField{}, // TODO; FIX
		SalesField:                        SalesField{[]AlternateCollectionId{exAltId}},
		DisposedField:                     DisposedField{&exampleTime},
		NotesField:                        NotesField{exampleNotes()},
		LastUpdatedField:                  LastUpdatedField{exampleTime},
		//PermsField:                        PermsField{}, // TODO; FIX
	}
	err = coll.FindOne(ctx, bson.D{{"_id", exAltId}}).Decode(&existingEntry)
	if err == nil {
		if reflect.DeepEqual(existingEntry, testItem) {
			return nil
		}
	}
	return testExistingEntry(ctx, coll, exAltId, testItem, existingEntry)
}
