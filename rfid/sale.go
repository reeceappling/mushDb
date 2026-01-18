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
)

//var (
//	_ Sellable = &Bag{}
//	_ Sellable = &Fruit{}
//	_ Sellable = &FruitingChamber{}
//	_ Sellable = &GrainJar{}
//	_ Sellable = &LiquidCulture{}
//	_ Sellable = &MSS{}
//	_ Sellable = &Plate{}
//	_ Sellable = &Slant{}
//	_ Sellable = &SporePrint{}
//	_ Sellable = &SporeSwab{}
//	_ Sellable = &StasisTube{}
//	_ Sellable = &SporePrint{}
//)
//
//type Sellable interface {
//	AddSale() error // TODO: likely get rid of?
//}

type SaleField struct { // TODO: sales is multiple only for LC!
	Sale *AlternateCollectionId `bson:"sale,omitempty" json:"sale,omitempty"`
}

type SalesField struct { // TODO: sales is multiple only for plugs!
	Sales []AlternateCollectionId `bson:"sales,omitempty" json:"sales,omitempty"`
}

func (field SalesField) AddSale() {
	// TODO: IMPL AND USE
}

type Sale struct {
	AlternateCollectionIdField `bson:"inline"`
	//Lot               AlternateCollectionId `bson:"_id" json:"_id"` // Lot number // TODO: REMOVED
	CreationDateField `bson:"inline"` // This is sale date
	NotesField        `bson:"inline"`
	LastUpdatedField  `bson:"inline"`
	AclField          `bson:"inline"` // TODO: handle EVERYWHERE
}

func (s Sale) EntryTypeField() *string {
	return nil
}

func initializeSales(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(SalesCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		newSimpleIndex("saleDate", "creationDate", true, false, false),
		//notes
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	existingEntry := Sale{}
	testItem := Sale{
		AlternateCollectionIdField: AlternateCollectionIdField{exAltId},
		CreationDateField:          exampleTime.asCreationDate(),
		NotesField:                 NotesField{exampleNotes()},
		LastUpdatedField:           LastUpdatedField{exampleTime},
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
	Items []SoldItem
	NotesField
	// TODO: USE PARENT PERMS
}

type SoldItem struct {
	Type string // TODO: fix
	ID   string
}

func createSaleHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: PERMISSIONS! REUSE THEM FROM PARENT
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	req := createSaleRequest{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	ctx, db := Db(r)
	coll := db.Collection(SalesCollectionName)

	now := unixTimeForNow()
	id := newAlternateCollectionId()
	toInsert := Sale{
		AlternateCollectionIdField: AlternateCollectionIdField{id},
		CreationDateField:          unixTimeForNow().asCreationDate(),
		NotesField:                 req.NotesField,
		LastUpdatedField:           LastUpdatedField{now},
		// TODO: USE PARENT PERMS?????
		//PermsField:                 PermsField{nil}, // TODO: THIS!!!!!!!!!!!!!
	}
	finishCreateAlternateEntry(ctx, coll, &toInsert, w)
}

type updateSaleRequest struct {
	Notes          AllEntries[Note]
	PermsOnRequest // TODO: ???? handle in typescript and handler!
}

func (req updateSaleRequest) modsFor(existing *Sale, aclField AclField) (bson.D, error) {
	return NewMods().
		updateNotesIfNeeded(req.Notes, existing.Notes).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
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
	ctx, db := Db(r)
	coll := db.Collection(SalesCollectionName)
	existing := Sale{}
	err = coll.FindOne(ctx, bson.M{"_id": id}).Decode(&existing)
	if err != nil {
		stat := http.StatusInternalServerError
		if err == mongo.ErrNoDocuments {
			stat = http.StatusNotFound
		}
		dbErr(w, err.Error(), stat)
		return
	}
	finishAltCollItemUpdate(ctx, w, coll, req.modsFor, &existing, req.PermsOnRequest)
}
