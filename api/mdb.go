package api

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
	if len(bs) == RfidByteSize {
		return MainCollectionId(bs), nil
	}
	if len(bs) < RfidByteSize {
		// TODO: ensure padding ok
		result := [RfidByteSize]byte{}
		for i, b := range bs {
			result[RfidByteSize-len(bs)+i] = b // TODO: validate ok
		}
		return result, nil
	}
	// TODO: what about too long?
	panic("too long not handled yet")
}

func (b58str Base58Str) toAltCollectionId() (AlternateCollectionId, error) {
	println("converting base58 string", b58str)
	out := [12]byte{}
	bs, err := b58str.Base2Bytes()
	if err != nil {
		return out, err
	}
	if len(bs) == 12 {
		return AlternateCollectionId(bs), nil
	}
	if len(bs) < 12 {
		// TODO: ensure padding ok
		result := [12]byte{}
		for i, b := range bs {
			result[12-len(bs)+i] = b // TODO: validate ok
		}
		return result, nil
	}
	// TODO: what about too long?
	panic("longer alts not handled yet")
	return AlternateCollectionId(bs), nil
}

// Tested, working
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

func (id BinaryCollectionId) AsBase58() Base58Str {
	b58bs, _ := Base2BytesToBase58(id.Bytes())
	return b58bs
}

func (id BinaryCollectionId) AsMainCollectionId() (MainCollectionId, error) {
	bs := []byte(id)
	if len(bs) != RfidByteSize {
		// TODO: PAD UP TO 8?
		result := [RfidByteSize]byte{}
		for i, b := range bs {
			result[RfidByteSize-len(bs)+i] = b // TODO: validate ok
		}

		return result, errors.Join(errors.New("bin as main failed"), ErrInvalidByteLength)
	}
	return MainCollectionId(bs), nil
}

func (id BinaryCollectionId) AsAltCollectionId() (AlternateCollectionId, error) {
	bs := []byte(id)
	if len(bs) != 12 {
		// TODO: PAD UP TO 12?
		result := [12]byte{}
		for i, b := range bs {
			result[12-len(bs)+i] = b // TODO: validate ok
		}
		return result, errors.Join(errors.New("bin as alt failed"), ErrInvalidByteLength)
	}
	return AlternateCollectionId(bs), nil
}

func (id BinaryCollectionId) Bytes() []byte {
	return []byte(id[:])
}

func (id BinaryCollectionId) ToBase58Bytes() []byte {
	return []byte(id.AsBase58())
}

func (id BinaryCollectionId) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, id.AsBase58())), nil
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

// newMcids is an internal channel holding new mainCollectionIds that have been pre-verified to not already exist
var newMcids chan MainCollectionId

// TODO: ctx should be cancellable here
// TODO: need to THOROUGHLY test this, ensure that it is generating properly, and that there are no duplicates across batches, and that it is properly cancelled when context is done, and that the channel is properly closed and drained when context is done
func StartGeneratingMCIDs(ctx context.Context, batchSize int) {
	if newMcids != nil {
		return
	}
	newMcids = make(chan MainCollectionId)
	go func() {
		// Defer closing and draining the channel
		defer func() {
			close(newMcids)
			channels.Drain(newMcids)
		}()

		// Initialize lastSet to ensure no duplicates across batches
		var (
			ids     []MainCollectionId
			err     error = nil
			lastSet       = utils.Set[string]{}
		)
		for {
			ids, lastSet, err = generateMainCollectionIds(ctx, batchSize, lastSet)
			if err != nil {
				println(err.Error())
				continue
			}
			for _, id := range ids {
				select {
				case <-ctx.Done():
					return
				case newMcids <- id:
					// Sent on channel, continue
					continue
				}
			}
		}
	}()
	return
}

func NextMainCollectionId() MainCollectionId { // TODO: use
	return <-newMcids
}
func NextMainCollectionIds(num int) []MainCollectionId { // TODO: use
	out := make([]MainCollectionId, num)
	for i := 0; i < num; i++ {
		out[i] = <-newMcids
	}
	return out
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
	return []byte(id.AsBase58())
}

func (id MainCollectionId) ToBinaryCollectionId() BinaryCollectionId {
	return BinaryCollectionId(id.dbIdStr())
}

func (id MainCollectionId) MarshalJSON() ([]byte, error) {
	//marshalling := [RfidByteSize]byte(id)
	//println("Marshalling " + string(marshalling[0:]))
	bs58 := id.AsBase58()
	//println(bs58) // TODO: CLEANUP
	out := []byte(`"` + string(bs58) + `"`)
	//println("Marshalled: " + string(out))
	return out, nil
}

