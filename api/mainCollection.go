package api

import (
	"context"
	"encoding/json"
	"errors"
	sliceutils "github.com/reeceappling/goUtils/v2/utils/slices"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"sync"
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
	_ MainCollectionItem = &PlugsJar{} // has multiple sales...
	_ MainCollectionItem = &SporeSwab{}
	_ MainCollectionItem = &WaterJar{}
)

type CollectionId interface {
	MainCollectionId | AlternateCollectionId
	AsBase58() Base58Str
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
	wg := &sync.WaitGroup{}
	wg.Add(1)
	_, txErr := newTxn(ctx, func(sessCtx mongo.SessionContext) (any, error) {
		defer wg.Done()
		db := mongo.SessionFromContext(sessCtx).Client().Database(dbName)
		_, err := DbFrom(ctx).
			Collection(idMapCollectionName).BulkWrite(ctx, sliceutils.Map(testItems, func(item T) mongo.WriteModel {
			return mongo.NewReplaceOneModel().SetReplacement(idMapEntry{
				Id:        item.DbId(),
				EntryType: item.EntryType(),
			}).SetFilter(BsonFindFilter(IDfld, item.DbId())).SetUpsert(true)
		}))
		if err != nil {
			return nil, errors.Join(errors.New("failed to bulk write id maps"), err)
		}
		_, err = db.Collection(testItems[0].CollectionName()).
			BulkWrite(ctx, sliceutils.Map(testItems, func(item T) mongo.WriteModel {
				return mongo.NewReplaceOneModel().
					SetReplacement(item).
					SetFilter(bson.M{IDfld: item.DbId()}).
					SetUpsert(true)
			}))
		if err != nil {
			return nil, errors.Join(errors.New("failed to bulk write"), err)
		}
		return nil, nil
	})
	wg.Wait()
	if testItems[0].CollectionName() != PlatesCollectionName {
		PrintMainCollectionItemIds("Built-in", testItems)
	}
	return txErr
}

func getTransferById(ctx context.Context, xferColl *mongo.Collection, id AlternateCollectionId) (*Transfer, error) {
	var xfer Transfer
	out := &xfer
	xferResult := xferColl.FindOne(ctx, BsonFindFilter(IDfld, id))
	if err := xferResult.Err(); err != nil {
		return nil, errors.Join(errors.New("failed to retrieve transfer by id"), err)
	}
	if err := xferResult.Decode(out); err != nil {
		return nil, errors.Join(errors.New("failed to decode transfer result"), err)
	}
	return out, nil
}

type ImagesUpdateField struct {
	Images SplitEntries[picWithNotesForm, PicWithNotesLessLocation] `json:"images"` //"newPic-1"
}
type ContamsUpdateField struct {
	Contams SplitEntries[contamForm, ContaminationLessLocation] `json:"contams"` //"newContam-1"
}
type FlushesUpdateField struct {
	Flushes SplitEntries[picWithNotesForm, PicWithNotesLessLocation] `json:"flushes"` //"newFlush-1"
}

func mainCollIdFromRequest(r *http.Request, w http.ResponseWriter) (b58id Base58Str, id MainCollectionId, err error) {
	var idStr string
	idStr, err = UrlDecodeString(r.PathValue("id"))
	if err != nil {
		http.Error(w, "failed to url decode string: "+err.Error(), http.StatusInternalServerError)
		return
	}
	mainCollId, err := StandardizeMainCollectionId(idStr)
	if err != nil {
		http.Error(w, "failed to standardize main collection id: "+err.Error(), http.StatusBadRequest)
		return
	}
	b58id, id = mainCollId.AsBase58(), *mainCollId
	return
}

func finishCreateMainCollectionEntry(ctx context.Context, toInsert MainCollectionItem, w http.ResponseWriter) {
	_, err := newTxn(ctx, func(sessCtx mongo.SessionContext) (any, error) {
		return nil, createMainCollectionEntryInTxn(sessCtx, toInsert)
	})
	if err != nil {
		http.Error(w, "failed to create main collection entry in txn:"+err.Error(), http.StatusInternalServerError)
		return
	}

	bsOut, err := json.Marshal(toInsert)
	if err != nil {
		http.Error(w, "failed to marshal result: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(bsOut)
	if err != nil {
		handleWriteErr(err, w)
	}
}

func createMainCollectionEntryInTxn(ctx mongo.SessionContext, toInsert MainCollectionItem) error {
	err := addToIdMapCollection(ctx, toInsert)
	if err != nil {
		return errors.Join(errors.New("failed to insert in map collection"), err)
	}
	_, err = mongo.SessionFromContext(ctx).Client().Database(dbName).Collection(toInsert.CollectionName()).InsertOne(ctx, toInsert)
	if err != nil {
		return errors.Join(errors.New("failed to insert main collection item"), err)
	}
	return nil
}
