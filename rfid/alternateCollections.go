package rfid

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"strconv"
)

var (
	_ AltCollectionItem = PCRun{}
	_ AltCollectionItem = AgarBatch{}
	_ AltCollectionItem = AgarRecipe{}
	_ AltCollectionItem = LCRecipe{}
	_ AltCollectionItem = JarRecipe{}
	_ AltCollectionItem = SubstrateRecipe{}
	_ AltCollectionItem = Transfer{}
	_ AltCollectionItem = Fruit{}
	_ AltCollectionItem = Species{}    // TODO: this is a str id case
	_ AltCollectionItem = Subspecies{} // TODO: this is a str id case
	_ AltCollectionItem = SporePrint{}
	_ AltCollectionItem = Project{}
	_ AltCollectionItem = Sale{}

	//_ AltCollectionItem = User{} // TODO: wtf to do with this
)

type AltCollectionItem interface {
	CollectionItem
}

type transferReason string // TODO: this (outgrew, contam, etc)

func ListEntriesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// TODO: DEPENDING ON VARIANT, EITHER DO LATEST OR LATEST AND STANDARD!!!!!

		var maxResults int = 10             // TODO: ENSURE OK!
		requested := r.PathValue("variant") // TODO: ENSURE OK!

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
			latestEntries, err = json.Marshal([]string{})
			if err != nil {
				http.Error(w, "Unexpected latest marshalling error: "+err.Error(), http.StatusInternalServerError)
			}
		}
		// TODO: DEPENDING ON VARIANT, MAY RETURN HERE!

		// TODO: parallelize?
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
		outObj := map[string]any{"standard": stdEntries, "latest": latestEntries}
		out, err := json.Marshal(outObj)
		if err != nil {
			http.Error(w, "Unexpected output marshalling error: "+err.Error(), http.StatusInternalServerError)
		}
		if _, err = w.Write(out); err != nil {
			HandleHttpWriteError(err)
		}
	}
}

func ListNewestEntriesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
}

func HandleHttpWriteError(err error) {
	println("http write errors are currently unhandled! Err: " + err.Error()) // TODO: THIS!!!!!!!!
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
