package api

//go:generate goGenerator/buildAndGenerate.sh

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/disintegration/imageorient"
	"github.com/reeceappling/goUtils/v2/utils"
	"github.com/reeceappling/mushDb/api/env"
	"github.com/reeceappling/mushDb/api/request/unix"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/exp/maps"
	"image/jpeg"
	"io"
	"iter"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"time"
)

const maxMultipartRequestSize = 32<<25 + 1024 //32<<20 + 1024 // TODO: is this max size ok?

const GoodTestRfidTag = "goodTestRdfidItem"

var ErrNoParentModifiedForTransfer = errors.New("parent not found for transfer update. Shouldnt occur")
var ErrMissingOptionalField = errors.New("missing optional field")
var ErrFailedToFinalizeMods = errors.New("failed to finalize mods")

// TODO: USE!
func UrlEncodeString(toEncode string) string {
	return url.QueryEscape(toEncode)
}

func UrlDecodeString(encoded string) (string, error) {
	return url.QueryUnescape(encoded)
}

var (
	_ CollectionItem = &Project{}
	_ CollectionItem = &Sale{}
)

type CollectionItem interface { // TODO: ADD USER TO THIS?
	CollectionName() string
	Decode(*mongo.SingleResult) (CollectionItem, error)
	IdValue() any // binary string id? DO NOT USE FOR ACTUALLY QUERYING THE DB DUE TO ANY TYPE
}

//var (
//	_ fruiter = FruitingChamber{}
//	_ fruiter = Bag{}
//)
//
//type fruiter interface {
//	basicFruit() Fruit
//	Permissioned
//}

var lastUpdatedIndexModel = mongo.IndexModel{
	Keys:    bson.D{{Key: "lastUpdated", Value: -1}},
	Options: options.Index().SetName("lastUpdated"),
}
var standardIndexModel = newSimpleIndex("standard", "standard", true, false, false)
var projectsIndexModel = newSimpleIndex("projects", "acl.projects.$**", false, false, false) // TODO: ensure actually indexes the correct thing! // TODO: this is a wildcard index!!!!
var saleIndexModel = newSimpleIndex("sale", "sale", false, true, false)
var transfersOutIndexModel = newSimpleIndex("transfersOut", "transfersOut", false, true, false) // TODO: do we even need to use this?
var creationDateIndexModel = newSimpleIndex("creationDate", "creationDate", true, false, false)
var disposedIndexModel = newSimpleIndex("disposed", "disposed", false, true, false) // TODO: USE?

var aliasesIndexModel = newSimpleIndex("aliases", "aliases", false, true, false) // THIS DOES NOT ENFORCE UNIQUENESS!!!!!

//func updateTogether() bson.D {
//	return []primitive.E{withUpdateNow()}
//}

func Initialize(ctx context.Context) error {
	if err := initializeItemMapCollection(ctx); err != nil {
		return err
	}
	for i, initializer := range map[string]func(context.Context) error{
		"db": initializeDb,
		// Initialize main collections
		"bags":             initializeBags,
		"fruiting chamber": initializeFruitingChamber,
		"jars":             initializeJars,
		"LCs":              initializeLCs,
		"LcSyringes":       initializeSyringes,
		"mss":              initializeMSS,
		"plate":            initializePlates,
		"slant":            initializeSlants,
		"stasis tube":      initializeStasisTubes,
		"spore swabs":      initializeSporeSwabs,
		// Initialize new main collections
		"fruit":       initializeFruits,
		"spore print": initializeSporePrints,
		"plugs":       initializePlugs,
		"waterJar":    initializeWaterJars,
		//Initialize Collections with predefined items
		"agar Recipe":      initializeAgarRecipes,
		"jar Recipe":       initializeJarRecipes,
		"lc Recipe":        initializeLcRecipes,
		"substrate Recipe": initializeSubstrates,
		"species":          initializeSpecies,
		"subspecies":       initializeSubspecies,
		// Initialize other alt collections
		"agar batch":        initializeAgarBatches,
		"grain batch":       initializeGrainBatches,
		"pc run":            initializePCRun,
		"sales":             initializeSales,
		"transfer":          initializeTransfers,
		"projects":          initializeProjects,
		"substrate batches": initializeSubstrateBatches,
		// initialize users
		"users": initializeUsers,
	} {
		println("trying", i, "initializer")
		if err := initializer(ctx); err != nil {
			return errors.Join(fmt.Errorf(`%s initializer failed`, i), err)
		}
		println("completed initializing", i)
	}
	// TODO: REENABLE!
	//if err := initializeAliasesCollection(ctx); err != nil { // TODO: when spec/subspec/subRec are created/modified/deleted, also update the aliases collection!
	//	println("aliases collection failed to initialize", err.Error())
	//	return errors.Join(errors.New("aliases initializer failed"), err)
	//}
	for name, b58IdStr := range map[string]string{
		// Mains IDs
		"plate":           string(exPlate.AsBase58()),
		"bag":             string(exBag.AsBase58()),
		"fruitingChamber": string(exFC.AsBase58()),
		"jar":             string(exJar.AsBase58()),
		"mss":             string(exMSS.AsBase58()),
		"slant":           string(exSlant.AsBase58()),
		"stasisTube":      string(exStasis.AsBase58()),
		"fruit":           string(exFruitId.AsBase58()),
		"sporePrint":      string(exSporePrint.AsBase58()),
		"waterJar":        string(exWaterId.AsBase58()),
		// Standard Alt IDs
		"agarBatch":       string(exAltId.AsBase58()),
		"agarRecipe":      string(exAltId.AsBase58()),
		"jarRecipe":       string(exAltId.AsBase58()),
		"lcRecipe":        string(exAltId.AsBase58()),
		"sale":            string(exAltId.AsBase58()),
		"substrateRecipe": string(exAltId.AsBase58()),
		"transfer":        string(exAltId.AsBase58()),
		// String Alt IDs
		"project":    testEntryStringId,
		"species":    testEntryStringId,
		"subspecies": testEntryStringId,
	} {
		println(fmt.Sprintf(`test %s can be found at /view/%s/%s`, name, name, b58IdStr))
	}
	// TODO: validateDbEntries(ctx)

	return nil
}

