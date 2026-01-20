package rfid

import (
	"context"
	"errors"
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

var mainCollectionEntryTypes = []string{
	LcSourceType,
	GrainJarSourceType,
	PlateSourceType,
	SlantSourceType,
	StasisTubeSourceType,
	BagSourceType,
	FruitingChamberSourceType,
	MssSourceType,
}

//func simpleInsertMainColl(ctx context.Context, item interface{}) (*MainCollectionId, error) {
//	out, err := ctx.Value(mongoClientContextKey).(*mongo.Client).
//		Database(dbName).
//		Collection(mainCollectionName).
//		InsertOne(ctx, item)
//	if err != nil {
//		return nil, err
//	}
//	res, ok := out.InsertedID.(MainCollectionId)
//	if !ok {
//		return nil, errors.New("failed to convert to primitive.ObjectID")
//	}
//
//	return &res, nil
//}

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

func initializeChildrenMethod(ctx context.Context) ([]geneticSource, *mongo.Database, *mongo.Collection) {
	out := []geneticSource{}
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	xferColl := db.Collection(TransfersCollName)
	return out, db, xferColl
}

func isMainCollItem(entryType string) bool {
	_, exists := mainCollMap(entryType)
	return exists
}

//func rawEntryTypeConversion(raw bson.Raw) (MainCollectionItem, error) {
//	entryType := raw.Lookup("entryType").String()
//	child, exists := mainCollMap(raw.Lookup("entryType").String())// TODO: unsure if this is correct
//	if !exists {
//		return nil, errors.Join(errors.New("invalid item entryType"), fmt.Errorf(`entryType: %s`, entryType))
//	}
//	return child, nil
//}

//func childrenOnlyToPlate(ctx context.Context, xferIds []AlternateCollectionId) ([]geneticSource, error) {
//	out, db, xferColl := initializeChildrenMethod(ctx)
//	mainColl := db.Collection(mainCollectionName)
//	for _, xferId := range xferIds {
//		xfer, err := getTransferById(ctx, xferColl, xferId)
//		if err != nil {
//			return nil, err
//		}
//		var dish Plate
//		item := mainColl.FindOne(ctx, bson.D{{"_id", xfer.To}})
//		if err = item.Err(); err != nil {
//			return nil, errors.Join(errors.New("failed to find plate by id"), err)
//		}
//		if err = item.Decode(&dish); err != nil {
//			return nil, errors.Join(errors.New("failed to decode plate"), err)
//		}
//		out = append(out, dish)
//	}
//	return out, nil
//}

//func childrenAreOnlyFruits(ctx context.Context, xferIds []AlternateCollectionId) ([]geneticSource, error) {
//	panic("not implemented")
//}

//func initializeMainCollection(ctx context.Context) error {
//	// Indices
//	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(mainCollectionName)
//	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
//		newSimpleIndex("entryType", "entryType", false, false, false),
//		newSimpleIndex("creationDate", "creationDate", true, false, false),
//		newSimpleIndex("agar", "agar", false, true, false),     // Plate, slant
//		newSimpleIndex("pcRun", "pcRun", false, true, false),   // TODO: only on _
//		newSimpleIndex("recipe", "recipe", false, true, false), // TODO: only on _
//		newSimpleIndex("species", "species", false, true, false),
//		newSimpleIndex("subspecies", "subspecies", false, true, false),
//		// filterSize (no index) (BAG ONLY)
//		newSimpleIndex("sealDate", "sealDate", true, true, false), // BAG ONLY
//		// flushes (no index) (BAG AND BOX ONLY)
//		newSimpleIndex("innoc", "innoc", false, true, false),
//		newSimpleIndex("genSinceSpore", "genSpore", true, true, false),
//		newSimpleIndex("genSinceFruitOrSpore", "genFruitOrSpore", true, true, false),
//		newSimpleIndex("transfersOut", "transfersOut", false, true, false),
//		newSimpleIndex("parent", "parent", false, true, false),
//		newSimpleIndex("parentType", "parentType", false, true, false),
//		// pics (no index)
//		// confirmedClean (no index) (LC Only)
//		// Contaminations (no index)
//		newSimpleIndex("knownFruitable", "knownFruitable", false, true, false),
//		// disposed (sparse)
//		newSimpleIndex("sale", "sale", true, true, false),   // All but LC
//		newSimpleIndex("sales", "sales", true, true, false), // LC Only
//		newSimpleIndex("disposed", "disposed", true, true, false),
//		projectsIndexModel,
//		// mostRecentImage (no index)
//		//Notes (no index unless tags)
//		lastUpdatedIndexModel,
//	})
//	return err
//}
