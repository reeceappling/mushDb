package rfid

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/reeceappling/goUtils/v2/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"reflect"
	"time"
)

// TODO: SPORE SWAB?!?!?!?!

const agarBatchCollectionName = "agarBatches"

type AgarBatch struct { // This is >=1 media bottles of the same recipe that went through the same PC cycle
	Id    alternateCollectionId `bson:"_id" json:"_id"`
	PcRun alternateCollectionId `bson:"pcRun,omitempty" json:"pcRun,omitempty"`
	// Recipe is the id of the Agar Recipe used
	Recipe      alternateCollectionId `bson:"recipe" json:"recipe"`
	Color       colorant              `bson:"color" json:"color"`
	Notes       []Note                `bson:"notes,omitempty" json:"notes,omitempty"`
	LastUpdated unixTime              `bson:"lastUpdated" json:"lastUpdated"`
}

func (batch AgarBatch) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := batch
	err := decodeItem(&out, encoded)
	return out, err
}

func (batch AgarBatch) clean() CollectionItem {
	return batch
}

func (batch AgarBatch) EntryTypeField() *string {
	return nil
}

func (batch AgarBatch) CollectionName() string {
	return agarBatchCollectionName
}

type NewAgarBatchRequest struct {
	PcRunId string // alt coll ID
	Recipe  string // alt coll ID
	Color   *string
	Notes   []string
}

func (req NewAgarBatchRequest) asAgarBatch() (AgarBatch, error) {
	pcRunId, err := StandardizeAltCollectionId(req.PcRunId)
	if err != nil {
		return AgarBatch{}, errors.New("invalid PC Run ID")
	}
	recipe, err := StandardizeAltCollectionId(req.Recipe)
	if err != nil {
		return AgarBatch{}, errors.New("invalid Recipe ID")
	}
	c := utils.Default(req.Color, string(clear))
	colorSelected, exists := colors[c]
	if !exists {
		return AgarBatch{}, errors.New("invalid color")
	}
	return AgarBatch{
		Id:          alternateCollectionId(primitive.NewObjectID()),
		PcRun:       *pcRunId,
		Recipe:      alternateCollectionId(*recipe),
		Color:       colorSelected,
		Notes:       stringsToNotes(req.Notes, time.Now()),
		LastUpdated: unixTime(time.Now().UnixMilli()),
	}, nil
}

func newAgarBatch(ctx mongo.SessionContext, batch AgarBatch) (*alternateCollectionId, error) {
	out, err := ctx.Client().Database(dbName).Collection(agarBatchCollectionName).InsertOne(ctx, batch)
	if err != nil {
		return nil, err
	}
	res, ok := out.InsertedID.(alternateCollectionId)
	if !ok {
		return nil, errors.New("failed to convert to primitive.ObjectID")
	}
	return &res, err
}

type updateAgarBatchRequest struct {
	Notes AllEntries[Note] `json:"notes"`
}

func updateAgarBatchHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	b58Id := Base58Str(r.PathValue("id"))
	req := updateAgarBatchRequest{}
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(bytes, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := b58Id.toAltCollectionId()
	if err != nil {
		http.Error(w, "Invalid id! "+err.Error(), http.StatusBadRequest)
		return
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		existing, err := GetAltCollectionItemInTxn(ctx, id.String(), AgarBatch{})
		if err != nil {
			stat := http.StatusInternalServerError
			if err == mongo.ErrNoDocuments {
				stat = http.StatusNotFound
			}
			http.Error(w, err.Error(), stat)
			return nil, nil

		}
		if !ableToModify(ctx) { // TODO: DO THIS EVERYWHERE!
			http.Error(w, "not permitted to modify", http.StatusForbidden)
			return nil, nil
		}
		// Do note changes
		mods, err := WithNotesUpdate(bson.D{}, req.Notes, existing.Notes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		if len(mods) == 0 {
			http.Error(w, "no changes made", http.StatusBadRequest)
			return nil, nil
		}
		coll := ctx.Value(mongoClientContextKey).(*mongo.Client).
			Database(dbName).
			Collection(agarBatchCollectionName)
		result := coll.
			FindOneAndUpdate(ctx, bson.D{{"_id", existing.Id}}, mods)
		err = result.Err()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, nil
		}
		return w.Write(alternateCollectionId(existing.Id).base58Bytes())
	})

	if err != nil {
		handleWriteErr(err, w)
	}
}

func initializeAgarBatches(ctx context.Context) error {
	// Indices
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(agarBatchCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		newSimpleIndex("pcRun", "pcRun", false, true, false),
		newSimpleIndex("recipe", "recipe", false, false, false),
		newSimpleIndex("color", "color", false, false, false),
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	existingEntry := AgarBatch{}
	testItem := AgarBatch{
		Id:          exAltId,
		PcRun:       altCollIdForint(0),
		Recipe:      altCollIdForint(idLmea),
		Color:       clear,
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

type createAgarBatchRequest struct {
	Color  colorant  `json:"color"`
	PcRun  Base58Str `json:"pcRun"`
	Recipe Base58Str `json:"recipe"`
	Notes  []Note    `json:"notes,omitempty"`
}

func createAgarBatchHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	req := createAgarBatchRequest{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pcRun, err := req.PcRun.toAltCollectionId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	recipe, err := req.Recipe.toAltCollectionId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id := alternateCollectionId(newAlternateCollectionId())
	if !ValidColor(req.Color) {
		http.Error(w, "Invalid agar color!", http.StatusBadRequest)
		return
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		out, err := newAgarBatch(ctx, AgarBatch{
			Id:          id,
			PcRun:       pcRun,
			Recipe:      alternateCollectionId(recipe),
			Color:       req.Color,
			Notes:       req.Notes,
			LastUpdated: unixTime(time.Now().UnixMilli()),
		})
		if err != nil {
			http.Error(w, "Agar batch creation failure: "+err.Error(), http.StatusInternalServerError)
			return nil, nil
		}
		return w.Write(alternateCollectionId(*out).base58Bytes())
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}