//func validateDbEntries(ctx context.Context) {
//	for _, _ = range map[string]
//}

func simplifyUpdates(elementsGroup ...bson.E) bson.D { // TODO: USE THIS
	return elementsGroup
}

func simpleUpdate(key string, value interface{}) bson.E {
	return bson.E{Key: key, Value: value}
}

func simplePointerUpdate[T any](mods []bson.E, key string, ptr *T) []bson.E {
	if ptr == nil {
		return mods
	}
	out := append(mods, simpleUpdate(key, ptr))
	return out
}

func getItemLatestImage(item CollectionItem) (*ImageLocation, unix.Time) { // TODO: consider using?
	var loc *ImageLocation = nil
	var latest unix.Time = 0
	if itemWithPicsField, ok := item.(interface{ getLatestPicFromPicsField() *PicWithNotes }); ok {
		if pwn := itemWithPicsField.getLatestPicFromPicsField(); pwn != nil {
			loc = &pwn.Location
			latest = pwn.Time
		}
	}
	if contamItem, ok := item.(interface{ getContamsLatestImage() *Contamination }); ok {
		if contam := contamItem.getContamsLatestImage(); contam != nil {
			if contam.Location != nil && contam.Time > latest {
				return contam.Location, contam.Time
			}
		}
	}
	// TODO: add getLatestFlush to Bags, FCs, etc...
	if itemWithFlushesField, ok := item.(interface{ getLatestFlush() *PicWithNotes }); ok {
		if pwn := itemWithFlushesField.getLatestFlush(); pwn != nil {
			loc = &pwn.Location
			latest = pwn.Time
		}
	}
	return loc, latest
}

// TODO: USE THIS IN UPDATES!
//func latestPicUpdate(latestPicPtrsForEachGroup []*PicWithNotes) []bson.E {
//	out := []bson.E{}
//	NonNilPics := utils.NonNil(latestPicPtrsForEachGroup)
//	if len(NonNilPics) == 0 {
//		return out
//	}
//	latest := NonNilPics[0]
//	for i := 1; i < len(NonNilPics); i++ {
//		toCheck := NonNilPics[i]
//		if toCheck.Time > latest.Time {
//			latest = toCheck
//		}
//	}
//	return append(out, simpleUpdate("mostRecentImage", latest))
//}

//func finishUpdate(ctx context.Context, coll *mongo.Collection, mods []bson.E, id any) error { // TODO: is just error ok?
//	if len(mods) == 0 {
//		return nil
//	}
//	mods = append(mods, withUpdate(nil))
//	res, err := coll.UpdateByID(ctx, id, mods)
//	if err != nil {
//		return err
//	}
//	switch res.ModifiedCount {
//	case 0:
//		return errors.New("not found") // TODO: move
//	case 1:
//		return nil
//	default:
//		return errors.New("modified more than 1 entry")
//	}
//}

func pushToArrayInline[T any](fieldName string, vals ...T) bson.D {
	switch len(vals) {
	case 1:
		return bson.D{{
			Key: "$push",
			Value: bson.D{{
				Key:   fieldName,
				Value: vals[0],
			}},
		}}
	case 0:
		return bson.D{}
	default:
		return bson.D{{
			Key: "$push",
			Value: bson.D{{
				Key: fieldName,
				Value: bson.D{{
					Key:   "$each",
					Value: vals,
				}},
			}},
		}} // TODO: ensure this works
	}
}

//func pushToArrayNew[T any](fieldName string, vals ...T) bson.D { // TODO: rename
//	switch len(vals) {
//	case 1:
//		return bson.D{{Key: fieldName, Value: vals[0]}}
//	case 0:
//		return bson.D{}
//	default:
//		return bson.D{{Key: fieldName, Value: bson.D{{Key: "$each", Value: vals}}}}
//	}
//}
//
//func withUpdate(t *time.Time) bson.E {
//	return bson.E{Key: "lastUpdated", Value: unixTimeFor(utils.Default(t, time.Now()))}
//}
//
//func withItemsRemoved[T any](field string, items ...T) bson.D {
//	itemsEquality := make([]bson.E, len(items))
//	for i, item := range items {
//		itemsEquality[i] = bson.E{Key: "$eq", Value: item}
//	}
//
//	//{ "$pull": { <field1>: <value|condition>, <field2>: <value|condition>, ... } }
//	return bson.D{{Key: "$pull", Value: bson.D{{Key: field, Value: itemsEquality}}}}
//}