func (id *MainCollectionId) UnmarshalJSON(bs []byte) error {
	var b58Str Base58Str
	// TODO: works with var b58Str string
	if err := json.Unmarshal(bs, &b58Str); err != nil {
		return err
	}
	val, err := b58Str.toMainCollectionId()
	if err != nil {
		return err
	}
	*id = val
	return nil
}

func (id MainCollectionId) dbIdStr() string { // Returns Most efficient string
	return string(id[:])
}
func (id MainCollectionId) AsBase58() Base58Str { // TODO: make sure that everywhere this is used, it is being used properly, and doesnt need to be in binary format
	return id.ToBinaryCollectionId().AsBase58() // TODO: make sure ok
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

func (id AlternateCollectionId) AsBase58() Base58Str {
	out, err := Base2BytesToBase58(id[:])
	if err != nil {
		panic("Error getting AltCollId str " + err.Error())
	}

	return out
}

func (id AlternateCollectionId) base58Bytes() []byte {
	return []byte(id.AsBase58())
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

// TODO: auth mechanisms https://www.mongodb.com/docs/drivers/go/current/fundamentals/auth/
func NewMongoDbClient(ctx context.Context, usern, pass, dbHostName string, dbPort int) (context.Context, *mongo.Client, error) {
	hostAndPort := dbHostName
	if dbPort != 0 && dbPort != 27017 {
		hostAndPort = fmt.Sprintf("%s:%d", dbHostName, dbPort)
	}

	//uri := fmt.Sprintf("mongodb://%s", hostAndPort)
	//uri := fmt.Sprintf("mongodb://%s:%s@%s", usern, pass, hostAndPort) // TODO: NAME OF DB
	hostPortAndParams := fmt.Sprintf("%s/?authSource=admin&replicaSet=rs0", hostAndPort)
	uri := fmt.Sprintf("mongodb://%s:%s@%s", usern, pass, hostPortAndParams)
	//uri := fmt.Sprintf("mongodb://%s:%s@%s", usern, pass, hostAndPort) // TODO: NAME OF DB 	// TODO: deleteMe

	println("trying to connect to database", usern, pass)
	opts := options.Client().
		ApplyURI(uri).
		SetDirect(true).

		//SetHosts([]string{hostAndPort}).
		SetServerSelectionTimeout(time.Second * 20).
		//SetAuth(options.Credential{Username: usern, Password: pass}). // TODO: get rid of?
		//SetAuth(options.Credential{Username: usern, Password: pass}).
		//SetAppName("mainApi").
		//SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1)).
		SetConnectTimeout(10 * time.Second). // TODO: no?
		SetTimeout(15 * time.Second)         // TODO: no?
	// TODO: ANY MORE?
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return ctx, nil, errors.Join(errors.New("failed to connect to db"), err)
	}
	connOk := false
	for i := 0; i < 2; i++ {
		println(fmt.Sprintf(`Testing connection attempt no.%d for user %s at %s`, i, usern, uri)) // TODO:" no uri
		err = client.Ping(ctx, nil)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		connOk = true
		break
	}
	if !connOk {
		errConn := errors.New("Ping failed to " + hostPortAndParams + ". " + err.Error())
		println("Ping failed to "+uri, err.Error()) // TODO: del
		return ctx, nil, errors.Join(errConn, err)
	}
	println("Client connected to db at " + hostPortAndParams)
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

func DbFrom(ctx context.Context) *mongo.Database {
	return GetMongoClient(ctx).Database(dbName)
}

func chanWithPostSendHook[T any](out chan<- T, afterFinalSend func(T)) (inpC chan<- T) {
	inp, out := make(chan T), make(chan T)
	go func() {
		defer close(out)
		for val := range inp {
			out <- val
			// Once validated sent, do the onSend handler
			afterFinalSend(val)
		}
	}()
	return inp
}

func generateMainCollectionIds(ctx context.Context, n int, lastSet utils.Set[string]) ([]MainCollectionId, utils.Set[string], error) {
	client := GetMongoClient(ctx)
	out := make([]MainCollectionId, 0, n)
	found := utils.Set[string]{}
	for ctFound := 0; ctFound < n; {
		select {
		case <-ctx.Done():
			// Break if context complete
			return nil, found, ctx.Err()
		default:
			newId := randomRFID(RfidByteSize)
			idStr := string(newId[:])
			if found.Contains(idStr) || lastSet.Contains(idStr) {
				continue
			}

			// get these ids from the map collection???
			err := client.Database(dbName).Collection(idMapCollectionName).FindOne(ctx, BsonFindFilter("_id", newId)).Err()
			//err := client.Database(dbName).Collection(collectionName).FindOne(ctx, BsonFindFilter("_id", newId)).Err()
			if err != nil {
				if errors.Is(err, mongo.ErrNoDocuments) {
					found.Add(string(newId[:]))
					ctFound++ // Increment num valid found
					out = append(out, MainCollectionId(newId))
				} else {
					return out, found, errors.Join(err, errors.New("failed to generate mainCollectionId"))
				}
			}
			// Item exists already, continue loop and do not add
		}
	}
	return out, found, nil
}

func getAllEntries[T CollectionItem](ctx context.Context, temp T) ([]T, error) {
	findBson := bson.D{{}}
	sortField := "$natural"
	// TODO: pagination?
	opts := options.Find().
		SetSort(bson.D{{Key: sortField, Value: -1}}) // Descending (latest first) // TODO: ensure -1 works with natural and that natural works with non-default ID types
	cursor, err := DbFrom(ctx).
		Collection(temp.CollectionName()).
		Find(ctx, findBson, opts)
	if err != nil {
		return nil, err
	}
	return getCollectionItemsFromCursor[T](ctx, cursor, nil)
}

func getLastNEntries[T CollectionItem](ctx context.Context, updated bool, nresults int, filterOutStandard bool, temp T) ([]T, error) {
	findBson := bson.D{{}}
	if filterOutStandard {
		findBson = BsonFindFilter("standard", false)
	}
	sortField := "$natural"
	if updated {
		sortField = "lastUpdated"
	}
	// TODO: pagination?
	opts := options.Find().
		//SetLimit(int64(nresults)). // no limit because user can be unable to view some items
		SetSort(bson.D{{Key: sortField, Value: -1}}) // Descending (latest first) // TODO: ensure -1 works with natural and that natural works with non-default IDs
	//opts.SetHint() // TODO: figure out if we need this (https://www.mongodb.com/docs/manual/reference/method/cursor.hint/#mongodb-method-cursor.hint)
	cursor, err := DbFrom(ctx).
		Collection(temp.CollectionName()).
		Find(ctx, findBson, opts)
	if err != nil {
		return nil, err
	}
	return getCollectionItemsFromCursor[T](ctx, cursor, &nresults)
}

func FindItemTypeForId(ctx context.Context, id MainCollectionId) (MainCollectionItem, error) {
	mapEntry := &idMapEntry{}
	err := DbFrom(ctx).Collection(idMapCollectionName).FindOne(ctx, bson.M{"_id": id}).Decode(mapEntry)
	if err != nil {
		return nil, err
	}
	return typeForEntryType(mapEntry.EntryType)
}

type InMemoryCache[T any] struct {
	TTL   time.Duration
	items []T
}

type InMemoryCacheItem[T any] struct {
	expiry time.Time
}

// TODO: CHANGESTREAMS to track most recently used/updated items in each collection!
// TODO: also want to track recipe ids and names in a cache, and invalidate the cache when additions happen

//type idNamePairCache struct {
//	*sync.RWMutex
//	timeout time.Time
//	IdNamePairArray
//}
//type IdNamePairArray struct {
//	names []string
//	ids   []string
//}
//func (arr IdNamePairArray) asMap() map[string]string{
//	out := map[string]string{}
//	for i, id := range arr.ids {
//		out[id] = arr.names[i]
//	}
//	return out
//}
//
//type RecipeCache struct {
//	Agar      *idNamePairCache
//	Jar       *idNamePairCache ``
//	LC        *idNamePairCache ``
//	Substrate *idNamePairCache ``
//}
//
//func (cache RecipeCache) GetAgarRecipes(ctx context.Context) map[string]string {
//	now := time.Now()
//	subcache := cache.Agar
//	if cache.Agar.timeout.Before(now) {
//		// TODO: go get them from the db
//		// TODO: update the cache
//	} else {
//		subcache.RLock()
//		defer subcache.RUnlock()
//		return cache.Agar.IdNamePairArray.asMap()
//	}
//}

var GetPageForIdHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "no id provided", http.StatusBadRequest)
		return
	}
	out := &idMapEntry{}
	err := DbFrom(ctx).
		Collection(idMapCollectionName).FindOne(ctx, BsonFindFilter("_id", id)).Decode(out)
	if err != nil {
		http.Error(w, "failed to get item by id: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write([]byte(fmt.Sprintf(`%s/%s`, out.EntryType, out.Id.AsBase58())))
	handleWriteErr(err, w)
}

var CreateHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	endpt := r.PathValue("variant")
	handle, exists := map[string]http.HandlerFunc{
		"agarBatch":       createAgarBatchHandler,
		"agarRecipe":      createAgarRecipeHandler,
		"bag":             createBagHandler,
		"lc":              createLiquidCultureHandler,
		"lcSyringe":       createSyringeHandler,
		"lcRecipe":        createLcRecipeHandler,
		"fruit":           createFruitHandler,
		"fruitingChamber": createFruitingChamberHandler,
		"grainBatch":      createGrainBatchHandler,
		"jar":             createJarHandler,
		"jarRecipe":       createJarRecipeHandler,
		"mss":             createMssHandler,
		"pcRun":           createPcRunHandler,
		"plate":           createPlateHandler,
		"plugs":           createPlugsHandler,
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
		"waterJar": createWaterJarHandler,
	}[endpt]
	if !exists {
		http.Error(w, "no handler for endpoint: "+endpt, http.StatusBadRequest)
		return
	}
	handle(w, r)
}

