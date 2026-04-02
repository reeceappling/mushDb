package rfid

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/itchyny/base58-go"
	"github.com/reeceappling/goUtils/v2/utils"
	"github.com/reeceappling/goUtils/v2/utils/channels"
	sliceutils "github.com/reeceappling/goUtils/v2/utils/slices"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
	"math/big"
	"net/http"
	"strconv"
	"time"
)

const (
	mongoClientContextKey = "mongoClient"
)

var (
	ErrInvalidByteLength = errors.New("invalid id bytes length")
)

type Base58Str string

func (b58str Base58Str) ToBinaryCollectionId() (BinaryCollectionId, error) {
	bs, err := b58str.Base2Bytes() // TODO: ensure this works fine for short values (ex: small base58 string should still output 12bytes for altId)
	if err != nil {
		return "", err
	}
	return BinaryCollectionId(bs), nil
}

// ConvertBase10StringToLittleEndianBytes converts a base-10 string to little-endian bytes.
func convertBase10StringToLittleEndianBytes(base10 string) ([]byte, error) { // TODO: AI-GENERATED, GO OVER IT
	// Parse the base-10 string into a big.Int.
	n := new(big.Int)
	_, ok := n.SetString(base10, 10)
	if !ok {
		return nil, fmt.Errorf("invalid base-10 string: %s", base10)
	}

	// Convert to bytes (big-endian by default).
	bigEndianBytes := n.Bytes()

	return sliceutils.ReverseOf(bigEndianBytes), nil
}

// Tested and works properly
func (b58str Base58Str) Base2Bytes() ([]byte, error) {
	baseTenBytes, err := base58.BitcoinEncoding.Decode([]byte(b58str))
	if err != nil {
		return baseTenBytes, errors.Join(errors.New("failed to btc decode"), err)
	}
	return convertBase10StringToLittleEndianBytes(string(baseTenBytes))
}

//func padBytes(src, dst []byte) {
//	if len(src) == len(dst) {
//		copy(dst, src)
//	}
//	if len(src) > len(dst) {
//		// TODO: handle
//	}
//	sizeDiff := len(dst) - len(src)
//
//	for i:=0; i<len(dst); i++ {
//		if i<sizeDiff {
//			dst[i]=0
//		}
//		dst[i]=src[i-sizeDiff]
//	}
//}

func (b58str Base58Str) toMainCollectionId() (MainCollectionId, error) {
	intVal, err := strconv.Atoi(string(b58str))
	if err == nil && intVal < 128 && intVal > 0 { // TODO: fix for zero???
		return [8]byte{0, 0, 0, 0, 0, 0, 0, uint8(intVal - 1)}, nil // TODO: validate works ok!
	}
	//if string(b58str) == "1" {
	//	return [8]byte{0, 0, 0, 0, 0, 0, 0, 0}, nil
	//} // TODO: fix this for
	out := [RfidByteSize]byte{}
	bs, err := b58str.Base2Bytes()
	if err != nil {
		return out, err
	}
	if len(bs) != RfidByteSize { // TODO: what about too long?
		// TODO: ensure padding ok
		result := [RfidByteSize]byte{}
		for i, b := range bs {
			result[RfidByteSize-len(bs)+i] = b // TODO: validate ok
		}
		return result, nil
	}
	return MainCollectionId(bs), nil
}

func (b58str Base58Str) toAltCollectionId() (AlternateCollectionId, error) {
	println("converting base58 string", b58str)
	out := [12]byte{}
	bs, err := b58str.Base2Bytes()
	if err != nil {
		return out, err
	}
	if len(bs) != 12 { // TODO: what about too long?
		// TODO: ensure padding ok
		result := [12]byte{}
		for i, b := range bs {
			result[12-len(bs)+i] = b // TODO: validate ok
		}
		return result, nil
	}
	return AlternateCollectionId(bs), nil
}