func createIndexes(ctx context.Context, coll *mongo.Collection, toCreate []mongo.IndexModel) error {
	if len(toCreate) == 0 {
		return nil
	}
	toCreateNames := map[string]mongo.IndexModel{}
	for _, idx := range toCreate {
		if idx.Options.Name == nil {
			return errors.New("cannot create an index without name")
		}
		toCreateNames[*idx.Options.Name] = idx
	}
	// Get current indices
	idxCursor, err := coll.Indexes().List(ctx)
	if err != nil {
		return err
	}
	for idxCursor.Next(ctx) {
		var existingIndex mongo.IndexModel
		if err = idxCursor.Decode(&existingIndex); err != nil {
			return errors.Join(errors.New("cursor decode error for existing index"), err)
		}
		// Remove all existing indexes from toCreate
		if existingIndex.Options != nil && existingIndex.Options.Name != nil {
			delete(toCreateNames, *existingIndex.Options.Name)
		}
	}
	if err = idxCursor.Err(); err != nil {
		return errors.Join(errors.New("mongo cursor error after UserPerms project iteration"), err)
	}
	_, err = coll.Indexes().CreateMany(ctx, maps.Values(toCreateNames))
	return err
}

func newSimpleIndex(indexName, key string, descending, sparse, unique bool) mongo.IndexModel {
	keyElement := bson.E{Key: key, Value: 1}
	if descending {
		keyElement.Value = -1
	}
	return mongo.IndexModel{
		Keys: bson.D{keyElement},
		Options: options.Index().
			SetName(indexName).
			SetSparse(sparse).
			SetUnique(unique),
	}
}

func indicesSame(a, b mongo.IndexModel) bool { // TODO: use?
	return (a.Keys.(bson.D)[0].Key == b.Keys.(bson.D)[0].Key) && reflect.DeepEqual(a.Options, b.Options)
}

var ErrInvalidEntryType = errors.New("invalid entry type")

func getStandardEntries[T CollectionItem](ctx context.Context, temp T) (out []T, err error) {
	cursor, err := GetMongoClient(ctx).
		Database(dbName).
		Collection(temp.CollectionName()).
		Find(ctx, BsonFindFilter("standard", true)) // TODO: NOT WORKING PROPERLY?!!!!! (check again)
	if err != nil {
		return nil, err
	}
	return getCollectionItemsFromCursor[T](ctx, cursor, nil)
}

func cursorIterator[T CollectionItem](ctx context.Context, cursor *mongo.Cursor) iter.Seq2[T, error] { // TODO: consider using!
	return func(yield func(T, error) bool) {
		defer cursor.Close(ctx) // TODO; ensure ok
		yieldCount := 0
		var tempResult T
		user, err := GetAuthInfo(ctx)
		if err != nil {
			if !yield(tempResult, err) {
				return
			}
			return
		}
		for {
			var result T
			if cursor.TryNext(ctx) {
				if err = cursor.Decode(&result); err != nil {
					if !yield(result, err) {
						return
					}
					return
				}
				//bs, err := json.Marshal(result)
				//if err == nil {
				//	println("CHECKING AN ITEM: " + string(bs)) // TODO: del
				//}
				// If item is permissioned, ensure the user can read it
				permedItem, ok := interface{}(result).(Permissioned)
				if ok {
					acl := permedItem.Permissions()
					// If user cannot read or write, do not add
					if acl.HighestPermFor(user) == nil {
						println("skipping entry, user does not have permission!") // TODO: del
						// Skip this entry
						continue
					}
				}
				//if !allowDisposed { // TODO: reenable if disposed isnt filtered out in query
				//	disposableItem, ok := interface{}(result).(Disposable)
				//	if ok && disposableItem.DisposalInfo() != nil {
				//		// Skip this entry
				//		continue
				//	}
				//}
				if !yield(result, nil) {
					return
				}
				yieldCount++
			}

			cursorClosed := cursor.ID() == 0
			if cursorClosed && yieldCount == 0 {
				yield(result, mongo.ErrNoDocuments)
				return
			}
			if err = cursor.Err(); err != nil {
				yield(result, err)
				return
			}
			if cursorClosed {
				return
			}
		}
	}
}

