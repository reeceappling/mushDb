package rfid

import (
	"context"
	"errors"
	sliceutils "github.com/reeceappling/goUtils/v2/utils/slices"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	_ MainCollectionItem = &LiquidCulture{}   // can go anywhere (in theory) except MSS
	_ MainCollectionItem = &GrainJar{}        // can go anywhere (in theory) except MSS
	_ MainCollectionItem = &Plate{}           // can go anywhere (in theory) except MSS
	_ MainCollectionItem = &Slant{}           // generally only goes to plate
	_ MainCollectionItem = &StasisTube{}      // generally only goes to plate
	_ MainCollectionItem = &Bag{}             // can only go to fruits
	_ MainCollectionItem = &FruitingChamber{} // can only go to fruits
	_ MainCollectionItem = &MSS{}             // generally only goes to plate
	_ MainCollectionItem = &SporePrint{}
	_ MainCollectionItem = &LcSyringe{}
	_ MainCollectionItem = &PlugsJar{} // TODO: has multiple sales
	_ MainCollectionItem = &SporeSwab{}
	_ MainCollectionItem = &WaterJar{}
)

type CollectionId interface { // TODO; USE????
	MainCollectionId | AlternateCollectionId // TODO: DELETEME
}

type MainCollectionItem interface {
	CollectionItem
	geneticSource
	EntryType() string
	Permissioned
}

func GetMainCollectionItemWithId(ctx context.Context, id MainCollectionId) (MainCollectionItem, error) {
	itemType, err := FindItemTypeForId(ctx, id)
	if err != nil {
		return nil, err
	}
	item, err := GetMainCollectionItem(ctx, id, itemType)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			err = errors.Join(errors.New("should never happen. Id table invalid"), err)
		}
		return nil, err
	}
	return item, nil
}

func addTestMainEntries[T MainCollectionItem](ctx context.Context, testItems ...T) error {
	if len(testItems) == 0 {
		return errors.New("testItems is empty for main collection")
	}
	_, txErr := newTxn(ctx, func(sessCtx mongo.SessionContext) (any, error) {
		db := mongo.SessionFromContext(sessCtx).Client().Database(dbName)
		_, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).
			Collection(idMapCollectionName).BulkWrite(ctx, sliceutils.Map(testItems, func(item T) mongo.WriteModel {
			return mongo.NewReplaceOneModel().SetReplacement(idMapEntry{
				Id:        item.DbId(),
				EntryType: item.EntryType(),
			}).SetFilter(bson.D{{"_id", item.DbId()}}).SetUpsert(true)
		}))
		if err != nil {
			return nil, errors.Join(errors.New("failed to bulk write id maps"), err)
		}
		// TODO: do something with the result?
		_, err = db.Collection(testItems[0].CollectionName()).
			BulkWrite(ctx, sliceutils.Map(testItems, func(item T) mongo.WriteModel {
				return mongo.NewReplaceOneModel().
					SetReplacement(item).
					SetFilter(bson.M{"_id": item.DbId()}).
					SetUpsert(true)
			}))
		if err != nil {
			return nil, errors.Join(errors.New("failed to bulk write"), err)
		}
		// TODO: do something with the result?
		return nil, nil
	})
	PrintMainCollectionItemIds("Built-in", testItems)
	return txErr
}

func getTransferById(ctx context.Context, xferColl *mongo.Collection, id AlternateCollectionId) (*Transfer, error) {
	var xfer Transfer
	out := &xfer
	xferResult := xferColl.FindOne(ctx, bson.D{{"_id", id}})
	if err := xferResult.Err(); err != nil {
		return nil, errors.Join(errors.New("failed to retrieve transfer by id"), err)
	}
	if err := xferResult.Decode(out); err != nil {
		return nil, errors.Join(errors.New("failed to decode transfer result"), err)
	}
	return out, nil
}
