package rfid

import (
	"fmt"
	"github.com/reeceappling/goUtils/v2/utils"
	"go.mongodb.org/mongo-driver/bson"
	"strconv"
	"strings"
	"time"
)

const dbName = "mushDb" // TODO: changeMe? GRAB FROM ENV VARS????
const mainCollectionName = "mainCollection"

type imageLocation string
type unixTime int64

var (
	_ subdocWithImage = PicWithNotes{}
	_ subdocWithImage = Contamination{}
	_ subdocWithImage = PicWithNotes{}
)

type subdocWithImage interface {
	getPicWithNotes() *PicWithNotes
}

func getLatestExistingImage(possibleSubdocs ...subdocWithImage) *PicWithNotes { // TODO: use
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

type AgarPour struct {
	Date       unixTime              `bson:"date" json:"date"`
	Wateriness string                `bson:"wateriness,omitempty" json:"wateriness,omitempty"`
	PourTempF  int                   `bson:"pourTempF" json:"pourTempF"`
	Recipe     alternateCollectionId `bson:"recipe" json:"recipe"` // This is from batch
	Batch      alternateCollectionId `bson:"batch" json:"batch"`
	Notes      []Note                `bson:"notes,omitempty" json:"notes,omitempty"`
}

type PicWithNotes struct {
	Time     unixTime      `bson:"time" json:"time"`
	Location imageLocation `bson:"location" json:"location"`
	Notes    []Note        `bson:"notes,omitempty" json:"notes,omitempty"`
}

type PicWithNotesLessLocation struct {
	Time  unixTime `bson:"time" json:"time"`
	Notes []Note   `bson:"notes,omitempty" json:"notes,omitempty"`
}

func (p PicWithNotesLessLocation) asPicWithNotes(location *string) PicWithNotes {
	return PicWithNotes{
		Time:     p.Time,
		Location: imageLocation(utils.Default(location, "")),
		Notes:    p.Notes,
	}
}

func (p PicWithNotes) getPicWithNotes() *PicWithNotes {
	return &p
}

type Contamination struct {
	Time      unixTime       `bson:"time" json:"time"` // TODO: NEW! HANDLE EVERYWHERE!
	Confirmed bool           `bson:"confirmed" json:"confirmed"`
	Bacteria  bool           `bson:"bacteria" json:"bacteria"`
	Mold      bool           `bson:"mold" json:"mold"`
	Location  *imageLocation `bson:"location,omitempty" json:"location,omitempty"` // TODO: NEW! HANDLE EVERYWHERE!
	Notes     []Note         `bson:"notes,omitempty" json:"notes,omitempty"`       // TODO: NEW! HANDLE EVERYWHERE!
}

type ContaminationLessLocation struct {
	Time      unixTime `bson:"time" json:"time"` // TODO: NEW! HANDLE EVERYWHERE!
	Confirmed bool     `bson:"confirmed" json:"confirmed"`
	Bacteria  bool     `bson:"bacteria" json:"bacteria"`
	Mold      bool     `bson:"mold" json:"mold"`
	Notes     []Note   `bson:"notes,omitempty" json:"notes,omitempty"` // TODO: NEW! HANDLE EVERYWHERE!
}

func (c ContaminationLessLocation) asContamination(location *imageLocation) Contamination {
	return Contamination{
		Time:      c.Time,
		Confirmed: c.Confirmed,
		Bacteria:  c.Bacteria,
		Mold:      c.Mold,
		Location:  location,
		Notes:     c.Notes,
	}
}

func (c Contamination) getPicWithNotes() *PicWithNotes {
	if c.Location == nil {
		return nil
	}
	return &PicWithNotes{
		Time:     c.Time,
		Location: *c.Location,
		Notes:    c.Notes,
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
	// TODO: ensure pct can never go over 100
	return liquid{
		Name: f,
		Pct:  val,
	}
}

type notesGroup []Note

func (notes notesGroup) GenerationIfExists() *int { // TODO: USE THIS
	for _, note := range notes {
		if out := note.GenerationIfExists(); out != nil {
			return out
		}
	}
	return nil
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

func newGenerationNote(n int) Note { // TODO: USE THIS!!!!
	return Note{
		Time: unixTime(time.Now().UnixMilli()),
		Note: fmt.Sprintf(`Generation: %d`, n),
	}
}

func newAltSourceNote(src string) Note { // TODO: USE THIS!!!!
	return Note{
		Time: unixTime(time.Now().UnixMilli()),
		Note: fmt.Sprintf(`Sample Source: %s`, src),
	}
}

// TODO:  // TODO: THIS!
// TODO: altSource (outside, store, etc) notes // TODO: THIS?

type nutrientMeasurement struct {
	Nutrient nutrient `bson:"nutrient" json:"nutrient"` // Nutrient name
	Amount   float64  `bson:"amount" json:"amount"`     // Amount per 1L agar
	Unit     string   `bson:"unit" json:"unit"`         // mL, tsp, tbsp, drop, pinch, cup, etc
}

type nutrient string

var (
	LME    nutrient = "light malt extract" // TODO: ok?
	Potato nutrient = "potato flakes"      // TODO: ok?
)

type sugarMeasurement struct {
	Type   sugar   `bson:"type" json:"type"`     // Sugar Name
	Amount float64 `bson:"amount" json:"amount"` // Amount per 1L agar
	Unit   string  `bson:"unit" json:"unit"`
}

type sugar string

func newSugarMeasurement(add sugar, amount float64, unit string) sugarMeasurement {
	return sugarMeasurement{
		Type:   add,
		Amount: amount,
		Unit:   unit,
	}
}

var (
	Dextrose   sugar = "dextrose" // This is corn syrup
	Honey      sugar = "honey"
	MapleSyrup sugar = "maple syrup"
)

type grain string

var (
	Oats     grain = "oats"
	Popcorn  grain = "popcorn"
	BirdSeed grain = "birdseed"
)

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

var (
	clear  colorant = "clear"
	black  colorant = "black"
	blue   colorant = "blue"
	yellow colorant = "yellow"
)

var colors = map[string]colorant{
	string(clear):  clear,
	string(black):  black,
	string(blue):   blue,
	string(yellow): yellow,
}

func ValidColor(c colorant) bool {
	_, ok := colors[string(c)]
	return ok
}

type additive string // TODO: ACCOUNT FOR THIS EVERYWHERE!
var (
	Vermiculite additive = "vermiculite"
	Perlite     additive = "perlite"
	Gypsum      additive = "gypsum"
)

type antibiotic string

var (
	HydrogenPeroxide            = "HydrogenPeroxide"
	Doxycycline      antibiotic = "doxycycline"
)

var antibioticDosages = map[antibiotic]string{ // TODO: USE THIS!
	Doxycycline: "N/A", // TODO: figure out measurements
}

type Generation int // TODO: fixMe, do sinceSpores and sinceClone

func (gen *Generation) Next() *Generation {
	if gen == nil {
		return nil
	}
	nextGen := (*gen) + 1
	return &nextGen
}

type changes []bson.E

func blankChanges() changes {
	return changes{}
}

func (chg changes) addIf(doChange bool, key string, newValue interface{}) changes {
	out := changes{}
	copy(out, chg)
	if doChange {
		out = append(out, bson.E{key, newValue})
	}
	return out
}

func (chg changes) resolve() bson.D {
	return bson.D(chg)
}

func areSame[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
