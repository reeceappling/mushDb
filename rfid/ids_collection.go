package rfid

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// centralized collection to map itemId to itemType
const idMapCollectionName = "itemIdMap"

type idMapEntry struct {
	Id        MainCollectionId `bson:"_id" json:"_id"`
	EntryType string           `bson:"entryType" json:"entryType"`
}

// TODO: can be bag, fruit, FC, jar, LC, LCSyr, MSS, plate, plugs, slant, sporePrint, sporeSwab, stasisTube

func initializeItemMapCollection(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(idMapCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		newSimpleIndex("entryType", "entryType", false, false, false),
	})
	return err
}

// TODO: USE
func getEntryTypeForId(ctx context.Context, id MainCollectionId) (string, error) {
	result := idMapEntry{}
	err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(idMapCollectionName).FindOne(ctx, bson.D{{"_id", id}}).Decode(&result)
	return result.EntryType, err
}

func addToIdMapCollection(ctx context.Context, item MainCollectionItem) (err error, rollback func() error) {
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(idMapCollectionName)
	id := item.DbId()
	_, err = coll.InsertOne(ctx, idMapEntry{
		Id:        id,
		EntryType: item.EntryType(),
	})
	if err != nil {
		return err, nil
	}
	return nil, func() error {
		return coll.FindOneAndDelete(ctx, bson.M{"_id": id}).Err()
	}
}
