package rfid

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// TODO: centralized collection to map itemId to itemType
const idMapCollectionName = "itemIdMap"

// Agar recipe to agar batch
// Agar batch to plate/slant
// slant/plate to slant/plate/jar/bag/lc/box/stasis
// bag to bag/plate/fruit/jar/box
// fruit to bag/jar/plate/box/lc/stasis/mss
// box to fruit/plate
// jar to jar/plate
// jar recipe to jar
// lc recipe to lc
// lc to plate,jar,stasis?,slant,box,bag
// pc run to agar batch?lc?
// project to ????????????
// sale to ?????????????
// species to ????????????
// subspecies to ?????????????
// substrate recipe to

// TODO: make all updates single requests

type idMappable interface {
}

type idMapEntry struct {
	Id        string `bson:"_id" json:"_id"`
	EntryType string `bson:"entryType" json:"entryType"`
}

// TODO: can be bag, fruit, FC, jar, LC, LCSyr, MSS, plate, plugs, slant, sporePrint, sporeSwab, stasisTube

func initializeItemMapCollection(ctx context.Context) error {
	// TODO: separate mainCollectionItems into their own collections?
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(idMapCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		//newSimpleIndex("entryType", "entryType", false, false, false),
	})
	return err
}

// TODO: USE
func newEntryInTypeMap(ctx context.Context, id BinaryCollectionId, entryType string) error {
	// TODO: ensure entry with this ID does not already exist
	_, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(idMapCollectionName).InsertOne(ctx, idMapEntry{
		Id:        string(id),
		EntryType: entryType,
	})
	return err
}

// TODO: USE
func getEntryTypeForId(ctx context.Context, id BinaryCollectionId) (string, error) {
	// TODO: separate mainCollectionItems into their own collections?
	result := &idMapEntry{}
	err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(idMapCollectionName).FindOne(ctx, bson.D{{"_id", id}}).Decode(&result)
	if err != nil {
		// mongo.ErrNoDocuments or something else
		return "", err
	}
	return result.EntryType, nil
}

func addToIdMapCollectionInTxn(ctx mongo.SessionContext, id BinaryCollectionId, entryType string) error { // TODO: MOVE
	_, err := ctx.Client().Database(dbName).Collection(idMapCollectionName).InsertOne(ctx, idMapEntry{
		Id:        string(id),
		EntryType: entryType,
	})
	return err
}
