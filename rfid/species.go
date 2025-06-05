package rfid

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/reeceappling/goUtils/v2/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"net/url"
	sliceutils "slices"
	"time"
)

const speciesCollectionName = "species"

type Species struct {
	Name              string                `bson:"_id" json:"_id"` // THIS IS THE COMMON NAME
	ScientificName    string                `bson:"scientificName" json:"scientificName"`
	Aliases           []string              `bson:"aliases,omitempty" json:"aliases,omitempty"`
	Has               bool                  `bson:"has" json:"has"` // TODO: ENSURE THIS IS SET UP RIGHT!!!!!
	StandardSubstrate alternateCollectionId `bson:"standardSubstrate,omitempty" json:"standardSubstrate,omitempty"`
	Notes             []Note                `bson:"notes,omitempty" json:"notes,omitempty"`
	LastUpdated       unixTime              `bson:"lastUpdated" json:"lastUpdated"`
}

func (sp Species) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := sp
	err := decodeItem(&out, encoded)
	return out, err
}

func (sp Species) clean() CollectionItem {
	out := sp
	// TODO: Change name
	// TODO: Change scientificName
	// TODO: Change aliases
	// TODO: Change has?
	// TODO: Change notes
	return out
}

func (sp Species) EntryTypeField() *string {
	return nil
}

func (sp Species) CollectionName() string {
	return speciesCollectionName
}

func initializeSpecies(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(speciesCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		newSimpleIndex("scientificName", "scientificName", false, false, true),
		newSimpleIndex("aliases", "aliases", false, true, false),
		//BackupSpecies (likely no index)      *string           `bson:"backupSpecies,omitempty" json:"backupSpecies,omitempty"` // a species that is similar to this one, somehow
		newSimpleIndex("standardSubstrate", "standardSubstrate", false, true, false),
		//Notes (no index) (maybe later with tags?)
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}

	// TODO: OTHER SPECIES
	inserted, updated := 0, 0
	woodPelletsId := alternateCollectionId(altCollIdForint(idWoodPellets))
	for _, species := range []Species{
		// King Oyester
		{
			Name:              "king oyster",
			ScientificName:    "Pleurotus Eryngii",
			Aliases:           nil,
			StandardSubstrate: woodPelletsId,
			Notes: []Note{
				{
					Time: ogTime,
					Note: "Colonization conditions: FIXME",
				},
				{
					Time: ogTime,
					Note: "Fruiting conditions: Prefer higher humidity than pinks FIXME",
				},
				{
					Time: ogTime,
					Note: "Best Agar: LMEA",
				}},
		},
		// Pink Oyster
		{
			Name:              "pink oyster", // TODO: this
			ScientificName:    "Pleurotus Djamor",
			Aliases:           nil,
			StandardSubstrate: woodPelletsId,
			Notes: []Note{{
				Time: ogTime,
				Note: "Colonization conditions: FIXME",
			},
				{
					Time: ogTime,
					Note: "Fruiting conditions: Prefer higher FAE than kings FIXME",
				},
				{
					Time: ogTime,
					Note: "Best Agar: LMEA",
				}},
		},
		// Enoki
		{
			Name:              "enoki", // TODO: this
			ScientificName:    "Flammulina filiformis",
			Aliases:           nil,
			StandardSubstrate: woodPelletsId,
			Notes: []Note{{
				Time: ogTime,
				Note: "Colonization conditions: FIXME",
			},
				{
					Time: ogTime,
					Note: "Fruiting conditions: Grow in a high-CO2 environment, with the only light being high-up in the enclosure to ensure they grow tall and thin, FAE==0, humidity=70+",
				},
				{
					Time: ogTime,
					Note: "Best Agar: LMEA",
				}},
		},
		// Shiitake
		{
			Name:              "shiitake", // TODO: this
			ScientificName:    "Lentinula edodes",
			Aliases:           nil,
			StandardSubstrate: woodPelletsId,
			Notes: []Note{{
				Time: ogTime,
				Note: "Colonization conditions: Hardwood sawdust with 15% bran, or pegs into a log. Indirect sun 8-12hrs. 80-90% humidity, 60-68degF, and regular FAE",
			},
				{
					Time: ogTime,
					Note: "Fruiting conditions: Needs roughly 3mo incubation, damage block to encourage fruiting.",
				},
				{
					Time: ogTime,
					Note: "Best Agar: LMEA",
				}},
		},
		// Maitake, Hen of the Woods
		{
			Name:              "maitake", // TODO: this
			ScientificName:    "Grifola frondosa",
			Aliases:           []string{"hen of the woods"},
			StandardSubstrate: woodPelletsId,
			Notes: []Note{{
				Time: ogTime,
				Note: "Colonization conditions: FIXME",
			},
				{
					Time: ogTime,
					Note: "Fruiting conditions: 50-70degF (64-66 is ideal). >90% humidity. Cold shock to begin fruiting ",
				},
				{
					Time: ogTime,
					Note: "Resistant to high temps",
				},
				{
					Time: ogTime,
					Note: "Best Agar: LMEA",
				}},
		},
		// Beech
		{ // TODO: WHITE AND BROWN
			Name:              "beech", // TODO: this
			ScientificName:    "",      // TODO: this
			Aliases:           []string{"hen of the woods"},
			StandardSubstrate: woodPelletsId,
			Notes: []Note{{
				Time: ogTime,
				Note: "Fruiting conditions: 90-100RH, 50-60degF, plenty of light, cold shock to begin",
			},
				{
					Time: ogTime,
					Note: "50-60DegF, 80-90RH, FAE",
				},
				{
					Time: ogTime,
					Note: "Best Agar: LMEA",
				}},
		},
	} {
		var existing Species
		err = coll.FindOne(ctx, bson.D{{"_id", species.Name}}).Decode(&existing)
		if err != nil {
			if err != mongo.ErrNoDocuments {
				println("ERROR WAS NOT ERRNODOCS") // TODO: deleteme
				return err
			}
			// if not exists, add it to the db
			_, err = coll.InsertOne(ctx, species)
			if err != nil {
				return err
			}
			inserted++
			continue
		}
		println("NO ERROR, UPDATING") // TODO: this
		// If exists, ensure it is the same as it was. Add notes if necessary
		update := false
		if species.ScientificName != existing.ScientificName {
			update = true
		}
		finalAliases := utils.Set[string]{}
		finalAliases.Add(existing.Aliases...)
		finalAliases.Add(species.Aliases...)
		if len(finalAliases) != len(existing.Aliases) {
			update = true
			species.Aliases = finalAliases.ToSlice()
		}
		//// Backup species
		//if !update {
		//	switch utils.CountNotNil(species.BackupSpecies, existing.BackupSpecies) {
		//	case 1:
		//		update = true
		//	case 2:
		//		if *species.BackupSpecies != *existing.BackupSpecies {
		//			update = true
		//		}
		//	default:
		//		// Do nothing
		//	}
		//}
		if !update && existing.StandardSubstrate != species.StandardSubstrate {
			update = true
		}

		// Notes
		finalNotes := []Note{}
		copy(finalNotes, existing.Notes)
		for _, note := range species.Notes {
			if !sliceutils.Contains(finalNotes, note) {
				finalNotes = append(finalNotes, note)
				update = true
			}
		}
		species.Notes = finalNotes

		// Update if necessary
		if update {
			err = coll.FindOneAndReplace(ctx, bson.D{{"_id", species.Name}}, species).Err()
			if err != nil {
				return err
			}
			updated++
		}
	}
	if inserted+updated > 0 { // TODO: ok?
		println(fmt.Sprintf(`Species entries: inserted %d, updated %d`, inserted, updated))
	}
	return nil
}