func GetPermsMiddleware(handler http.HandlerFunc) http.Handler {

	return handler // TODO: replace with old GetPermsMiddleware once perms are reenabled
}

func DenyGuestMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resolvedPerms, err := GetAuthInfo(r.Context()) // TODO: is err still needed here?
		if err != nil {
			http.Error(w, "failed to load permissions", http.StatusInternalServerError)
			return
		}
		if resolvedPerms.isGuest() {
			http.Error(w, "guest users cannot utilize this endpoint", http.StatusUnauthorized) // TODO: forbidden?
			return
		}
		handler.ServeHTTP(w, r)
	})

}

var ImportHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	endpt := r.PathValue("endpt")
	handler, exists := map[string]http.HandlerFunc{
		"bag":             importBagHandler,
		"lc":              importLiquidCultureHandler,
		"lcSyringe":       importLcSyringeHandler,
		"plugs":           importPlugsHandler,
		"fruit":           importFruitHandler,
		"fruitingChamber": importFruitingChamberHandler,
		"jar":             importJarHandler,
		"mss":             importMssHandler,
		"plate":           importPlateHandler,
		"slant":           importSlantHandler,
		"sporePrint":      importSporePrintHandler,
		"sporeSwab":       importSporeSwabHandler,
		"stasisTube":      importStasisTubeHandler,
		//"waterJar":      importWaterJarHandler, // TODO: import water jar handler?

	}[endpt]
	if !exists {
		http.Error(w, "no import handler for endpoint: "+endpt, http.StatusBadRequest)
		return
	}
	handler.ServeHTTP(w, r)
}

var UpdateHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	println("hit updateById handler")
	resolvedPerms, err := GetResolvedUserPerms(r.Context())
	if err != nil {
		println("failed to load permissions")
		http.Error(w, "failed to load permissions", http.StatusInternalServerError)
		return
	}
	//bs, _ := json.Marshal(resolvedPerms)  // TODO; del
	//println("resolved perms", string(bs)) // TODO: del
	if resolvedPerms.isGuest() { // TODO: reenable
		println("guest tried to edit")
		http.Error(w, "guest users cannot edit entries", http.StatusForbidden)
		return
	}
	endpt := r.PathValue("endpt")
	handler, exists := map[string]http.HandlerFunc{
		"agarBatch":       updateAgarBatchHandler,
		"agarRecipe":      updateAgarRecipeHandler,
		"bag":             updateBagHandler,
		"grainBatch":      updateGrainBatchHandler,
		"lc":              updateLiquidCultureHandler,
		"lcRecipe":        updateLcRecipeHandler,
		"lcSyringe":       updateSyringeHandler,
		"plugs":           updatePlugsHandler,
		"fruit":           updateFruitHandler,
		"fruitingChamber": updateFruitingChamberHandler,
		"jar":             updateJarHandler,
		"jarRecipe":       updateJarRecipeHandler,
		"mss":             updateMssHandler,
		"pcRun":           updatePcRunHandler,
		"plate":           updatePlateHandler,
		"plug":            updatePlugsHandler,
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
		"waterJar": updateWaterJarHandler,
	}[endpt]
	if !exists {
		http.Error(w, "no handler for endpoint: "+endpt, http.StatusBadRequest)
	}
	handler.ServeHTTP(w, r)
}

func dbErr(w http.ResponseWriter, txt string, status int) {
	println("txnErr " + txt)
	http.Error(w, txt, status)
	return
}
