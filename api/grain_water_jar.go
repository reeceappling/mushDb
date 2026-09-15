package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/reeceappling/mushDb/api/env"
	"github.com/reeceappling/mushDb/api/request"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
)

// TODO: implement in ts, and reveal in endpoints!

type GrainWaterJar struct {
	MainCollectionIdField `bson:"inline"`
	GrainBatchField       `bson:"inline"`
	NotesField            `bson:"inline"`
	CreationDateField     `bson:"inline"`
	LastUpdatedField      `bson:"inline"`
	DisposedField         `bson:"inline"`
	AclField              `bson:"inline"`
}

func (f GrainWaterJar) DbId() MainCollectionId {
	return f.Id
}

func (wj *GrainWaterJar) GeneticInfoAsParent() (GeneticParentInfo, error) {
	return GeneticParentInfo{}, errors.New("grain water jars have no genetics, and cannot be used as a parent") // TODO: ok?
}

func (wj *GrainWaterJar) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
	return errors.New("grainwater jars cannot be used as a parent for transfers")
}

func (wj *GrainWaterJar) generation() (sinceSpore, sinceSporeOrClone *Generation) {
	return nil, nil
}

func (wj *GrainWaterJar) CanTransferTo(dst geneticSource) error {
	return errors.New("not transferrable") // TODO: var?
}

func (wj *GrainWaterJar) Innoculatable() error {
	return errors.New("not innoculatable") // TODO: var?
}

func (wj *GrainWaterJar) EntryType() string {
	return "grainWaterJar" // TODO: const
}

func (wj *GrainWaterJar) Children(ctx context.Context) ([]MainCollectionItem, error) {
	panic("not implemented")
}

func (wj *GrainWaterJar) GetParent(ctx context.Context) (MainCollectionItem, error) {
	return MainCollectionItem(nil), errors.New("GWJs do not support GetParent") // TODO: ok?
}

func (gwj *GrainWaterJar) SourceType() string {
	return "grainWaterJar" // TODO: const
}

func createGrainWaterJarHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctx, now := request.UnixTime(r.Context())
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	req := createGrainWaterJarRequest{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	if _, err = req.GrainBatchField.Get(ctx); err != nil { // TODO; ensure ok
		http.Error(w, "failed to validate grain batch: "+err.Error(), http.StatusBadRequest)
		return
	}
	id := <-newMcids
	toInsert := GrainWaterJar{
		MainCollectionIdField: MainCollectionIdField{id},
		GrainBatchField:       req.GrainBatchField,
		NotesField:            req.NotesField,
		CreationDateField:     CreationDateField{now},
		LastUpdatedField:      LastUpdatedField{now},
		AclField:              allCanWriteAcl(),
	}
	finishImportMainCollectionEntry(ctx, &toInsert, w)
}

func initializeGrainWaterJars(ctx context.Context) error { // TODO: USE!
	// Indices
	coll := DbFrom(ctx).Collection(GrainWaterJarCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		// Id index already exists
		//Notes (no index unless tags)
		// TODO: grain batch?
		newSimpleIndex("creationDate", "creationDate", true, false, false),
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}

	return env.IfNotProd(ctx, func() error {
		id := mainCollIdForint(idTestGrainWaterJar)
		testItem := GrainWaterJar{
			MainCollectionIdField: MainCollectionIdField{id},
			GrainBatchField:       GrainBatchField{GrainBatch: defGrainBatchId}, // TODO: ??? default to 0-batch for imports?
			CreationDateField:     CreationDateField{},                          // TODO: ???
			NotesField:            NotesField{exampleNotes()},
			LastUpdatedField:      LastUpdatedField{exampleTime},
		}
		println("test Grain Water Jar:", id.AsBase58())
		return addTestMainEntries(ctx, &testItem)
	})
}

type updateGrainWaterJarRequest struct {
	NotesUpdateField
	PermsOnRequest `json:"acl"`
	DisposedField
}

