package rfid

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/reeceappling/goUtils/v2/utils"
	sliceutils "github.com/reeceappling/goUtils/v2/utils/slices"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

var dbName string

func init() {
	var ok bool = true
	dbName, ok = os.LookupEnv("MONGO_INITDB_DATABASE")
	if !ok {
		dbName = "mushdb"
		println("missing env var MONGO_INITDB_DATABASE")
	}
}

type imageLocation string
type unixTime int64 // unixMilli!

func unixTimeFor(t time.Time) unixTime {
	return unixTime(t.UnixMilli())
}
func unixTimeForNow() unixTime {
	return unixTimeFor(time.Now())
}

func (t unixTime) asCreationDate() CreationDateField {
	return CreationDateField{t}
}

var (
	_ subdocWithImage = PicWithNotes{}
	_ subdocWithImage = Contamination{}
	_ subdocWithImage = PicWithNotes{}
)

type subdocWithImage interface {
	getPicWithNotes() *PicWithNotes
}

// TODO: USE
func getLatestExistingImage(possibleSubdocs ...subdocWithImage) *PicWithNotes { // TODO: use?
	var out *PicWithNotes = nil
	latestTime := unixTime(time.Date(1995, 12, 29, 0, 0, 0, 0, nil).UnixMilli())
	for _, subdoc := range possibleSubdocs {
		pwn := subdoc.getPicWithNotes()
		if pwn.Time > latestTime {
			latestTime = pwn.Time
			out = pwn
		}
	}
	return out
}

type PicsField struct {
	Pics []PicWithNotes `bson:"pics,omitempty" json:"pics,omitempty"`
}

type FlushesField struct {
	Flushes []PicWithNotes `bson:"flushes,omitempty" json:"flushes,omitempty"`
}

type MostRecentImageField struct {
	MostRecentImage *PicWithNotes `bson:"mostRecentImage,omitempty" json:"mostRecentImage,omitempty"`
}

type PicWithNotes struct {
	Time       unixTime      `bson:"time" json:"time"`
	Location   imageLocation `bson:"location" json:"location"`
	NotesField `bson:"inline"`
}

func picsWithoutNotes(inp []PicWithNotes) []PicWithNotes {
	out := make([]PicWithNotes, len(inp))
	for i, pic := range inp {
		out[i] = PicWithNotes{
			Time:       pic.Time,
			Location:   pic.Location,
			NotesField: NotesField{[]Note{}},
		}
	}
	return out
}

func (pwn PicWithNotes) withoutNotes() PicWithNotes {
	return PicWithNotes{
		Time:       pwn.Time,
		Location:   pwn.Location,
		NotesField: NotesField{[]Note{}},
	}
}

type PicWithNotesLessLocation struct {
	Time       unixTime `bson:"time" json:"time"`
	NotesField `bson:"inline"`
}

func (p PicWithNotesLessLocation) asPicWithNotes(location *string) PicWithNotes {
	return PicWithNotes{
		Time:       p.Time,
		Location:   imageLocation(utils.Default(location, "")),
		NotesField: p.NotesField,
	}
}

func (p PicWithNotes) getPicWithNotes() *PicWithNotes {
	return &p
}

type ContaminationsField struct {
	Contaminations []Contamination `bson:"contamination,omitempty" json:"contamination,omitempty"`
}

type Contamination struct {
	Time       unixTime       `bson:"time" json:"time"`
	Confirmed  bool           `bson:"confirmed" json:"confirmed"`
	Bacteria   bool           `bson:"bacteria" json:"bacteria"`
	Mold       bool           `bson:"mold" json:"mold"`
	Location   *imageLocation `bson:"location,omitempty" json:"location,omitempty"`
	NotesField `bson:"inline"`
}

type ContaminationLessLocation struct {
	Time      unixTime `bson:"time" json:"time"`
	Confirmed bool     `bson:"confirmed" json:"confirmed"`
	Bacteria  bool     `bson:"bacteria" json:"bacteria"`
	Mold      bool     `bson:"mold" json:"mold"`
	NotesField
}

