package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	sliceutils "github.com/reeceappling/goUtils/v2/utils/slices"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type AltCollectionIdType interface {
	AlternateCollectionId | string
}

type AltCollectionItem[T AltCollectionIdType] interface {
	CollectionItem
	DbId() T
}

type PermissionedAltCollectionItem[T AltCollectionIdType] interface {
	AltCollectionItem[T]
	Permissioned
}

type ListResponse[T any] struct {
	Latest   []T `json:"latest"`
	Standard []T `json:"standard,omitempty"`
}

func listEntriesHandlerInternal[T CollectionItem](ctx context.Context, updated bool, maxResults int, doStandardToo bool, temp T, allowDisposed bool) (bs []byte, err error) {
	latestEntries, err := getLastNEntries(ctx, updated, maxResults, doStandardToo, temp, allowDisposed)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			println("ERROR: listEntriesHandlerInternal found a non-ErrNoDocs", err) // TODO: this
			return nil, err
		}
		println("error getting entries: " + err.Error())
		latestEntries, err = []T{}, nil
	}
	//println(fmt.Sprintf("listEntriesHandlerInternal found %d entries", len(latestEntries))) // TODO: del
	if !doStandardToo {
		bs, err = json.Marshal(latestEntries)
	} else {
		outObj := map[string][]T{"recent": latestEntries}
		// TODO: do we want to also display repeats on standard entries?
		outObj["standard"], err = getStandardEntries(ctx, temp)
		if err != nil {
			if !errors.Is(err, mongo.ErrNoDocuments) {
				println("ERROR: listEntriesHandlerInternal found a non-ErrNoDocs", err) // TODO: this
				return nil, err
			}
			println("error getting std entries: " + err.Error()) // TODO: del
			outObj["standard"], err = []T{}, nil
		}
		// Standard is filtered out from latest already
		//println(fmt.Sprintf("listEntriesHandlerInternal found %d std entries", len(outObj["standard"]))) // TODO: del

		bs, err = json.Marshal(outObj)
	}
	if err != nil {
		return nil, err
	}
	return bs, nil
}
func listProjectsHandlerInternal(ctx context.Context, updated bool) (bs []byte, err error) {
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
		Collection(ProjectsCollectionName).
		Find(ctx, bson.D{{}}, opts)
	if err != nil {
		return nil, err
	}
	projects, err := getUserProjectsFromCursor(ctx, cursor, nil)
	if err != nil {
		return nil, err
	}
	return json.Marshal(projects)
}
func ListUsersHandler(ctx context.Context, removeGuests bool) ([]byte, error) {
	findBson := bson.D{{}}
	if removeGuests { // TODO: REMOVE ALL GUESTS FROM THE LIST
		findBson = bson.D{{"perms", bson.M{"$ne": nil}}} // TODO; ensure works!
	}
	// TODO: pagination?
	opts := options.Find().
		//SetLimit(int64(nresults)). // no limit because user can be unable to view some items
		SetSort(bson.D{{"_id", 1}}) // 1 = Ascending, -1 = Descending
	//opts.SetHint() // TODO: figure out if we need this (https://www.mongodb.com/docs/manual/reference/method/cursor.hint/#mongodb-method-cursor.hint)
	cursor, err := DbFrom(ctx).
		Collection(UserCollName).
		Find(ctx, findBson, opts)
	if err != nil {
		return nil, err
	}
	results := []*User{}
	err = cursor.All(ctx, &results)
	if err != nil {
		return nil, err
	}
	return json.Marshal(results)
}

var ListEntriesHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	// TODO: DEPENDING ON VARIANT, EITHER DO LATEST OR LATEST AND STANDARD!!!!!
	allowDisposed := true
	allowDisposed = r.URL.Query().Get("hideDisposed") != "true"
	var maxResults int = 10 // TODO: extend where needed?
	requested := r.PathValue("variant")
	doStandardToo := strings.Contains(requested, "Recipe") // "agarRecipe", "jarRecipe", "lcRecipe", "substrateRecipe"

	if maxNum := r.URL.Query().Get("n"); maxNum != "" {
		n, err := strconv.Atoi(maxNum)
		if err != nil {
			http.Error(w, fmt.Sprintf(`param n must be a number, or nonexistent (defaults to %d)`, maxResults), http.StatusBadRequest)
			return
		}
		maxResults = n
	}

	// TODO: parallelize?
	var bs []byte
	var err error
	switch strings.ToLower(requested) {
	case "agarbatch", "agar batch",
		"agarbatches", "agar batches":
		bs, err = listEntriesHandlerInternal[*AgarBatch](r.Context(), true, maxResults, doStandardToo, &AgarBatch{}, allowDisposed)
	case "agarrecipe", "agar recipe",
		"agarrecipes", "agar recipes":
		bs, err = listEntriesHandlerInternal[*AgarRecipe](r.Context(), true, maxResults, doStandardToo, &AgarRecipe{}, allowDisposed)
	case "bag",
		"bags":
		bs, err = listEntriesHandlerInternal[*Bag](r.Context(), true, maxResults, doStandardToo, &Bag{}, allowDisposed)
	case "fruit",
		"fruits":
		bs, err = listEntriesHandlerInternal[*Fruit](r.Context(), true, maxResults, doStandardToo, &Fruit{}, allowDisposed)
	case "fruitingchamber", "box", "chamber", "fruiting chamber",
		"boxes", "fruitingchambers", "chambers", "fruiting chambers":
		bs, err = listEntriesHandlerInternal[*FruitingChamber](r.Context(), true, maxResults, doStandardToo, &FruitingChamber{}, allowDisposed)
	case "grainbatch", "grainbatches":
		bs, err = listEntriesHandlerInternal[*GrainBatch](r.Context(), true, maxResults, doStandardToo, &GrainBatch{}, allowDisposed)
	case "jar", "grainjar", "grain jar",
		"jars", "grainjars", "grain jars":
		bs, err = listEntriesHandlerInternal[*GrainJar](r.Context(), true, maxResults, doStandardToo, &GrainJar{}, allowDisposed)
	case "jarrecipe", "jar recipe",
		"jarrecipes", "jar recipes":
		bs, err = listEntriesHandlerInternal[*JarRecipe](r.Context(), true, maxResults, doStandardToo, &JarRecipe{}, allowDisposed)
	case "lc", "liquidculture", "liquid culture",
		"lcs", "liquidcultures", "liquid cultures":
		bs, err = listEntriesHandlerInternal[*LiquidCulture](r.Context(), true, maxResults, doStandardToo, &LiquidCulture{}, allowDisposed)
	case "lcrecipe", "lc recipe", "liquidculturerecipe", "liquid culture recipe",
		"lcrecipes", "lc recipes", "liquidculturerecipes", "liquid culture recipes":
		bs, err = listEntriesHandlerInternal[*LcRecipe](r.Context(), true, maxResults, doStandardToo, &LcRecipe{}, allowDisposed)
	case "lcsyringe", "lcsyringes":
		bs, err = listEntriesHandlerInternal[*LcSyringe](r.Context(), true, maxResults, doStandardToo, &LcSyringe{}, allowDisposed)
	case "mss", "sporesyringe", "spore syringe", "multisporesyringe", "multi spore syringe",
		"msss", "sporesyringes", "spore syringes", "multisporesyringes", "multi spore syringes":
		bs, err = listEntriesHandlerInternal[*MSS](r.Context(), true, maxResults, doStandardToo, &MSS{}, allowDisposed)
	case "pcrun", "pc run", "pressure cooker run", "pressure cooker", "pc", "pressurecooker", "run",
		"pcruns", "pc runs", "pcRuns", "pressure cooker runs", "pressure cookers", "pcs", "pressurecookers", "runs":
		bs, err = listEntriesHandlerInternal[*PCRun](r.Context(), true, maxResults, doStandardToo, &PCRun{}, allowDisposed)
	case "plate", "dish", "agarplate", "agar plate", "agardish", "agar dish", "petri", "petridish", "petri dish",
		"plates", "dishes", "agarplates", "agar plates", "agardishes", "agar dishes", "petris", "petridishes", "petri dishes":
		bs, err = listEntriesHandlerInternal[*Plate](r.Context(), true, maxResults, doStandardToo, &Plate{}, allowDisposed)
	case "plugs", "plug", "peg", "pegs":
		bs, err = listEntriesHandlerInternal[*PlugsJar](r.Context(), true, maxResults, doStandardToo, &PlugsJar{}, allowDisposed)
	case "project", "projects":
		bs, err = listProjectsHandlerInternal(r.Context(), true) // TODO: true ok here? TEST HEAVILY!
		//bs, err = listEntriesHandlerInternal[*Project](r.Context(), true, maxResults, doStandardToo, &Project{}, allowDisposed)
	case "sale", "sales":
		bs, err = listEntriesHandlerInternal[*Sale](r.Context(), true, maxResults, doStandardToo, &Sale{}, allowDisposed)
	case "slant", "slants":
		bs, err = listEntriesHandlerInternal[*Slant](r.Context(), true, maxResults, doStandardToo, &Slant{}, allowDisposed)
	case "species":
		bs, err = listEntriesHandlerInternal[*Species](r.Context(), true, -1, doStandardToo, &Species{}, allowDisposed) // TODO: ensure showing all species (that can be viewed) instead of just 10
	case "sporeprint", "spore print", "print",
		"sporeprints", "spore prints", "prints":
		bs, err = listEntriesHandlerInternal[*SporePrint](r.Context(), true, maxResults, doStandardToo, &SporePrint{}, allowDisposed)
	case "sporeswab", "sporeswabs", "swab", "swabs":
		bs, err = listEntriesHandlerInternal[*SporeSwab](r.Context(), true, maxResults, doStandardToo, &SporeSwab{}, allowDisposed)
	case "stasistube", "stasis tube", "stasis", "tube",
		"stasistubes", "stasis tubes", "tubes":
		bs, err = listEntriesHandlerInternal[*StasisTube](r.Context(), true, maxResults, doStandardToo, &StasisTube{}, allowDisposed)
	case "subspecies":
		bs, err = listEntriesHandlerInternal[*Subspecies](r.Context(), true, -1, doStandardToo, &Subspecies{}, allowDisposed) // TODO: ensure showing all subspecies (that can be viewed) instead of just 10
	case "substrate", "substraterecipe", "substrate recipe",
		"substrates", "substraterecipes", "substrate recipes":
		bs, err = listEntriesHandlerInternal[*SubstrateRecipe](r.Context(), true, maxResults, doStandardToo, &SubstrateRecipe{}, allowDisposed)
	case "substratebatch", "substratebatches":
		bs, err = listEntriesHandlerInternal[*SubstrateBatch](r.Context(), true, maxResults, doStandardToo, &SubstrateBatch{}, allowDisposed)
	case "transfer", "xfer",
		"transfers", "xfers":
		bs, err = listEntriesHandlerInternal[*Transfer](r.Context(), true, maxResults, doStandardToo, &Transfer{}, allowDisposed)
	case "user", "users":
		bs, err = listEntriesHandlerInternal[*User](r.Context(), true, maxResults, doStandardToo, &User{}, allowDisposed)
	case "nonguest", "nonguests":
		bs, err = ListUsersHandler(r.Context(), true) // TODO: validate working!
		//bs, err = listEntriesHandlerInternal[*User](r.Context(), true, maxResults, doStandardToo, &User{})
	case "waterjar", "waterjars", "water jar", "water jars", "sterilizedwater", "sterilizedwaterjar", "sterilewater", "sterilewaterjar":
		bs, err = listEntriesHandlerInternal[*WaterJar](r.Context(), true, maxResults, doStandardToo, &WaterJar{}, allowDisposed)
	default:
		http.Error(w, errors.Join(ErrInvalidEntryType, errors.New("invalid collection input. Does not map to a collection name")).Error(), http.StatusBadRequest)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(bs)
	handleWriteErr(err, w)
}
var ListSubspeciesHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	spec, err := UrlDecodeString(r.PathValue("variant"))
	if err != nil {
		http.Error(w, "got bad species name. "+err.Error(), http.StatusBadRequest)
		return
	}
	findBson := BsonFindFilter("species", spec)
	sortField := "$natural" // TODO: FIX to sort for name!
	// TODO: pagination?
	opts := options.Find().
		SetSort(bson.D{{Key: sortField, Value: -1}}) // Descending (latest first) // TODO: ensure -1 works with natural
	//opts.SetHint() // TODO: figure out if we need this (https://www.mongodb.com/docs/manual/reference/method/cursor.hint/#mongodb-method-cursor.hint)
	cursor, err := DbFrom(ctx).
		Collection(SubspeciesCollectionName).
		Find(ctx, findBson, opts)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) { // TODO: NOT WORKING PROPERLY FOR BEECH!
			_, err = w.Write([]byte("[]"))
			handleWriteErr(err, w)
			return
		}
		http.Error(w, "failed to list subspecies. "+err.Error(), http.StatusInternalServerError)
		return
	}
	subspecs, err := getCollectionItemsFromCursor[Subspecies](ctx, cursor, nil, true)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			_, err = w.Write([]byte("[]"))
			handleWriteErr(err, w)
			return
		}
		http.Error(w, "failed to list subspecies after getting. "+err.Error(), http.StatusInternalServerError)
		return
	}
	bs, err := json.Marshal(subspecs)
	if err != nil {
		http.Error(w, "failed to marshal subspecies after getting. "+err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(bs)
	handleWriteErr(err, w)
}

//	func ListNewestEntriesHandler() http.Handler {
//		handler := func(w http.ResponseWriter, r *http.Request) {
//			var maxResults int = 10
//			requested := r.PathValue("variant")
//
//			createdOrUpdated, ok := map[string]string{
//				"":        "updated",
//				"updated": "updated",
//				"created": "created",
//			}[r.URL.Query().Get("createdOrUpdated")]
//			if !ok {
//				http.Error(w, "param createdOrUpdated must be created, updated, or nonexistent", http.StatusBadRequest)
//				return
//			}
//			if maxNum := r.URL.Query().Get("n"); maxNum != "" {
//				n, err := strconv.Atoi(maxNum)
//				if err != nil {
//					http.Error(w, fmt.Sprintf(`param n must be a number, or nonexistent (defaults to %d)`, maxResults), http.StatusBadRequest)
//					return
//				}
//				maxResults = n
//			}
//
//			entries, err := getLastNEntries(r.Context(), requested, createdOrUpdated == "updated", maxResults)
//			if err != nil {
//				code := http.StatusInternalServerError
//				if errors.Is(err, mongo.ErrNoDocuments) {
//					code = http.StatusNotFound
//				}
//				http.Error(w, err.Error(), code)
//				return
//			}
//			bs, err := json.Marshal(entries)
//			if err != nil {
//				http.Error(w, "Unexpected latest marshalling error: "+err.Error(), http.StatusInternalServerError)
//				return
//			}
//			if _, err = w.Write(bs); err != nil {
//				HandleHttpWriteError(err)
//			}
//		}
//		return http.HandlerFunc(handler)
//		//return GetPermsMiddleware(handler)
//	}
func HandleHttpWriteError(err error) {
	if err != nil {
		println("http write errors are currently unhandled! Err: " + err.Error())
	}
}

//
//func ListStandardEntriesHandler() http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.Request) {
//		requested := r.PathValue("variant")
//		entries, err := getStandardEntries(r.Context(), requested)
//		if err != nil {
//			code := http.StatusInternalServerError
//			if errors.Is(err, mongo.ErrNoDocuments) {
//				code = http.StatusNotFound
//			}
//			http.Error(w, err.Error(), code)
//			return
//		}
//		bs, err := json.Marshal(entries)
//		if err != nil {
//			http.Error(w, "Unexpected latest marshalling error: "+err.Error(), http.StatusInternalServerError)
//			return
//		}
//		if _, err = w.Write(bs); err != nil {
//			HandleHttpWriteError(err)
//		}
//	}
//}

func PrintAltCollectionItemIds[T AltCollectionItem[U], U AltCollectionIdType](Prefix string, testItems []T) error {
	if len(testItems) == 0 {
		println("Warning, no test items passed into PrintAltCollectionItemIds")
		return errors.New("no test items passed into PrintAltCollectionItemIds")
	}
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf(`%s %s entries:`, Prefix, testItems[0].CollectionName()) + "\n")
	for _, item := range testItems {
		switch id := item.IdValue().(type) {
		case AlternateCollectionId:
			sb.WriteString(string(id.AsBase58()) + "\n")
		case string:
			sb.WriteString(id + "\n")
		default:
			return errors.New("invalid basic alt entry id, must be string or Alt Id")
		}
	}
	println(sb.String())
	return nil
}

