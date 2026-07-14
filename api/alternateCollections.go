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

func listEntriesHandlerInternal[T CollectionItem, U any](ctx context.Context, updated bool, maxResults int, doStandardToo bool, temp T, disposed *bool, startAfterId *U) (bs []byte, err error) {
	latestEntries, err := getLastNEntries(ctx, updated, maxResults, doStandardToo, temp, disposed, startAfterId)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			//println("ERROR: listEntriesHandlerInternal found a non-ErrNoDocs", err) // TODO: this
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
		// TODO: do we want to also display repeats on standard entries? NO
		outObj["standard"], err = getStandardEntries(ctx, temp)
		if err != nil {
			if !errors.Is(err, mongo.ErrNoDocuments) {
				//println("ERROR: listEntriesHandlerInternal found a non-ErrNoDocs", err) // TODO: this
				return nil, err
			}
			//println("error getting std entries: " + err.Error()) // TODO: del
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
		SetSort(bson.D{{IDfld, 1}}) // 1 = Ascending, -1 = Descending
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
	// TODO: handle disposedFilter
	var disposedFilter *bool = nil
	disposedParam := r.URL.Query().Get("disposed")
	var tempDisposed = false
	switch disposedParam {
	case "only": // TODO: USE THIS!
		tempDisposed = true
		disposedFilter = &tempDisposed
	case "hide": // TODO: USE THIS!
		disposedFilter = &tempDisposed
	default:
		disposedFilter = nil
	}
	startAfterParam := r.URL.Query().Get("startAfter")
	if startAfterParam != "" {
		// TODO: THIS!
	}
	//allowDisposed = r.URL.Query().Get("hideDisposed") != "true" // TODO: REMOVE IF USED
	var maxResults = 30 // TODO: extend where needed?
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
	getAcid := func(sap string, w http.ResponseWriter) (*AlternateCollectionId, error) {
		if startAfterParam == "" {
			return nil, nil
		}
		acid, err := Base58Str(startAfterParam).toAltCollectionId()
		if err != nil {
			http.Error(w, "failed to convert base58 to altCollId: "+err.Error(), http.StatusBadRequest)
			return nil, err
		}
		return &acid, nil
	}
	getMcid := func(sap string, w http.ResponseWriter) (*MainCollectionId, error) {
		if startAfterParam == "" {
			return nil, nil
		}
		mcid, err := Base58Str(startAfterParam).ToMainCollectionId()
		if err != nil {
			http.Error(w, "invalid mainCollectionId: "+err.Error(), http.StatusBadRequest)
			return nil, err
		}
		return &mcid, nil
	}
	getParamStringDecoded := func(sap string, w http.ResponseWriter) (*string, error) {
		if startAfterParam == "" {
			return nil, nil
		}
		decoded, err := UrlDecodeString(startAfterParam)
		if err != nil {
			http.Error(w, "invalid altCollectionId: "+err.Error(), http.StatusBadRequest)
			return nil, err
		}
		return &decoded, nil

	}
	switch strings.ToLower(requested) {
	case "agarbatch", "agar batch",
		"agarbatches", "agar batches":
		acid, err := getAcid(startAfterParam, w)
		if err != nil {
			return // already wrote
		}
		bs, err = listEntriesHandlerInternal[*AgarBatch, AlternateCollectionId](r.Context(), true, maxResults, doStandardToo, &AgarBatch{}, disposedFilter, acid)
	case "agarrecipe", "agar recipe",
		"agarrecipes", "agar recipes":
		acid, err := getAcid(startAfterParam, w)
		if err != nil {
			return // already wrote
		}
		bs, err = listEntriesHandlerInternal[*AgarRecipe, AlternateCollectionId](r.Context(), true, maxResults, doStandardToo, &AgarRecipe{}, disposedFilter, acid)
	case "bag",
		"bags":
		mcid, err := getMcid(startAfterParam, w)
		if err != nil {
			return // already wrote
		}
		bs, err = listEntriesHandlerInternal[*Bag, MainCollectionId](r.Context(), true, maxResults, doStandardToo, &Bag{}, disposedFilter, mcid)
	case "fruit",
		"fruits":
		mcid, err := getMcid(startAfterParam, w)
		if err != nil {
			return // already wrote
		}
		bs, err = listEntriesHandlerInternal[*Fruit, MainCollectionId](r.Context(), true, maxResults, doStandardToo, &Fruit{}, disposedFilter, mcid)
	case "fruitingchamber", "box", "chamber", "fruiting chamber",
		"boxes", "fruitingchambers", "chambers", "fruiting chambers":
		mcid, err := getMcid(startAfterParam, w)
		if err != nil {
			return // already wrote
		}
		bs, err = listEntriesHandlerInternal[*FruitingChamber, MainCollectionId](r.Context(), true, maxResults, doStandardToo, &FruitingChamber{}, disposedFilter, mcid)
	case "grainbatch", "grainbatches":
		acid, err := getAcid(startAfterParam, w)
		if err != nil {
			return // already wrote
		}
		bs, err = listEntriesHandlerInternal[*GrainBatch, AlternateCollectionId](r.Context(), true, maxResults, doStandardToo, &GrainBatch{}, disposedFilter, acid)
	case "jar", "grainjar", "grain jar",
		"jars", "grainjars", "grain jars":
		mcid, err := getMcid(startAfterParam, w)
		if err != nil {
			return // already wrote
		}
		bs, err = listEntriesHandlerInternal[*GrainJar, MainCollectionId](r.Context(), true, maxResults, doStandardToo, &GrainJar{}, disposedFilter, mcid)
	case "jarrecipe", "jar recipe",
		"jarrecipes", "jar recipes":
		acid, err := getAcid(startAfterParam, w)
		if err != nil {
			return // already wrote
		}
		bs, err = listEntriesHandlerInternal[*JarRecipe, AlternateCollectionId](r.Context(), true, maxResults, doStandardToo, &JarRecipe{}, disposedFilter, acid)
	case "lc", "liquidculture", "liquid culture",
		"lcs", "liquidcultures", "liquid cultures":
		mcid, err := getMcid(startAfterParam, w)
		if err != nil {
			return // already wrote
		}
		bs, err = listEntriesHandlerInternal[*LiquidCulture, MainCollectionId](r.Context(), true, maxResults, doStandardToo, &LiquidCulture{}, disposedFilter, mcid)
	case "lcrecipe", "lc recipe", "liquidculturerecipe", "liquid culture recipe",
		"lcrecipes", "lc recipes", "liquidculturerecipes", "liquid culture recipes":
		acid, err := getAcid(startAfterParam, w)
		if err != nil {
			return // already wrote
		}
		bs, err = listEntriesHandlerInternal[*LcRecipe, AlternateCollectionId](r.Context(), true, maxResults, doStandardToo, &LcRecipe{}, disposedFilter, acid)
	case "lcsyringe", "lcsyringes":
		mcid, err := getMcid(startAfterParam, w)
		if err != nil {
			return // already wrote
		}
		bs, err = listEntriesHandlerInternal[*LcSyringe, MainCollectionId](r.Context(), true, maxResults, doStandardToo, &LcSyringe{}, disposedFilter, mcid)
	case "mss", "sporesyringe", "spore syringe", "multisporesyringe", "multi spore syringe",
		"msss", "sporesyringes", "spore syringes", "multisporesyringes", "multi spore syringes":
		mcid, err := getMcid(startAfterParam, w)
		if err != nil {
			return // already wrote
		}
		bs, err = listEntriesHandlerInternal[*MSS, MainCollectionId](r.Context(), true, maxResults, doStandardToo, &MSS{}, disposedFilter, mcid)
	case "pcrun", "pc run", "pressure cooker run", "pressure cooker", "pc", "pressurecooker", "run",
		"pcruns", "pc runs", "pcRuns", "pressure cooker runs", "pressure cookers", "pcs", "pressurecookers", "runs":
		acid, err := getAcid(startAfterParam, w)
		if err != nil {
			return // already wrote
		}
		bs, err = listEntriesHandlerInternal[*PCRun, AlternateCollectionId](r.Context(), true, maxResults, doStandardToo, &PCRun{}, disposedFilter, acid)
	case "plate", "dish", "agarplate", "agar plate", "agardish", "agar dish", "petri", "petridish", "petri dish",
		"plates", "dishes", "agarplates", "agar plates", "agardishes", "agar dishes", "petris", "petridishes", "petri dishes":
		mcid, err := getMcid(startAfterParam, w)
		if err != nil {
			return // already wrote
		}
		bs, err = listEntriesHandlerInternal[*Plate, MainCollectionId](r.Context(), true, maxResults, doStandardToo, &Plate{}, disposedFilter, mcid)
	case "plugs", "plug", "peg", "pegs":
		mcid, err := getMcid(startAfterParam, w)
		if err != nil {
			return // already wrote
		}
		bs, err = listEntriesHandlerInternal[*PlugsJar, MainCollectionId](r.Context(), true, maxResults, doStandardToo, &PlugsJar{}, disposedFilter, mcid)
	case "project", "projects":
		// TODO: projects list after!
		bs, err = listProjectsHandlerInternal(r.Context(), true) // TODO: true ok here? TEST HEAVILY!
		//bs, err = listEntriesHandlerInternal[*Project](r.Context(), true, maxResults, doStandardToo, &Project{}, disposedFilter)
	case "sale", "sales":
		acid, err := getAcid(startAfterParam, w)
		if err != nil {
			return // already wrote
		}
		bs, err = listEntriesHandlerInternal[*Sale, AlternateCollectionId](r.Context(), true, maxResults, doStandardToo, &Sale{}, disposedFilter, acid)
	case "slant", "slants":
		mcid, err := getMcid(startAfterParam, w)
		if err != nil {
			return // already wrote
		}
		bs, err = listEntriesHandlerInternal[*Slant, MainCollectionId](r.Context(), true, maxResults, doStandardToo, &Slant{}, disposedFilter, mcid)
	case "species":
		// Species are not paginated, we always return ALL OF THEM
		startAfterName, err := getParamStringDecoded(startAfterParam, w)
		if err != nil {
			return // already wrote
		}
		bs, err = listEntriesHandlerInternal[*Species, string](r.Context(), true, -1, doStandardToo, &Species{}, disposedFilter, startAfterName)
	case "sporeprint", "spore print", "print",
		"sporeprints", "spore prints", "prints":
		mcid, err := getMcid(startAfterParam, w)
		if err != nil {
			return // already wrote
		}
		bs, err = listEntriesHandlerInternal[*SporePrint, MainCollectionId](r.Context(), true, maxResults, doStandardToo, &SporePrint{}, disposedFilter, mcid)
	case "sporeswab", "sporeswabs", "swab", "swabs":
		mcid, err := getMcid(startAfterParam, w)
		if err != nil {
			return // already wrote
		}
		bs, err = listEntriesHandlerInternal[*SporeSwab, MainCollectionId](r.Context(), true, maxResults, doStandardToo, &SporeSwab{}, disposedFilter, mcid)
	case "stasistube", "stasis tube", "stasis", "tube",
		"stasistubes", "stasis tubes", "tubes":
		mcid, err := getMcid(startAfterParam, w)
		if err != nil {
			return // already wrote
		}
		bs, err = listEntriesHandlerInternal[*StasisTube, MainCollectionId](r.Context(), true, maxResults, doStandardToo, &StasisTube{}, disposedFilter, mcid)
	case "subspecies":
		// TODO: string ok?
		startAfterName, err := getParamStringDecoded(startAfterParam, w)
		if err != nil {
			return // already wrote
		}
		bs, err = listEntriesHandlerInternal[*Subspecies, string](r.Context(), true, -1, doStandardToo, &Subspecies{}, disposedFilter, startAfterName)
	case "substrate", "substraterecipe", "substrate recipe",
		"substrates", "substraterecipes", "substrate recipes":
		acid, err := getAcid(startAfterParam, w)
		if err != nil {
			return // already wrote
		}
		bs, err = listEntriesHandlerInternal[*SubstrateRecipe, AlternateCollectionId](r.Context(), true, maxResults, doStandardToo, &SubstrateRecipe{}, disposedFilter, acid)
	case "substratebatch", "substratebatches":
		acid, err := getAcid(startAfterParam, w)
		if err != nil {
			return // already wrote
		}
		bs, err = listEntriesHandlerInternal[*SubstrateBatch, AlternateCollectionId](r.Context(), true, maxResults, doStandardToo, &SubstrateBatch{}, disposedFilter, acid)
	case "transfer", "xfer",
		"transfers", "xfers":
		acid, err := getAcid(startAfterParam, w)
		if err != nil {
			return // already wrote
		}
		bs, err = listEntriesHandlerInternal[*Transfer, AlternateCollectionId](r.Context(), true, maxResults, doStandardToo, &Transfer{}, disposedFilter, acid)
	case "user", "users":
		// TODO: string ok?
		startAfterEmail, err := getParamStringDecoded(startAfterParam, w)
		if err != nil {
			return // already wrote
		}
		bs, err = listEntriesHandlerInternal[*User, string](r.Context(), true, maxResults, doStandardToo, &User{}, disposedFilter, startAfterEmail)
	case "nonguest", "nonguests":
		// TODO: start after handler?
		bs, err = ListUsersHandler(r.Context(), true) // TODO: validate working!
		//bs, err = listEntriesHandlerInternal[*User](r.Context(), true, maxResults, doStandardToo, &User{})
	case "waterjar", "waterjars", "water jar", "water jars", "sterilizedwater", "sterilizedwaterjar", "sterilewater", "sterilewaterjar":
		mcid, err := getMcid(startAfterParam, w)
		if err != nil {
			return // already wrote
		}
		bs, err = listEntriesHandlerInternal[*WaterJar, MainCollectionId](r.Context(), true, maxResults, doStandardToo, &WaterJar{}, disposedFilter, mcid)
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
		if errors.Is(err, mongo.ErrNoDocuments) {
			_, err = w.Write([]byte("[]"))
			handleWriteErr(err, w)
			return
		}
		http.Error(w, "failed to list subspecies. "+err.Error(), http.StatusInternalServerError)
		return
	}
	subspecs, err := getCollectionItemsFromCursor[Subspecies](ctx, cursor, nil)
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

func HandleHttpWriteError(err error) {
	if err != nil {
		println("http write errors are currently unhandled! Err: " + err.Error())
	}
}

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
			return mongo.NewReplaceOneModel().SetReplacement(item).SetFilter(bson.M{IDfld: item.DbId()}).SetUpsert(true)
		}))
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

