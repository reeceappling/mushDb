package rfid

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"testing"
)

func TestCommon(t *testing.T) {
	ctx := ctxWithClient
	require.NoError(t, Initialize(ctx))
	client := ctx.Value(mongoClientContextKey).(*mongo.Client)
	db := client.Database(dbName)

	t.Run("Default collections should exist", func(t *testing.T) {
		colls, err := db.ListCollectionNames(ctx, bson.D{})
		assert.NoError(t, err)
		assert.Equal(t, 12, len(colls))
		assert.Contains(t, colls, PlatesCollectionName)
		assert.Contains(t, colls, PcRunCollectionName)
	})

	t.Run("Built-in items", func(t *testing.T) {
		t.Run("Recipes", func(t *testing.T) {
			// TODO: this
		})
		t.Run("Sub/Species", func(t *testing.T) {
			spec := Species{}
			amt, err := db.Collection(SpeciesCollectionName).CountDocuments(ctx, bson.D{}) //.FindOne(ctx, bson.D{{"_id", "maitake"}})
			assert.NoError(t, err)
			assert.NotEqual(t, int64(0), amt)

			res := db.Collection(SpeciesCollectionName).FindOne(ctx, bson.D{{"_id", "maitake"}})
			assert.NotNil(t, res)
			assert.NoError(t, res.Err())
			assert.NoError(t, res.Decode(&spec))
			assert.Equal(t, "maitake", spec.Name)
			assert.Equal(t, "Grifola frondosa", spec.ScientificName)
			// TODO: subspecies
		})

		// TODO: ensure WE DONT DO UPDATING HERE. SHOULD BE DONE VIA WEBAPP
	})

	//t.Run("DB helper functions", func(t *testing.T) {
	//	t.Run("pushToArray", func(t *testing.T) {
	//		coll := db.Collection(mainCollectionName)
	//      id := NextMainCollectionId()
	//		//id, err := newMainCollectionId(ctx)
	//		//assert.NoError(t, err)
	//		now := time.Now()
	//		before := now.AddDate(0, 0, -1)
	//		nowUnix := unixTimeFor(now)
	//		_, err = coll.InsertOne(ctx, Plate{
	//			EntryTypeStructField:  EntryTypeStructField{"plate"},
	//			MainCollectionIdField: MainCollectionIdField{id},
	//			Notes: []Note{
	//				{Time: unixTimeFor(before), Note: "preexisting"},
	//			},
	//		})
	//		assert.NoError(t, err)
	//		noteToAdd := Note{Time: nowUnix, Note: "NEW NOTE!"}
	//		notesToAdd := []Note{
	//			{Time: nowUnix, Note: "NEW NOTE 2"},
	//			{Time: nowUnix, Note: "NEW NOTE 3"},
	//		}
	//		idBson := bson.D{{"_id", id}}
	//		var resA, resB, resC Plate
	//		ress := coll.FindOneAndUpdate(ctx, idBson, pushToArray("notes", noteToAdd))
	//		assert.NoError(t, ress.Decode(&resA))
	//		assert.Equal(t, 1, len(resA.Notes))
	//		resss := coll.FindOneAndUpdate(ctx, idBson, pushToArray("notes", notesToAdd...))
	//		assert.NoError(t, resss.Decode(&resB))
	//		assert.Equal(t, 2, len(resB.Notes))
	//		finRes := coll.FindOne(ctx, idBson)
	//		assert.NoError(t, finRes.Decode(&resC))
	//		assert.Equal(t, 4, len(resC.Notes))
	//		t.Run("removeFromArray", func(t *testing.T) {
	//			ress = coll.FindOneAndUpdate(ctx, idBson, withItemsRemoved("notes", noteToAdd))
	//			assert.NoError(t, ress.Decode(&resB))
	//			rem := coll.FindOne(ctx, idBson)
	//			assert.NoError(t, rem.Decode(&resC))
	//			assert.Equal(t, 3, len(resC.Notes))
	//			ress = coll.FindOneAndUpdate(ctx, idBson, withItemsRemoved("notes", notesToAdd...))
	//			assert.NoError(t, ress.Decode(&resB))
	//			rem = coll.FindOne(ctx, idBson)
	//			assert.NoError(t, rem.Decode(&resC))
	//			assert.Equal(t, 1, len(resC.Notes))
	//		})
	//	})
	//})
}
