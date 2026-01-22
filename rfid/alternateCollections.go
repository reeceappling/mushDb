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

func ListEntriesHandler() http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		// TODO: DEPENDING ON VARIANT, EITHER DO LATEST OR LATEST AND STANDARD!!!!!

		var maxResults int = 10                                // TODO: ENSURE OK!
		requested := r.PathValue("variant")                    // TODO: ENSURE OK!
		doStandardToo := strings.Contains(requested, "Recipe") // "agarRecipe", "jarRecipe", "lcRecipe", "substrateRecipe"

		if maxNum := r.URL.Query().Get("n"); maxNum != "" { // TODO: ENSURE OK!
			n, err := strconv.Atoi(maxNum)
			if err != nil {
				http.Error(w, fmt.Sprintf(`param n must be a number, or nonexistent (defaults to %d)`, maxResults), http.StatusBadRequest)
				return
			}
			maxResults = n
		}

		// TODO: parallelize?
		latestEntries, err := getLastNEntries(r.Context(), requested, true, maxResults)
		if err != nil {
			if !errors.Is(err, mongo.ErrNoDocuments) {
				code := http.StatusInternalServerError

				http.Error(w, err.Error(), code)
				return
			}
			latestEntries = nil
		}
		var bs []byte
		if !doStandardToo {
			bs, err = json.Marshal(latestEntries)
		} else {
			stdEntries, err := getStandardEntries(r.Context(), requested)
			if err != nil {
				if !errors.Is(err, mongo.ErrNoDocuments) {
					code := http.StatusInternalServerError
					http.Error(w, err.Error(), code)
					return
				}
				stdEntries = nil
			}
			outObj := map[string]any{"latest": latestEntries, "standard": stdEntries}
			outObj["latest"] = latestEntries
			outObj["standard"] = stdEntries
			bs, err = json.Marshal(outObj)
		}
		if err != nil {
			http.Error(w, "Unexpected latest marshalling error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = w.Write(bs)
		handleWriteErr(err, w)
	}
	return http.HandlerFunc(handler)
}

func ListNewestEntriesHandler() http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		var maxResults int = 10
		requested := r.PathValue("variant")

		createdOrUpdated, ok := map[string]string{
			"":        "updated",
			"updated": "updated",
			"created": "created",
		}[r.URL.Query().Get("createdOrUpdated")]
		if !ok {
			http.Error(w, "param createdOrUpdated must be created, updated, or nonexistent", http.StatusBadRequest)
			return
		}
		if maxNum := r.URL.Query().Get("n"); maxNum != "" {
			n, err := strconv.Atoi(maxNum)
			if err != nil {
				http.Error(w, fmt.Sprintf(`param n must be a number, or nonexistent (defaults to %d)`, maxResults), http.StatusBadRequest)
				return
			}
			maxResults = n
		}

		entries, err := getLastNEntries(r.Context(), requested, createdOrUpdated == "updated", maxResults)
		if err != nil {
			code := http.StatusInternalServerError
			if errors.Is(err, mongo.ErrNoDocuments) {
				code = http.StatusNotFound
			}
			http.Error(w, err.Error(), code)
			return
		}
		bs, err := json.Marshal(entries)
		if err != nil {
			http.Error(w, "Unexpected latest marshalling error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err = w.Write(bs); err != nil {
			HandleHttpWriteError(err)
		}
	}
	return http.HandlerFunc(handler)
	//return GetPermsMiddleware(handler)
}

func HandleHttpWriteError(err error) {
	println("http write errors are currently unhandled! Err: " + err.Error())
}

func ListStandardEntriesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requested := r.PathValue("variant")
		entries, err := getStandardEntries(r.Context(), requested)
		if err != nil {
			code := http.StatusInternalServerError
			if errors.Is(err, mongo.ErrNoDocuments) {
				code = http.StatusNotFound
			}
			http.Error(w, err.Error(), code)
			return
		}
		bs, err := json.Marshal(entries)
		if err != nil {
			http.Error(w, "Unexpected latest marshalling error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err = w.Write(bs); err != nil {
			HandleHttpWriteError(err)
		}
	}
}

func addTestAltEntries[T AltCollectionItem[U], U AltCollectionIdType](ctx context.Context, testItems ...T) error {
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(testItems[0].CollectionName())
	_, err := coll.BulkWrite(ctx, sliceutils.Map(testItems, func(item T) mongo.WriteModel {
		return mongo.NewReplaceOneModel().SetReplacement(item).SetFilter(bson.M{"_id": item.DbId()}).SetUpsert(true)
	}))
	// TODO: do something with the result?
	return err
}

func addBasicAltEntries[T AltCollectionItem[U], U AltCollectionIdType](ctx context.Context, testItems ...T) error {
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(testItems[0].CollectionName())
	for _, item := range testItems {
		_, err := coll.InsertOne(ctx, item, options.InsertOne())
		// TODO: do something with the result?
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
