package rfid

import (
	"context"
	"encoding/json"
	"github.com/reeceappling/goUtils/v2/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"net/url"
	sliceutils "slices"
	"time"
)

const subSpeciesCollectionName = "subspecies"

type Subspecies struct {
	Name        string   `bson:"_id" json:"_id"`
	Species     string   `bson:"species" json:"species"`
	Aliases     []string `bson:"aliases,omitempty" json:"aliases,omitempty"`
	Notes       []Note   `bson:"notes,omitempty" json:"notes,omitempty"`
	LastUpdated unixTime `bson:"lastUpdated" json:"lastUpdated"`
}

func (subsp Subspecies) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := subsp
	err := decodeItem(&out, encoded)
	return out, err
}

func (subsp Subspecies) clean() CollectionItem {
	out := subsp
	// TODO: Change name
	// TODO: change species
	// TODO: remove aliases
	// TODO: remove notes
	return out
}

func (subsp Subspecies) EntryTypeField() *string {
	return nil
}

func (subsp Subspecies) CollectionName() string {
	return subSpeciesCollectionName
}

type NewSubspeciesRequest struct {
	Name    string   `json:"name"`
	Aliases []string `json:"aliases,omitempty"`
	Species string   `json:"species"`
	Notes   []string `json:"notes,omitempty"`
}

func (req NewSubspeciesRequest) asSubspecies() Subspecies {
	return Subspecies{
		Name:    req.Name,
		Aliases: req.Aliases,
		Species: req.Species,
		Notes:   stringsToNotes(req.Notes, time.Now()),
	}
}

func initializeSubspecies(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(subSpeciesCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		newSimpleIndex("species", "species", false, false, false),
		//BackupSubspecies (likely no index) *string  `bson:"backupSubspecies,omitempty" json:"backupSubspecies,omitempty"`
		newSimpleIndex("aliases", "aliases", false, true, false),
		//Notes (no index) (maybe later with tags?)
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}

	inserted, updated := 0, 0
	// TODO: OTHER SUBSPECIES
	for _, subsp := range []Subspecies{
		// White Beech
		{
			Name:    "white beech",
			Species: "beech",
			Aliases: nil,
			Notes:   nil, // TODO: something to do with light?
		},
		// Brown Beech
		{
			Name:    "brown beech",
			Species: "beech",
			Aliases: nil,
			Notes:   nil, // TODO: something to do with light?
		},
	} {
		var existing Subspecies
		err = coll.FindOne(ctx, bson.D{{"_id", subsp.Name}}).Decode(&existing)
		if err != nil {
			if err != mongo.ErrNoDocuments {
				return err
			}
			// if not exists, add it to the db
			_, err = coll.InsertOne(ctx, subsp)
			if err != nil {
				return err
			}
			inserted++
			continue
		}
		// If exists, ensure it is the same as it was. Add notes if necessary
		update := false
		if subsp.Species != subsp.Species {
			update = true
		}
		finalAliases := utils.Set[string]{}
		finalAliases.Add(existing.Aliases...)
		finalAliases.Add(subsp.Aliases...)
		if len(finalAliases) != len(existing.Aliases) {
			update = true
			subsp.Aliases = finalAliases.ToSlice()
		}

		// Notes
		finalNotes := []Note{}
		copy(finalNotes, existing.Notes)
		for _, note := range subsp.Notes {
			if !sliceutils.Contains(finalNotes, note) {
				finalNotes = append(finalNotes, note)
				update = true
			}
		}
		subsp.Notes = finalNotes

		// Update if necessary
		if update {
			err = coll.FindOneAndReplace(ctx, bson.D{{"_id", subsp.Name}}, subsp).Err()
			if err != nil {
				return err
			}
			updated++
		}
	}
	//if inserted+updated > 0 { // TODO: ok?
	//	println(fmt.Sprintf(`Subspecies entries: inserted %d, updated %d`, inserted, updated))
	//}
	return nil
}

type createSubspeciesRequest struct {
	Name    string   `json:"name"`
	Species string   `json:"species"`
	Aliases []string `json:"aliases,omitempty"`
	Notes   []Note   `json:"notes,omitempty"`
}

func createSubspeciesHandler(w http.ResponseWriter, r *http.Request) {
	req := createSubspeciesRequest{}
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(subSpeciesCollectionName)
		_, err = coll.InsertOne(r.Context(), Subspecies{
			Name:        req.Name,
			Species:     req.Species, // TODO: ENSURE EXISTS!
			Aliases:     req.Aliases,
			Notes:       req.Notes,
			LastUpdated: unixTime(time.Now().UnixMilli()),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, nil
		}
		return w.Write([]byte(req.Name))
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}

type updateSubspeciesRequest struct {
	Notes   AllEntries[Note] `json:"notes,omitempty"`
	Aliases []string         `json:"aliases,omitempty"`
}

func updateSubspeciesHandler(w http.ResponseWriter, r *http.Request) {
	urlEncodedSpeciesName := r.PathValue("id")
	speciesName, err := url.QueryUnescape(urlEncodedSpeciesName)
	if err != nil {
		http.Error(w, "failed to decode subspecies name from url: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req := updateSubspeciesRequest{}
	err = json.Unmarshal(bs, &req)
	if err != nil {
		http.Error(w, "failed to unmarshal body: "+err.Error(), http.StatusBadRequest)
		return
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(subSpeciesCollectionName)
		existing, err := GetAltCollectionItemInTxn(ctx, speciesName, Species{})
		if err != nil {
			stat := http.StatusInternalServerError
			if err == mongo.ErrNoDocuments {
				stat = http.StatusNotFound
			}
			http.Error(w, err.Error(), stat)
			return nil, nil
		}
		mods := bson.D{}
		// Do alias changes
		mods = setStringArrayIfUnequal(mods, req.Aliases, existing.Aliases, "aliases")
		// Do note changes
		mods, err = WithNotesUpdate(bson.D{}, req.Notes, existing.Notes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		if len(mods) == 0 {
			http.Error(w, "no changes made", http.StatusBadRequest)
			return nil, nil
		}
		result := coll.FindOneAndUpdate(ctx, bson.D{{"_id", speciesName}}, mods)
		err = result.Err()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, nil
		}
		return w.Write([]byte(speciesName))
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}
