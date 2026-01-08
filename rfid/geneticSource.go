package rfid

import (
	"errors"
	"github.com/reeceappling/goUtils/v2/utils"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	// Alts
	// SporePrint
	_ geneticSource = Fruit{} // TODO: only clones via transfer
)

type GeneticParentInfo struct { // TODO: FIGURE OUT WHERE WE CAN USE THESE
	SpeciesOptionalField
	SubspeciesOptionalField
	KnownFruitableField
	GenerationsFields // TODO: NEW, FIX
	// TODO: used to be gensSinceSpore, now genSpore
	//TODO: used to be gensSinceFruitOrSpore, now genFruitOrSpore
}

var (
	_ geneticSource = &Bag{}
	_ geneticSource = &Fruit{}
	_ geneticSource = &FruitingChamber{}
	_ geneticSource = &GrainJar{}
	_ geneticSource = &LiquidCulture{}
	_ geneticSource = &LcSyringe{}
	_ geneticSource = &MSS{}
	_ geneticSource = &Plate{}
	_ geneticSource = &PlugsJar{}
	_ geneticSource = &Slant{}
	_ geneticSource = &SporePrint{} // TODO: sporeSwab?
	_ geneticSource = &SporeSwab{}  // TODO: sporeSwab?
	_ geneticSource = &StasisTube{}
)

type geneticSource interface {
	SourceType() string
	GeneticInfoAsParent() (GeneticParentInfo, error)
	DbId() BinaryCollectionId
	setTransferParent(ctx mongo.SessionContext, xfer Transfer) error
	setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error
	generation() (sinceSpore *Generation, sinceSporeOrClone *Generation)
	//Permissioned // TODO: get rid of?
	CanTransferTo(dst geneticSource) error
	Innoculatable() bool
}

func childGensForParent(parent geneticSource) (parentInfo GeneticParentInfo, genSpore, genFruitSpore *Generation, err error) {

	parentInfo, err = parent.GeneticInfoAsParent()
	if err != nil {
		return parentInfo, nil, nil, err
	}
	if parentInfo.Species == nil {
		return parentInfo, nil, nil, errors.New("parent must have a species")
	}
	switch parent.SourceType() {
	case MssSourceType:
		genSpore = utils.Pointer(Generation(0))
		genFruitSpore = utils.Pointer(Generation(0))
	case FruitSourceType:
		genSpore = parentInfo.GenSinceSpore.Next()
		genFruitSpore = utils.Pointer(Generation(0))
	default:
		genSpore = parentInfo.GenSinceSpore.Next()
		genFruitSpore = parentInfo.GenSinceFruitOrSpore.Next()
	}
	return
}
