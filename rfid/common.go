package rfid

//go:generate goGenerator/buildAndGenerate.sh

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/reeceappling/goUtils/v2/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/exp/maps"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

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
	//EntryTypeField() *string // TODO: GET RID OF // "entryType" field. Non-nil for main collection items
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
	Keys:    bson.D{{"lastUpdated", -1}},
	Options: options.Index().SetName("lastUpdated"),
}
var standardIndexModel = newSimpleIndex("standard", "standard", true, false, false)
var projectsIndexModel = newSimpleIndex("projects", "acl.projects.$**", false, false, false) // TODO: ensure actually indexes the correct thing! // TODO: this is a wildcard index!!!!
var saleIndexModel = newSimpleIndex("sale", "sale", false, true, false)
var transfersOutIndexModel = newSimpleIndex("transfersOut", "transfersOut", false, true, false)
var creationDateIndexModel = newSimpleIndex("creationDate", "createDate", true, false, false)
var disposedIndexModel = newSimpleIndex("disposed", "disposed", false, true, false)

var aliasesIndexModel = newSimpleIndex("aliases", "aliases", false, true, false)

//// TODO: searching in a specific index
//func latestNUpdatedB(ctx context.Context) error { // TODO: fixMe
//	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
//	//indx := // TODO: use correct index
//	opts := options.Find().SetHint(
//		mongo.IndexModel{Keys: bson.D{{"transfersOut", 1}}}, // TODO: ????????????
//		)
//	coll := db.Collection(FruitsCollName)
//	_, err := coll.Find(ctx, bson.D{}, opts)
//	return err
//	//coll.UpdateByID(ctx, bson.D{{"_id": "someId"}}, ) // TODO: use this
//}

func withUpdateNow() primitive.E {
	return primitive.E{
		Key:   "lastUpdated",
		Value: unixTimeForNow(),
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
		//Initialize Collections with predefined items
		"agar Recipe":      initializeAgarRecipes,
		"jar Recipe":       initializeJarRecipes,
		"lc Recipe":        initializeLcRecipes,
		"substrate Recipe": initializeSubstrates,
		"species":          initializeSpecies,
		"subspecies":       initializeSubspecies,
		// Initialize other alt collections
		"agar batch":        initializeAgarBatches,
		"pc run":            initializePCRun,
		"sales":             initializeSales,
		"transfer":          initializeTransfers,
		"projects":          initializeProjects,
		"substrate batches": initializeSubstrateBatches,
		// initialize users
		"users": initializeUsers,
	} {
		if err := initializer(ctx); err != nil {
			return errors.Join(fmt.Errorf(`%s initializer failed`, i), err)
		}
	}
	for name, b58IdStr := range map[string]string{
		// Mains IDs
		"plate":           string(exPlate.asBase58()),
		"bag":             string(exBag.asBase58()),
		"fruitingChamber": string(exFC.asBase58()),
		"jar":             string(exJar.asBase58()),
		"mss":             string(exMSS.asBase58()),
		"slant":           string(exSlant.asBase58()),
		"stasisTube":      string(exStasis.asBase58()),
		"fruit":           string(exFruitId.base58Bytes()),
		"sporePrint":      string(exSporePrint.base58Bytes()),
		// Standard Alt IDs
		"agarBatch":       string(exAltId.base58Bytes()),
		"agarRecipe":      string(exAltId.base58Bytes()),
		"jarRecipe":       string(exAltId.base58Bytes()),
		"lcRecipe":        string(exAltId.base58Bytes()),
		"sale":            string(exAltId.base58Bytes()),
		"substrateRecipe": string(exAltId.base58Bytes()),
		"transfer":        string(exAltId.base58Bytes()),
		// String Alt IDs
		"project":    testEntryStringId,
		"species":    testEntryStringId,
		"subspecies": testEntryStringId,
	} {
		println(fmt.Sprintf(`test %s can be found at /view/%s/%s`, name, name, b58IdStr))
	}

	return nil
}

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
			"$push", bson.D{{fieldName, vals[0]}},
		}}
	case 0:
		return bson.D{}
	default:
		return bson.D{{"$push", bson.D{{fieldName, bson.D{{"$each", vals}}}}}} // TODO: ensure this works
	}
}

