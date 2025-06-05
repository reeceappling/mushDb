package rfid

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"reflect"
	"time"
)

const pcRunCollectionName = "pcRuns"

type PCRun struct { // TODO: add most recently updated field, add creation date field
	Id          alternateCollectionId `bson:"_id" json:"_id"`
	Date        unixTime              `bson:"date" json:"date"`
	RunTime     string                `bson:"runtime" json:"runtime"`
	Notes       []Note                `bson:"notes,omitempty" json:"notes,omitempty"`
	LastUpdated unixTime              `bson:"lastUpdated" json:"lastUpdated"`
}

func (run PCRun) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := run
	err := decodeItem(&out, encoded)
	return out, err
}

func (run PCRun) clean() CollectionItem {
	return run
}

func (run PCRun) EntryTypeField() *string {
	return nil
}

func (run PCRun) CollectionName() string {
	return pcRunCollectionName
}

func initializePCRun(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(pcRunCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		newSimpleIndex("date", "date", true, false, false),
		//RunTime (likely no index)    string                `bson:"runtime" json:"runtime"`
		//Notes (no index unless tags)
		lastUpdatedIndexModel,
	})
	// TODO: project, sale, sporePrint
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	existingEntry := PCRun{}
	testItem := PCRun{
		Id:          exAltId,
		Date:        exampleTime,
		RunTime:     "1 hour",
		Notes:       exampleNotes(),
		LastUpdated: exampleTime,
	}
	err = coll.FindOne(ctx, bson.D{{"_id", exAltId}}).Decode(&existingEntry)
	if err == nil {
		if reflect.DeepEqual(existingEntry, testItem) {
			return nil
		}
	}
	return renameMe(ctx, coll, exAltId, testItem, existingEntry)
}

type createPcRunRequest struct {
	CreationDate unixTime
	RunTime      string
	Notes        []Note `json:"notes,omitempty"`
}

func createPcRunHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req := createPcRunRequest{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id := newAlternateCollectionId()

	entry := PCRun{
		Id:          id,
		Date:        req.CreationDate,
		RunTime:     req.RunTime, // TODO: VALIDATE
		Notes:       req.Notes,
		LastUpdated: unixTime(time.Now().UnixMilli()),
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(pcRunCollectionName)
		_, err := coll.InsertOne(ctx, entry)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return nil, nil
		}
		return w.Write(id.base58Bytes())
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}

type updatePcRunRequest struct {
	Notes AllEntries[Note] `json:"notes"`
}

func updatePcRunHandler(w http.ResponseWriter, r *http.Request) {
	b58Id := Base58Str(r.PathValue("id"))
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req := updatePcRunRequest{}
	err = json.Unmarshal(bs, &req)
	if err != nil {
		http.Error(w, "failed to unmarshal body: "+err.Error(), http.StatusBadRequest)
		return
	}
	id, err := b58Id.toAltCollectionId()
	if err != nil {
		http.Error(w, "Invalid id! "+err.Error(), http.StatusBadRequest)
		return
	}
	existing, err := GetAltCollectionItem(r.Context(), id.String(), PCRun{})
	if err != nil {
		stat := http.StatusInternalServerError
		if err == mongo.ErrNoDocuments {
			stat = http.StatusNotFound
		}
		http.Error(w, err.Error(), stat)
		return
	}
	if !ableToModify(r.Context()) { // TODO: DO THIS EVERYWHERE!
		http.Error(w, "not permitted to modify", http.StatusForbidden)
		return
	}
	mods := bson.D{}
	// Do note changes
	mods, err = WithNotesUpdate(bson.D{}, req.Notes, existing.Notes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(mods) == 0 {
		http.Error(w, "no changes made", http.StatusBadRequest)
		return
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(pcRunCollectionName)
		result := coll.FindOneAndUpdate(ctx, bson.D{{"_id", existing.Id}}, mods)
		if err := result.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, nil
		}
		return w.Write([]byte(b58Id))
	})
	if err != nil {
		// TODO: WHAT HERE?
	}
}