func (c ContaminationLessLocation) asContamination(location *imageLocation) Contamination {
	return Contamination{
		Time:       c.Time,
		Confirmed:  c.Confirmed,
		Bacteria:   c.Bacteria,
		Mold:       c.Mold,
		Location:   location,
		NotesField: c.NotesField,
	}
}

func (c Contamination) getPicWithNotes() *PicWithNotes {
	if c.Location == nil {
		return nil
	}
	return &PicWithNotes{
		Time:       c.Time,
		Location:   *c.Location,
		NotesField: c.NotesField,
	}
}

func contamUpdates(contams SplitEntries[contamForm, ContaminationLessLocation]) SplitEntries[contamForm, Contamination] {
	return SplitEntries[contamForm, Contamination]{
		Existing: contams.Existing,
		New: sliceutils.Map(contams.New, func(i ContaminationLessLocation) Contamination {
			return i.asContamination(nil)
		}),
	}
}

func imageUpdates(Images SplitEntries[picWithNotesForm, PicWithNotesLessLocation]) SplitEntries[picWithNotesForm, PicWithNotes] {
	return SplitEntries[picWithNotesForm, PicWithNotes]{
		Existing: Images.Existing,
		New: sliceutils.Map(Images.New, func(i PicWithNotesLessLocation) PicWithNotes {
			return i.asPicWithNotes(nil)
		}),
	}
}

type liquid struct {
	Name fluid   `bson:"name" json:"name"`
	Pct  float64 `bson:"pct" json:"pct"`
}

func (l liquid) withPct(pct float64) liquid {
	return liquid{
		Name: l.Name,
		Pct:  pct,
	}
}

type fluid string

var fluids = []fluid{Water, DistilledWater, GrainWater}

// TODO: add all of these to autogenned
var (
	Water          = fluid("water")
	DistilledWater = fluid("distilledWater")
	GrainWater     = fluid("grain water")
)

func (f fluid) AsLiquid(pct ...float64) liquid {
	val := 100.0
	if len(pct) != 0 {
		val = pct[0]
	}
	if val > 100.0 || val <= 0.0 {
		panic("invalid fluid percentage") // TODO: ok?
	}
	return liquid{
		Name: f,
		Pct:  val,
	}
}

type Note struct {
	Time unixTime `bson:"time" json:"time"`
	Note string   `bson:"note" json:"note"`
}

func (n Note) GenerationIfExists() *int {
	if !strings.Contains(n.Note, "Generation: ") {
		return nil
	}
	spl := strings.Split(n.Note, " ")
	if len(spl) != 2 {
		return nil
	}
	out, err := strconv.Atoi(spl[1])
	if err != nil {
		return nil
	}
	return &out
}

// TODO: add all of these to autogenned
type nutrientMeasurement struct {
	Nutrient nutrient `bson:"nutrient" json:"nutrient"` // Nutrient name
	Amount   float64  `bson:"amount" json:"amount"`     // Amount per 1L agar
	Unit     string   `bson:"unit" json:"unit"`         // mL, tsp, tbsp, drop, pinch, cup, etc
}

type nutrient string

// TODO: add all of these to autogenned
var nutrients = []nutrient{LME, Potato}
var (
	LME    nutrient = "light malt extract"
	Potato nutrient = "potato flakes"
)

// TODO: add all of these to autogenned
type sugarMeasurement struct {
	Type   sugar   `bson:"type" json:"type"`     // Sugar Username
	Amount float64 `bson:"amount" json:"amount"` // Amount per 1L agar
	Unit   string  `bson:"unit" json:"unit"`
}

type sugar string

var sugars = []sugar{Dextrose, Honey, MapleSyrup}

func newSugarMeasurement(add sugar, amount float64, unit string) sugarMeasurement {
	return sugarMeasurement{
		Type:   add,
		Amount: amount,
		Unit:   unit,
	}
}

// TODO: add all of these to autogenned
var (
	Dextrose   sugar = "dextrose" // This is corn syrup
	Honey      sugar = "honey"
	MapleSyrup sugar = "maple syrup"
)

type grain string

var grains = []grain{Rye, Wheat, Oats, Millett, Popcorn, BirdSeed}

func (g grain) Validate() error {
	if !slices.Contains(grains, g) {
		return errors.New("invalid grain")
	}
	return nil
}