// Tested, working // TODO: remove comment
func Base2BytesToBase58(littleEndianBytes []byte) (Base58Str, error) {
	if len(littleEndianBytes)%4 != 0 {
		return "", errors.Join(errors.New("Base2BytesToBase58 failed"), ErrInvalidByteLength) // TODO: unnecesary?
	}
	baseTenStr := []byte(new(big.Int).SetBytes(sliceutils.ReverseOf(littleEndianBytes)).Text(10))
	encoded, err := base58.BitcoinEncoding.Encode(baseTenStr)
	println(string(baseTenStr), string(encoded)) // TODO; DEL
	if err != nil {
		return "", errors.Join(errors.New("base58 encoding fault"), err)
	}
	return Base58Str(encoded), nil
}

type BinaryCollectionId string // THIS IS ALWAYS IN BINARY FORMAT SERVERSIDE

func (id BinaryCollectionId) asBase58() Base58Str {
	b58bs, _ := Base2BytesToBase58(id.Bytes())
	return b58bs
}

func (id BinaryCollectionId) AsMainCollectionId() (MainCollectionId, error) {
	bs := []byte(id)
	if len(bs) != RfidByteSize {
		// TODO: PAD UP TO 8?
		return MainCollectionId{}, errors.Join(errors.New("bin as main failed"), ErrInvalidByteLength)
	}
	return MainCollectionId(bs), nil
}

func (id BinaryCollectionId) AsAltCollectionId() (AlternateCollectionId, error) {
	bs := []byte(id)
	if len(bs) != 12 {
		// TODO: PAD UP TO 12!
		return AlternateCollectionId{}, errors.Join(errors.New("bin as alt failed"), ErrInvalidByteLength)
	}
	return AlternateCollectionId(bs), nil
}

func (id BinaryCollectionId) Bytes() []byte {
	return []byte(id[:])
	//return []byte(id)
}

func (id BinaryCollectionId) ToBase58Bytes() []byte {
	return []byte(id.asBase58())
}

func (id BinaryCollectionId) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, id.asBase58())), nil
}

func (id *BinaryCollectionId) UnmarshalJSON(bs []byte) (err error) {
	*id, err = Base58Str(bs).ToBinaryCollectionId()
	return err
}

const RfidByteSize = 8

//var _ bson.ValueMarshaler = MainCollectionId{}
//var _ bson.ValueUnmarshaler = &MainCollectionId{}

// type MainCollectionId string // TODO: WILL ALWAYS BE THE BINARY ID!!!!
type MainCollectionId [RfidByteSize]byte

type mcidNode struct {
	val  MainCollectionId
	prev *mcidNode
	next *mcidNode
}

func startGeneratingIds(ctx context.Context, batchSize int) <-chan MainCollectionId {
	out := make(chan MainCollectionId)
	go func() {
		_ = <-ctx.Done()
		// TODO: DRAIN THE CHANNEL. First or last?
		close(out)
		channels.Drain(out)
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				ids, err := generateMainCollectionIds(ctx, batchSize)
				if err != nil {
					println(err.Error())
					continue
				}
				for _, id := range ids {
					out <- id
				}
			}
		}
	}()
	return out
}

func NextMainCollectionIdChan(ctx context.Context) <-chan MainCollectionId { // TODO: use
	val, ok := ctx.Value("newMcidChan").(<-chan MainCollectionId)
	if !ok {
		panic("no mainCollIdChan found on context")
	}
	return val
}

// TODO: ctx should be cancellable here
func StartGeneratingMCIDs(ctx context.Context, bufferSize int) context.Context {
	var final <-chan MainCollectionId = startGeneratingIds(ctx, bufferSize)
	return context.WithValue(ctx, "newMcidChan", final)
}

//// MarshalBSONValue implements the bson.ValueMarshaler interface
//func (id MainCollectionId) MarshalBSONValue() (bsontype.Type, []byte, error) {
//	// Format the id as a string for BSON storage
//	return bson.TypeString, bsoncore.AppendString(nil, string(id[:])), nil
//}
//
//// UnmarshalBSONValue implements the bson.ValueUnmarshaler interface
//func (id *MainCollectionId) UnmarshalBSONValue(t bsontype.Type, Data []byte) error {
//	if t != bson.TypeString {
//		return fmt.Errorf("invalid bson type %s, expected string", t.String())
//	}
//	// Read the BSON string and parse it back into a time.Time
//	s, _, ok := bsoncore.ReadString(Data)
//	if !ok {
//		return fmt.Errorf("invalid bson string value")
//	}
//	if len(s) != RfidByteSize {
//		return fmt.Errorf("invalid bson string value for main collection id. Length should be 8")
//	}
//	*id = MainCollectionId([]byte(s[0:RfidByteSize]))
//	return nil
//}