func altCollIdFromRequest(r *http.Request, w http.ResponseWriter) (b58id Base58Str, id AlternateCollectionId, err error) {
	var idStr string
	idStr, err = UrlDecodeString(r.PathValue("id"))
	if err != nil {
		http.Error(w, "failed to url decode altCollId string: "+err.Error(), http.StatusInternalServerError)
		return
	}
	altCollId, err := StandardizeAltCollectionId(idStr)
	if err != nil {
		http.Error(w, "failed to standardize alt collection id: "+err.Error(), http.StatusBadRequest)
		return
	}
	b58id, id = altCollId.AsBase58(), *altCollId
	return
}

func finishCreateAlternateEntry[T CollectionItem](ctx context.Context, toInsert T, w http.ResponseWriter) {
	coll := DbFrom(ctx).Collection(toInsert.CollectionName())
	_, err := coll.InsertOne(ctx, toInsert)
	if err != nil {
		http.Error(w, "failed to insert one: "+err.Error(), http.StatusInternalServerError)
		return
	}
	bsOut, err := json.Marshal(toInsert)
	if err != nil {
		return
	}
	_, err = w.Write(bsOut)
	if err != nil {
		handleWriteErr(err, w)
	}
}

func finishImportMainCollectionEntry(ctx context.Context, toInsert MainCollectionItem, w http.ResponseWriter) {
	finishCreateMainCollectionEntry(ctx, toInsert, w)
}
