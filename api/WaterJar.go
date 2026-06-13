package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/reeceappling/goUtils/v2/utils"
	"github.com/reeceappling/mushDb/api/request"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
)

/* required for:
MSS
Stasis tube (if filled with water later, probably not)
*/

type WaterJar struct {
	MainCollectionIdField `bson:"inline"`
	CreationDateField     `bson:"inline"` // From PcRun
	PcRunField            `bson:"inline"` // Creation date assumed to be the same as pc run date
	NotesField            `bson:"inline"`
	DisposedField         `bson:"inline"`
	LastUpdatedField      `bson:"inline"`
	AclField              `bson:"inline"`
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

func (wj WaterJar) Permissions() ACL {
	// Water jars always have full write perms
	return ACL{Users: nil, Projects: nil, BlanketPerm: utils.Pointer(true)}
}

type WaterJarOptionalField struct {
	WaterSource *MainCollectionId `bson:"waterSource,omitempty" json:"waterSource,omitempty"`
}

func (field WaterJarOptionalField) Get(ctx context.Context) (out WaterJar, err error) {
	if field.WaterSource == nil {
		err = ErrMissingOptionalField
		return
	}
	return WaterJarField{*field.WaterSource}.Get(ctx)
}

type WaterJarField struct {
	WaterSource MainCollectionId `bson:"waterSource" json:"waterSource"`
}

func (field WaterJarField) Get(ctx context.Context) (out WaterJar, err error) {
	err = DbFrom(ctx).Collection(WaterJarsCollectionName).FindOne(ctx, BsonFindFilter("_id", field.WaterSource)).Decode(&out)
	return out, err
}

func initializeWaterJars(ctx context.Context) error {
	// Indices
	coll := DbFrom(ctx).Collection(WaterJarsCollectionName)
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
		AclField:              allCanWriteAcl(),
	}
	println("binary water jar id initial:"+string(exWaterId[:]), len(exWaterId[:]))
	println("created waterJar with id: " + exWaterId.AsBase58())
	return addTestMainEntries(ctx, testItem)
}

type createWaterJarRequest struct {
	PcRunField
	NotesField
	WriteTagToField
	// Default perms are allCanWrite
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
	ctx, now := request.UnixTime(r.Context())
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
		LastUpdatedField:      LastUpdatedField{now},
		AclField:              allCanWriteAcl(),
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
	PermsOnRequest `json:"acl"`
}

func (req updateWaterJarRequest) modsFor(existing *WaterJar, aclField AclField) (bson.D, error) {
	return NewMods().
		updateNotesIfNeeded(req, existing).
		updateDisposedIfNeeded(req, existing).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateWaterJarHandler(w http.ResponseWriter, r *http.Request) {
	_, id, err := mainCollIdFromRequest(r, w)
	if err != nil {
		return
	}
	req := updateWaterJarRequest{}
	if err = ReadSimpleStructuredBody(r, w, &req); err != nil {
		return
	}
	ctx := r.Context()
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
	finishMainCollItemUpdate(ctx, w, req.modsFor, wj, req.PermsOnRequest)
}