func PrintMainCollectionItemIds[T MainCollectionItem](Prefix string, testItems []T) {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf(`%s %s entries:`, Prefix, testItems[0].CollectionName()) + "\n")
	for _, item := range testItems {
		sb.WriteString(string(item.DbId().AsBase58()) + "\n")
	}
	println(sb.String())
}

func addTestAltEntries[T AltCollectionItem[U], U AltCollectionIdType](ctx context.Context, testItems ...T) error {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	_, err := newTxn(ctx, func(sessCtx mongo.SessionContext) (any, error) {
		defer wg.Done()
		coll := mongo.SessionFromContext(sessCtx).Client().Database(dbName).Collection(testItems[0].CollectionName())
		return coll.BulkWrite(ctx, sliceutils.Map(testItems, func(item T) mongo.WriteModel {
			// TODO: bson.M vs bson.D?
			return mongo.NewReplaceOneModel().SetReplacement(item).SetFilter(bson.M{"_id": item.DbId()}).SetUpsert(true)
		}))
		// TODO: do something with the result?
	})
	wg.Wait()
	if err := PrintAltCollectionItemIds("Test", testItems); err != nil {
		return err
	}
	return err
}

func addBasicAltEntries[T AltCollectionItem[U], U AltCollectionIdType](ctx context.Context, testItems ...T) error {
	if err := PrintAltCollectionItemIds("Builtin", testItems); err != nil {
		return err
	}
	// TODO: txn or no?
	coll := DbFrom(ctx).Collection(testItems[0].CollectionName())
	for _, item := range testItems {
		_, err := coll.InsertOne(ctx, item, options.InsertOne())
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				continue
			}
			println("error adding basic alt entries: " + err.Error())
			return err
		}

	}

	return nil
}
