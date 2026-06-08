package api

import (
	"errors"
	"github.com/reeceappling/goUtils/v2/utils"
	"go.mongodb.org/mongo-driver/mongo"
)

type GeneticParentInfo struct {
	SpeciesOptionalField
	SubspeciesOptionalField
	KnownFruitableField
	GenerationsFields
}

//func (genetics GeneticParentInfo) GetSpeciesSubspecies(ctx context.Context) (*Species, *Subspecies, error) {
//	if genetics.Species != nil {
//		return nil, nil, errors.New("no species present on entry")
//	}
//	sp, subsp, err := getSpeciesAndSubspecies(ctx, *genetics.Species, genetics.Subspecies)
//	if err != nil {
//		return nil, nil, err
//	}
//	return &sp, subsp, err
//}

var ( // TODO: all used to be non-pointer. Ensure they all still work
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
	_ geneticSource = &SporePrint{}
	_ geneticSource = &SporeSwab{}
	_ geneticSource = &StasisTube{}
)

type geneticSource interface {
	SourceType() string
	GeneticInfoAsParent() (GeneticParentInfo, error)
	DbId() MainCollectionId
	//setTransferParent(ctx context.Context, xfer Transfer) error
	setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error
	generation() (sinceSpore *Generation, sinceSporeOrClone *Generation)
	Permissioned
	CanTransferTo(dst geneticSource) error
	Innoculatable() bool
	CollectionItem
	SetPerms(AclField) // MUST be a pointer reciever
}

func setTransferParent(ctx mongo.SessionContext, parent geneticSource, xfer Transfer) error {
	coll := mongo.SessionFromContext(ctx).Client().Database(dbName).Collection(parent.CollectionName())
	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
	// TODO: if transfer has a fromPic on it, can we add it to the parent?
	if err != nil {
		return err
	}
	res, err := coll.UpdateByID(ctx, parent.DbId(), upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer
	}
	return nil
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
