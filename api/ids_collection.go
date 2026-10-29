package api

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
	"slices"
)

// centralized collection to map itemId to itemType
const idMapCollectionName = "itemIdMap"

type idMapEntry struct {
	Id        MainCollectionId `bson:"_id" json:"_id"`
	EntryType string           `bson:"entryType" json:"entryType"`
}

func initializeItemMapCollection(ctx context.Context) error {
	// Indices
	coll := DbFrom(ctx).Collection(idMapCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		newSimpleIndex("entryType", "entryType", false, false, false),
	})
	return err
}

func GetEntryTypeForId(ctx context.Context, id MainCollectionId) (string, error) {
	result := idMapEntry{}
	err := DbFrom(ctx).Collection(idMapCollectionName).FindOne(ctx, BsonFindFilter(IDfld, id)).Decode(&result)
	return result.EntryType, err
}

func addToIdMapCollection(ctx mongo.SessionContext, item MainCollectionItem) (err error) {
	coll := mongo.SessionFromContext(ctx).Client().Database(dbName).Collection(idMapCollectionName)
	id := item.DbId()
	_, err = coll.InsertOne(ctx, idMapEntry{
		Id:        id,
		EntryType: item.EntryType(),
	})
	if err != nil {
		return err
	}
	return nil
}

var (
	errStringIdDefaultId  = errors.New("string ID entry types are not valid in DefaultId()")
	errDefaultIdUnhandled = errors.New("unhandled entry type for DefaultAltId")
	errIsNotDefault       = errors.New("result is not a default entry, but an example or dev entry")
)

// TODO: USE THIS!
func DefaultId(entryType string) (BinaryCollectionId, error) {
	isMain, isEither := entryTypeCollectionIdType(entryType)
	if !isEither {
		return "", errDefaultIdUnhandled
	}
	if isMain {
		mid, err := DefaultMainId(entryType) // TODO: use in places
		if err != nil {
			if !errors.Is(err, errIsNotDefault) {
				// Id does not exist
				return "", err
			}
		}
		return mid.ToBinaryCollectionId(), err // TODO: ensure this is ok and non-defaults are handled properly
	}
	temp, err := defaultAltId(entryType)
	if err != nil {
		return "", err
	}
	return temp.ToBinaryCollectionId(), err // TODO: use in places
}

// TODO: make a library for constantMaps?
func entryTypeCollectionIdType(entryType string) (isMain bool, exists bool) {
	switch entryType {
	case "agarBatch", "agarRecipe", "grainBatch", "jarRecipe", "lcRecipe", "pcRun", "sale", "substrateBatch", "substrateRecipe", "transfer", "project", "species", "subspecies", "user":
		return false, true
	case "bag", "fruit", "fruitingChamber", "grainJar", "grainWaterJar", "lc", "lcSyringe", "mss", "plate", "plugs" /* TODO; ensure ok*/, "slant", "sporePrint", "sporeSwab", "stasisTube", "waterJar":
		return true, true
	}
	return false, false
}
func DefaultAltId(entryType string) (id AlternateCollectionId, isAlt bool, err error) {
	id, err = defaultAltId(entryType)
	isAlt = err != nil && errors.Is(err, errDefaultIdUnhandled)
	return id, isAlt, err
}
func defaultAltId(entryType string) (id AlternateCollectionId, err error) {
	switch entryType {
	case "agarBatch", "agarRecipe", "grainBatch", "jarRecipe", "lcRecipe", "pcRun", "sale", "substrateBatch", "substrateRecipe", "transfer":
		return exAltId, nil // TODO: use in places
	case "project", "species", "subspecies", "user":
		return exAltId, errStringIdDefaultId
	default:
		return exAltId, errDefaultIdUnhandled
	}
}
func DefaultMainId(entryType string) (id MainCollectionId, err error) {
	mainEntryTypesWithDefaults := []string{"grainWaterJar", "waterJar"}
	if idInt, exists := map[string]int{
		"bag":             idTestBag,
		"fruit":           idTestFruit,
		"fruitingChamber": idTestFC,
		"grainJar":        idTestJar,
		"grainWaterJar":   idTestGrainWaterJar,
		"lc":              idTestLC,
		"lcSyringe":       idTestLCS,
		"mss":             idTestMSS,
		"plate":           idTestPlate,
		"plugs":           idTestPlug, // TODO; ensure ok
		"slant":           idTestSlant,
		"sporePrint":      idTestSp,
		"sporeSwab":       idTestSwab,
		"stasisTube":      idTestStasis,
		"waterJar":        idTestWaterJar,
	}[entryType]; exists {
		isDefault := slices.Contains(mainEntryTypesWithDefaults, entryType)
		if !isDefault {
			err = errIsNotDefault
		}
		return mainCollIdForint(idInt), err
	}
	return MainCollectionId{ /* TODO: ok?*/ }, errDefaultIdUnhandled

}
