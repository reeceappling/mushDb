package api

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
)

// centralized collection to map itemId to itemType
const idMapCollectionName = "itemIdMap"

type idMapEntry struct {
	Id        MainCollectionId `bson:"_id" json:"_id"`
	EntryType string           `bson:"entryType" json:"entryType"`
}

func initializeItemMapCollection(ctx context.Context) error { // TODO: USE!
	// Indices
	coll := DbFrom(ctx).Collection(idMapCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		newSimpleIndex("entryType", "entryType", false, false, false),
	})
	return err
}

// TODO: USE
func getEntryTypeForId(ctx context.Context, id MainCollectionId) (string, error) {
	result := idMapEntry{}
	err := DbFrom(ctx).Collection(idMapCollectionName).FindOne(ctx, BsonFindFilter(IDfld, id)).Decode(&result)
	return result.EntryType, err
}

func addToIdMapCollection(ctx mongo.SessionContext, item MainCollectionItem) (err error) {
	coll := mongo.SessionFromContext(ctx).Client().Database(dbName).Collection(idMapCollectionName)
	id := item.DbId()
	_, err = coll.InsertOne(ctx, idMapEntry{
		Id:        id,
		EntryType: item.EntryType(),
	})
	if err != nil {
		return err
	}
	return nil
}
