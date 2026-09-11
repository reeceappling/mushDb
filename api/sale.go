package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/reeceappling/mushDb/api/env"
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

/*

Products catalog table: defining products (which can have multiple serial numbers under one product)
CREATE TABLE products ( // TODO: mark items in entries collections as not available for sale? not-sellable-yet, available, sold, returned?
    id SERIAL PRIMARY KEY, -- product ID (SKU?)
    sku VARCHAR(50) UNIQUE NOT NULL, -- (Stock Keeping Unit): A unique identifier for each product, often used in inventory management.
    name VARCHAR(255) NOT NULL, -- The name of the product.
	description TEXT, --  A detailed description of the product.
    base_price DECIMAL(10, 2) NOT NULL,
    is_serialized BOOLEAN DEFAULT FALSE NOT NULL,
    -- unsure if needed! product_class_id INT REFERENCES product_classes(id) ON DELETE SET NULL -- MOST SPECIFIC CLASS?
);

serialized_items catalog: Tracks the individual physical assets (WE PROBABLY DONT WANT THIS! COVERED BY THE ENTRIES TABLES)
CREATE TABLE serialized_items (
    id SERIAL PRIMARY KEY,
    product_id INT REFERENCES products(id) ON DELETE CASCADE,
    serial_number VARCHAR(100) UNIQUE NOT NULL,
    status VARCHAR(50) DEFAULT 'in_stock'
        CHECK (status IN ('in_stock', 'allocated', 'sold', 'returned'))
);

sales_orders table: Contains transaction headers
CREATE TABLE sales_orders (
    id SERIAL PRIMARY KEY,
    sales_channel VARCHAR(20) NOT NULL CHECK (sales_channel IN ('pos', 'online')),
    order_status VARCHAR(50) DEFAULT 'pending'
        CHECK (order_status IN ('pending', 'processing', 'completed', 'cancelled')),
    total_amount DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


order_line_items table: Contains individual items within a transaction
CREATE TABLE order_line_items (
    id SERIAL PRIMARY KEY,
    sales_order_id INT REFERENCES sales_orders(id) ON DELETE CASCADE, -- (FK)
    product_id INT REFERENCES products(id), -- (FK to products)
    quantity INT NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(10, 2) NOT NULL -- Stores actual final price after promotions (Reflects promotional pricing or $0 for the free item)
);

order_line_serialized_items table: The bridge mapping specific physical serial numbers to the transaction.
-- Keeps order lines flexible. Null initially for online orders before picking.
CREATE TABLE order_line_serialized_items (
    id SERIAL PRIMARY KEY,
    order_line_item_id INT REFERENCES order_line_items(id) ON DELETE CASCADE, // TODO: is this just a replica of the line item field in each item?
    // TODO: how to mark sales on returns?
	serialized_item_id INT REFERENCES serialized_items(id) ON DELETE SET NULL,
    -- UNSURE IF NEEDED! scanned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

Handling the "Mix and Match" Bundle Rules:
	Because your bundles are pool-based ("Buy 2 of A, B, or C, get 1 of A, B, or C free"),
	you should not hardcode bundles as fixed products. Instead, use a Rule-Based Engine Architecture.

product_subclasses table: defines subclasses (plates will have subclasses commonPlate and rarePlate)
	id (unique row id)
	className (can have multiple rows)
	subclass_name (subClassification)


product_classes table: one row for each class a SKU falls into?  defines classes of products, eg: plates, rare plates, common plates, sporePrints, etc.
CREATE TABLE product_classes (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL
	product ID (SKU) (PRODUCTS SHOULD BE PLACED INTO THEIR MOST SPECIFIC CLASS?)
);


promotions table: Defines the overall promotional campaign (bundles)
CREATE TABLE promotions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL, -- (e.g., 'Mix & Match Triple Deal')
    promo_type VARCHAR(50) NOT NULL CHECK (promo_type IN ('buy_x_get_y', 'bundle_set_price')), -- // Buy X get Y, or bundle_set_price (bundle with set price) // TODO: full-order discount? bundle-discount?
    -- discount_type (e.g., '100_percent_off_cheapest', '50% off all')
	set_price DECIMAL(10, 2), -- Only used if promo_type is 'bundle_set_price'
    is_active BOOLEAN DEFAULT TRUE NOT NULL -- TODO: consider making an active timeframe
	-- TODO: consider just putting condition_group_id here!
	-- TODO: consider also putting reward id here?
);

promotion_condition_groups table: This table now simply lists the components that are allowed to fill up bucket #55.
-- Each contains the rules for the entire promotion inputs (2A and 3B get a C free == 2A and 3B in this case)
CREATE TABLE promotion_condition_groups (
    id SERIAL PRIMARY KEY, -- TODO: this is the group id???
    promotion_id INT REFERENCES promotions(id) ON DELETE CASCADE,
    required_quantity INT DEFAULT 0 NOT NULL, -- Total required across 'any' strategy pools
    selection_strategy VARCHAR(50) DEFAULT 'any'
        CHECK (selection_strategy IN ('any', 'all'))
        -- 'any' = Pool mix-and-match (Class A or Class B satisfies pool)
        -- 'all' = Checklist style (Must meet every row's specific required_quantity)
);

-- Defines the item triggers (The "Buy" or the "Set" contents)
promotion_conditions table: defines conditions required for a specific promotion
-- For "Buy 2A, 1B, get 1C free" there will be 2 rows, one for 2xA 1xB
-- EACH ROW IS A CONDITION! So for group ID of Plates and Prints, there'd be 2 rows - ID:GrpName:null:Plate, and ID:GrpName:null:Print
CREATE TABLE promotion_conditions (
    id SERIAL PRIMARY KEY,
	promotion_id INT REFERENCES promotions(id) ON DELETE CASCADE,
    condition_group_id INT REFERENCES promotion_condition_groups(id) ON DELETE CASCADE,
	-- Make both nullable: a condition can look for a specific SKU OR a whole class
    product_id INT REFERENCES products(id) ON DELETE CASCADE,
    product_class_id INT REFERENCES product_classes(id) ON DELETE CASCADE,
    required_quantity INT DEFAULT 1 NOT NULL, -- Used primarily when group strategy is 'all'
    -- TODO: maybe not: is_scaling_factor BOOLEAN DEFAULT FALSE NOT NULL, -- True if reward scales with this item's count

    -- Enforce that a condition row targets either a specific SKU or a Class, never both or neither
    CONSTRAINT chk_condition_target CHECK (
        (product_id IS NOT NULL AND product_class_id IS NULL) OR
        (product_id IS NULL AND product_class_id IS NOT NULL)
    )
);
////minimum_quantity_or_required_quantity String?

promotion_rewards table: defines the rewards for a specific promotion
-- Defines the resulting rewards (The "Get" contents - used only for buy_x_get_y)
CREATE TABLE promotion_rewards (
    id SERIAL PRIMARY KEY,
    promotion_id INT REFERENCES promotions(id) ON DELETE CASCADE,
	-- must contain ONLY 1 of the following 2!
    product_id INT REFERENCES products(id) ON DELETE CASCADE,
    product_class_id INT REFERENCES product_classes(id) ON DELETE CASCADE,

    -- Quantity Strategies
    reward_qty_strategy VARCHAR(50) DEFAULT 'fixed'
        CHECK (reward_qty_strategy IN ('fixed', 'all_purchased_items', 'multiply_by_condition_qty')), -- TODO: remove the third?
    fixed_quantity INT, -- Used only if strategy is 'fixed'
    -- TODO: maybe not: matching_condition_class_id INT REFERENCES product_classes(id), -- Used if scaling by a specific class

    -- Discount Evaluation
    discount_type VARCHAR(50) NOT NULL CHECK (discount_type IN ('percent_off', 'fixed_price')),
    discount_value DECIMAL(10, 2) NOT NULL,

    -- Enforce that a reward targets either a specific SKU or an entire Class
    CONSTRAINT chk_reward_target CHECK (
        (product_id IS NOT NULL AND product_class_id IS NULL) OR
        (product_id IS NULL AND product_class_id IS NOT NULL)
    )
);
*/

