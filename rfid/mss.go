package rfid

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/reeceappling/goUtils/v2/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"reflect"
	"time"
)

const MssSourceType = "mss"

type MSS struct {
	// ALWAYS assume contaminated
	EntryType    string           `bson:"entryType" json:"entryType"` // always mss
	Id           MainCollectionId `bson:"_id" json:"_id"`
	CreationDate unixTime         `bson:"creationDate" json:"creationDate"`
	Species      string           `bson:"species" json:"species"`
	SubSpecies   *string          `bson:"subSpecies,omitempty" json:"subSpecies,omitempty"`
	// NOTE: parentType is always either sporePrint or purchased
	Parent       *alternateCollectionId  `bson:"parent,omitempty" json:"parent,omitempty"` // no parent means purchased, traded-for, or imported
	Projects     []string                `bson:"projects,omitempty" json:"projects,omitempty"`
	TransfersOut []alternateCollectionId `bson:"transfersOut,omitempty" json:"transfersOut,omitempty"`
	Sale         *alternateCollectionId  `bson:"sale,omitempty" json:"sale,omitempty"`
	Disposed     *unixTime               `bson:"disposed,omitempty" json:"disposed,omitempty"`
	Notes        []Note                  `bson:"notes,omitempty" json:"notes,omitempty"`
	LastUpdated  unixTime                `bson:"lastUpdated" json:"lastUpdated"`
}

func (M MSS) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := M
	err := decodeItem(&out, encoded)
	return out, err
}

func (M MSS) clean() CollectionItem {
	out := M
	// TODO: Change species
	// TODO: change subspecies
	// TODO: remove parentType and Parent
	// TODO: remove projects
	// TODO: remove pic notes
	// TODO: remove mostRecentImage notes
	// TODO: remove notes
	return out
}

func (M MSS) DbId() string {
	return M.Id.dbIdStr()
}

func (M MSS) projects() []string {
	return M.Projects
}

func (M MSS) GeneticInfoAsParent() (GeneticParentInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (M MSS) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return utils.Pointer(Generation(0)), utils.Pointer(Generation(0))
}

func (M MSS) SourceType() string {
	return MssSourceType
}

func (M MSS) setTransferParent(ctx mongo.SessionContext, xfer Transfer) error {
	upd := pushToArray("transfersOut", xfer.Id)
	res, err := ctx.Client().Database(dbName).Collection(mainCollectionName).UpdateByID(ctx, M.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("Parent not found for transfer update. Should never happen!")
	}
	return nil
}

func (M MSS) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
	return errors.New("mss cannot be a child in a normal transfer") // TODO: ensure ok
}

func (M MSS) EntryTypeField() *string {
	return utils.Pointer(MssSourceType)
}

func (M MSS) CollectionName() string {
	return mainCollectionName
}

func (M MSS) id() []byte {
	return M.Id[:]
}

func (M MSS) knownFruitable() bool {
	return false
}

func (M MSS) children(ctx context.Context) ([]geneticSource, error) {
	return childrenOnlyToPlate(ctx, M.TransfersOut)
}

func (M MSS) idAsStr() string {
	return M.Id.dbIdStr()
}

func initializeMSS(ctx context.Context) error {
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	coll := db.Collection(mainCollectionName)
	// If test agar batch does not exist, then create it
	existingEntry := MSS{}
	testId := mainCollIdForint(idTestMSS)
	testItem := MSS{
		EntryType:    *existingEntry.EntryTypeField(),
		Id:           testId,
		CreationDate: exampleTime,
		Species:      exampleSpecies,
		SubSpecies:   exampleSubspecies,
		TransfersOut: exAlts,
		Parent:       &exAltId,
		Projects:     exProjects,
		Disposed:     &exampleTime,
		Notes:        exampleNotes(),
		LastUpdated:  exampleTime,
	}
	err := coll.FindOne(ctx, bson.D{{"_id", testId}}).Decode(&existingEntry)
	if err == nil {
		if reflect.DeepEqual(existingEntry, testItem) {
			return nil
		}
	}
	return renameMe(ctx, coll, testId, testItem, existingEntry)
}

type createMssRequest struct {
	SporePrintId Base58Str
	Notes        []Note `json:"notes,omitempty"`
	WriteTagTo   *string
}