func (req updateGrainWaterJarRequest) modsFor(existing *GrainWaterJar, acl AclField) (bson.D, error) {
	return NewMods().
		updateNotesIfNeeded(req, existing).
		updatePermsIfNeeded(acl.ACL, existing.ACL).
		updateDisposedIfNeeded(req, existing).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateGrainWaterJarHandler(w http.ResponseWriter, r *http.Request) {
	_, id, err := mainCollIdFromRequest(r, w)
	if err != nil {
		return
	}
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req := updateGrainWaterJarRequest{}
	err = json.Unmarshal(bs, &req)
	if err != nil {
		http.Error(w, "failed to unmarshal body: "+err.Error(), http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	existing, err := GetMainCollectionItemSpecific(ctx, id, &GrainWaterJar{})
	if err != nil {
		stat := http.StatusInternalServerError
		if err == mongo.ErrNoDocuments {
			stat = http.StatusNotFound
		}
		dbErr(w, err.Error(), stat)
		return
	}
	finishMainCollItemUpdate(ctx, w, req.modsFor, existing, req.PermsOnRequest)
}

type importGrainWaterJarRequest struct {
	GrainBatchOptionalField
	NotesField
}

func importGrainWaterJarHandler(w http.ResponseWriter, r *http.Request) { // TODO: USE!
	defer r.Body.Close()
	ctx, now := request.UnixTime(r.Context())
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	req := importGrainWaterJarRequest{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	if req.GrainBatch == nil {
		// Use the default batch
		req.GrainBatch = &exAltId
	} else {
		if _, err = req.GrainBatchOptionalField.Get(ctx); err != nil { // TODO; ensure ok
			http.Error(w, "failed to validate grain batch: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	id := <-newMcids
	toInsert := &GrainWaterJar{
		MainCollectionIdField: MainCollectionIdField{id},
		GrainBatchField:       GrainBatchField{GrainBatch: *req.GrainBatch},
		NotesField:            req.NotesField,
		CreationDateField:     CreationDateField{now},
		LastUpdatedField:      LastUpdatedField{now},
		AclField:              allCanWriteAcl(),
	}
	_ = finishCreateAlternateEntry(ctx, toInsert, w)
}

type GrainWaterJarsField struct {
	GrainWaterJars []AlternateCollectionId `bson:"grainWaterJars,omitempty" json:"grainWaterJars,omitempty"`
}

func (gwjs GrainWaterJarsField) EnsureExist(ctx context.Context, w http.ResponseWriter) error {
	if gwjs.GrainWaterJars == nil || len(gwjs.GrainWaterJars) == 0 {
		return nil
	}
	return ensureIdsExist(ctx, w, GrainWaterJarCollectionName, gwjs.GrainWaterJars)
}

// TODO: validate works, and use it everywhere necessary
func ensureIdsExist[T CollectionId](ctx context.Context, w http.ResponseWriter, collectionName string, ids []T) error {
	db := DbFrom(ctx)
	nExisting, err := db.Collection(collectionName).CountDocuments(ctx, bson.M{IDfld: bson.M{"$in": ids}})
	if err != nil {
		err = errors.Join(errors.New("failed to check ids in collection "+collectionName), err)
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	if int(nExisting) != len(ids) {
		err = errors.New("not all IDs provided were found in " + collectionName)
		dbErr(w, err.Error(), http.StatusBadRequest)
		return err
	}
	//for _, gwj := range ids { // TODO: this will work if we cant check them all at once
	//	err = db.Collection(GrainWaterJarCollectionName).FindOne(ctx, bson.M{
	//		IDfld: gwj,
	//	}).Err()
	//	if err != nil {
	//		err = errors.Join(errors.New("failed to get collection "+collectionName+" id "+string(gwj.AsBase58())), err)
	//		dbErr(w, err.Error(), http.StatusBadRequest)
	//		return err
	//	}
	//}
	return nil
}

type GrainWaterJarsDisposableField struct {
	GrainWaterJarsField
	GrainWaterJarsDisposed []bool `json:"grainWaterJarsDisposed,omitempty"`
}

func (f GrainWaterJarsDisposableField) Dispose(ctx mongo.SessionContext) (int, error) {
	db := DbFrom(ctx)
	_, now := request.UnixTime(ctx)
	toDispose := make([]AlternateCollectionId, 0, len(f.GrainWaterJars))
	for i := 0; i < len(f.GrainWaterJars); i++ {
		if f.GrainWaterJarsDisposed[i] {
			toDispose = append(toDispose, f.GrainWaterJars[i])
		}
	}
	upd, err := NewMods().
		setDisposedTo(now).    // TODO; ensure ok
		setLastUpdatedTo(now). // TODO; ensure ok
		Finalized()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	updateResult, err := db.Collection(GrainWaterJarCollectionName).UpdateMany(ctx, bson.M{
		IDfld: bson.M{"$in": toDispose}}, upd)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if int(updateResult.MatchedCount) != len(toDispose) {
		// TODO: 404 ok here?
		return http.StatusNotFound, fmt.Errorf(`tried to dispose %d but only matched %d`, len(toDispose), int(updateResult.MatchedCount))
	}
	return http.StatusOK, nil
}

/// TODO: use next func?
//func unstatusTxnFunc(f func(ctx mongo.SessionContext)(int,error))func(ctx mongo.SessionContext)(error){
//	return func(sess mongo.SessionContext)(error){
//		_,err := f(sess)
//		return err
//	}
//}
//func withStatus(f func(mongo.SessionContext)error, errorStatusCode int) func(mongo.SessionContext)(int,error){
//	return func(ctx mongo.SessionContext)(int,error){
//		if err := f(ctx); err != nil {
//			return errorStatusCode,err
//		}
//		return 200, nil
//	}
//}
