package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/reeceappling/goUtils/v2/utils"
	sliceutils "github.com/reeceappling/goUtils/v2/utils/slices"
	"github.com/reeceappling/mushDb/api/request/unix"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"os"
	"slices"
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

type ImageLocation string

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
	latestTime := unix.Time(time.Date(1995, 12, 29, 0, 0, 0, 0, nil).UnixMilli())
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

func (pics PicsField) getLatestPicFromPicsField() *PicWithNotes {
	var out *PicWithNotes = nil
	var latest unix.Time = 0
	for _, pic := range pics.Pics {
		if pic.Time > latest {
			latest = pic.Time
			out = &pic
		}
	}
	return out
}

type FlushesField struct {
	Flushes []PicWithNotes `bson:"flushes,omitempty" json:"flushes,omitempty"`
}

type MostRecentImageField struct {
	MostRecentImage *PicWithNotes `bson:"mostRecentImage,omitempty" json:"mostRecentImage,omitempty"`
}

type PicWithNotes struct {
	PicWithNotesLessLocation `bson:"inline"`
	Location                 ImageLocation `bson:"location" json:"location"`
}

func (pwn PicWithNotes) EqualTo(other PicWithNotes) bool {
	if pwn.Location != other.Location || pwn.Time != other.Time {
		return false
	}
	return pwn.NotesField.EqualTo(other.NotesField)
}

func newPicWithNotes(tim unix.Time, notes []Note, location ImageLocation) PicWithNotes {
	return PicWithNotes{
		PicWithNotesLessLocation: newPicWithNotesLessLocation(tim, notes),
		Location:                 location,
	}
}

func picsWithoutNotes(inp []PicWithNotes) []PicWithNotes {
	out := make([]PicWithNotes, len(inp))
	for i, pic := range inp {
		out[i] = PicWithNotes{
			PicWithNotesLessLocation: PicWithNotesLessLocation{
				RequiredTimeField: RequiredTimeField{Time: pic.Time},
				NotesField:        NotesField{[]Note{}},
			},
			Location: pic.Location,
		}
	}
	return out
}

func (pwn PicWithNotes) withoutNotes() PicWithNotes {
	return PicWithNotes{
		PicWithNotesLessLocation: *pwn.PicWithNotesLessLocation.withoutNotes(),
		Location:                 pwn.Location,
	}
}

type RequiredTimeField struct {
	Time unix.Time `bson:"time" json:"time"`
}

func newRequiredTimeField(t unix.Time) RequiredTimeField {
	return RequiredTimeField{t}
}

type PicWithNotesLessLocation struct {
	RequiredTimeField `bson:"inline"`
	NotesField        `bson:"inline"`
}

func newPicWithNotesLessLocation(t unix.Time, notes []Note) PicWithNotesLessLocation {
	return PicWithNotesLessLocation{
		RequiredTimeField: newRequiredTimeField(t),
		NotesField:        NotesField{notes},
	}
}
func (p PicWithNotesLessLocation) withoutNotes() *PicWithNotesLessLocation {
	return &PicWithNotesLessLocation{
		RequiredTimeField: p.RequiredTimeField,
		NotesField:        NotesField{Notes: []Note{}},
	}
}

func (p PicWithNotesLessLocation) asPicWithNotes(location *string) PicWithNotes {
	return PicWithNotes{
		PicWithNotesLessLocation: p,
		Location:                 ImageLocation(utils.Default(location, "")),
	}
}

func (p PicWithNotes) getPicWithNotes() *PicWithNotes {
	return &p
}

type ContaminationsField struct {
	Contaminations []Contamination `bson:"contamination,omitempty" json:"contamination,omitempty"`
}

func (contams ContaminationsField) getContamsLatestImage() *Contamination {
	var out *Contamination = nil
	var latest unix.Time = 0
	for _, contam := range contams.Contaminations {
		if contam.Location != nil && contam.Time > latest {
			latest = contam.Time
			out = &contam
		}
	}
	return out
}

type Contamination struct {
	ContaminationLessLocation `bson:"inline"` // TODO: new, ensure ok
	Location                  *ImageLocation  `bson:"location,omitempty" json:"location,omitempty"`
}