func (id MainCollectionId) base58Bytes() []byte {
	return []byte(id.asBase58())
}

func (id MainCollectionId) ToBinaryCollectionId() BinaryCollectionId {
	return BinaryCollectionId(id.dbIdStr())
}

func (id MainCollectionId) MarshalJSON() ([]byte, error) {
	//marshalling := [RfidByteSize]byte(id)
	//println("Marshalling " + string(marshalling[0:]))
	bs58 := id.asBase58()
	//println(bs58) // TODO: CLEANUP
	out := []byte(`"` + string(bs58) + `"`)
	//println("Marshalled: " + string(out))
	return out, nil
}

func (id *MainCollectionId) UnmarshalJSON(bs []byte) error {
	var b58Str string
	if err := json.Unmarshal(bs, &b58Str); err != nil {
		return err
	}
	val, err := Base58Str(b58Str).toMainCollectionId()
	if err != nil {
		return err
	}
	*id = val
	return nil
}

func (id MainCollectionId) dbIdStr() string { // Returns Most efficient string
	return string(id[:])
}
func (id MainCollectionId) asBase58() Base58Str { // TODO: make sure that everywhere this is used, it is being used properly, and doesnt need to be in binary format
	return id.ToBinaryCollectionId().asBase58() // TODO: make sure ok
}
func (id MainCollectionId) IdField() MainCollectionIdField { // Returns Most efficient string
	return MainCollectionIdField{id}
}

type AlternateCollectionId primitive.ObjectID // == [12]byte
func GuestEmail() string                      { return "guest" }

func (id AlternateCollectionId) MarshalBSONValue() (bsontype.Type, []byte, error) {
	// Format the id as a string for BSON storage
	return bson.TypeString, bsoncore.AppendString(nil, id.String()), nil
}

// UnmarshalBSONValue implements the bson.ValueUnmarshaler interface
func (id *AlternateCollectionId) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	if t != bson.TypeString {
		return fmt.Errorf("invalid bson type %s, expected string", t.String())
	}
	// Read the BSON string and parse it back into a time.Time
	s, _, ok := bsoncore.ReadString(data)
	if !ok {
		return fmt.Errorf("invalid bson string value")
	}
	if len(s) != 12 {
		return fmt.Errorf("invalid bson string value for alt collection id. Length should be 12")
	}
	*id = AlternateCollectionId([]byte(s[0:12]))
	return nil
}

func (id AlternateCollectionId) ToBinaryCollectionId() BinaryCollectionId {
	return BinaryCollectionId(id.String())
}

func (id AlternateCollectionId) MarshalJSON() ([]byte, error) { // Turn to base58 before outputting
	return []byte(`"` + string(id.base58Bytes()) + `"`), nil
}

func (id *AlternateCollectionId) UnmarshalJSON(bs []byte) error { // Turn from base58 to server-type (binary)
	var b58Str Base58Str
	if err := json.Unmarshal(bs, &b58Str); err != nil {
		return err
	}
	val, err := b58Str.toAltCollectionId()
	if err != nil {
		return err
	}
	*id = val
	return nil
}

func (id AlternateCollectionId) String() string {
	return string(id[:])
}

func (id AlternateCollectionId) asBase58() Base58Str {
	out, err := Base2BytesToBase58(id[:])
	if err != nil {
		panic("Error getting AltCollId str " + err.Error())
	}

	return out
}

func (id AlternateCollectionId) base58Bytes() []byte {
	return []byte(id.asBase58())
}
func (id AlternateCollectionId) asIdField() AlternateCollectionIdField {
	return AlternateCollectionIdField{id}
}

