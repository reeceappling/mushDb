package rfid

import (
	"context"
	"encoding/json"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"reflect"
	"time"
)

const salesCollectionName = "sales"

type Sale struct {
	Lot         alternateCollectionId `bson:"_id" json:"_id"` // Lot number
	SaleDate    unixTime              `bson:"saleDate" json:"saleDate"`
	Notes       []Note                `bson:"notes,omitempty" json:"notes,omitempty"`
	LastUpdated unixTime              `bson:"lastUpdated" json:"lastUpdated"`
}

func (s Sale) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := s
	err := decodeItem(&out, encoded)
	return out, err
}

func (s Sale) clean() CollectionItem {
	return s
}

func (s Sale) CollectionName() string {
	return salesCollectionName
}

func (s Sale) EntryTypeField() *string {
	return nil
}

func initializeSales(ctx context.Context) error { // TODO: USE!!!!
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(salesCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		newSimpleIndex("saleDate", "saleDate", true, false, false),
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	existingEntry := Sale{}
	testItem := Sale{
		Lot:         exAltId,
		SaleDate:    exampleTime,
		Notes:       exampleNotes(),
		LastUpdated: exampleTime,
	}
	err = coll.FindOne(ctx, bson.D{{"_id", exAltId}}).Decode(&existingEntry)
	if err == nil {
		if reflect.DeepEqual(existingEntry, testItem) {
			return nil
		}
	}
	res, err := coll.InsertOne(ctx, testItem)
	if err != nil {
		return err
	}
	if res == nil {
		return errors.New("result should not be nil")
	}
	if res.InsertedID != exAltId {
		return errors.New("entry id did not match")
	}
	return nil
}

type createSaleRequest struct {
	Notes []Note `json:"notes"`
}

func createSaleHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	req := createSaleRequest{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		db := ctx.Client().Database(dbName)
		coll := db.Collection(salesCollectionName)
		now := unixTime(time.Now().UnixMilli())
		id := newAlternateCollectionId()
		_, err = coll.InsertOne(r.Context(), Sale{
			Lot:         alternateCollectionId(id),
			SaleDate:    now,
			Notes:       req.Notes,
			LastUpdated: now,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, nil
		}
		_, err = w.Write(id.base58Bytes())
		return nil, err
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}

type updateSaleRequest struct {
	Notes AllEntries[Note]
}

func updateSaleHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req := updateSaleRequest{}
	err = json.Unmarshal(bs, &req)
	if err != nil {
		http.Error(w, "failed to unmarshal body: "+err.Error(), http.StatusBadRequest)
		return
	}
	b58Id := Base58Str(r.PathValue("id"))
	id, err := b58Id.toAltCollectionId()
	if err != nil {
		http.Error(w, "failed to convert id: "+err.Error(), http.StatusBadRequest)
		return
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		db := ctx.Client().Database(dbName)
		coll := db.Collection(salesCollectionName)
		existing := Sale{}
		err = coll.FindOne(ctx, bson.M{"_id": id}).Decode(&existing)
		if err != nil {
			stat := http.StatusInternalServerError
			if err == mongo.ErrNoDocuments {
				stat = http.StatusNotFound
			}
			http.Error(w, err.Error(), stat)
			return nil, nil
		}
		mods := bson.D{}
		// Do note changes
		mods, err = WithNotesUpdate(bson.D{}, req.Notes, existing.Notes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		if len(mods) == 0 {
			http.Error(w, "no changes made", http.StatusBadRequest)
			return nil, nil
		}
		result := coll.FindOneAndUpdate(ctx, bson.D{{"_id", b58Id}}, mods)
		err = result.Err()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, nil
		}
		_, err = w.Write([]byte(b58Id))
		return nil, err
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}
