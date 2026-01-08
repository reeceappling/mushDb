package rfid

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/exp/slices"
	"net/http"
	"strconv"
)

var (
	_ AltCollectionItem = PCRun{}
	_ AltCollectionItem = AgarBatch{}
	_ AltCollectionItem = AgarRecipe{}
	_ AltCollectionItem = LcRecipe{}
	_ AltCollectionItem = JarRecipe{}
	_ AltCollectionItem = SubstrateRecipe{}
	_ AltCollectionItem = Transfer{}
	_ AltCollectionItem = Fruit{}      // TODO: main or alt?
	_ AltCollectionItem = Species{}    // this is a str id case
	_ AltCollectionItem = Subspecies{} // this is a str id case

	_ AltCollectionItem = Project{}
	_ AltCollectionItem = Sale{}
	//_ AltCollectionItem = User{}
	_ AltCollectionItem = SubstrateBatch{}
)

type AltCollectionItem interface {
	CollectionItem
}

func ListEntriesHandler() http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		// TODO: DEPENDING ON VARIANT, EITHER DO LATEST OR LATEST AND STANDARD!!!!!

		var maxResults int = 10             // TODO: ENSURE OK!
		requested := r.PathValue("variant") // TODO: ENSURE OK!
		doStandardToo := slices.Contains([]string{"agarRecipe", "jarRecipe", "lcRecipe", "substrateRecipe"}, requested)

		if maxNum := r.URL.Query().Get("n"); maxNum != "" { // TODO: ENSURE OK!
			n, err := strconv.Atoi(maxNum)
			if err != nil {
				http.Error(w, fmt.Sprintf(`param n must be a number, or nonexistent (defaults to %d)`, maxResults), http.StatusBadRequest)
				return
			}
			maxResults = n
		}

		// TODO: parallelize?
		outObj := map[string]any{}
		latestEntries, err := getLastNEntries(r.Context(), requested, true, maxResults)
		if err != nil {
			if !errors.Is(err, mongo.ErrNoDocuments) {
				code := http.StatusInternalServerError

				http.Error(w, err.Error(), code)
				return
			}
			latestEntries, err = json.Marshal([]string{})
			if err != nil {
				http.Error(w, "Unexpected latest marshalling error: "+err.Error(), http.StatusInternalServerError)
			}
		}
		outObj["latest"] = latestEntries
		// TODO: DEPENDING ON VARIANT, MAY RETURN HERE!

		// TODO: parallelize?
		// TODO: only for agarRecipe, jarRecipe, lcRecipe, substrateRecipe

		if doStandardToo {
			stdEntries, err := getStandardEntries(r.Context(), requested)
			if err != nil {
				if !errors.Is(err, mongo.ErrNoDocuments) {
					code := http.StatusInternalServerError
					http.Error(w, err.Error(), code)
					return
				}
				stdEntries, err = json.Marshal([]string{})
				if err != nil {
					http.Error(w, "Unexpected standard marshalling error: "+err.Error(), http.StatusInternalServerError)
				}
			}
			outObj["standard"] = stdEntries
		}

		out, err := json.Marshal(outObj)
		if err != nil {
			http.Error(w, "Unexpected output marshalling error: "+err.Error(), http.StatusInternalServerError)
		}
		if _, err = w.Write(out); err != nil {
			HandleHttpWriteError(err)
		}
	}
	return http.HandlerFunc(handler)
	//return GetPermsMiddleware(handler)
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
		if _, err = w.Write(entries); err != nil {
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
		if _, err = w.Write(entries); err != nil {
			HandleHttpWriteError(err)
		}
	}
}