func newAlternateCollectionId() AlternateCollectionId {
	return AlternateCollectionId(primitive.NewObjectID())
}

/*
TODO: everything in here
single-chromosome-set (cant think of the word)
monoculture
source (cutting, MSS, Transfer, LC, Grain, TODO:etc)
*/

/*
use DATABASE_NAME
db.createCollection(name, options)
autoIndexId	Boolean	(Optional) If true, automatically create index on _id field.s Default value is false.
*/

// TODO: auth mechanisms https://www.mongodb.com/docs/drivers/go/current/fundamentals/auth/
func NewMongoDbClient(ctx context.Context, usern, pass, dbHostName string, dbPort int) (context.Context, *mongo.Client, error) {
	hostAndPort := dbHostName
	if dbPort != 0 && dbPort != 27017 {
		hostAndPort = fmt.Sprintf("%s:%d", dbHostName, dbPort)
	}

	//uri := fmt.Sprintf("mongodb://%s", hostAndPort)
	//uri := fmt.Sprintf("mongodb://%s:%s@%s", usern, pass, hostAndPort) // TODO: NAME OF DB
	uri := fmt.Sprintf("mongodb://%s:%s@%s", usern, pass, hostAndPort) // TODO: NAME OF DB 	// TODO: deleteMe

	// TODO: SET UP INITIAL USER IF USER DOES NOT EXIST!
	// TODO: THIS SHOULD BE DONE VIA: https://stackoverflow.com/questions/42912755/how-to-create-a-db-for-mongodb-container-on-start-up

	println(fmt.Sprintf(`trying to connect with %s %s`, usern, pass))
	opts := options.Client().
		ApplyURI(uri).
		SetDirect(true).

		//SetHosts([]string{hostAndPort}).
		SetServerSelectionTimeout(time.Second * 20).
		//SetAuth(options.Credential{Username: usern, Password: pass}). // TODO: get rid of?
		//SetAuth(options.Credential{Username: usern, Password: pass}).
		//SetAppName("mainApi").
		//SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1)).
		SetConnectTimeout(5 * time.Second). // TODO: no?
		SetTimeout(10 * time.Second)        // TODO: no?
	// TODO: ANY MORE?
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return ctx, nil, errors.Join(errors.New("failed to connect to db"), err)
	}
	connOk := false
	for i := 0; i < 5; i++ {
		println(fmt.Sprintf(`Testing connection attempt no.%d`, i))
		err = client.Ping(ctx, nil)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		connOk = true
		break
	}
	if !connOk {
		errConn := errors.New("Ping failed to " + uri)
		return ctx, nil, errors.Join(errConn, err)

	}
	println("Client connected to db at " + uri)
	return context.WithValue(ctx, mongoClientContextKey, client), client, nil
}

func mongoClientForURI(ctx context.Context, uri string) (context.Context, error) {
	// Use the SetServerAPIOptions() method to set the Stable API version to 1
	// To configure auth via URI instead of a Credential, use
	// "mongodb://email:password@localhost:27017".

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return ctx, errors.Join(errors.New("failed to connect to db"), err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		return ctx, errors.Join(errors.New("Ping failed to "+uri), err)
	}
	println("Client connected to db at " + uri)
	return context.WithValue(ctx, mongoClientContextKey, client), nil
}

func GetMongoClient(ctx context.Context) *mongo.Client {
	val, exists := ctx.Value(mongoClientContextKey).(*mongo.Client)
	if !exists {
		panic("no mongo client found on context")
	}
	return val
}

func generateMainCollectionIds(ctx context.Context, n int) ([]MainCollectionId, error) {
	client := ctx.Value(mongoClientContextKey).(*mongo.Client)
	out := make([]MainCollectionId, n)
	for i, _ := range out {
		found := utils.Set[string]{}
		for { // TODO: break eventually
			newId := randomRFID(RfidByteSize)
			if found.Contains(string(newId[:])) {
				continue
			}

			// get these ids from the map collection???
			err := client.Database(dbName).Collection(idMapCollectionName).FindOne(ctx, bson.D{{"_id", newId}}).Err()
			//err := client.Database(dbName).Collection(collectionName).FindOne(ctx, bson.D{{"_id", newId}}).Err()
			if err != nil {
				if errors.Is(err, mongo.ErrNoDocuments) {
					found.Add(string(newId[:]))
					out[i] = MainCollectionId(newId)
					break
				}
				return out, errors.Join(err, errors.New("failed to generate mainCollectionId"))
			}
		}
	}
	return out, nil
}

