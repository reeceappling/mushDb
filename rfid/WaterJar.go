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
)

// TODO: required for:
// TODO: MSS
// TODO:
// TODO: Stasis tube (if filled with water later, probably not)

type WaterJar struct { // TODO: HANDLE THIS EVERYWHERE! DO ALL TYPESCRIPT FOR THIS!
	MainCollectionIdField `bson:"inline"`
	CreationDateField     `bson:"inline"` // From PcRun
	PcRunField            `bson:"inline"` // Creation date assumed to be the same as pc run date
	NotesField            `bson:"inline"`
	DisposedField         `bson:"inline"`
	LastUpdatedField      `bson:"inline"`
}

func (wj WaterJar) GeneticInfoAsParent() (GeneticParentInfo, error) {
	return GeneticParentInfo{}, errors.New("WaterJar has no genetic info")
}

func (wj WaterJar) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
	return errors.New("WaterJar is not genetic, so it cannot setTransferChild")
}

func (wj WaterJar) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	panic("WaterJar is not genetic, so it cannot have a generation")
}

func (wj WaterJar) CanTransferTo(dst geneticSource) error {
	return errors.New("WaterJar is not genetic, so it cannot transfer to anything")
}

func (wj WaterJar) Innoculatable() bool {
	// WaterJar is not genetic, so it cannot be innoculated
	return false
}

func (wj WaterJar) Permissions() *ACL {
	// Water jars always have full write perms
	return nil
}

type WaterJarOptionalField struct {
	WaterSource *MainCollectionId `bson:"waterSource,omitempty" json:"waterSource,omitempty"`
}

func (field WaterJarOptionalField) Get(ctx context.Context) (out PCRun, err error) {
	if field.WaterSource == nil {
		err = ErrMissingOptionalField
		return
	}
	return WaterJarField{*field.WaterSource}.Get(ctx)
}

type WaterJarField struct {
	WaterSource MainCollectionId `bson:"waterSource" json:"waterSource"`
}

func (field WaterJarField) Get(ctx context.Context) (out PCRun, err error) {
	err = ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(WaterJarsCollectionName).FindOne(ctx, bsonFindFilter("_id", field.WaterSource)).Decode(&out)
	return out, err
}

func initializeWaterJars(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(WaterJarsCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		creationDateIndexModel,
		newSimpleIndex("pcRun", "pcRun", false, false, false),
		//Notes (no index unless tags)
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	testItem := &WaterJar{
		MainCollectionIdField: MainCollectionIdField{exWaterId},
		CreationDateField:     CreationDateField{exampleTime},
		PcRunField:            PcRunField{exAltId},
		NotesField:            NotesField{exampleNotes()},
		LastUpdatedField:      LastUpdatedField{exampleTime},
	}
	println("binary water jar id initial:"+string(exWaterId[:]), len(exWaterId[:]))
	println("created waterJar with id: " + exWaterId.asBase58())
	return addTestMainEntries(ctx, testItem)
}

type createWaterJarRequest struct {
	PcRunField
	NotesField
	WriteTagToField
}

func createWaterJarHandler(w http.ResponseWriter, r *http.Request) { // TODO: THIS! test ts!
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req := createWaterJarRequest{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id := NextMainCollectionId()
	ctx, _ := Db(r)
	pcRun, err := req.PcRunField.Get(ctx)
	if err != nil {
		http.Error(w, "failed to get pc run: "+err.Error(), http.StatusInternalServerError)
		return
	}

	toInsert := &WaterJar{
		MainCollectionIdField: MainCollectionIdField{id},
		CreationDateField:     pcRun.CreationDateField,
		PcRunField:            req.PcRunField,
		NotesField:            req.NotesField,
		DisposedField:         DisposedField{nil},
		LastUpdatedField:      LastUpdatedField{unixTimeForNow()},
	}

	err = writeRfidTagIfNecessary(r.Context(), req.WriteTagTo, id)
	if err != nil {
		http.Error(w, "failed to write tag: "+err.Error(), http.StatusInternalServerError)
		return
	}
	finishCreateMainCollectionEntry(ctx, toInsert, w)
}

type updateWaterJarRequest struct {
	NotesUpdateField
	DisposedField
}

func (req updateWaterJarRequest) modsFor(existing *WaterJar, aclField AclField) (bson.D, error) {
	return NewMods().
		updateNotesIfNeeded(req, existing).
		updateDisposedIfNeeded(req, existing).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateWaterJarHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: make this a func that can be called from most mainCollItem updates ------
	_, id, err := mainCollIdFromRequest(r, w)
	if err != nil {
		return
	}
	req := updateWaterJarRequest{}
	if err = ReadSimpleStructuredBody(r, w, &req); err != nil {
		return
	}
	ctx, db := Db(r)
	coll := db.Collection(WaterJarsCollectionName)
	println("binary water jar id:"+string(id[:]), len(id[:])) // TODO: del
	existing, err := GetMainCollectionItem(r.Context(), id, &WaterJar{})
	if err != nil {
		stat := http.StatusInternalServerError
		if err == mongo.ErrNoDocuments {
			println("FAILED TO FIND WATER JAR")
			stat = http.StatusNotFound
		}
		http.Error(w, err.Error(), stat)
		return
	}
	wj, ok := existing.(*WaterJar)
	if !ok {
		http.Error(w, "mcItem was not WaterJar", http.StatusInternalServerError)
		return
	}
	unnecessaryPerms := PermsOnRequest{BlanketPerm: utils.Pointer(ReadWritePerm(true))}
	finishMainCollItemUpdate(ctx, w, coll, req.modsFor, wj, unnecessaryPerms)
}