func createMssHandler(w http.ResponseWriter, r *http.Request) { // TODO: ONLY CALLED FROM SPORE PRINT PAGE!
	data := createMssRequest{}
	id, err := generateMainCollectionId(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	b58id := id.asBase58()
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(bs, &data)
	if err != nil {
		http.Error(w, "failed to unmarshal request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	parentId, err := data.SporePrintId.toAltCollectionId()
	if err != nil {
		http.Error(w, "failed to parse sporePrintId: "+err.Error(), http.StatusBadRequest)
		return
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		db := ctx.Client().Database(dbName)
		coll := db.Collection(mainCollectionName)
		var parent SporePrint
		err = db.Collection(sporePrintCollectionName).FindOne(ctx, bson.D{{"_id", parentId}}).Decode(&parent)
		if err != nil {
			http.Error(w, "failed to find sporePrint: "+err.Error(), http.StatusBadRequest)
			return nil, nil
		}
		err = writeRfidTagIfNecessary(ctx, data.WriteTagTo, id)
		if err != nil {
			http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
			return nil, nil
		}
		now := unixTime(time.Now().UnixMilli())
		_, err := coll.InsertOne(ctx, MSS{
			EntryType:    "mss",
			Id:           id,
			CreationDate: now,
			Species:      parent.Species,
			SubSpecies:   parent.Subspecies,
			Parent:       &parent.Id,
			Projects:     parent.Projects,
			Notes:        data.Notes,
			LastUpdated:  now,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, nil
		}
		return w.Write([]byte(b58id))
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}

type importMssRequest struct { // TODO: THIS!
	CreationDate unixTime
	Species      string    // TODO: VALIDATE ON INSERT
	Recipe       Base58Str // lc recipe
	Subspecies   *string
	Notes        []Note `json:"notes,omitempty"`
	WriteTagTo   *string
	// image as "img"
}

func importMssHandler(w http.ResponseWriter, r *http.Request) {
	data := importMssRequest{}
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(bs, &data)
	if err != nil {
		http.Error(w, "failed to unmarshal request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	id, err := generateMainCollectionId(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	b58id := id.asBase58()
	if speciesIsSpecial(r.Context(), &data.Species) && !userIsAdmin(r.Context()) { // TODO: DO THIS EVERYWHERE!
		http.Error(w, "not permitted to modify", http.StatusForbidden)
		return
	}
	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	out := MSS{
		EntryType:    "mss",
		Id:           id,
		CreationDate: data.CreationDate,
		Species:      data.Species,
		SubSpecies:   data.Subspecies,
		Notes:        data.Notes,
		LastUpdated:  unixTime(time.Now().UnixMilli()),
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(mainCollectionName)
		_, err := coll.InsertOne(ctx, out)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, nil
		}
		return w.Write([]byte(b58id))
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}

type updateMssRequest struct {
	Projects   []string `json:"projects,omitempty"`
	Notes      AllEntries[Note]
	Disposed   *unixTime              `json:"disposed,omitempty"`
	Sale       *alternateCollectionId `json:"sale,omitempty"`
	WriteTagTo *string
}

func updateMssHandler(w http.ResponseWriter, r *http.Request) {
	data := updateMssRequest{}
	b58Id := Base58Str(r.PathValue("id"))
	defer r.Body.Close()
	id, err := b58Id.toMainCollectionId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(bs, &data)
	if err != nil {
		http.Error(w, "failed to unmarshal request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	err = writeRfidTagIfNecessary(r.Context(), data.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(mainCollectionName)
		// go get current plate
		current := MSS{}
		err := coll.FindOne(ctx, bson.D{{"_id", id}}).Decode(&current)
		if err != nil {
			http.Error(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
			return nil, nil
		}
		if speciesIsSpecial(ctx, &current.Species) && !userIsAdmin(ctx) { // TODO: DO THIS EVERYWHERE!
			http.Error(w, "not permitted to modify", http.StatusForbidden)
			return nil, nil
		}
		upd := bson.D{}
		// Compare SALES
		upd = setUnsetUnequalPointers("sale", data.Sale, current.Sale, upd)
		// Compare PROJECTS
		upd = setProjectsIfUnequal(upd, data.Projects, current.Projects)
		upd = setUnsetUnequalPointers("disposed", data.Disposed, current.Disposed, upd)
		// Do note changes
		mods, err := WithNotesUpdate(bson.D{}, data.Notes, current.Notes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return nil, nil
		}

		if len(mods) == 0 {
			http.Error(w, "no changes made", http.StatusBadRequest)
			return nil, nil
		}

		// write updates to db
		res := coll.FindOneAndUpdate(ctx, bson.D{{"_id", id}}, mods)
		if err = res.Err(); err != nil {
			http.Error(w, "failed to write update to db: "+err.Error(), http.StatusInternalServerError)
			return nil, nil
		}
		return w.Write([]byte(b58Id))
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}