func newMainCollectionId(ctx context.Context) (MainCollectionId, error) {
	ids, err := generateMainCollectionIds(ctx, 1)
	if err != nil {
		return MainCollectionId{}, err
	}
	return ids[0], nil
}

//func getLastNEntriesForType[T]() { // TODO: do this
//
//}

func getLastNEntries[T CollectionItem](ctx context.Context, updated bool, nresults int, filterOutStandard bool, temp T) ([]T, error) {
	findBson := bson.D{{}}
	if filterOutStandard {
		findBson = bson.D{{"standard", false}}
	}
	sortField := "$natural"
	if updated {
		sortField = "lastUpdated"
	}
	// TODO: pagination?
	opts := options.Find().
		//SetLimit(int64(nresults)). // TODO: no limit because user can be unable to view some items
		SetSort(bson.D{{Key: sortField, Value: -1}}) // Descending (latest first) // TODO: ensure -1 works with natural
	//opts.SetHint() // TODO: figure out if we need this (https://www.mongodb.com/docs/manual/reference/method/cursor.hint/#mongodb-method-cursor.hint)
	cursor, err := ctx.Value(mongoClientContextKey).(*mongo.Client).
		Database(dbName).
		Collection(temp.CollectionName()).
		Find(ctx, findBson, opts)
	if err != nil {
		return nil, err
	}
	// TODO: ensure that user can read each item!!!!!!!!!!
	return getCollectionItemsFromCursor[T](ctx, cursor, &nresults)
}

func FindItemTypeForId(ctx context.Context, id MainCollectionId) (MainCollectionItem, error) {
	mapEntry := &idMapEntry{}
	err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(idMapCollectionName).FindOne(ctx, bson.M{"_id": id}).Decode(mapEntry)
	if err != nil {
		return nil, err
	}
	return typeForEntryType(mapEntry.EntryType)
}

func HandleCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resolvedPerms, err := GetResolvedUserPerms(r.Context())
		if err != nil {
			http.Error(w, "failed to load permissions", http.StatusInternalServerError)
			return
		}
		if resolvedPerms.isGuest() {
			http.Error(w, "guest users cannot create entries", http.StatusForbidden)
			return
		}
		endpt := r.PathValue("variant")
		//switch endpt {
		//case "agarBatch": createAgarBatchHandler(w,r)
		//case "agarRecipe": createAgarRecipeHandler(w,r)
		//case "bag": createBagHandler(w,r)
		//}
		handle, exists := map[string]http.HandlerFunc{
			"agarBatch":  createAgarBatchHandler,
			"agarRecipe": createAgarRecipeHandler,
			"bag":        createBagHandler,
			"lc":         createLiquidCultureHandler,
			"lcSyringe":  createSyringeHandler,
			//"plugs": createPlugsHandler, // TODO: FIX!
			"lcRecipe":        createLcRecipeHandler,
			"fruit":           createFruitHandler,
			"fruitingChamber": createFruitingChamberHandler,
			"jar":             createJarHandler,
			"jarRecipe":       createJarRecipeHandler,
			"mss":             createMssHandler,
			"pcRun":           createPcRunHandler,
			"plate":           createPlateHandler,
			"project":         createProjectHandler,
			"sale":            createSaleHandler,
			"slant":           createSlantHandler,
			"species":         createSpeciesHandler,
			"sporePrint":      createSporePrintHandler,
			"sporeSwab":       createSporeSwabHandler,
			"stasisTube":      createStasisTubeHandler,
			"subspecies":      createSubspeciesHandler,
			"substrateRecipe": createSubstrateRecipeHandler,
			"substrateBatch":  createSubstrateBatchHandler,
			"transfer":        createTransferHandler,
			//"user":"", // TODO: probably don't need
		}[endpt]
		if !exists {
			http.Error(w, "no handler for endpoint: "+endpt, http.StatusBadRequest)
			return
		}
		handle(w, r)
		//GetPermsMiddleware(handler).ServeHTTP(w, r)
	}
}