func getCollectionItemsFromCursor[T CollectionItem](ctx context.Context, cursor *mongo.Cursor, numItems *int) ([]T, error) {
	defer cursor.Close(ctx) // TODO; ensure ok
	user, err := GetAuthInfo(ctx)
	if err != nil {
		return nil, err
	}
	results := []T{}
	if numItems != nil && *numItems > 0 {
		results = make([]T, 0, *numItems)
	} else {
		if user.IsAdmin() {
			err = cursor.All(ctx, &results) // TODO: only results size???
			return results, err
		}
	}
	//for result, err := range cursorIterator[T](ctx, cursor) { // TODO: consider using
	//	if err != nil {
	//		return nil, err // TODO: ok?
	//	} else {
	//		results = append(results, result)
	//	}
	//}

	for numItems == nil || len(results) < *numItems {
		if cursor.TryNext(ctx) {
			var result T
			if err = cursor.Decode(&result); err != nil {
				return nil, err
			}
			bs, err := json.Marshal(result)
			if err == nil {
				println("CHECKING AN ITEM: " + string(bs)) // TODO: del
			}
			// If item is permissioned, ensure the user can read it
			permedItem, ok := interface{}(result).(Permissioned)
			if ok {
				acl := permedItem.Permissions()
				// If user cannot read or write, do not add
				if acl.HighestPermFor(user) == nil {
					println("skipping entry, user does not have permission!") // TODO: del
					// Skip this entry
					continue
				}
			}
			//if !allowDisposed { // TODO: reenable if disposed isnt filtered out in query
			//	disposableItem, ok := interface{}(result).(Disposable)
			//	if ok && disposableItem.DisposalInfo() != nil {
			//		// Skip this entry
			//		continue
			//	}
			//}

			results = append(results, result)
			continue
		}
		cursorClosed := cursor.ID() == 0
		if cursorClosed {
			if len(results) == 0 {
				return results, mongo.ErrNoDocuments
			}
		}
		if err = cursor.Err(); err != nil {
			return nil, err
		}
		if cursorClosed {
			break
		}
	}
	return results, nil
}
func getUserProjectsFromCursor(ctx context.Context, cursor *mongo.Cursor, numItems *int) ([]*Project, error) {
	defer cursor.Close(ctx) // TODO; ensure ok
	user, err := GetAuthInfo(ctx)
	if err != nil {
		return nil, err
	}
	results := []*Project{}
	if numItems != nil {
		results = make([]*Project, 0, *numItems)
	} else {
		if user.IsAdmin() {
			err = cursor.All(ctx, &results) // TODO: only results size???
			return results, err
		}
	}

	for numItems == nil || len(results) < *numItems {
		if cursor.TryNext(ctx) {
			var result *Project
			if err = cursor.Decode(&result); err != nil {
				return nil, err
			}
			if result.Private && !result.Perms.ForUser(user.Email).CanRead() {
				continue
			}
			results = append(results, result)
			continue
		}
		cursorClosed := cursor.ID() == 0
		if cursorClosed {
			if len(results) == 0 {
				return results, mongo.ErrNoDocuments
			}
		}
		if err = cursor.Err(); err != nil {
			return nil, err
		}
		if cursorClosed {
			break
		}
	}
	return results, nil
}

type picWithNotesForm struct {
	Time unix.Time `json:"time"`
	Img  string    `json:"img"` // Same as Location on PicWithNotes
	NotesUpdateField
}

func (pwn picWithNotesForm) convert() PicWithNotes {
	return PicWithNotes{
		PicWithNotesLessLocation: newPicWithNotesLessLocation(pwn.Time, pwn.Notes.asEntries()),
		Location:                 ImageLocation(pwn.Img),
	}
}

type contamForm struct {
	Time      unix.Time `json:"time"`
	Confirmed bool      `json:"confirmed"`
	Bacteria  bool      `json:"bacteria"`
	Mold      bool      `json:"mold"`
	NotesUpdateField
	Location *string `json:"location,omitempty"` // MAY OR MAY NOT EXIST ON RESPONSE
}

func (cf contamForm) convert() Contamination {
	var loc *ImageLocation = nil
	if cf.Location != nil {
		loc = utils.Pointer(ImageLocation(*cf.Location))
	}
	return Contamination{
		ContaminationLessLocation: ContaminationLessLocation{
			PicWithNotesLessLocation: PicWithNotesLessLocation{
				RequiredTimeField: RequiredTimeField{cf.Time},
				NotesField:        NotesField{cf.Notes.asEntries()},
			},
			Confirmed: cf.Confirmed,
			Bacteria:  cf.Bacteria,
			Mold:      cf.Mold,
		},
		Location: loc,
	}
}

func compareContamUpdate(a contamForm, b Contamination) (equal bool) {
	if a.Confirmed != b.Confirmed || a.Mold != b.Mold || a.Bacteria != b.Bacteria || len(a.Notes.New) > 0 {
		return false
	}
	for i, updatedNote := range a.Notes.Existing {
		if DataStripped(updatedNote) != b.Notes[i] {
			return false
		}
	}
	return true
}

func picWithNotesWasModified(existing PicWithNotes, updated picWithNotesForm) (wasModified bool) {
	return updated.Img != string(existing.Location) ||
		updated.Time != existing.Time ||
		notesWereModified(existing.Notes, updated.Notes)
}

func compareImageUpdate(updated picWithNotesForm, existing PicWithNotes) (equal bool) {
	if len(updated.Notes.New) > 0 {
		return false
	}
	return notesWereModified(existing.Notes, updated.Notes)
}

const IDfld = "_id"

func BsonFindFilter(key string, value any) bson.D {
	return bson.D{bson.E{Key: key, Value: value}}
}
func BsonFindByIdFilterOrdered[T CollectionId](id T) bson.D { // TODO: ensure ok
	return bson.D{bson.E{Key: IDfld, Value: id}}
}

