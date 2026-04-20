package rfid

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

func listEntriesHandlerInternal[T CollectionItem](ctx context.Context, updated bool, maxResults int, doStandardToo bool, temp T) (bs []byte, err error) {
	println("getting latest entries")
	latestEntries, err := getLastNEntries(ctx, true, maxResults, doStandardToo, temp)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		latestEntries = nil
	}
	if !doStandardToo {
		bs, err = json.Marshal(latestEntries)
		if err != nil {
			return nil, err
		}
		tempBs, errr := json.MarshalIndent(latestEntries, "", " ")
		if errr != nil {
			return nil, errr
		}
		println(string(tempBs)) // TODO: del
	} else {
		// TODO: do we want to also display repeats on standard entries?
		println("getting standard entries")
		stdEntries, err := getStandardEntries(ctx, temp)
		if err != nil {
			if !errors.Is(err, mongo.ErrNoDocuments) {
				return nil, err
			}
			stdEntries = nil
		}
		tempBs, errr := json.MarshalIndent(stdEntries, "", " ")
		if errr != nil {
			return nil, errr
		}
		println("standard entries: ", string(tempBs)) // TODO: del
		// Standard is filtered out from latest already
		outObj := map[string][]T{"standard": stdEntries, "recent": latestEntries}
		bs, err = json.Marshal(outObj)
		if err != nil {
			return nil, err
		}
		tempBs, errr = json.MarshalIndent(outObj, "", " ")
		if errr != nil {
			return nil, errr
		}
		println(string(tempBs)) // TODO: del
	}
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func ListEntriesHandler() http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		// TODO: DEPENDING ON VARIANT, EITHER DO LATEST OR LATEST AND STANDARD!!!!!

		var maxResults int = 10
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
		case "bag",
			"bags":
			bs, err = listEntriesHandlerInternal(r.Context(), true, maxResults, doStandardToo, &Bag{})
		case "box", "fruitingchamber", "chamber", "fruiting chamber",
			"boxes", "fruitingchambers", "chambers", "fruiting chambers":
			bs, err = listEntriesHandlerInternal[*FruitingChamber](r.Context(), true, maxResults, doStandardToo, &FruitingChamber{})
		case "jar", "grainjar", "grain jar",
			"jars", "grainjars", "grain jars":
			bs, err = listEntriesHandlerInternal[*GrainJar](r.Context(), true, maxResults, doStandardToo, &GrainJar{})
		case "lc", "liquidculture", "liquid culture",
			"lcs", "liquidcultures", "liquid cultures":
			bs, err = listEntriesHandlerInternal[*LiquidCulture](r.Context(), true, maxResults, doStandardToo, &LiquidCulture{})
		case "lcSyringe", "lcSyringes":
			bs, err = listEntriesHandlerInternal[*LcSyringe](r.Context(), true, maxResults, doStandardToo, &LcSyringe{})
		case "plugs", "plug", "peg", "pegs":
			bs, err = listEntriesHandlerInternal[*PlugsJar](r.Context(), true, maxResults, doStandardToo, &PlugsJar{})
		case "mss", "sporesyringe", "spore syringe", "multisporesyringe", "multi spore syringe",
			"msss", "sporesyringes", "spore syringes", "multisporesyringes", "multi spore syringes":
			bs, err = listEntriesHandlerInternal[*MSS](r.Context(), true, maxResults, doStandardToo, &MSS{})
		case "plate", "dish", "agarplate", "agar plate", "agardish", "agar dish", "petri", "petridish", "petri dish",
			"plates", "dishes", "agarplates", "agar plates", "agardishes", "agar dishes", "petris", "petridishes", "petri dishes":
			bs, err = listEntriesHandlerInternal[*Plate](r.Context(), true, maxResults, doStandardToo, &Plate{})
		case "slant", "slants":
			bs, err = listEntriesHandlerInternal[*Slant](r.Context(), true, maxResults, doStandardToo, &Slant{})
		case "stasistube", "stasis tube", "stasis", "tube",
			"stasistubes", "stasis tubes", "tubes":
			bs, err = listEntriesHandlerInternal[*StasisTube](r.Context(), true, maxResults, doStandardToo, &StasisTube{})
		case "agarbatch", "agar batch",
			"agarbatches", "agar batches":
			bs, err = listEntriesHandlerInternal[*AgarBatch](r.Context(), true, maxResults, doStandardToo, &AgarBatch{})
		case "agarrecipe", "agar recipe",
			"agarrecipes", "agar recipes":
			bs, err = listEntriesHandlerInternal[*AgarRecipe](r.Context(), true, maxResults, doStandardToo, &AgarRecipe{})
		case "fruit",
			"fruits":
			bs, err = listEntriesHandlerInternal[*Fruit](r.Context(), true, maxResults, doStandardToo, &Fruit{})
		case "jarrecipe", "jar recipe",
			"jarrecipes", "jar recipes":
			bs, err = listEntriesHandlerInternal[*JarRecipe](r.Context(), true, maxResults, doStandardToo, &JarRecipe{})
		case "lcrecipe", "lc recipe", "liquidculturerecipe", "liquid culture recipe",
			"lcrecipes", "lc recipes", "liquidculturerecipes", "liquid culture recipes":
			bs, err = listEntriesHandlerInternal[*LcRecipe](r.Context(), true, maxResults, doStandardToo, &LcRecipe{})
		case "pcrun", "pc run", "pressure cooker run", "pressure cooker", "pc", "pressurecooker", "run",
			"pcruns", "pc runs", "pressure cooker runs", "pressure cookers", "pcs", "pressurecookers", "runs":
			bs, err = listEntriesHandlerInternal[*PCRun](r.Context(), true, maxResults, doStandardToo, &PCRun{})
		case "project", "Projects":
			bs, err = listEntriesHandlerInternal[*Project](r.Context(), true, maxResults, doStandardToo, &Project{})
		case "sale", "sales":
			bs, err = listEntriesHandlerInternal[*Sale](r.Context(), true, maxResults, doStandardToo, &Sale{})
		case "species":
			bs, err = listEntriesHandlerInternal[*Species](r.Context(), true, maxResults, doStandardToo, &Species{})
		case "subspecies":
			bs, err = listEntriesHandlerInternal[*Subspecies](r.Context(), true, maxResults, doStandardToo, &Subspecies{})
		case "sporeprint", "spore print", "print",
			"sporeprints", "spore prints", "prints":
			bs, err = listEntriesHandlerInternal[*SporePrint](r.Context(), true, maxResults, doStandardToo, &SporePrint{})
		case "substrate", "substraterecipe", "substrate recipe",
			"substrates", "substraterecipes", "substrate recipes":
			bs, err = listEntriesHandlerInternal[*SubstrateRecipe](r.Context(), true, maxResults, doStandardToo, &SubstrateRecipe{})
		case "transfer", "xfer",
			"transfers", "xfers":
			bs, err = listEntriesHandlerInternal[*Transfer](r.Context(), true, maxResults, doStandardToo, &Transfer{})
		case "user", "users":
			bs, err = listEntriesHandlerInternal[*User](r.Context(), true, maxResults, doStandardToo, &User{})
		case "waterJar", "waterjar", "water jar", "Water jar", "sterilizedWater", "sterile water":
			bs, err = listEntriesHandlerInternal[*WaterJar](r.Context(), true, maxResults, doStandardToo, &WaterJar{})
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
	return http.HandlerFunc(handler)
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
	println("http write errors are currently unhandled! Err: " + err.Error())
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
			sb.WriteString(string(id.asBase58()) + "\n")
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
		sb.WriteString(string(item.DbId().asBase58()) + "\n")
	}
	println(sb.String())
}

