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

const projectsCollectionName = "projects"

type Project struct {
	Name         string    `bson:"_id" json:"_id"`
	CreationDate unixTime  `bson:"creationDate" json:"creationDate"`
	Completed    *unixTime `bson:"completed,omitempty" json:"completed,omitempty"`
	Notes        []Note    `bson:"notes,omitempty" json:"notes,omitempty"`
	LastUpdated  unixTime  `bson:"lastUpdated" json:"lastUpdated"`
}

func (p Project) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := p
	err := decodeItem(&out, encoded)
	return out, err
}

func (p Project) clean() CollectionItem {
	return p
}

func (p Project) CollectionName() string {
	return projectsCollectionName
}

func (p Project) EntryTypeField() *string {
	return nil
}

func initializeProjects(ctx context.Context) error { // TODO: USE!!!!
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(projectsCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		newSimpleIndex("creationDate", "creationDate", true, false, false),
		newSimpleIndex("completed", "creationDate", true, true, false),
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	existingEntry := Project{}
	testItem := Project{
		Name:         exProj,
		CreationDate: exampleTime,
		Completed:    &exampleTime,
		Notes:        exampleNotes(),
		LastUpdated:  exampleTime,
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

type createProjectRequest struct {
	Name  string `json:"name"`
	Notes []Note `json:"notes"`
}

func createProjectHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req := createProjectRequest{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(projectsCollectionName)
		now := unixTime(time.Now().UnixMilli())
		_, err := coll.InsertOne(r.Context(), Project{
			Name:         req.Name,
			CreationDate: now,
			Notes:        req.Notes,
			LastUpdated:  now,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, nil
		}
		return w.Write([]byte(req.Name))
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}

type updateProjectRequest struct {
	Name      string           `json:"name"`
	Completed *unixTime        `json:"completed,omitempty"`
	Notes     AllEntries[Note] `json:"notes"`
}

func updateProjectHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req := updateProjectRequest{}
	err = json.Unmarshal(bs, &req)
	if err != nil {
		http.Error(w, "failed to unmarshal body: "+err.Error(), http.StatusBadRequest)
		return
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(projectsCollectionName)
		existing := Project{}
		err := coll.FindOne(ctx, bson.M{"_id": req.Name}).Decode(&existing)
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
		mods := bson.D{}
		// change standard if needed
		if req.Completed != existing.Completed {
			mods = bson.D{{"$set", bson.D{{"completed", req.Completed}}}}
		}
		// Do note changes
		mods, err = WithNotesUpdate(bson.D{}, req.Notes, existing.Notes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return nil, nil
		}
		if len(mods) == 0 {
			http.Error(w, "no changes made", http.StatusBadRequest)
			return nil, nil
		}
		result := coll.FindOneAndUpdate(ctx, bson.D{{"_id", existing.Name}}, mods)
		err = result.Err()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, nil
		}
		return w.Write([]byte(existing.Name))
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}