func BsonFindByIdFilterUnordered[T CollectionId](id T) bson.M { // TODO: ensure ok
	return bson.M{IDfld: id}
}

func BsonItemsAfterFilter(key string, prevValue any) bson.M { // TODO: use and validate ok
	return bson.M{key: bson.M{"$gt": prevValue}}
}
func BsonItemsStartingWithFilter(key string, prevValue any) bson.M { // TODO: use and validate ok
	return bson.M{key: bson.M{"$gte": prevValue}}
}
func BsonItemsUntilIncludingFilter(key string, prevValue any) bson.M { // TODO: use and validate ok
	return bson.M{key: bson.M{"$lte": prevValue}}
}
func BsonItemsUntilExcludingFilter(key string, prevValue any) bson.M { // TODO: use and validate ok
	return bson.M{key: bson.M{"$lt": prevValue}}
}
func BsonPredicateFilter(predicate string, value any) bson.M { // TODO: use and validate ok
	/* Predicate options:
	https://www.mongodb.com/docs/manual/reference/mql/query-predicates/
	"$X" where X in one of the following arrays.
		comparison: [eq, ne, gt,gte,lt,lte, in, nin]
		logical: [and, nor, not, or]
		array queries: [all, elemMatch, size]
		bitwise: [bitsAllClear,bitsAllSet,bitsAnyClear,bitsAnySet]
		data types:  [exists, type]
		misc: [expr, jsonSchema, mod, regex, where]
		geo: [geoIntersects, geoWithin, near, nearSphere]
	*/
	return bson.M{predicate: value}
}

//func setUnsetUnequalPointers[T comparable](key string, update *T, current *T, modsIn bson.D) bson.D {
//	if (update == nil && current == nil) || ((update != nil && current != nil) && (*(update)) == (*(current))) {
//		return modsIn
//	}
//	out := modsIn
//	if update != nil {
//		out = append(out, bson.E{"$set", bson.D{{key, *update}}})
//	} else {
//		out = append(out, bson.E{"$unset", bson.D{{key, 1}}})
//	}
//	return out
//}
//
//func WithImageChanges(currentMods bson.D, fieldName string, outImages SplitEntries[picWithNotesForm, PicWithNotes], currentPics []PicWithNotes) (bson.D, error) {
//	mods, err := WithExistingEntriesChange(currentMods, fieldName, outImages.Existing, currentPics, compareImageUpdate)
//	if err != nil {
//		return bson.D{}, err
//	}
//	mods = append(mods, pushToArray(fieldName, outImages.New...)...)
//	return mods, nil
//}
//
//func WithContamChanges(currentMods bson.D, fieldName string, outContams SplitEntries[contamForm, Contamination], currentContams []Contamination) (bson.D, error) {
//	mods, err := WithExistingEntriesChange(currentMods, fieldName, outContams.Existing, currentContams, compareContamUpdate)
//	if err != nil {
//		return bson.D{}, err
//	}
//	mods = append(mods, pushToArrayInline(fieldName, outContams.New...)...)
//	return mods, nil
//}

//// WithEntriesChanges Is to be used with notes, and things formatted like them (no image-holders)
//func WithEntriesChanges[T any](currentMods bson.D, id string, updatedEntries AllEntries[T], existing []T, areEqual func(a, b T) bool) (mods bson.D, err error) {
//	mods, err = WithExistingEntriesChange(currentMods, id, updatedEntries.Existing, existing, areEqual)
//	if err != nil {
//		return nil, err
//	}
//	// add new items
//	mods = append(mods, pushToArrayInline("notes", updatedEntries.New...)...)
//	return mods, nil
//}

//// WithExistingEntriesChange is to be used with Images, Contams, etc
//func WithExistingEntriesChangeNew[T, U any](upd *Mods, id string, updatedExisting []Data[T], existing []U, areEqual func(a T, b U) bool) *Mods {
//	if upd.err != nil {
//		return upd
//	}
//	// INCOMING SIZE MUST BE THE SAME!
//	if len(existing) != len(updatedExisting) {
//		upd.err = errors.New("incorrect amount of incoming existing " + id + "s")
//		return upd
//	}
//	// Do changes/removals
//	for i, newExisting := range updatedExisting {
//		indexKey := fmt.Sprintf(`%s.%d`, id, i)
//		if newExisting.Disabled {
//			upd.Unset(indexKey) // TODO: value of 1 was here?
//			//removals = append(removals, bson.E{currentKey, 1}) // TODO: make sure ok
//			continue
//		}
//		if !areEqual(newExisting.Data, existing[i]) {
//			upd.Set(indexKey, newExisting.Data)
//		}
//	}
//	// TODO: Changes (sets) first if exist (not sure if possible the way we do it)
//	// TODO: Removals second if exist
//	return upd
//}