func addTestAltEntries[T AltCollectionItem[U], U AltCollectionIdType](ctx context.Context, testItems ...T) error {
	if err := PrintAltCollectionItemIds("Test", testItems); err != nil {
		return err
	}
	_, err := newTxn(ctx, func(sessCtx mongo.SessionContext) (any, error) {
		coll := mongo.SessionFromContext(sessCtx).Client().Database(dbName).Collection(testItems[0].CollectionName())
		return coll.BulkWrite(ctx, sliceutils.Map(testItems, func(item T) mongo.WriteModel {
			return mongo.NewReplaceOneModel().SetReplacement(item).SetFilter(bson.M{"_id": item.DbId()}).SetUpsert(true)
		}))
		// TODO: do something with the result?
	})

	return err
}

func addBasicAltEntries[T AltCollectionItem[U], U AltCollectionIdType](ctx context.Context, testItems ...T) error {
	if err := PrintAltCollectionItemIds("Builtin", testItems); err != nil {
		return err
	}
	// TODO: txn or no?
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(testItems[0].CollectionName())
	for _, item := range testItems {
		switch id := item.IdValue().(type) {
		case AlternateCollectionId:
			println(id.asBase58())
		case string:
			println(id)
		default:
			return errors.New("invalid basic alt entry id, must be string or Alt Id")
		}
		_, err := coll.InsertOne(ctx, item, options.InsertOne())
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				continue
			}
			// TODO: update existing if needed?
			println("error adding basic alt entries: " + err.Error())
			return err
		}
	}

	return nil
}
