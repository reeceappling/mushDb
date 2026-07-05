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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/exp/maps"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"time"
)

const GoodTestRfidTag = "goodTestRdfidItem"

var ErrNoParentModifiedForTransfer = errors.New("parent not found for transfer update. Shouldnt occur")
var ErrMissingOptionalField = errors.New("missing optional field")
var ErrFailedToFinalizeMods = errors.New("failed to finalize mods")

// TODO: ALL CREATION ENDPOINTS
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
	IdValue() any // binary string id?
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
var disposedIndexModel = newSimpleIndex("disposed", "disposed", false, true, false)

var aliasesIndexModel = newSimpleIndex("aliases", "aliases", false, true, false) // TODO: THIS DOES NOT ENFORCE UNIQUENESS!!!!!

// TODO: USE!
func SetEnv(ctx context.Context, isProd bool) context.Context {
	return context.WithValue(ctx, "isProd", isProd)
}
func GetEnv(ctx context.Context) bool {
	isProd, ok := ctx.Value("isProd").(bool)
	if !ok {
		panic("Env not found")
	}
	return isProd
}

//// TODO: searching in a specific index
//func latestNUpdatedB(ctx context.Context) error { // TODO: fixMe
//	db := DbFrom(ctx)
//	//indx := // TODO: use correct index
//	opts := options.Find().SetHint(
//		mongo.IndexModel{Keys: bson.D{{"transfersOut", 1}}}, // TODO: ????????????
//		)
//	coll := db.Collection(FruitsCollName)
//	_, err := coll.Find(ctx, bson.D{}, opts)
//	return err
//	//coll.UpdateByID(ctx, bson.D{bson.E{Key: "_id": "someId"}}, ) // TODO: use this
//}

func withUpdateNow() primitive.E { // TODO: FIXME!
	return primitive.E{
		Key:   "lastUpdated",
		Value: unix.TimeForNow(), // TODO: FIXME!
	}
}

func updateTogether() bson.D {
	return []primitive.E{withUpdateNow()}
}

// TODO: HOW TO SORT AND STUFF IS ABOVE