//// WithExistingEntriesChange is to be used with Images, Contams, etc
//func WithExistingEntriesChange[T, U any](currentMods bson.D, id string, updatedExisting []Data[T], existing []U, areEqual func(a T, b U) bool) (mods bson.D, err error) {
//	mods = currentMods
//	// INCOMING SIZE MUST BE THE SAME!
//	if len(existing) != len(updatedExisting) {
//		err = errors.New("incorrect amount of incoming existing " + id + "s")
//	}
//	// Do changes
//	removals := []bson.E{}
//	chgs := []bson.E{}
//	for i, newExisting := range updatedExisting {
//		if newExisting.Disabled {
//			removals = append(removals, bson.E{fmt.Sprintf(`%s.%d`, id, i), 1})
//			continue
//		}
//		if !areEqual(newExisting.Data, existing[i]) {
//			chgs = append(chgs, bson.E{fmt.Sprintf(`%s.%d`, id, i), newExisting.Data})
//		}
//	}
//	// Changes first if exist
//	if len(chgs) > 0 {
//		mods = append(mods, bson.E{"$set", chgs})
//	}
//	// Removals second if exist
//	if len(removals) > 0 {
//		mods = append(mods, bson.E{"$unset", removals})
//	}
//	return mods, nil
//}

func multipartToImageBytes(p *multipart.Part, w http.ResponseWriter) ([]byte, error) {
	// Get field bytes as an image
	println("decoding jpg")
	img, _, err := imageorient.Decode(p)
	// If using mac screenshots (cmd+shift+5), you'll need to do this:
	// defaults write com.apple.screencapture type jpg; killall SystemUIServer
	if err != nil {
		http.Error(w, "failed to read image as either jpeg as png! "+err.Error(), http.StatusBadRequest)
		return nil, err
	}
	//img, err := jpeg.Decode(p)
	//if err != nil {
	//	println("decoding png")
	//	img, err = png.Decode(p)
	//	if err != nil {
	//		http.Error(w, "failed to read image as either jpeg as png! "+err.Error(), http.StatusBadRequest)
	//		return nil, err
	//	}
	//}
	buf := new(bytes.Buffer)
	println("re-encoding as jpg")
	err = jpeg.Encode(buf, img, nil) // TODO: JPEG OR PNG??????
	if err != nil {
		http.Error(w, "failed to encode image to save! "+err.Error(), http.StatusInternalServerError)
		return nil, err
	}
	return buf.Bytes(), nil
}

func handleWriteErr(err error, w http.ResponseWriter) {
	if err != nil {
		env.LogAlways("failed to write! " + err.Error())
	}
}

func handleFileDeleteErr(err error) {
	env.LogAlways("failed to delete file! " + err.Error())
}

var (
	testEntryStringId    = "TestEntry"
	exAltId              = altCollIdForint(0)
	exFruitId            = mainCollIdForint(idTestFruit)
	exampleTime          = unix.TimeFor(time.Date(2024, 12, 29, 0, 0, 0, 0, time.UTC))
	exReqTimeField       = RequiredTimeField{exampleTime}
	exampleSpecies       = "Beech"
	exampleSubspecies    = utils.Pointer("Brown Beech")
	exGenSinceSpore      = Generation(2)
	exGenSinceFruitSpore = Generation(1)
	exParentType         = "plate"
	exPlate              = MainCollectionId([8]byte{0, 0, 0, 0, 0, 0, 0, 0})
	exBag                = mainCollIdForint(idTestBag)
	exBatch              = altCollIdForint(idTestBatch)
	exJar                = mainCollIdForint(idTestJar)
	exSporePrint         = mainCollIdForint(idTestSp)
	exFC                 = mainCollIdForint(idTestFC)
	exLC                 = mainCollIdForint(idTestLC)
	exLCS                = mainCollIdForint(idTestLCS)
	exMSS                = mainCollIdForint(idTestMSS)
	exPlugId             = mainCollIdForint(idTestPlug)
	exSlant              = mainCollIdForint(idTestSlant)
	exStasis             = mainCollIdForint(idTestStasis)
	exSwabId             = mainCollIdForint(idTestSwab)
	exWaterId            = mainCollIdForint(idTestWaterJar)
	exAlts               = []AlternateCollectionId{exAltId, exAltId}
	exProjWrite          = projectName("example write project ")
	exProjRead           = projectName("example read project ")
	testProj             = projectName("test project")
	exProjects           = []projectName{exProjWrite, exProjWrite}
	exUserRead           = altCollIdForint(0)
	exUserWrite          = altCollIdForint(1)
	exUserAdmin          = altCollIdForint(2)
	exUserNoProjectRead  = "1@example.com"
	exUserNoProjectWrite = "2@example.com"
	exProjPerms          = map[string]*bool{
		string(exUserRead[:]):  nil,
		string(exUserWrite[:]): utils.Pointer(false),
		string(exUserAdmin[:]): utils.Pointer(true),
	}
	exAcl = ACL{
		Users: map[string]bool{
			exUserNoProjectRead:  false,
			exUserNoProjectWrite: true,
		},
		Projects: map[projectName]bool{
			exProjRead:  false,
			exProjWrite: true,
		},
		BlanketPerm: RWPermWrite(),
	}
	testAcl = ACL{
		Users: map[string]bool{
			//testUserEmailGoogleNormal:     false, // TODO: fix for correct user! may be 2409?
			//testUserEmailSelf: true,
			testUserEmailPAA: true,
			testUserEmailPWA: true,
			testUserEmailPRA: false,
			// testUserEmailPNA: nil,
		},
		Projects: map[projectName]bool{
			exProjRead:  false,
			exProjWrite: true,
		},
		BlanketPerm: RWPermRead(),
	}
	exBool                     = utils.Pointer(true)
	exPicLoc                   = "test.jpg" // Uses a picture from another site...
	exPicWithNotesLessLocation = PicWithNotesLessLocation{
		RequiredTimeField: exReqTimeField,
		NotesField:        NotesField{exampleNotes()},
	}
	exPic = PicWithNotes{
		PicWithNotesLessLocation: exPicWithNotesLessLocation,
		Location:                 ImageLocation(exPicLoc),
	}
	exPics          = []PicWithNotes{exPic, exPic}
	exContamLessLoc = ContaminationLessLocation{
		PicWithNotesLessLocation: exPicWithNotesLessLocation,
		Confirmed:                false,
		Bacteria:                 false,
		Mold:                     true,
	}
	ec = Contamination{
		ContaminationLessLocation: exContamLessLoc,
		Location:                  (*ImageLocation)(&exPicLoc),
	}
	exContams = []Contamination{ec, ec}
)