func GetPermsMiddleware(handler http.HandlerFunc) http.Handler {
	return handler // TODO: replace with old GetPermsMiddleware once perms are reenabled
}
func ImportHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resolvedPerms, err := GetResolvedUserPerms(r.Context())
		if err != nil {
			http.Error(w, "failed to load permissions", http.StatusInternalServerError)
			return
		}
		if resolvedPerms.isGuest() {
			http.Error(w, "guest users cannot import entries", http.StatusForbidden)
			return
		}
		endpt := r.PathValue("endpt")
		handler, exists := map[string]http.HandlerFunc{
			"bag":       importBagHandler,
			"lc":        importLiquidCultureHandler,
			"lcSyringe": importLcSyringeHandler,
			//"plugs": importPlugsHandler, // TODO: FIX!
			"fruit":           importFruitHandler,
			"fruitingChamber": importFruitingChamberHandler,
			"jar":             importJarHandler,
			"mss":             importMssHandler,
			"plate":           importPlateHandler,
			"slant":           importSlantHandler,
			"sporePrint":      importSporePrintHandler,
			"sporeSwab":       importSporeSwabHandler,
			"stasisTube":      importStasisTubeHandler,
		}[endpt]
		if !exists {
			http.Error(w, "no import handler for endpoint: "+endpt, http.StatusBadRequest)
		}
		GetPermsMiddleware(handler).ServeHTTP(w, r)
	}
}

func UpdateById() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		println("hit updateById handler")
		resolvedPerms, err := GetResolvedUserPerms(r.Context())
		if err != nil {
			println("failed to load permissions")
			http.Error(w, "failed to load permissions", http.StatusInternalServerError)
			return
		}
		bs, _ := json.Marshal(resolvedPerms)  // TODO; del
		println("resolved perms", string(bs)) // TODO: del
		//if resolvedPerms.isGuest() { // TODO: reenable
		//	println("guest tried to edit")
		//	http.Error(w, "guest users cannot edit entries", http.StatusForbidden)
		//	return
		//}
		endpt := r.PathValue("endpt")
		handler, exists := map[string]http.HandlerFunc{
			"agarBatch":  updateAgarBatchHandler,
			"agarRecipe": updateAgarRecipeHandler,
			"bag":        updateBagHandler,
			"lc":         updateLiquidCultureHandler,
			"lcRecipe":   updateLcRecipeHandler,
			"lcSyringe":  updateSyringeHandler,
			//"plugs": updatePlugsHandler, // TODO: FIX!
			"fruit":           updateFruitHandler,
			"fruitingChamber": updateFruitingChamberHandler,
			"jar":             updateJarHandler,
			"jarRecipe":       updateJarRecipeHandler,
			"mss":             updateMssHandler,
			"pcRun":           updatePcRunHandler,
			"plate":           updatePlateHandler,
			"project":         updateProjectHandler,
			"sale":            updateSaleHandler,
			"slant":           updateSlantHandler,
			"species":         updateSpeciesHandler,
			"sporePrint":      updateSporePrintHandler,

			"sporeSwab":       updateSporeSwabHandler,
			"stasisTube":      updateStasisTubeHandler,
			"subspecies":      updateSubspeciesHandler,
			"substrateRecipe": updateSubstrateRecipeHandler,
			"substrateBatch":  updateSubstrateBatchHandler,
			"transfer":        updateTransferHandler,
			//"user":             updateUserHandler,
		}[endpt]
		if !exists {
			http.Error(w, "no handler for endpoint: "+endpt, http.StatusBadRequest)
		}
		GetPermsMiddleware(handler).ServeHTTP(w, r)
	}
}

func dbErr(w http.ResponseWriter, txt string, status int) {
	println("txnErr " + txt)
	http.Error(w, txt, status)
	return
}
