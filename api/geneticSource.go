package api

import (
	"errors"
	"github.com/reeceappling/goUtils/v2/utils"
	"github.com/reeceappling/mushDb/api/request"
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
	_ geneticSource = &SporePrint{}
	_ geneticSource = &SporeSwab{}
	_ geneticSource = &StasisTube{}
)

type geneticSource interface {
	SourceType() string
	GeneticInfoAsParent() (GeneticParentInfo, error)
	DbId() MainCollectionId
	setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error
	generation() (sinceSpore *Generation, sinceSporeOrClone *Generation)
	Permissioned
	CanTransferTo(dst geneticSource) error
	Innoculatable() error
	CollectionItem
	Disposable
	SetPerms(AclField) // MUST be a pointer reciever
}

func setTransferParent[T MainCollectionItem](ctx mongo.SessionContext, parent T, xfer Transfer, dispose bool) error {
	coll := mongo.SessionFromContext(ctx).Client().Database(dbName).Collection(parent.CollectionName())
	ctx, now := request.UnixTimeInTxn(ctx)
	var temp any = parent // TODO: will this actually work with geneticSource rather than MainCollectionItem?
	mods := &Mods{}
	if xfer.FromImage != nil {
		if parentWithPics, ok := temp.(HasPicsField); ok {
			mods = xfer.PicsModsForParent(parentWithPics) // TODO: VALIDATE WORKS!
		}
	}

	mods.addTransferOut(xfer.Id)
	doDispose := dispose || parent.SourceType() == StasisTubeSourceType // TODO: Validate stasis tube source type ok here
	if doDispose {
		mods = mods.updateDisposedIfNeeded(DisposedField{Disposed: &now}, parent)
	}

	upd, err := mods.updateLastUpdatedIfNeeded().Finalized()
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