type ContaminationLessLocation struct {
	PicWithNotesLessLocation `bson:"inline"` // TODO: new, ensure ok
	Confirmed                bool            `bson:"confirmed" json:"confirmed"`
	Bacteria                 bool            `bson:"bacteria" json:"bacteria"`
	Mold                     bool            `bson:"mold" json:"mold"`
}

func (c ContaminationLessLocation) asContamination(location *ImageLocation) Contamination {
	return Contamination{
		ContaminationLessLocation: c,
		Location:                  location,
	}
}

func (c Contamination) getPicWithNotes() *PicWithNotes {
	if c.Location == nil {
		return nil
	}
	return &PicWithNotes{
		PicWithNotesLessLocation: c.PicWithNotesLessLocation,
		Location:                 *c.Location,
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

type Liquid struct {
	Name Fluid   `bson:"name" json:"name"`
	Pct  float64 `bson:"pct" json:"pct"`
}

func (l Liquid) withPct(pct float64) Liquid {
	return Liquid{
		Name: l.Name,
		Pct:  pct,
	}
}

type Fluid string

var fluids = []Fluid{Water, DistilledWater, GrainWater}

// TODO: add all of these to autogenned
var (
	Water          = Fluid("water")
	DistilledWater = Fluid("distilledWater")
	GrainWater     = Fluid("grain water")
)

func (f Fluid) AsLiquid(pct ...float64) Liquid {
	val := 100.0
	if len(pct) != 0 {
		val = pct[0]
	}
	if val > 100.0 || val <= 0.0 {
		panic("invalid fluid percentage") // TODO: ok?
	}
	return Liquid{
		Name: f,
		Pct:  val,
	}
}

type Note struct {
	RequiredTimeField `bson:"inline"`
	Note              string `bson:"note" json:"note"`
}

func (n Note) EqualTo(other Note) bool {
	if n.Time != other.Time || n.Note != other.Note {
		return false
	}
	return true
}

func newNote(tim unix.Time, txt string) Note {
	return Note{
		RequiredTimeField: RequiredTimeField{tim},
		Note:              txt,
	}
}

type NotesUpdateField struct {
	Notes AllEntries[Note] `json:"notes,omitempty"`
}

func (nuf NotesUpdateField) NoteChanges() AllEntries[Note] {
	return nuf.Notes
}

type NoteMods interface {
	NoteChanges() AllEntries[Note]
}

//func (n Note) GenerationIfExists() *int {
//	if !strings.Contains(n.Note, "Generation: ") {
//		return nil
//	}
//	spl := strings.Split(n.Note, " ")
//	if len(spl) != 2 {
//		return nil
//	}
//	out, err := strconv.Atoi(spl[1])
//	if err != nil {
//		return nil
//	}
//	return &out
//}

// TODO: add all of these to autogenned
type NutrientMeasurement struct {
	Nutrient Nutrient `bson:"nutrient" json:"nutrient"` // Nutrient name
	Amount   float64  `bson:"amount" json:"amount"`     // Amount per 1L agar
	Unit     string   `bson:"unit" json:"unit"`         // mL, tsp, tbsp, drop, pinch, cup, etc
}

type Nutrient string

// TODO: add all of these to autogenned
var nutrients = []Nutrient{LME, Potato, BRF}
var (
	LME    Nutrient = "LME"
	Potato Nutrient = "potato flakes"
	BRF    Nutrient = "Brown rice flour"
)

// TODO: add all of these to autogenned
type SugarMeasurement struct {
	Type   Sugar   `bson:"type" json:"type"`     // Sugar Username
	Amount float64 `bson:"amount" json:"amount"` // Amount per 1L agar
	Unit   string  `bson:"unit" json:"unit"`
}

type Sugar string

var sugars = []Sugar{Dextrose, Honey, MapleSyrup}

func newSugarMeasurement(add Sugar, amount float64, unit string) SugarMeasurement {
	return SugarMeasurement{
		Type:   add,
		Amount: amount,
		Unit:   unit,
	}
}

// TODO: add all of these to autogenned
var (
	Dextrose   Sugar = "dextrose" // This is corn syrup
	Honey      Sugar = "honey"
	MapleSyrup Sugar = "maple syrup"
)

type Grain string

var grains = []Grain{Rye, Wheat, Oats, Millett, Popcorn, BirdSeed}

func (g Grain) Validate() error {
	if !slices.Contains(grains, g) {
		return errors.New("invalid grain")
	}
	return nil
}

// TODO: add all of these to autogenned
var (
	Rye      Grain = "rye"
	Wheat    Grain = "wheat"
	Millett  Grain = "millett"
	Oats     Grain = "oats"
	Popcorn  Grain = "popcorn"
	BirdSeed Grain = "birdseed"
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

type AdditiveMeasurement struct {
	Additive Additive `bson:"additive" json:"additive"` // Nutrient name
	Amount   float64  `bson:"amount" json:"amount"`     // Amount per 1L agar
	Unit     string   `bson:"unit" json:"unit"`         // mL, tsp, tbsp, drop, pinch, cup, etc
}

func newAdditiveMeasurement(add Additive, amount float64, unit string) AdditiveMeasurement {
	return AdditiveMeasurement{
		Additive: add,
		Amount:   amount,
		Unit:     unit,
	}
}

type Colorant string

var colorants = []Colorant{clearColor, black, blue, yellow, orange, red}

// TODO: add all of these to autogenned
var (
	clearColor Colorant = "Clear"
	black      Colorant = "Black"
	blue       Colorant = "Blue"
	green      Colorant = "Green"
	yellow     Colorant = "Yellow"
	orange     Colorant = "Orange"
	red        Colorant = "Red" // MOST REDS ARE FUNGICIDAL
)

var colors = map[string]Colorant{
	string(clearColor): clearColor,
	string(black):      black,
	string(blue):       blue,
	string(green):      green,
	string(yellow):     yellow,
	string(orange):     orange,
}

func ValidColor(c Colorant) bool {
	_, ok := colors[string(c)]
	return ok
}

type Additive string // TODO: ACCOUNT FOR THIS EVERYWHERE!
var additives = []Additive{Vermiculite, Perlite, Gypsum, YeastNutrient, CoffeeGrounds}

// TODO: add all of these to autogenned
var (
	Vermiculite   Additive = "vermiculite"
	Perlite       Additive = "perlite"
	Gypsum        Additive = "gypsum"
	YeastNutrient Additive = "yeast nutrient"
	CoffeeGrounds Additive = "Coffee Grounds"
)

type Antibiotic string

// TODO: add all of these to autogenned
var antibiotics = []Antibiotic{HydrogenPeroxide, Doxycycline}

var (
	HydrogenPeroxide Antibiotic = "HydrogenPeroxide"
	Doxycycline      Antibiotic = "Doxycycline"
)

// TODO: use
var antibioticDosages = map[Antibiotic]string{ // TODO: USE THIS!
	Doxycycline:      "unknown as of right now", // TODO: figure out measurements
	HydrogenPeroxide: "unknown as of right now", // TODO: figure out measurements
}

type Generation int

func (gen *Generation) validate() error {
	if gen != nil {
		if *gen < 1 {
			return errors.New("invalid generation. Cannot be less than 1")
		}
	}
	return nil
}

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
		out = append(out, bson.E{Key: "$set", Value: upd.sets})
	}
	if upd.pushes != nil && len(upd.pushes) != 0 {
		out = append(out, bson.E{Key: "$push", Value: upd.pushes})
	}
	if upd.pulls != nil && len(upd.pulls) != 0 {
		out = append(out, bson.E{Key: "$pull", Value: upd.pulls})
	}
	if upd.unsets != nil && len(upd.unsets) != 0 {
		out = append(out, bson.E{Key: "$unset", Value: upd.unsets})
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
		x := utils.Set[string]{}
		x.Add(future...)
		if len(x) != len(future) {
			// TODO: validate no repeats in future
			upd.err = errors.New("aliases cannot contain replica values")
			return upd
		}
		return upd.Set("aliases", future)
	}
	return upd
}

func (upd *Mods) updateDisposedIfNeeded(future, existing Disposable) *Mods {
	exist, fut := existing.DisposalInfo(), future.DisposalInfo()
	if exist == nil && fut != nil {
		upd.Set("disposed", *fut)
	}
	return upd
}

func (upd *Mods) updatePcRunIfNeeded(next, current pcRunOptional) *Mods {
	a, b := current.pcRunId(), next.pcRunId()
	if a == nil && b != nil {
		upd.Set("pcRun", b)
	}
	return upd
}

func (upd *Mods) updateKnownFruitableIfNeeded(future, existing hasKnownFruitableField) *Mods {
	ex := existing.knownToBeFruitable()
	return updatePointerIfNeeded(upd, "knownFruitable", future.knownToBeFruitable(), ex)
}
func (upd *Mods) updateWetnessIfNeeded(future, existing *int) *Mods {
	if existing == nil && future != nil {
		if *future < 0 || *future > 10 { // TODO: validate
			upd.err = errors.New("wetness must be 1-10") // TODO: validate range ok
			return upd
		}
		return upd.Set("wetness", *future)
	}
	return upd
}
func (upd *Mods) updateBurstGrainsIfNeeded(future, existing *int) *Mods {
	if existing == nil && future != nil {
		if *future < 0 || *future > 10 { // TODO: validate
			upd.err = errors.New("burst grains must be 0-10") // TODO: validate range ok
			return upd
		}
		return upd.Set("burstGrains", *future)
	}
	return upd
}
func (upd *Mods) updateCondensationCoverageIfNeeded(future, existing hasCondensCov) *Mods {
	exist, fut := existing.condensationCoverage(), future.condensationCoverage()
	if exist != nil {
		if fut == nil || *fut != *exist {
			upd.err = errors.New("condensationConverage mismatch")
		}
		return upd
	} else {
		if exist == fut { // Still empty
			return upd
		}
	}
	return updatePointerIfNeeded(upd, "condensationCoverageAtSealTime", fut, exist)
}
func (upd *Mods) updatePourCoverageIfNeeded(future, existing hasPourCoverage) *Mods {
	exist, fut := existing.pourCoverage(), future.pourCoverage()
	if exist != nil {
		if fut == nil || *fut != *exist {
			upd.err = errors.New("pourCoverage mismatch")
		}
		return upd
	} else {
		if exist == fut { // Still empty
			return upd
		}
	}
	return updatePointerIfNeeded(upd, "pourCoverage", fut, exist)
}
func (upd *Mods) updateWetAtCooledTimeIfNeeded(future, existing hasWact) *Mods {
	// Only update if existing value is nil and new value is not
	exist, fut := existing.wetAtCool(), future.wetAtCool()
	if exist == nil && fut != exist {
		return updatePointerIfNeeded(upd, "wetAtCooledTime", fut, exist)
	}
	return upd
}
func (upd *Mods) updateAgarOnOutsideAtPourTimeIfNeeded(future, existing hasAgarOutside) *Mods {
	// Only update if existing value is nil and new value is not
	exist, fut := existing.agarOutside(), future.agarOutside()
	if exist == nil && fut != exist {
		return updatePointerIfNeeded(upd, "agarOnOutsideAtPourTime", fut, exist)
	}
	return upd
}

func (upd *Mods) updateConfirmedCleanIfNeeded(future, existing *bool) *Mods {
	return updatePointerIfNeeded(upd, "confirmedClean", future, existing)
}

func (upd *Mods) updateProjectCompletedIfNeeded(future, existing *unix.Time) *Mods {
	return updatePointerIfNeeded(upd, "completed", future, existing)
}

func (upd *Mods) updateProjectPrivateIfNeeded(future, existing bool) *Mods {
	if future == existing {
		return upd
	}
	return updateValueIfNeeded(upd, "private", future, existing)
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
	return upd.Set("lastUpdated", unix.TimeFor(time.Now()))
}

func (upd *Mods) updatePermsIfNeeded(next, current ACL) *Mods {
	if current.Equivalent(next) {
		return upd
	}
	return upd.Set("acl", next)
}

func (upd *Mods) updateDefaultAclIfNeeded(nextReq PermsOnRequest, current ACL) *Mods {
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

func PrettyPrintJson(prefix string, toPrint any) {
	outBs, err := json.MarshalIndent(toPrint, "", " ")
	if err != nil {
		println(prefix, "failed to get bytes for pretty print")
		return
	}
	println(prefix, string(outBs))
}

func (upd *Mods) updateNotesIfNeeded(updatedIn NoteMods, existingIn HasNotesField) *Mods { // TODO: make sure this always works the way we want it to!!! // TODO: lower-down notes?
	if upd.err != nil {
		return upd
	}
	println("updating notes if needed")
	existing := existingIn.GetNotes()
	updated := updatedIn.NoteChanges()
	if len(updated.Existing) != len(existing) {
		upd.err = errors.Join(errors.New("length of existing notes must match"), upd.err)
		exNotesBs, err := json.Marshal(existing) // TODO: delete later
		if err != nil {
			println("failed to get bytes for existing notes")
			return upd
		}
		upNotesBs, err := json.Marshal(updated.Existing)
		if err != nil {
			println("failed to get bytes for updated notes")
			return upd
		}
		toOutput := struct {
			Existing string
			Updated  string
		}{
			Existing: string(exNotesBs),
			Updated:  string(upNotesBs),
		}
		PrettyPrintJson("noteUpdDiscrepancy", toOutput) // TODO: delete later
		return upd
	}
	if !notesWereModified(existing, updated) {
		println("notes not found to be modified...") // TODO: del

		return upd
	}
	finalNotes := []Note{}
	for _, final := range updated.Existing {
		if !final.Disabled {
			finalNotes = append(finalNotes, final.Data)
		}
	}
	finalNotes = append(finalNotes, sliceutils.Map(updated.New, func(nt Data[Note]) Note { return nt.Data })...)
	// Set notes
	PrettyPrintJson("finalNotes", finalNotes) // TODO: delete later
	return upd.Set("notes", finalNotes)
}

func (upd *Mods) updateTimeIfNoLongerNil(fieldName string, updated *int, existing *int) *Mods { // TODO: make sure this works as anticipated
	if updated == nil {
		return upd
	}
	if *updated < 0 {
		upd.err = errors.New("time cannot be negative")
		return upd
	}
	return updateTimeIfWasNil(upd, fieldName, updated, existing)
}

func (upd *Mods) updatePicsIfNeeded(updatedEntries SplitEntries[picWithNotesForm, PicWithNotes], existing []PicWithNotes) *Mods { // TODO: make sure this works as anticipated
	return upd.updatePwnIfNeeded("pics", updatedEntries, existing)
}

// Flatten takes a 2D slice and returns a flattened 1D slice
func Flatten[T any](lists [][]T) []T {
	var res []T
	for _, list := range lists {
		res = append(res, list...) // The ... unpacks the inner slice
	}
	return res
}

// TODO: ADD TO ALL PLACES!
func (upd *Mods) updateMostRecentImageIfNeeded(existing *PicWithNotes, updatedPicsGroups ...[]PicWithNotes) *Mods { // TODO: make sure this works as anticipated!
	// TODO: CONSIDER USING getItemLatestImage from common.go
	updatedPics := Flatten(updatedPicsGroups)
	if len(updatedPics) == 0 {
		if existing != nil {
			// if pic already exists, remove it
			return upd.withMostRecentImage(nil)
		}
		return upd
	}

	getLatestPic := func(updated []PicWithNotes) PicWithNotes {
		var latestPic PicWithNotes = updated[0]
		for i := 1; i < len(updated); i++ {
			candidate := updated[i]
			if latestPic.Time < candidate.Time {
				latestPic = candidate
			}
		}
		return latestPic
	}
	latestPic := getLatestPic(updatedPics)
	if existing != nil {
		if existing.EqualTo(latestPic) {
			return upd
		}
	}
	return upd.withMostRecentImage(&latestPic)
}

func (upd *Mods) updateFlushesIfNeeded(updatedEntries SplitEntries[picWithNotesForm, PicWithNotes], existing []PicWithNotes) *Mods { // TODO: make sure this works as anticipated
	return upd.updatePwnIfNeeded("flushes", updatedEntries, existing)
}

func (upd *Mods) updatePwnIfNeeded(fieldName string, updatedEntries SplitEntries[picWithNotesForm, PicWithNotes], existing []PicWithNotes) *Mods { // TODO: make sure this works as anticipated
	if upd.err != nil {
		return upd
	}
	if len(updatedEntries.Existing) != len(existing) {
		// TODO: print this out????
		upd.err = errors.Join(errors.New("length of existing "+fieldName+" must match"), upd.err)
		return upd
	}
	if !picsWereModified(existing, updatedEntries) {
		return upd
	}
	// TODO: THIS IS NOT WORKING???? I think this is working as of 7/11/26
	finalPics := make([]PicWithNotes, 0, len(existing)+len(updatedEntries.New))
	for _, final := range updatedEntries.Existing {
		if !final.Disabled {
			finalPics = append(finalPics, final.Data.convert())
		}
	}
	finalPics = append(finalPics, updatedEntries.New...)
	// Set field
	return upd.Set(fieldName, finalPics)

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
		//println("contams were NOT modified!------------------------") // TODO: del

		return upd
	}
	//println("CONTAMS WERE MODIFIED, SHOULD BE CHANGING!-------------------") // TODO: del
	finalEntries := make([]Contamination, 0, len(existing)+len(updatedEntries.New))
	for _, final := range updatedEntries.Existing {
		if !final.Disabled {
			finalEntries = append(finalEntries, final.Data.convert())
		}
	}
	finalEntries = append(finalEntries, updatedEntries.New...)
	// Set field
	return upd.Set("contamination", finalEntries)
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
func (upd *Mods) withLastUpdated(lastUpdatedTime unix.Time) *Mods {
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
func (upd *Mods) withPerms(acl ACL) *Mods {
	return upd.Set("acl", acl)
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
	return setPointerIfNonNil(upd, "subspecies", subsp)
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
func updateTimeIfWasNil[T comparable](upd *Mods, fieldName string, future, existing *T) *Mods {
	if upd.err != nil {
		return upd
	}
	if existing != nil {
		if *existing != *future {
			upd.err = errors.New(fieldName + "was already set to a different value. Cannot change value once set")
		}
		return upd
	}
	if future == nil {
		return upd
	}
	return upd.Set(fieldName, *future)
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
	switch strings.ToLower(opt) {
	case "additives", "additive":
		writeAsJson(w, additives)
		return
	case "antibiotics", "antibiotic":
		writeAsJson(w, antibiotics)
		return
	case strings.ToLower("bagFilterSizes"):
		writeAsJson(w, bagFilterSizes)
		return
	case "colors", "color",
		"colorants", "colorant":
		writeAsJson(w, colorants)
		return
	case "grains", "grain":
		writeAsJson(w, grains)
		return
	case "liquids", "liquid",
		"fluids", "fluid":
		writeAsJson(w, fluids)
		return
	case "nutrients", "nutrient":
		writeAsJson(w, nutrients)
		return
	case strings.ToLower("sporePrintColors"), strings.ToLower("sporePrintColor"):
		writeAsJson(w, sporePrintColors)
		return
	case strings.ToLower("sporePrintDensities"), strings.ToLower("sporePrintDensity"):
		writeAsJson(w, sporePrintDensities)
		return
	case "sugars", "sugar":
		writeAsJson(w, sugars)
		return
	case strings.ToLower("transferReasons"), strings.ToLower("transferReason"):
		writeAsJson(w, transferReasons)
		return
	case "woods", "wood":
		writeAsJson(w, woods)
		return
	default:
		http.Error(w, fmt.Sprintf(`invalid option provided: "%s" is not one of [bagFilterSizes, color,liquid,nutrient,sugar,grain,additive,transferReason]`, opt), http.StatusBadRequest)
		return
	}
}
