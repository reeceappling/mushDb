package rfid

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
)

// TODO: this whole thing whenever we need to

const (
	PlugSourceType      = "plug"
	plugsCollectionName = "plugs"
	plugIdPrefix        = "pl"
)

type Plugs struct { // TODO: do this whole file! This should be an alt, not a main, due to multi-sales
	// TODO; any more?
	AlternateCollectionIdField
	ParentTypeField           // None, plugs, jars, lcSyringe, plate, slant, etc (both alt and main
	BinaryOptionalParentField // TODO: empty=bought, from plugs, jar, LC, plate/slant
	CreationDateField
	DowelTypes []Dowel `bson:"dowelTypes" json:"dowelTypes"`
	SpeciesOptionalField
	SubspeciesOptionalField
	InnocField
	PcRunField // TODO: created before innoculation
	SalesField
	DisposedField
	NotesField
	LastUpdatedField
	PermsField

	// TODO: this whole thing whenever we need to
}

func (pl Plugs) EntryTypeField() *string {
	return nil
}

func (pl Plugs) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
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

func (pl Plugs) CollectionName() string {
	return plugsCollectionName
}

func (pl Plugs) projects() []projectName {
	return pl.Perms.Projects.Ids
}

func (pl Plugs) setTransferParent(ctx mongo.SessionContext, xfer Transfer) error {
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
		//newSimpleIndex("parent", "parent", false, false, false),
		//newSimpleIndex("creationDate", "creationDate", true, false, false), // TODO: INDEX CREATION DATES EVERYWHERE!
		//newSimpleIndex("species", "species", false, false, false),
		//newSimpleIndex("subSpecies", "subSpecies", false, true, false),
		projectsIndexModel,
		//saleIndexModel,
		//Notes (no index unless tags)
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	existingEntry := Plugs{}
	testItem := Plugs{
		AlternateCollectionIdField: AlternateCollectionIdField{exAltId},
		ParentTypeField:            ParentTypeField{
			// TODO; FIX!
		},
		BinaryOptionalParentField: BinaryOptionalParentField{}, // TODO; FIX
		CreationDateField:         exampleTime.asCreationDate(),
		DowelTypes:                nil, // TODO; FIX
		SpeciesOptionalField:      SpeciesOptionalField{&testEntryStringId},
		SubspeciesOptionalField:   SubspeciesOptionalField{&testEntryStringId},
		InnocField:                InnocField{}, // TODO; FIX
		PcRunField:                PcRunField{}, // TODO; FIX
		SalesField:                SalesField{[]AlternateCollectionId{exAltId}},
		DisposedField:             DisposedField{&exampleTime},
		NotesField:                NotesField{exampleNotes()},
		LastUpdatedField:          LastUpdatedField{exampleTime},
		PermsField:                PermsField{}, // TODO; FIX
	}
	err = coll.FindOne(ctx, bson.D{{"_id", exAltId}}).Decode(&existingEntry)
	if err == nil {
		if reflect.DeepEqual(existingEntry, testItem) {
			return nil
		}
	}
	return testExistingEntry(ctx, coll, exAltId, testItem, existingEntry)
}