func pushToArrayNew[T any](fieldName string, vals ...T) bson.D { // TODO: rename
	switch len(vals) {
	case 1:
		return bson.D{{fieldName, vals[0]}}
	case 0:
		return bson.D{}
	default:
		return bson.D{{fieldName, bson.D{{"$each", vals}}}}
	}
}

//func addNotes(notes ...Note) bson.D {
//	return pushToArray("notes", notes...)
//}

func withUpdate(t *time.Time) bson.E {
	return bson.E{"lastUpdated", unixTimeFor(utils.Default(t, time.Now()))}
}

func withItemsRemoved[T any](field string, items ...T) bson.D {
	itemsEquality := make([]bson.E, len(items))
	for i, item := range items {
		itemsEquality[i] = bson.E{"$eq", item}
	}

	//{ "$pull": { <field1>: <value|condition>, <field2>: <value|condition>, ... } }
	return bson.D{{"$pull", bson.D{{field, itemsEquality}}}}
}

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
		var existingIndex mongo.IndexModel // TODO: ensure ok
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

func entryTypeFor(inp string) (CollectionItem, error) { // TODO: does not work for Projects?
	switch strings.ToLower(inp) {
	case "bag",
		"bags":
		return &Bag{}, nil
	case "box", "fruitingchamber", "chamber", "fruiting chamber",
		"boxes", "fruitingchambers", "chambers", "fruiting chambers":
		return &FruitingChamber{}, nil
	case "jar", "grainjar", "grain jar",
		"jars", "grainjars", "grain jars":
		return &GrainJar{}, nil
	case "lc", "liquidculture", "liquid culture",
		"lcs", "liquidcultures", "liquid cultures":
		return &LiquidCulture{}, nil
	case "lcSyringe", "lcSyringes":
		return &LcSyringe{}, nil
	case "plugs", "plug", "peg", "pegs":
		return &PlugsJar{}, nil
	case "mss", "sporesyringe", "spore syringe", "multisporesyringe", "multi spore syringe",
		"msss", "sporesyringes", "spore syringes", "multisporesyringes", "multi spore syringes":
		return &MSS{}, nil
	case "plate", "dish", "agarplate", "agar plate", "agardish", "agar dish", "petri", "petridish", "petri dish",
		"plates", "dishes", "agarplates", "agar plates", "agardishes", "agar dishes", "petris", "petridishes", "petri dishes":
		return &Plate{}, nil
	case "slant", "slants":
		return &Slant{}, nil
	case "stasistube", "stasis tube", "stasis", "tube",
		"stasistubes", "stasis tubes", "tubes":
		return &StasisTube{}, nil
	case "agarbatch", "agar batch",
		"agarbatches", "agar batches":
		return &AgarBatch{}, nil
	case "agarrecipe", "agar recipe",
		"agarrecipes", "agar recipes":
		return &AgarRecipe{}, nil
	case "fruit",
		"fruits":
		return &Fruit{}, nil
	case "jarrecipe", "jar recipe",
		"jarrecipes", "jar recipes":
		return &JarRecipe{}, nil
	case "lcrecipe", "lc recipe", "liquidculturerecipe", "liquid culture recipe",
		"lcrecipes", "lc recipes", "liquidculturerecipes", "liquid culture recipes":
		return &LcRecipe{}, nil
	case "pcrun", "pc run", "pressure cooker run", "pressure cooker", "pc", "pressurecooker", "run",
		"pcruns", "pc runs", "pressure cooker runs", "pressure cookers", "pcs", "pressurecookers", "runs":
		return &PCRun{}, nil
	case "project", "Projects":
		return &Project{}, nil
	case "sale", "sales":
		return &Sale{}, nil
	case "species":
		return &Species{}, nil
	case "subspecies":
		return &Subspecies{}, nil
	case "sporeprint", "spore print", "print",
		"sporeprints", "spore prints", "prints":
		return &SporePrint{}, nil
	case "substrate", "substraterecipe", "substrate recipe",
		"substrates", "substraterecipes", "substrate recipes":
		return &SubstrateRecipe{}, nil
	case "transfer", "xfer",
		"transfers", "xfers":
		return &Transfer{}, nil
	case "user", "users":
		return &User{}, nil
	default:
		return nil, errors.Join(ErrInvalidEntryType, errors.New("invalid collection input. Does not map to a collection name"))
	}
}

