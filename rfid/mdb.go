package rfid

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/itchyny/base58-go"
	sliceutils "github.com/reeceappling/goUtils/v2/utils/slices"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"math/big"
	"net/http"
	"reflect"
	"slices"
	"time"
)

const (
	mongoClientContextKey = "mongoClient"
)

var (
	_ idAsDbStr = Fruit{}           // Sec // TODO: may not be needed
	_ idAsDbStr = LiquidCulture{}   // Prim
	_ idAsDbStr = GrainJar{}        // Prim
	_ idAsDbStr = Plate{}           // Prim
	_ idAsDbStr = Slant{}           // Prim
	_ idAsDbStr = StasisTube{}      // Prim
	_ idAsDbStr = Bag{}             // Prim
	_ idAsDbStr = FruitingChamber{} // Prim
	_ idAsDbStr = MSS{}
)

type idAsDbStr interface { // TODO: THIS?
	idAsStr() string
}
type Base58Str string // TODO: need a way to ensure we can get both b58 version and byte version

// TODO: make sure everywhere this is called is using it correctly
func (b58str Base58Str) Base58Bytes() []byte {
	return []byte(b58str)
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

func (b58str Base58Str) toMainCollectionId() (MainCollectionId, error) {
	out := [RfidByteSize]byte{}
	bs, err := b58str.Base2Bytes()
	if err != nil {
		return out, err
	}
	if len(bs) != RfidByteSize {
		return MainCollectionId{}, errors.New("FIXME") // TODO: this
	}
	return MainCollectionId(bs), nil
}

func (b58str Base58Str) toAltCollectionId() (alternateCollectionId, error) {
	out := [12]byte{}
	bs, err := b58str.Base2Bytes()
	if err != nil {
		return out, err
	}
	if len(bs) != 12 {
		return alternateCollectionId{}, errors.New("string was not an alt collection id") // TODO: this
	}
	return alternateCollectionId(bs), nil
}

// Tested, working
func Base2BytesToBase58(littleEndianBytes []byte) (Base58Str, error) {
	if len(littleEndianBytes)%4 != 0 {
		return "", errors.New("invalid tag data length") // TODO: unnecesary?
	}
	baseTenStr := new(big.Int).SetBytes(sliceutils.ReverseOf(littleEndianBytes)).Text(10)
	println(baseTenStr) // TODO: delete
	encoded, err := base58.BitcoinEncoding.Encode([]byte(baseTenStr))
	if err != nil {
		return "", errors.Join(errors.New("base58 encoding fault"), err)
	}
	return Base58Str(encoded), nil
}

const RfidByteSize = 8

type MainCollectionId [RfidByteSize]byte

func (id MainCollectionId) MarshalJSON() ([]byte, error) {
	marshalling := [RfidByteSize]byte(id)
	println("Marshalling " + string(marshalling[0:]))
	bs58 := id.asBase58()
	println(bs58) // TODO: DELETEME
	out := []byte(`"` + string(bs58) + `"`)
	println("Marshalled: " + string(out)) // TODO: DEL
	return out, nil
}

func (id *MainCollectionId) UnmarshalJSON(bs []byte) error {
	var b58Str string
	err := json.Unmarshal(bs, &b58Str)
	val, err := Base58Str(b58Str).toMainCollectionId()
	if err != nil {
		return err
	}
	*id = val
	return nil
}

func (id MainCollectionId) dbIdStr() string { // Returns Most efficient string
	bytes := [RfidByteSize]byte(id)
	return string(bytes[:])
}
func (id MainCollectionId) asBase58() Base58Str {
	arr := [RfidByteSize]byte(id)
	bs := arr[0:]
	println("bytes asBase58 preBase: " + string(bs))
	out, err := Base2BytesToBase58(bs)
	if err != nil {
		errOut := "Error getting MainCollId str " + err.Error()
		println(errOut)
		panic(errOut) // TODO: EW, GET RID OF
	}
	println("bytes asBase58 out: " + string(out))
	return out
}

type alternateCollectionId primitive.ObjectID // == [12]byte

func (id alternateCollectionId) MarshalJSON() ([]byte, error) {
	marshalling := [12]byte(id)
	println("Marshalling " + string(marshalling[0:]))
	bs58 := id.base58Bytes()
	println(bs58) // TODO: DELETEME
	out := []byte(`"` + string(bs58) + `"`)
	println("Marshalled: " + string(out)) // TODO: DEL
	return out, nil
}

func (id *alternateCollectionId) UnmarshalJSON(bs []byte) error {
	var b58Str string
	err := json.Unmarshal(bs, &b58Str)
	val, err := Base58Str(b58Str).toAltCollectionId()
	if err != nil {
		return err
	}
	*id = val
	return nil
}

func (id alternateCollectionId) String() string {
	return string(id[:])
}

func (id alternateCollectionId) base58() Base58Str {
	out, err := Base2BytesToBase58(id[:])
	if err != nil {
		panic("Error getting AltCollId str " + err.Error())
	}

	return out
}

func (id alternateCollectionId) base58Bytes() []byte {
	return []byte(id.base58())
}

func newAlternateCollectionId() alternateCollectionId {
	return alternateCollectionId(primitive.NewObjectID())
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
	//credential := options.Credential{
	//	Username: usern,
	//	Password: pass,
	//}
	//if len(authDb) > 0 {
	//	credential.AuthSource = authDb[0] // TODO: was "<authenticationDb>", defaults to "admin"
	//}
	//dbUri := fmt.Sprintf("mongodb://%s:%d", dbHostName, dbPort) // TODO: was "mongodb://<hostname>:<port>"
	//clientOpts := options.Client().ApplyURI(dbUri).SetAuth(credential)
	//client, err := mongo.Connect(context.TODO(), clientOpts)
	//if err != nil {
	//	return nil, err
	//}
	//return context.WithValue(ctx, mongoClientContextKey, client), nil
	//var DBName = "testDbName" // TODO: FIXME!
	// TODO: FIX DB HOSTNAME
	hostAndPort := dbHostName
	if dbPort != 0 {
		hostAndPort = fmt.Sprintf("%s:%d", dbHostName, dbPort)
	}

	//uri := fmt.Sprintf("mongodb://%s", hostAndPort)
	uri := fmt.Sprintf("mongodb://%s:%s@%s", usern, pass, hostAndPort) // TODO: NAME OF DB
	println("creating client to " + uri)                               // TODO: deleteMe

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
		SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1)).
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
		panic("Ping failed to " + uri + " ... " + err.Error())
	}
	println("Client connected to db at " + uri)
	return context.WithValue(ctx, mongoClientContextKey, client), client, nil

	// TODO: ? return mongoClientForURI(ctx, uri)
}