// TODO: add all of these to autogenned
var (
	Rye      grain = "rye"
	Wheat    grain = "wheat"
	Millett  grain = "millett"
	Oats     grain = "oats"
	Popcorn  grain = "popcorn"
	BirdSeed grain = "birdseed"
)

func writeAsJson(w http.ResponseWriter, obj any) {
	bs, err := json.Marshal(obj)
	if err != nil {
		http.Error(w, "failed to marshal output: "+err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(bs)
	handleWriteErr(err, w)
}

type additiveMeasurement struct {
	Additive additive `bson:"additive" json:"additive"` // Nutrient name
	Amount   float64  `bson:"amount" json:"amount"`     // Amount per 1L agar
	Unit     string   `bson:"unit" json:"unit"`         // mL, tsp, tbsp, drop, pinch, cup, etc
}

func newAdditiveMeasurement(add additive, amount float64, unit string) additiveMeasurement {
	return additiveMeasurement{
		Additive: add,
		Amount:   amount,
		Unit:     unit,
	}
}

type colorant string

var colorants = []colorant{clearColor, black, blue, yellow, orange, red}

// TODO: add all of these to autogenned
var (
	clearColor colorant = "Clear"
	black      colorant = "Black"
	blue       colorant = "Blue"
	green      colorant = "Green"
	yellow     colorant = "Yellow"
	orange     colorant = "Orange"
	red        colorant = "Red" // MOST REDS ARE FUNGICIDAL
)

var colors = map[string]colorant{
	string(clearColor): clearColor,
	string(black):      black,
	string(blue):       blue,
	string(green):      green,
	string(yellow):     yellow,
	string(orange):     orange,
}

func ValidColor(c colorant) bool {
	_, ok := colors[string(c)]
	return ok
}

type additive string // TODO: ACCOUNT FOR THIS EVERYWHERE!
var additives = []additive{Vermiculite, Perlite, Gypsum}

// TODO: add all of these to autogenned
var (
	Vermiculite additive = "vermiculite"
	Perlite     additive = "perlite"
	Gypsum      additive = "gypsum"
)

type antibiotic string

// TODO: add all of these to autogenned
var antibiotics = []antibiotic{HydrogenPeroxide, Doxycycline}

var (
	HydrogenPeroxide antibiotic = "HydrogenPeroxide"
	Doxycycline      antibiotic = "Doxycycline"
)

// TODO: use
var antibioticDosages = map[antibiotic]string{ // TODO: USE THIS!
	Doxycycline:      "unknown as of right now", // TODO: figure out measurements
	HydrogenPeroxide: "unknown as of right now", // TODO: figure out measurements
}

type Generation int

func (gen *Generation) Next() *Generation {
	if gen == nil {
		return nil
	}
	nextGen := (*gen) + 1
	return &nextGen
}

func NewMods() *Mods {
	return &Mods{}
}

type Mods struct {
	err    error    // SKIP WHEN THIS IS NON-NIL
	unsets []bson.E // Key, "" // TODO: ?
	sets   []bson.E // Key, value
	pushes []bson.E // FieldName, valueToPush
	pulls  []bson.E //{ "$pull": { <field1>: <value|condition>, <field2>: <value|condition>, ... } }
}

func (upd *Mods) Add(sets, unsets, pushes, pulls []bson.E) *Mods {
	if sets != nil && len(sets) > 0 {
		upd.sets = append(upd.sets, sets...)
	}
	if unsets != nil && len(unsets) > 0 {
		upd.unsets = append(upd.unsets, unsets...)
	}
	if pushes != nil && len(pushes) > 0 {
		upd.pushes = append(upd.pushes, pushes...)
	}
	if pulls != nil && len(pulls) > 0 {
		upd.pulls = append(upd.pulls, pulls...)
	}
	return upd
}

func (upd *Mods) Set(key string, value interface{}) *Mods { // TODO: interface ok?
	upd.sets = append(upd.sets, bson.E{Key: key, Value: value})
	return upd
}
func (upd *Mods) SetNew(key string, value interface{}) { // TODO: interface ok?
	upd.sets = append(upd.sets, bson.E{Key: key, Value: value})
}

//func (upd *Mods) UpdatePointerIfNeeded(key string, future, current *interface{}) *Mods {
//	return updatePointerIfNeeded(upd, key, future, current)
//}

func (upd *Mods) UpdateValueIfNeeded(key string, future, current interface{}) *Mods { // TODO: interface ok?
	return updateValueIfNeeded(upd, key, future, current)
}

func (upd *Mods) Unset(key string) *Mods {
	upd.unsets = append(upd.unsets, bson.E{Key: key, Value: ""}) // TODO: "" ok here?
	return upd
}

func (upd *Mods) Push(key string, value interface{}) *Mods { // TODO: interface ok? or slice?
	upd.pushes = append(upd.pushes, bson.E{Key: key, Value: value})
	return upd
}

// TODO: write what this actually does here!
func (upd *Mods) pushBson(pushValues ...bson.E) *Mods { // TODO: make sure ok
	if len(pushValues) == 0 {
		return upd
	}
	upd.pushes = append(upd.pushes, pushValues...)
	return upd
}

func (upd *Mods) Pull(key string, value interface{}) *Mods { // TODO: interface ok? or slice?
	upd.pulls = append(upd.pulls, bson.E{Key: key, Value: value})
	return upd
}

func (upd *Mods) IsEmpty() bool {
	return upd == nil || len(upd.sets)+len(upd.unsets)+len(upd.pushes)+len(upd.pulls) == 0
}

func (upd *Mods) Finalized() (bson.D, error) { // TODO: validate this works as intended (is bson.D ok?)
	if upd.err != nil {
		return nil, upd.err
	}
	out := bson.D{}
	if upd.sets != nil && len(upd.sets) != 0 {
		out = append(out, bson.E{"$set", upd.sets})
	}
	if upd.pushes != nil && len(upd.pushes) != 0 {
		out = append(out, bson.E{"$push", upd.pushes})
	}
	if upd.pulls != nil && len(upd.pulls) != 0 {
		out = append(out, bson.E{"$pull", upd.pulls})
	}
	if upd.unsets != nil && len(upd.unsets) != 0 {
		out = append(out, bson.E{"$unset", upd.unsets})
	}
	return out, nil
}

func (upd *Mods) addTransferOut(xferId AlternateCollectionId) *Mods {
	return upd.Push("transfersOut", xferId) // TODO: will this work for nonexisting field?
}

func (upd *Mods) addSaleToSales(saleId AlternateCollectionId, currentSales []AlternateCollectionId) *Mods { // TODO; USE!
	// TODO: ENSURE NOT ALREADY EXISTS
	return upd.Push("sales", saleId) // TODO: will this work for nonexisting field?
}
func (upd *Mods) addOnlySale(saleId AlternateCollectionId) *Mods { // TODO; USE!
	// TODO: ensure the sale did not already exist first
	return upd.Push("sale", saleId) // TODO: will this work for nonexisting field?
}

func (upd *Mods) updateAliasesIfNeeded(future, existing []string) *Mods {
	futureIsEmpty := future == nil || len(future) == 0
	existingIsEmpty := existing == nil || len(existing) == 0
	if existingIsEmpty {
		if futureIsEmpty {
			// Do nothing
			return upd
		}
		// future is not empty, so set it
		return upd.Set("aliases", future)
	}
	if futureIsEmpty {
		return upd.Unset("aliases") // unset aliases
	}
	if len(future) != len(existing) {
		return upd.Set("aliases", future)
	}
	for i := 0; i < len(future); i++ {
		if !slices.Contains(existing, future[i]) {
			return upd.Set("aliases", future)
		}
	}
	return upd
}

func (upd *Mods) updateDisposedIfNeeded(future, existing *unixTime) *Mods {
	return updatePointerIfNeeded(upd, "disposed", future, existing) // TODO: ok if this can be rolled back????
}

func (upd *Mods) updateKnownFruitableIfNeeded(future, existing *bool) *Mods {
	return updatePointerIfNeeded(upd, "knownFruitable", future, existing)
}

func (upd *Mods) updateConfirmedCleanIfNeeded(future, existing *bool) *Mods {
	return updatePointerIfNeeded(upd, "confirmedClean", future, existing)
}

func (upd *Mods) updateProjectCompletedIfNeeded(future, existing *unixTime) *Mods {
	return updatePointerIfNeeded(upd, "completed", future, existing)
}

func (upd *Mods) updateProjectPermsIfNeeded(future, existing ProjectPerms) *Mods {
	if future.Equal(existing) { // TODO: validate works
		return upd
	}
	// TODO: THIS IS NOT WORKING PROPERLY!!!!!
	bs, err := json.MarshalIndent(existing, "", " ") // TODO: del
	if err != nil {                                  // TODO: del
		panic(err) // TODO; del // TODO: del
	} // TODO: del
	println("existing", string(bs))               // TODO: del
	bs, err = json.MarshalIndent(future, "", " ") // TODO: del
	if err != nil {                               // TODO: del
		panic(err) // TODO; del
	} // TODO: del
	println("final", string(bs)) // TODO: del
	return upd.Set("perms", future)
}

func (upd *Mods) updateLastUpdatedIfNeeded() *Mods {
	if upd.IsEmpty() {
		return upd
	}
	return upd.Set("lastUpdated", unixTimeFor(time.Now()))
}

func (upd *Mods) updatePermsIfNeeded(next, current *ACL) *Mods {
	if current.Equivalent(next) {
		return upd
	}
	return upd.Set("acl", next)
}

func (upd *Mods) updateDefaultEntryPermsIfNeeded(nextReq PermsOnRequest, current *ACL) *Mods {
	next := nextReq.DefaultAcl()
	if current.Equivalent(next) {
		return upd
	}
	return upd.Set("defaultAcl", next)
}

func (upd *Mods) updateNameIfNeeded(future, existing string) *Mods {
	return updateValueIfNeeded(upd, "name", future, existing)
}

func notesWereModified(existing []Note, updated AllEntries[Note]) (hasChanged bool) {
	if len(updated.New) > 0 {
		return true
	}
	for i, finalExisting := range updated.Existing {
		if finalExisting.Disabled {
			return true
		}
		if finalExisting.Data.Note != existing[i].Note {
			return true
		}
		if finalExisting.Data.Time != existing[i].Time { // TODO: do we even want this?
			return true
		}
	}
	return false
}

func picsWereModified(existing []PicWithNotes, updated SplitEntries[picWithNotesForm, PicWithNotes]) (hasChanged bool) {
	if len(updated.New) > 0 {
		return true
	}
	for i, finalExisting := range updated.Existing {
		if finalExisting.Disabled ||
			finalExisting.Data.Img != string(existing[i].Location) || // TODO: ensure ok
			finalExisting.Data.Time != existing[i].Time || // TODO: do we even want this?
			notesWereModified(existing[i].Notes, finalExisting.Data.Notes) {
			return true
		}
	}
	return false

}

func contamsWereModified(existing []Contamination, updated SplitEntries[contamForm, Contamination]) bool {
	if len(updated.New) > 0 {
		return true
	}
	for i, finalExisting := range updated.Existing {
		if finalExisting.Disabled ||
			finalExisting.Data.Time != existing[i].Time || // TODO: do we even want this?
			finalExisting.Data.Mold != existing[i].Mold ||
			finalExisting.Data.Bacteria != existing[i].Bacteria ||
			finalExisting.Data.Confirmed != existing[i].Confirmed ||
			notesWereModified(existing[i].Notes, finalExisting.Data.Notes) {
			return true
		}
		if existing[i].Location != nil {
			if finalExisting.Data.Location == nil || *finalExisting.Data.Location != string(*existing[i].Location) { // TODO: ensure ok
				return true
			}
		}
	}
	return false
}

func (upd *Mods) updateNotesIfNeeded(updatedEntries AllEntries[Note], existing []Note) *Mods { // TODO: make sure this always works the way we want it to!!! // TODO: lower-down notes?
	if upd.err != nil {
		return upd
	}
	if len(updatedEntries.Existing) != len(existing) {
		upd.err = errors.Join(errors.New("length of existing notes must match"), upd.err)
		return upd
	}
	if !notesWereModified(existing, updatedEntries) {
		return upd
	}
	finalNotes := make([]Note, 0, len(existing)+len(updatedEntries.New))
	for _, final := range updatedEntries.Existing {
		if !final.Disabled {
			finalNotes = append(finalNotes, final.Data)
		}
	}
	for _, final := range updatedEntries.New {
		finalNotes = append(finalNotes, final.Data)
	}
	// Set notes
	return upd.Set("notes", finalNotes) // TODO: ensure ok
}

func (upd *Mods) updatePicsIfNeeded(updatedEntries SplitEntries[picWithNotesForm, PicWithNotes], existing []PicWithNotes) *Mods { // TODO: make sure this works as anticipated
	return upd.updatePwnIfNeeded("pics", updatedEntries, existing)
}

func (upd *Mods) updateFlushesIfNeeded(updatedEntries SplitEntries[picWithNotesForm, PicWithNotes], existing []PicWithNotes) *Mods { // TODO: make sure this works as anticipated
	return upd.updatePwnIfNeeded("flushes", updatedEntries, existing)
}

func (upd *Mods) updatePwnIfNeeded(fieldName string, updatedEntries SplitEntries[picWithNotesForm, PicWithNotes], existing []PicWithNotes) *Mods { // TODO: make sure this works as anticipated
	if upd.err != nil {
		return upd
	}
	if len(updatedEntries.Existing) != len(existing) {
		upd.err = errors.Join(errors.New("length of existing "+fieldName+" must match"), upd.err)
		return upd
	}
	if !picsWereModified(existing, updatedEntries) {
		return upd
	}
	// TODO: THIS IS NOT WORKING????
	finalPics := make([]PicWithNotes, 0, len(existing)+len(updatedEntries.New))
	for _, final := range updatedEntries.Existing {
		if !final.Disabled {
			finalPics = append(finalPics, final.Data.convert())
		}
	}
	finalPics = append(finalPics, updatedEntries.New...)
	// Set field
	return upd.Set(fieldName, finalPics) // TODO: ensure ok

}

func (upd *Mods) updateContamsIfNeeded(updatedEntries SplitEntries[contamForm, Contamination], existing []Contamination) *Mods {
	if upd.err != nil {
		return upd
	}
	if len(updatedEntries.Existing) != len(existing) {
		upd.err = errors.Join(errors.New("length of existing contams must match"), upd.err)
		return upd
	}
	if !contamsWereModified(existing, updatedEntries) {
		return upd
	}
	finalEntries := make([]Contamination, 0, len(existing)+len(updatedEntries.New))
	for _, final := range updatedEntries.Existing {
		if !final.Disabled {
			finalEntries = append(finalEntries, final.Data.convert())
		}
	}
	finalEntries = append(finalEntries, updatedEntries.New...)
	// Set field
	return upd.Set("contamination", finalEntries) // TODO: ensure ok
}

func (upd *Mods) updateSaleIfNeeded(future, existing *AlternateCollectionId) *Mods {
	return updatePointerIfNeeded(upd, "sale", future, existing)
}

func (upd *Mods) updateSporePrintColorIfNeeded(future, existing *SporePrintColor) *Mods {
	return updatePointerIfNeeded(upd, "color", future, existing)
}

func (upd *Mods) updateSporePrintDensityIfNeeded(future, existing *SporePrintDensity) *Mods {
	return updatePointerIfNeeded(upd, "density", future, existing)
}

// TODO: Only used on plugs?
func (upd *Mods) updateSalesIfNeeded(future, existing []AlternateCollectionId) *Mods { // TODO: validate works as intended (Should this be structured like notes?)
	if upd.err != nil {
		return upd
	}
	if len(future) != len(existing) {
		return upd.Set("sales", future)
	}
	for i, sale := range future {
		if string(sale[:]) != string(existing[i][:]) {
			return upd.Set("sales", future)
		}
	}
	return upd
}

func (upd *Mods) updateStandardIfNeeded(future, existing bool) *Mods {
	return updateValueIfNeeded(upd, "standard", future, existing)
}

func (upd *Mods) withGens(genSpore, genFruitSpore *Generation) *Mods {
	out := setPointerIfNonNil(upd, "genSpore", genSpore)
	return setPointerIfNonNil(out, "genFruitOrSpore", genFruitSpore)
}

func (upd *Mods) withInnoc(xfer Transfer) *Mods {
	return upd.Set("innoc", xfer.Id)
}
func (upd *Mods) withKnownFruitable(knownFruitable *bool) *Mods {
	return setPointerIfNonNil(upd, "knownFruitable", knownFruitable)
}
func (upd *Mods) withLastUpdated(lastUpdatedTime unixTime) *Mods {
	return upd.Set("lastUpdated", lastUpdatedTime)
}
func (upd *Mods) withMostRecentImage(parentType *PicWithNotes) *Mods {
	return setPointerIfNonNil(upd, "mostRecentImage", parentType)
}
func (upd *Mods) withParentType(parentType *string) *Mods {
	return setPointerIfNonNil(upd, "parentType", parentType)
}
func (upd *Mods) withParent(parentId *MainCollectionId) *Mods {
	return setPointerIfNonNil(upd, "parent", parentId)
}
func (upd *Mods) withPerms(acl *ACL) *Mods {
	return setPointerIfNonNil(upd, "acl", acl)
}
func (upd *Mods) withPics(pics []PicWithNotes) *Mods {
	if len(pics) == 0 {
		return upd
	}
	return upd.Set("pics", pics)
}
func (upd *Mods) withSpecies(species *string) *Mods {
	return setPointerIfNonNil(upd, "species", species)
}
func (upd *Mods) withSubspecies(subsp *string) *Mods {
	return setPointerIfNonNil(upd, "subspecies", subsp) // TODO: make sure string is correct!!!
}

func setPointerIfNonNil[T any](upd *Mods, fieldName string, val *T) *Mods {
	if val == nil {
		return upd
	}
	return upd.Set(fieldName, *val)
}

func updateValueIfNeeded[T comparable](upd *Mods, fieldName string, future, existing T) *Mods {
	if upd.err != nil || future == existing {
		return upd
	}
	return upd.Set(fieldName, future)
}
func updatePointerIfNeeded[T comparable](upd *Mods, fieldName string, future, existing *T) *Mods {
	if upd.err != nil {
		return upd
	}
	existingIsNil := existing == nil
	if future == nil {
		if existingIsNil {
			return upd
		}
		return upd.Unset(fieldName)
	}
	if existingIsNil {
		return upd.Set(fieldName, *future)
	}
	return updateValueIfNeeded(upd, fieldName, *future, *existing)
}
func updatePointerIfNeededNew[T comparable](upd *Mods, fieldName string, future, existing *T) *Mods {
	if upd.err != nil {
		return upd
	}
	futureIsNil := future == nil
	existingIsNil := existing == nil
	if futureIsNil {
		if existingIsNil {
			return upd
		}
		return upd.Unset(fieldName)
	}
	if existingIsNil {
		return upd.Set(fieldName, *future)
	}
	return updateValueIfNeeded(upd, fieldName, *future, *existing)
}

var GetOptionsHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	opt := r.PathValue("optionsType")
	var toWrite any
	switch strings.ToLower(opt) {
	case "colors", "color",
		"colorants", "colorant":
		toWrite = colorants
		break
	case "liquids", "liquid",
		"fluids", "fluid":
		toWrite = fluids
		break
	case "nutrients", "nutrient":
		toWrite = nutrients
		break
	case "sugars", "sugar":
		toWrite = sugars
		break
	case "grains", "grain":
		toWrite = grains
		break
	case "additives", "additive":
		toWrite = additives
		break
	case "antibiotics", "antibiotic":
		toWrite = antibiotics
		break
	case "transferreasons", "transferreason":
		toWrite = transferReasons
		break
	case "sporePrintColors", "sporePrintColor":
		toWrite = sporePrintColors
		break
	case "sporePrintDensities", "sporePrintDensity":
		toWrite = sporePrintDensities
		break
		// TODO: any other cases???
	default:
		http.Error(w, fmt.Sprintf(`invalid option provided: "%s" is not one of [color,liquid,nutrient,sugar,grain,additive,transferReason]`, opt), http.StatusBadRequest)
		return
	}
	writeAsJson(w, toWrite)
}