func getStandardEntries[T CollectionItem](ctx context.Context, temp T) (out []T, err error) {
	cursor, err := ctx.Value(mongoClientContextKey).(*mongo.Client).
		Database(dbName).
		Collection(temp.CollectionName()).
		Find(ctx, bson.D{{"standard", true}}) // TODO: NOT WORKING PROPERLY!!!!!
	if err != nil {
		return nil, err
	}
	return getCollectionItemsFromCursor[T](ctx, cursor, nil)
}

func getCollectionItemsFromCursor[T CollectionItem](ctx context.Context, cursor *mongo.Cursor, numItems *int) ([]T, error) {
	results := []T{}
	if numItems != nil {
		results = make([]T, 0, *numItems)
	}
	user, err := GetAuthInfo(ctx)
	if err != nil {
		return nil, err
	}
	for numItems == nil || len(results) < *numItems {
		if cursor.TryNext(ctx) {
			var result T

			//result := reflect.New(reflect.TypeOf(entryType)). // TODO: elem ok here?
			if err := cursor.Decode(&result); err != nil {
				return nil, err
			} // TODO: no pointer ok here?
			permedItem, ok := interface{}(result).(Permissioned)
			if ok {
				if permedItem.Permissions().HighestPermFor(user) == nil {
					// Skip this entry
					continue
				}
			}
			// TODO: ok to get rid of this?
			//resultCollItem, ok := interface{}(result).(CollectionItem)
			//if !ok {
			//	err = fmt.Errorf(`invalid collection result at index %d. THIS SHOULD NEVER HAPPEN`, len(results))
			//	bs, errr := json.MarshalIndent(result, "", " ")
			//	if errr != nil {
			//		err = errors.Join(err, errr)
			//	} else {
			//		println(string(bs))
			//	}
			//	return nil, fmt.Errorf(`invalid collection result at index %d. THIS SHOULD NEVER HAPPEN`, len(results))
			//}

			results = append(results, result)
			continue
		}
		cursorClosed := cursor.ID() == 0
		if cursorClosed && len(results) == 0 {
			return results, mongo.ErrNoDocuments // TODO: ok? or will this cause other problems?
		}
		if err := cursor.Err(); err != nil {
			return nil, err
		}
		if cursorClosed {
			break
		}
	}
	return results, nil
}

type picWithNotesForm struct {
	Time  unixTime         `json:"time"`
	Img   string           `json:"img"`
	Notes AllEntries[Note] `json:"notes"`
}

func (pwn picWithNotesForm) convert() PicWithNotes {
	return PicWithNotes{
		Time:       pwn.Time,
		Location:   imageLocation(pwn.Img),
		NotesField: NotesField{pwn.Notes.asEntries()},
	}
}

type contamForm struct {
	Time      unixTime         `json:"time"`
	Confirmed bool             `json:"confirmed"`
	Bacteria  bool             `json:"bacteria"`
	Mold      bool             `json:"mold"`
	Notes     AllEntries[Note] `json:"notes"`
	Location  *string          `json:"location,omitempty"` // MAY OR MAY NOT EXIST ON RESPONSE
}