type createSpeciesRequest struct {
	Name           string    `json:"name"`
	ScientificName string    `json:"scientificName"`
	Aliases        []string  `json:"aliases,omitempty"`
	Have           bool      `json:"have"`
	Sub            Base58Str `json:"sub"`
	Notes          []Note    `json:"notes,omitempty"`
}

func createSpeciesHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	req := createSpeciesRequest{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	subId, err := req.Sub.toAltCollectionId() // TODO: CONFIRM EXISTS
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(speciesCollectionName)
		_, err = coll.InsertOne(r.Context(), Species{
			Name:              req.Name,
			ScientificName:    req.ScientificName,
			Aliases:           req.Aliases,
			Has:               req.Have,
			StandardSubstrate: alternateCollectionId(subId),
			Notes:             req.Notes,
			LastUpdated:       unixTime(time.Now().UnixMilli()),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, nil
		}
		return w.Write([]byte(req.Name))
	})
	if err != nil {
		// TODO: handle write failure
	}
}

type updateSpeciesRequest struct {
	Substrate Base58Str        `json:"substrate"`
	Notes     AllEntries[Note] `json:"notes,omitempty"`
	Aliases   []string         `json:"aliases,omitempty"`
	Have      bool             `json:"have"`
}

func updateSpeciesHandler(w http.ResponseWriter, r *http.Request) {
	urlEncodedSpeciesName := r.PathValue("id")
	speciesName, err := url.QueryUnescape(urlEncodedSpeciesName)
	if err != nil {
		http.Error(w, "failed to decode species name from url: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req := updateSpeciesRequest{}
	err = json.Unmarshal(bs, &req)
	if err != nil {
		http.Error(w, "failed to unmarshal body: "+err.Error(), http.StatusBadRequest)
		return
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(speciesCollectionName)
		existing, err := GetAltCollectionItemInTxn(ctx, speciesName, Species{})
		if err != nil {
			stat := http.StatusInternalServerError
			if err == mongo.ErrNoDocuments {
				stat = http.StatusNotFound
			}
			http.Error(w, err.Error(), stat)
			return nil, nil
		}
		// TODO: HOW TO HANDLE MODIFYING SPECIAL SPECIES?
		reqStandardSubstrate, err := req.Substrate.toAltCollectionId() // TODO: check exists
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return nil, nil
		}
		mods := bson.D{}
		// change substrate if needed
		if alternateCollectionId(reqStandardSubstrate) != existing.StandardSubstrate {
			mods = bson.D{{"$set", bson.D{{"standardSubstrate", reqStandardSubstrate}}}}
		}
		// Do note changes
		mods, err = WithNotesUpdate(bson.D{}, req.Notes, existing.Notes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		// Do alias changes
		mods = setStringArrayIfUnequal(mods, req.Aliases, existing.Aliases, "aliases")
		// Do have changes
		if req.Have != existing.Has {
			mods = append(mods, bson.E{"$set", bson.D{{"has", reqStandardSubstrate}}}) // TODO: ensure ok! May be HAVE elsewhere!
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
		// TODO: WRITE ERR
	}
}