type Sale struct { // TODO: THIS IS A LINE ITEM! SHOULD ONLY CONTAIN ONE ITEM! OR MAYBE A MULTIPLE OF THE SAME ITEM FOR THE CASE OF PLUGS??
	AlternateCollectionIdField `bson:"inline"`
	// TODO: price????? other info?????
	SoldItems         []entryTypeIdCount
	CreationDateField `bson:"inline"` // This is sale date
	NotesField        `bson:"inline"`
	LastUpdatedField  `bson:"inline"`
	AclField          `bson:"inline"`
}

//func (s Sale) Blank() CollectionItem {
//	return &Sale{}
//}

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
	return env.IfNotProd(ctx, func() error {
		// If test agar batch does not exist, then create it
		testItem := &Sale{
			AlternateCollectionIdField: AlternateCollectionIdField{exAltId},
			CreationDateField:          CreationDateField{exampleTime},
			NotesField:                 NotesField{exampleNotes()},
			LastUpdatedField:           LastUpdatedField{exampleTime},
			AclField:                   allCanReadAcl(nil),
		}
		return addTestAltEntries(ctx, testItem)
	})
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

	ctx, now := request.UnixTime(r.Context())
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
	_, id, err := altCollIdFromRequest(r, w)
	if err != nil {
		return
	}
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
	ctx, db := Db(r)
	coll := db.Collection(SalesCollectionName)
	existing := Sale{}
	err = coll.FindOne(ctx, bson.M{IDfld: id}).Decode(&existing)
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

func deleteSaleHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "delete sale not allowed...", http.StatusForbidden)
}