func mongoClientForURI(ctx context.Context, uri string) (context.Context, error) {
	// Use the SetServerAPIOptions() method to set the Stable API version to 1
	// To configure auth via URI instead of a Credential, use
	// "mongodb://user:password@localhost:27017".

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return ctx, errors.Join(errors.New("failed to connect to db"), err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		panic("Ping failed to " + uri)
	}
	println("Client connected to db at " + uri)
	return context.WithValue(ctx, mongoClientContextKey, client), nil
}

func GetMongoClient(ctx context.Context) *mongo.Client {
	return ctx.Value(mongoClientContextKey).(*mongo.Client) // TODO, ensure ok that this may not be set
}

func doTxn(ctx context.Context, txnFunc func(ctx mongo.SessionContext) (interface{}, error)) (interface{}, error) {
	client := ctx.Value(mongoClientContextKey).(*mongo.Client)

	// Starts a session on the client
	session, err := client.StartSession()
	if err != nil {
		return nil, errors.Join(errors.New("failed to start transaction session"), err)
	}
	// Defers ending the session after the transaction is committed or ended
	defer session.EndSession(ctx)

	txnOptions := options.Transaction().SetWriteConcern(writeconcern.Majority()) // TODO: other concerns?
	result, err := session.WithTransaction(ctx, txnFunc, txnOptions)
	if err != nil {
		newErr := errors.New("failed to execute transaction") // TODO: move
		err = errors.Join(err, newErr)
	}
	return result, err
}

func generateMainCollectionIds(ctx context.Context, n int) ([]MainCollectionId, error) {
	client := ctx.Value(mongoClientContextKey).(*mongo.Client)
	out := make([]MainCollectionId, n)
	for i, _ := range out {
		for { // TODO: break eventually
			newId := randomRFID(RfidByteSize)
			err := client.Database(dbName).Collection(mainCollectionName).FindOne(ctx, bson.D{{"_id", newId}}).Err()
			if err != nil {
				if errors.Is(err, mongo.ErrNoDocuments) {
					out[i] = MainCollectionId(newId)
					break
				}
				return out, errors.Join(err, errors.New("failed to generate mainCollectionId"))
			}
		}
	}
	return out, nil
}

func generateMainCollectionId(ctx context.Context) (MainCollectionId, error) {
	ids, err := generateMainCollectionIds(ctx, 1)
	if err != nil {
		return MainCollectionId{}, err
	}
	return ids[0], nil
}

func getLastNEntries(ctx context.Context, variant string, updated bool, nresults int) ([]byte, error) {
	entryType, err := entryTypeFor(variant)
	if err != nil {
		return nil, err
	}
	findBson := bson.D{{}}
	if etf := entryType.EntryTypeField(); etf != nil {
		findBson = bson.D{{"entryType", *etf}} // TODO: ensure ok?
	}
	sortField := "$natural"
	if updated {
		sortField = "lastUpdated"
	}
	// TODO: pagination?
	opts := options.Find().
		SetLimit(int64(nresults)).
		SetSort(bson.D{{Key: sortField, Value: -1}}) // Descending (latest first) // TODO: ensure -1 works with natural
	//opts.SetHint() // TODO: figure out if we need this (https://www.mongodb.com/docs/manual/reference/method/cursor.hint/#mongodb-method-cursor.hint)

	cursor, err := ctx.Value(mongoClientContextKey).(*mongo.Client).
		Database(dbName).
		Collection(entryType.CollectionName()).
		Find(ctx, findBson, opts)
	if err != nil {
		return nil, err
	}
	results, err := getCollectionItemsFromCursor(ctx, cursor, reflect.TypeOf(entryType))
	if err != nil {
		return nil, err
	}
	return json.Marshal(results)
}

func HandleCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !ableToModify(r.Context()) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		endpt := r.PathValue("endpt")
		handler, exists := map[string]http.HandlerFunc{
			"agarBatch":       createAgarBatchHandler,
			"agarRecipe":      createAgarRecipeHandler,
			"bag":             createBagHandler,
			"lc":              createLiquidCultureHandler,
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
			"stasisTube":      createStasisTubeHandler,
			"subspecies":      createSubspeciesHandler,
			"substrateRecipe": createSubstrateRecipeHandler,
			"transfer":        createTransferHandler,
			//"User":"", // TODO: probably don't need
		}[endpt]
		if !exists {
			http.Error(w, "no handler for endpoint: "+endpt, http.StatusBadRequest)
		}
		handler.ServeHTTP(w, r)
	}
}
func ImportHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !ableToModify(r.Context()) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		endpt := r.PathValue("endpt")
		handler, exists := map[string]http.HandlerFunc{
			"bag":             importBagHandler,
			"lc":              importLiquidCultureHandler,
			"fruit":           importFruitHandler,
			"fruitingChamber": importFruitingChamberHandler,
			"jar":             importJarHandler,
			"mss":             importMssHandler,
			"plate":           importPlateHandler,
			"slant":           importSlantHandler,
			"sporePrint":      importSporePrintHandler,
			"stasisTube":      importStasisTubeHandler,
		}[endpt]
		if !exists {
			http.Error(w, "no import handler for endpoint: "+endpt, http.StatusBadRequest)
		}
		handler.ServeHTTP(w, r)
	}
}

func UpdateById() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !ableToModify(r.Context()) { // TODO: remove all other uses of ableToModify?
			http.Error(w, "User does not have update permissions", http.StatusForbidden)
			return
		}
		endpt := r.PathValue("endpt")
		handler, exists := map[string]http.HandlerFunc{
			"agarBatch":       updateAgarBatchHandler,
			"agarRecipe":      updateAgarRecipeHandler,
			"bag":             updateBagHandler,
			"lc":              updateLiquidCultureHandler,
			"lcRecipe":        updateLcRecipeHandler,
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
			"stasisTube":      updateStasisTubeHandler,
			"subspecies":      updateSubspeciesHandler,
			"substrateRecipe": updateSubstrateRecipeHandler,
			"transfer":        updateTransferHandler,
			//"User":"", // TODO: probably don't need
		}[endpt]
		if !exists {
			http.Error(w, "no handler for endpoint: "+endpt, http.StatusBadRequest)
		}
		handler.ServeHTTP(w, r)
	}
}

func speciesIsSpecial(ctx context.Context, sp *string) bool { // TODO: rename, move?
	if sp == nil {
		return false
	}
	// TODO: check against all special species
	// TODO: Hold special species in map? slice?
	// TODO: update special species on db updates
	return false // TODO: CHANGE!
}

func userIsAdmin(ctx context.Context) bool {
	return GetAuthInfo(ctx).Opts.Contains(MaxAuthKey)
}

func ableToModify(ctx context.Context) bool {
	perms := GetAuthInfo(ctx)
	for _, key := range []string{MaxAuthKey, ChangeKey} {
		if perms.Opts.Contains(key) {
			return true
		}
	}
	return false
}

func setStringArrayIfUnequal(upd bson.D, new []string, current []string, key string) bson.D {
	out := upd
	if len(new) != len(current) {
		out = append(out, bson.E{"$set", bson.D{{key, new}}})
		return out
	}
	for i := 0; i < len(current); i++ {
		if !slices.Contains(new, current[i]) {
			out = append(out, bson.E{"$set", bson.D{{key, new}}})
			return out
		}
	}
	return out
}

func setProjectsIfUnequal(upd bson.D, new []string, current []string) bson.D {
	return setStringArrayIfUnequal(upd, new, current, "projects")
}
func setSalesIfUnequal(upd bson.D, new []alternateCollectionId, current []alternateCollectionId) bson.D {
	out := upd
	if len(new) != len(current) {
		out = append(out, bson.E{"$set", bson.D{{"sales", new}}})
		return out
	}
	for i := 0; i < len(current); i++ {
		if !slices.Contains(new, current[i]) {
			out = append(out, bson.E{"$set", bson.D{{"sales", new}}})
			return out
		}
	}
	return out
}
