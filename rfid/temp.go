package rfid

import (
	"context"
	"crypto/rand"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func randomRFID(bytes int) []byte {
	b := make([]byte, bytes)
	_, _ = rand.Read(b)
	return b
}

func writeRfidTagIfNecessary(ctx context.Context, writeTagTo *string, id MainCollectionId) error {
	if writeTagTo == nil {
		return nil // Don't write
	}
	// TODO: DO THIS!!!!! IDEAS IN test/main
	return errors.New("not yet implemented") // TODO: write tag
}

func StandardizeMainCollectionId(id string) (*MainCollectionId, error) {
	var out MainCollectionId
	idBytes := []byte(id)
	if len(idBytes) == 8 {
		out = [8]byte(idBytes)
		return &out, nil
	}
	realId, err := Base58Str(idBytes).toMainCollectionId()
	if err != nil {
		return nil, err
	}
	return &realId, nil
}

func StandardizeAltCollectionId(id string) (*alternateCollectionId, error) {
	var out alternateCollectionId
	idBytes := []byte(id)
	if len(idBytes) == 12 {
		out = [12]byte(idBytes)
		return &out, nil
	}
	realId, err := Base58Str(idBytes).toAltCollectionId()
	if err != nil {
		return nil, err
	}
	return &realId, nil
}

func GetMainCollectionItem[T MainCollectionItem](ctx context.Context, id MainCollectionId, resultItemType T) (out MainCollectionItem, err error) {
	encodedResult := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(mainCollectionName).FindOne(ctx, bson.D{{"_id", id}})
	if encodedResult.Err() != nil {
		return resultItemType, encodedResult.Err() // mongo.ErrNoDocuments if 404
	}
	//if resultItem != nil { // TODO: DELETE OR USE ELSEWHERE
	//	out = *resultItem
	//} else {
	//	raw := bson.Raw{}
	//	raw, err = encodedResult.Raw()
	//	if err != nil {
	//		return nil, errors.Join(err, errors.New("failed to decode encoded result to raw"))
	//	}
	//	out, err = rawEntryTypeConversion(raw)
	//	if err != nil {
	//		return nil, err
	//	}
	//}
	temp, err := resultItemType.Decode(encodedResult)
	if err != nil {
		err = errors.Join(errors.New("failed to decode"), err)
		return resultItemType, err
	}
	return temp.(MainCollectionItem), err
}

func GetMainCollectionItemInTxn(ctx mongo.SessionContext, id MainCollectionId, optionalItemForType *MainCollectionItem) (out MainCollectionItem, err error) {
	encodedResult := ctx.Client().Database(dbName).Collection(mainCollectionName).FindOne(ctx, bson.D{{"_id", id}})
	if encodedResult.Err() != nil {
		return nil, encodedResult.Err() // mongo.ErrNoDocuments if 404
	}
	if optionalItemForType != nil {
		out = *optionalItemForType
	} else {
		raw := bson.Raw{}
		raw, err = encodedResult.Raw()
		if err != nil {
			return nil, err
		}
		out, err = rawEntryTypeConversion(raw)
		if err != nil {
			return nil, err
		}
	}
	err = encodedResult.Decode(&out)
	return
}

func GetAltCollectionItem[T AltCollectionItem](ctx context.Context, id string, item T) (out T, err error) {
	out = item
	encodedResult := ctx.Value(mongoClientContextKey).(*mongo.Client).
		Database(dbName).
		Collection(item.CollectionName()).
		FindOne(ctx, bson.D{{"_id", id}})
	if encodedResult.Err() != nil {
		return out, encodedResult.Err() // mongo.ErrNoDocuments if 404
	}
	err = encodedResult.Decode(&out)
	if err != nil {
		return out, err
	}
	// TODO: IF NOT CORRECT PERMS, THEN CHANGE OUTPUT AND CLEAN
	return out, nil
}

func GetAltCollectionItemInTxn[T AltCollectionItem](ctx mongo.SessionContext, id string, item T) (out T, err error) {
	out = item
	encodedResult := ctx.Client().
		Database(dbName).
		Collection(item.CollectionName()).
		FindOne(ctx, bson.D{{"_id", id}})
	if encodedResult.Err() != nil {
		return out, encodedResult.Err() // mongo.ErrNoDocuments if 404
	}
	err = encodedResult.Decode(&out)
	if err != nil {
		return out, err
	}
	return out, nil
}