func exampleNotes() []Note {
	return []Note{{
		RequiredTimeField: exReqTimeField,
		Note:              "test/example entry note 1",
	}, {
		RequiredTimeField: exReqTimeField,
		Note:              "test/example entry note 2",
	}}
}

func decodeItem[T any](item *T, encoded *mongo.SingleResult) (err error) {
	err = encoded.Decode(item)
	if err != nil {
		err = errors.Join(errors.New("failed to decode"), err)
	}
	return
}

//func CollectionFor(item CollectionItem, db *mongo.Database) *mongo.Collection {
//	return db.Collection(item.CollectionName())
//}

func finishMainCollItemUpdate[T MainCollectionItem](ctx context.Context, w http.ResponseWriter, modsFor func(T, AclField) (bson.D, error), existing T, reqPerms PermsOnRequest) {
	coll := DbFrom(ctx).Collection(existing.CollectionName())
	user, err := GetAuthInfo(ctx)
	if err != nil {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if user.isGuest() {
		dbErr(w, "guests cannot edit", http.StatusForbidden)
		return
	}
	if !user.HasPermissionToEdit(existing) {
		dbErr(w, "unauthorized to edit", http.StatusForbidden)
		return
	}
	// TODO: dont allow user to remove their own perms!
	//if !reqPerms.AsACL().HighestPermFor(user).CanWrite() {
	//	http.Error(w, "user cannot remove their own ability to write", http.StatusBadRequest)
	//	return
	//}
	aclField, err := reqPerms.AclForUser(ctx, user)
	if err != nil {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	upd, err := modsFor(existing, aclField)
	handleUpdateMods(ctx, w, coll, existing, existing.DbId(), upd, err)
	return
}

func finishMainCollItemUpdateInTxn[T MainCollectionItem](ctx mongo.SessionContext, w http.ResponseWriter, modsFor func(T, AclField) (bson.D, error), existing T, reqPerms PermsOnRequest) (T, error) {
	db := mongo.SessionFromContext(ctx).Client().Database(dbName)
	coll := db.Collection(existing.CollectionName())
	user, err := GetAuthInfo(ctx)
	if err != nil {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return existing, err
	}
	if user.isGuest() { // TODO: do we even need this here?
		dbErr(w, "guests cannot edit", http.StatusForbidden)
		return existing, err
	}
	if !user.HasPermissionToEdit(existing) {
		dbErr(w, "unauthorized to edit", http.StatusForbidden)
		return existing, err
	}
	// TODO: dont allow user to remove their own perms!
	//if !reqPerms.AsACL().HighestPermFor(user).CanWrite() {
	//	http.Error(w, "user cannot remove their own ability to write", http.StatusBadRequest)
	//	return
	//}
	aclField, err := reqPerms.AclForUser(ctx, user)
	if err != nil {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return existing, err
	}
	upd, err := modsFor(existing, aclField)
	if err != nil {
		dbErr(w, "err in modsFor: "+err.Error(), http.StatusInternalServerError)
		return existing, err
	}
	err = handleUpdateModsInTxn(ctx, coll, existing, existing.DbId(), upd, err)
	if err != nil {
		dbErr(w, "failed to update: "+err.Error(), http.StatusInternalServerError)
		return existing, err
	}
	return existing, err
}

func finishAltCollItemUpdate[T PermissionedAltCollectionItem[AlternateCollectionId]](ctx context.Context, w http.ResponseWriter, coll *mongo.Collection, modsFor func(T, AclField) (bson.D, error), existing T, reqPerms PermsOnRequest) {
	user, err := GetAuthInfo(ctx)
	if err != nil {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !user.HasPermissionToEdit(existing) {
		dbErr(w, "unauthorized to edit", http.StatusUnauthorized)
		return
	}
	if !reqPerms.AsACL().HighestPermFor(user).CanWrite() {
		http.Error(w, "user cannot remove their own ability to write", http.StatusBadRequest)
		return
	}
	aclField, err := reqPerms.AclForUser(ctx, user)
	if err != nil {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	upd, err := modsFor(existing, aclField)
	handleUpdateMods(ctx, w, coll, existing, existing.DbId(), upd, err)
	return
}

func finishStringIdAltCollItemUpdate[T PermissionedAltCollectionItem[string]](ctx context.Context, w http.ResponseWriter, coll *mongo.Collection, modsFor func(T, AclField) (bson.D, error), existing T, reqPerms PermsOnRequest) {
	user, err := GetAuthInfo(ctx)
	if err != nil {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if user.isGuest() {
		dbErr(w, "guests cannot edit", http.StatusForbidden)
		return
	}
	if !user.HasPermissionToEdit(existing) {
		dbErr(w, "unauthorized to edit", http.StatusForbidden)
		return
	}
	// TODO: dont allow user to remove their own perms!
	//if !reqPerms.AsACL().HighestPermFor(user).CanWrite() {
	//	http.Error(w, "user cannot remove their own ability to write", http.StatusBadRequest)
	//	return
	//}
	aclField, err := reqPerms.AclForUser(ctx, user)
	if err != nil {
		dbErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	upd, err := modsFor(existing, aclField)
	handleUpdateMods(ctx, w, coll, existing, existing.DbId(), upd, err)
	return
}

func ReadSimpleStructuredBody[T any](r *http.Request, w http.ResponseWriter, req *T) error {
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		println("failed to read body: " + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	if err = json.Unmarshal(bs, &req); err != nil {
		println("bad body format: " + string(bs))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	return nil
}

func MarshalAndReturn(ctx context.Context, w http.ResponseWriter, toReturn any) {
	bs, err := json.Marshal(toReturn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	env.LogIfDev(ctx, "sending response item: "+string(bs))
	_, err = w.Write(bs)
	if err != nil {
		handleWriteErr(err, w)
	}
}

func ImportFinalPerms(ctx context.Context, spec string, subspec *string) (ACL, error) {
	var finalPerms ACL
	sp, subsp, err := getSpeciesAndSubspecies(ctx, spec, subspec)
	if err != nil {
		return ACL{}, errors.New("failed to get species or subspecies: " + err.Error())
	}
	if subsp != nil {
		finalPerms = subsp.DefaultAcl.Clone()
	} else {
		finalPerms = sp.DefaultAcl.Clone()
	}
	userEmail := GetUserEmail(ctx)
	if finalPerms.Users == nil {
		finalPerms.Users = map[string]bool{}
	}
	finalPerms.Users[userEmail] = true

	return finalPerms, nil
}

//func TimeFromId(id AlternateCollectionId) time.Time { // TODO: USE AND MOVE
//	return primitive.ObjectID(id).Timestamp()
//}
//
//func Ternary[T any](val bool, ifTrue, ifFalse T) T {
//	if val {
//		return ifTrue
//	}
//	return ifFalse
//}

func TernaryPtr[T any](val *bool, ifTrue, ifFalse, ifNil T) T {
	if val == nil {
		return ifNil
	}
	if *val {
		return ifTrue
	}
	return ifFalse
}

// TODO: validate working
func validateAliasesUnused(ctx context.Context, coll *mongo.Collection, existingName string, existingAliases, updatedAliases []string) error {
	newAliases := utils.SetFrom(updatedAliases...)
	for _, al := range existingAliases {
		newAliases.Remove(al)
	}
	if newAliases.Contains(existingName) {
		return errors.New("aliases cannot match name")
	}
	if len(newAliases) == 0 {
		return nil // No alias changes, return early successfully
	}
	brandNewAliases := newAliases.ToSlice()
	// TODO: VALIDATE WORKS!
	aliasesFilter := bson.M{ // TODO: probably super inefficient, so use sparingly in spec, subspec, and subRec
		"$or": bson.A{
			bson.M{"_id": bson.M{"$in": brandNewAliases}},     // Matches if _id is in the list
			bson.M{"aliases": bson.M{"$in": brandNewAliases}}, // Matches if any array item is in the list // TODO: ensure ok
		},
	}

	if err := coll.FindOne(ctx, aliasesFilter).Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// Not found, success! all unused aliases
			return nil
		}
		return err
	}
	return errors.New("at least one entry contained a new alias the user was trying to add")
}

// TODO: validate working
func validateAliasesNameUnused(ctx context.Context, coll *mongo.Collection, newName string, aliases []string) error {
	newAliases := utils.SetFrom(aliases...)
	newAliases.Add(newName)
	if len(newAliases) == 0 {
		return nil // No alias changes, return early successfully
	}
	brandNewAliases := newAliases.ToSlice()
	aliasesFilter := bson.M{ // TODO: probably super inefficient, so use sparingly in spec, subspec, and subRec
		"$or": bson.A{
			bson.M{"_id": bson.M{"$in": brandNewAliases}},     // Matches if _id is in the list
			bson.M{"aliases": bson.M{"$in": brandNewAliases}}, // Matches if any array item is in the list // TODO: ensure ok
		},
	}

	if err := coll.FindOne(ctx, aliasesFilter).Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// Not found, success! all unused aliases
			return nil
		}
		return err
	}
	return errors.New("at least one entry contained a new alias the user was trying to add")
}