func (cf contamForm) convert() Contamination {
	var loc *imageLocation = nil
	if cf.Location != nil {
		loc = utils.Pointer(imageLocation(*cf.Location))
	}
	return Contamination{
		Time:       cf.Time,
		Confirmed:  cf.Confirmed,
		Bacteria:   cf.Bacteria,
		Mold:       cf.Mold,
		Location:   loc,
		NotesField: NotesField{cf.Notes.asEntries()},
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
	for i, updatedNote := range updated.Notes.Existing {
		if updatedNote.Disabled {
			return false
		}
		if updatedNote.Data.Note != existing.Notes[i].Note {
			return false
		}
	}
	return true
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
	img, err := jpeg.Decode(p)
	if err != nil {
		img, err = png.Decode(p)
		if err != nil {
			http.Error(w, "failed to read image as either jpeg as png! "+err.Error(), http.StatusBadRequest)
			return nil, err
		}
	}
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, img, nil) // TODO: JPEG OR PNG??????
	if err != nil {
		http.Error(w, "failed to encode image to save! "+err.Error(), http.StatusInternalServerError)
		return nil, err
	}
	return buf.Bytes(), nil
}

func handleWriteErr(err error, w http.ResponseWriter) {
	if err != nil {
		println("failed to write! " + err.Error()) // TODO: SOMETHING HERE!
	}
}

func handleFileDeleteErr(err error) {
	println("failed to delete file! " + err.Error()) // TODO: SOMETHING HERE!
}

var (
	testEntryStringId    = "testEntry"
	exAltId              = altCollIdForint(0)
	exFruitId            = mainCollIdForint(idTestFruit)
	exampleTime          = unixTimeFor(time.Date(2024, 12, 29, 0, 0, 0, 0, time.UTC))
	exampleSpecies       = "beech"
	exampleSubspecies    = utils.Pointer("brown beech")
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
	exMSS                = mainCollIdForint(idTestMSS)
	exPlugId             = mainCollIdForint(idTestPlug)
	exSlant              = mainCollIdForint(idTestSlant)
	exStasis             = mainCollIdForint(idTestStasis)
	exSwabId             = mainCollIdForint(idTestSwab)
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
		BlanketPerm: false,
	}
	testAcl = ACL{
		// TODO: ensure ok
		Users: map[string]bool{
			exUserNoProjectRead:  false,
			exUserNoProjectWrite: true,
		},
		Projects: map[projectName]bool{
			exProjRead:  false,
			exProjWrite: true,
		},
		BlanketPerm: false,
	}
	exBool   = utils.Pointer(true)
	exPicLoc = "test.jpg" // TODO: make sure this exists
	exPic    = PicWithNotes{
		Time:       exampleTime,
		Location:   imageLocation(exPicLoc),
		NotesField: NotesField{exampleNotes()},
	}
	exPics = []PicWithNotes{exPic, exPic}
	ec     = Contamination{
		Time:       exampleTime,
		Confirmed:  false,
		Bacteria:   false,
		Mold:       true,
		Location:   (*imageLocation)(&exPicLoc),
		NotesField: NotesField{exampleNotes()},
	}
	exContams = []Contamination{ec, ec}
)

func exampleNotes() []Note {
	return []Note{{
		Time: exampleTime,
		Note: "example note A",
	}, {
		Time: exampleTime,
		Note: "example note B",
	}}
}

func decodeItem[T any](item *T, encoded *mongo.SingleResult) (err error) {
	err = encoded.Decode(item) // TODO: was pointer
	if err != nil {
		err = errors.Join(errors.New("failed to decode"), err)
	}
	return
}

func CollectionFor(item CollectionItem, db *mongo.Database) *mongo.Collection {
	return db.Collection(item.CollectionName())
}
func Refresh[T CollectionItem](ctx context.Context, db *mongo.Database, item *T) error {
	return CollectionFor(*item, db).FindOne(ctx, bson.D{{"_id", (*item).IdValue( /* TODO: PROBABLY WONT WORK*/ )}}).Decode(item)
}

func finishMainCollItemUpdate[T MainCollectionItem](ctx context.Context, w http.ResponseWriter, coll *mongo.Collection, modsFor func(T, AclField) (bson.D, error), existing T, reqPerms PermsOnRequest) {
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

func finishAltCollItemUpdate[T PermissionedAltCollectionItem[AlternateCollectionId]](ctx context.Context, w http.ResponseWriter, coll *mongo.Collection, modsFor func(T, AclField) (bson.D, error), existing T, reqPerms PermsOnRequest) {
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
