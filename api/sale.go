package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/reeceappling/mushDb/api/request"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
)

type entryTypeWithId struct {
	EntryType string `json:"entryType"`
	Id        string `json:"id"`
}
type entryTypeIdCount struct {
	EntryType string `json:"entryType"`
	Id        string `json:"id"`
	Count     *int   `json:"count,omitempty"` // Nil is the same as 1
}

type createSaleRequestNew struct {
	// TODO: notes?
	EntryTypes []string `json:"entryTypes"`
	Ids        []string `json:"ids"`                  // TODO: plugs will have a -# to specify how many????
	ClosePlugs []int    `json:"closePlugs,omitempty"` // If entryType is plugs, then it will not close out the entry unless specified
}

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

func (f SaleField) RequireUnsold() error {
	if f.Sale != nil {
		return errors.New("must be unsold")
	}
	return nil
}

type SalesField struct { // TODO: sales is multiple only for plugs?!
	Sales []AlternateCollectionId `bson:"sales,omitempty" json:"sales,omitempty"`
}

func (field SalesField) AddSale() {
	// TODO: IMPL AND USE
}

//type ItemSaleData struct {
//	UndiscountedPrice float64
//	DiscountedPrice   float64
//}
//
//type Discount struct {
//	// TODO: ???????
//}

/*
To model sales and discounts effectively in a database, use a dedicated Product Pricing table with start and end timestamps.
This prevents tampering with historical order data, tracks price changes over time, and allows you to apply either percentage-based or flat-rate currency discounts.
1. The Core Data Model. Instead of modifying the base Product table whenever there is a sale, structure your tables as follows:
	Product: Stores the core, static product details (Name, SKU, Base Price).

2. Best Practices for Order HistoryStore Price Snapshot on the Order:
	When a customer makes a purchase, hardcode the final calculated sale price into your Order_Line_Items table.
	Do not look up the price live later, as base prices and discounts change over time.
	Keep Promotions Separated: If you are running a multi-item promotion or using promo codes, create a Discounts or Promotions table.
	Link this to the Order or Order_Line_Items table so you can track how much revenue a specific campaign or coupon generated.

Create tables:
Product: Stores the core, static product details (Name, SKU, Base Price).
Product_Pricing: Manages all price fluctuations. It includes:
		price_id: id for this very specific price
		product_id (Foreign Key to Product)
		price (The new sale price or base price)
		discount_id? which discount(s) this price is part of?
		discount_type (e.g., 'PERCENTAGE', 'FLAT_AMOUNT')
		discount_value (e.g., 10 for 10%)
		valid_from (Timestamp for when the deal starts)
		valid_until (Timestamp for when the deal ends)
Order_Line_Items table, each line item for an order (multi-bundles show up as multiple line items, where each line item points to the same txn)
	itemType, item id, order id, link to base product, final price after any discounts
Promotions and/or discounts table.
	HOW ARE RULES STORED IN HERE? How can we assert that only plates can be bought, or just 2 plates and a print?
	valid_from (Timestamp for when the deal starts)
	valid_until (Timestamp for when the deal ends)
*/

//type EntryTypesSoldMap map[string]map[MainCollectionId]int // map of entryType to mainCollectionId to count

//type SaleFileContent struct {
//	PurchaseDate time.Time
//	ShipDate     *time.Time
//	SoldItems    EntryTypesSoldMap
//	Discounts    []Discount
//	TotalPrice   float64
//	// TODO: how to account for
//}

type Sale struct { // TODO: THIS IS A LINE ITEM! SHOULD ONLY CONTAIN ONE ITEM! OR MAYBE A MULTIPLE OF THE SAME ITEM FOR THE CASE OF PLUGS??
	AlternateCollectionIdField `bson:"inline"`
	// TODO: price????? other info?????
	SoldItems         []entryTypeIdCount
	CreationDateField `bson:"inline"` // This is sale date
	NotesField        `bson:"inline"`
	LastUpdatedField  `bson:"inline"`
	AclField          `bson:"inline"`
}

//type Sale struct {
//	AlternateCollectionIdField `bson:"inline"`
//	CreationDateField          `bson:"inline"` // This is sale date
//	// TODO: price?????
//	// TODO: shipDate?
//	// TODO: refunded?
//	// TODO: destination (location, person, etc), nil means in-person
//	// TODO: soldAt (location)
//	// TODO: soldTo (destination state if done online, address, who bought it?)
//	SoldItems []entryTypeIdCount
//
//	NotesField       `bson:"inline"`
//	LastUpdatedField `bson:"inline"`
//	AclField         `bson:"inline"`
//}
//
//type Order struct { // TODO: new inPerson order vs packing pre-existing order?
//
//	AlternateCollectionIdField `bson:"inline"`
//	CreationDateField          `bson:"inline"`
//	ExternalOrderId            *string `json:"externalOrderId,omitempty"` // CC processing id?
//	TotalPrice                 float64 // TODO: incl tax?
//	SaleLocation               string  // TODO: ????
//	Destination                string  // TODO: if sold online, then who/where to. If sold in-person, use that location...
//	// TODO: Shipped bool `bson:"shipped" json:"shipped"`
//	Refunded *unix.Time // TODO: ???
//	//Transactions are usually split into two parts:
//	//	the Header (which stores the total invoice, global order discounts, and overarching transaction fields)
//	// 	Line items: SKU (itemType, species, subspecies), price_id (includes possible bundles/promotions that are not order-wide), item_id (from db?)
//
//	// Line items point to the order, not the other way around?????
//}
//
//// TODO: func to create plugs sale?
//// TODO: func to create a sale for anything else?

func initializeSales(ctx context.Context) error {
	// Indices
	coll := DbFrom(ctx).Collection(SalesCollectionName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		newSimpleIndex("saleDate", "creationDate", true, false, false),
		//notes
		projectsIndexModel,
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	testItem := &Sale{
		AlternateCollectionIdField: AlternateCollectionIdField{exAltId},
		CreationDateField:          CreationDateField{exampleTime},
		NotesField:                 NotesField{exampleNotes()},
		LastUpdatedField:           LastUpdatedField{exampleTime},
		AclField:                   allCanReadAcl(nil),
	}
	return addTestAltEntries(ctx, testItem)
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
	// TODO: HOW TO HANDLE PERMS? FOR NOW, JUST DO ONLY USER?

	ctx, now := request.UnixTime(r.Context()) // TODO: no more r.Context below
	id := newAlternateCollectionId()
	toInsert := &Sale{
		AlternateCollectionIdField: AlternateCollectionIdField{id},
		CreationDateField:          CreationDateField{now},
		NotesField:                 req.NotesField,
		LastUpdatedField:           LastUpdatedField{now},
		// TODO: USE PARENT PERMS?????
	}
	finishCreateAlternateEntry(ctx, toInsert, w)
}

type updateSaleRequest struct {
	NotesUpdateField
	PermsOnRequest `json:"acl"` // TODO: ???? handle in typescript and handler!
}

func (req updateSaleRequest) modsFor(existing *Sale, aclField AclField) (bson.D, error) {
	return NewMods().
		updateNotesIfNeeded(req, existing).
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