func Initialize(ctx context.Context) error {
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

//func entryTypeFor(inp string) (CollectionItem, error) { // TODO: does not work for Projects?
//	switch strings.ToLower(inp) {
//	case "agarbatch", "agar batch",
//		"agarbatches", "agar batches":
//		return &AgarBatch{}, nil
//	case "agarrecipe", "agar recipe",
//		"agarrecipes", "agar recipes":
//		return &AgarRecipe{}, nil
//	case "bag",
//		"bags":
//		return &Bag{}, nil
//	case "fruit",
//		"fruits":
//		return &Fruit{}, nil
//	case "fruitingchamber", "box", "chamber", "fruiting chamber",
//		"boxes", "fruitingchambers", "chambers", "fruiting chambers":
//		return &FruitingChamber{}, nil
//	case "jar", "grainjar", "grain jar",
//		"jars", "grainjars", "grain jars":
//		return &GrainJar{}, nil
//	case "jarrecipe", "jar recipe",
//		"jarrecipes", "jar recipes":
//		return &JarRecipe{}, nil
//	case "lc", "liquidculture", "liquid culture",
//		"lcs", "liquidcultures", "liquid cultures":
//		return &LiquidCulture{}, nil
//	case "lcrecipe", "lc recipe", "liquidculturerecipe", "liquid culture recipe",
//		"lcrecipes", "lc recipes", "liquidculturerecipes", "liquid culture recipes":
//		return &LcRecipe{}, nil
//	case "lcSyringe", "lcSyringes":
//		return &LcSyringe{}, nil
//	case "mss", "sporesyringe", "spore syringe", "multisporesyringe", "multi spore syringe",
//		"msss", "sporesyringes", "spore syringes", "multisporesyringes", "multi spore syringes":
//		return &MSS{}, nil
//	case "pcrun", "pc run", "pressure cooker run", "pressure cooker", "pc", "pressurecooker", "run",
//		"pcruns", "pc runs", "pressure cooker runs", "pressure cookers", "pcs", "pressurecookers", "runs":
//		return &PCRun{}, nil
//	case "plate", "dish", "agarplate", "agar plate", "agardish", "agar dish", "petri", "petridish", "petri dish",
//		"plates", "dishes", "agarplates", "agar plates", "agardishes", "agar dishes", "petris", "petridishes", "petri dishes":
//		return &Plate{}, nil
//	case "plugs", "plug", "peg", "pegs":
//		return &PlugsJar{}, nil
//	case "project", "Projects":
//		return &Project{}, nil
//	case "sale", "sales":
//		return &Sale{}, nil
//	case "slant", "slants":
//		return &Slant{}, nil
//	case "species":
//		return &Species{}, nil
//	case "sporeprint", "spore print", "print",
//		"sporeprints", "spore prints", "prints":
//		return &SporePrint{}, nil
//	case "stasistube", "stasis tube", "stasis", "tube",
//		"stasistubes", "stasis tubes", "tubes":
//		return &StasisTube{}, nil
//	case "subspecies":
//		return &Subspecies{}, nil
//	case "substrate", "substraterecipe", "substrate recipe",
//		"substrates", "substraterecipes", "substrate recipes":
//		return &SubstrateRecipe{}, nil
//	case "transfer", "xfer",
//		"transfers", "xfers":
//		return &Transfer{}, nil
//	case "user", "users":
//		return &User{}, nil
//	case "waterJar", "waterJars":
//		return &WaterJar{}, nil
//	default:
//		return nil, errors.Join(ErrInvalidEntryType, errors.New("invalid collection input. Does not map to a collection name"))
//	}
//}

func getStandardEntries[T CollectionItem](ctx context.Context, temp T) (out []T, err error) {
	cursor, err := GetMongoClient(ctx).
		Database(dbName).
		Collection(temp.CollectionName()).
		Find(ctx, BsonFindFilter("standard", true)) // TODO: NOT WORKING PROPERLY?!!!!! (check again)
	if err != nil {
		return nil, err
	}
	return getCollectionItemsFromCursor[T](ctx, cursor, nil, true)
}

func getCollectionItemsFromCursor[T CollectionItem](ctx context.Context, cursor *mongo.Cursor, numItems *int, allowDisposed bool) ([]T, error) {
	defer cursor.Close(ctx)       // TODO; ensure ok
	user, err := GetAuthInfo(ctx) // TODO: unsure if needed anymore
	if err != nil {
		return nil, err
	} // TODO: del if unneeded
	results := []T{}
	if numItems != nil {
		results = make([]T, 0, *numItems)
	} else {
		if user.IsAdmin() {
			err = cursor.All(ctx, &results) // TODO: only results size???
			return results, err
		}
	}

	for numItems == nil || len(results) < *numItems {
		if cursor.TryNext(ctx) {
			var result T
			//err = checkIdTypeWithRawOnCursor(cursor) // TODO: DEL!
			//if err != nil {                          // TODO: del!
			//	panic("failed to check id type! " + err.Error()) // TODO: del!
			//} // TODO: del!
			if err = cursor.Decode(&result); err != nil {
				return nil, err
			}
			// If item is permissioned, ensure the user can read it
			permedItem, ok := interface{}(result).(Permissioned)
			if ok {
				// If user cannot read or write, do not add
				if permedItem.Permissions().HighestPermFor(user) == nil {
					println("skipping entry, user does not have permission!") // TODO: del
					// Skip this entry
					continue
				}
			}
			if !allowDisposed {
				disposableItem, ok := interface{}(result).(Disposable)
				if ok && disposableItem.DisposalInfo() != nil {
					// Skip this entry
					continue
				}
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
	Img  string    `json:"img"`
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

func BsonFindFilter(key string, value any) bson.D {
	return bson.D{bson.E{Key: key, Value: value}}
}

func BsonFindByIdFilterOrdered[T CollectionId](id T) bson.D { // TODO: ensure ok
	return BsonFindFilter("_id", id)
}

func BsonFindByIdFilterUnordered[T CollectionId](id T) bson.M { // TODO: ensure ok
	return bson.M{"_id": id}
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

func CollectionFor(item CollectionItem, db *mongo.Database) *mongo.Collection {
	return db.Collection(item.CollectionName())
}

//func Refresh[T CollectionItem](ctx context.Context, db *mongo.Database, item T) error {
//	return CollectionFor(item, db).FindOne(ctx, bson.D{{Key: "_id", Value: item.IdValue( /* TODO: PROBABLY WONT WORK*/ )}}).Decode(item)
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
	if user.isGuest() { // TODO: isnt this done elsewhere?
		dbErr(w, "guests cannot edit", http.StatusUnauthorized)
		return
	}
	if !user.HasPermissionToEdit(existing) {
		dbErr(w, "unauthorized to edit", http.StatusUnauthorized)
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
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		println("failed to read body: " + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}
	if err = json.Unmarshal(bytes, &req); err != nil {
		println("bad body format: " + string(bytes))
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
