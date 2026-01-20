package rfid

import (
	"errors"
	"fmt"
	sliceutils "github.com/reeceappling/goUtils/v2/utils/slices"
	"math"
	"slices"
	"time"
)

type AdditivesField struct {
	Additives []additiveMeasurement `bson:"additives,omitempty" json:"additives,omitempty"`
}

func (field AdditivesField) Validate() error {
	for i, item := range field.Additives {
		if !slices.Contains(additives, item.Additive) {
			return fmt.Errorf(`invalid additive at position %d: %s`, i, item.Additive)
		}
	}
	return nil
}

type AliasesField struct {
	Aliases []string `bson:"aliases,omitempty" json:"aliases,omitempty"`
}

func (field AliasesField) updateIfNeeded(existing AliasesField, upd *Mods) *Mods {
	if slices.Compare(existing.Aliases, field.Aliases) != 0 {
		upd.Set("aliases", field.Aliases)
	}
	return upd
}

type AlternateCollectionIdField struct {
	Id AlternateCollectionId `bson:"_id" json:"_id"`
}

func (field AlternateCollectionIdField) DbId() AlternateCollectionId {
	return field.Id
}
func (field AlternateCollectionIdField) IdValue() any { return field.DbId() } // TODO: DEL!

type AlternateCollectionOptionalParentField struct {
	Parent *AlternateCollectionId `bson:"parent,omitempty" json:"parent,omitempty"`
}

type WetnessField struct {
	Wetness *int `bson:"wetness,omitempty" json:"wetness,omitempty"` // nil==unknown, 0== very dry, 10==veryWet, 5==perfect fieldCapacity, normal range 4-6
}

func (field WetnessField) Validate() error {
	if field.Wetness == nil {
		return nil
	}
	if *field.Wetness < 0 || *field.Wetness > 10 {
		return errors.New("Invalid wetness, must either be nonexistent or 0-10")
	}
	return nil
}

type AntibioticsField struct {
	Antibiotics []antibiotic `bson:"antibiotics,omitempty" json:"antibiotics,omitempty"`
}

func (field AntibioticsField) Validate() error {
	for i, item := range field.Antibiotics {
		if !slices.Contains(antibiotics, item) {
			return fmt.Errorf(`invalid antibiotic at position %d: %s`, i, item)
		}
	}
	return nil
}

//type BinaryOptionalParentField struct {
//	Parent *BinaryCollectionId `bson:"parent,omitempty" json:"parent,omitempty"`
//}

type LiquidsField struct {
	Liquids []liquid `bson:"liquids" json:"liquids"`
}

func (field LiquidsField) Validate() error {
	totalPct := 0.0 // TODO: make percentage int rather than float?
	for i, item := range field.Liquids {
		if !slices.Contains(fluids, item.Name) {
			return fmt.Errorf(`invalid antibiotic at position %d: %s`, i, item.Name)
		}
		totalPct += item.Pct
	}
	if math.Abs(totalPct-100.0) > 1.0 { // TODO: ok?
		return fmt.Errorf(`total liquids percentage too high or low. Expected 100.0, got %f`, totalPct)
	}
	return nil
}

type MainCollectionIdField struct {
	Id MainCollectionId `bson:"_id" json:"_id"` // TODO; INLINE ALL!
}

func (field MainCollectionIdField) DbId() BinaryCollectionId {
	return field.Id.ToBinaryCollectionId()
}
func (field MainCollectionIdField) IdValue() any { return string(field.DbId()) }

type MainCollectionOptionalParentField struct {
	Parent *MainCollectionId `bson:"parent,omitempty" json:"parent,omitempty"`
}
type MainCollectionParentField struct {
	Parent MainCollectionId `bson:"parent,omitempty" json:"parent,omitempty"`
}

type NameIdField struct {
	Name string `bson:"_id" json:"_id"`
}

func (field NameIdField) IdValue() any {
	return field.Name
} // TODO: DELETE?
func (field NameIdField) DbId() string {
	return field.Name
}

type NameField struct {
	Name string `bson:"name" json:"name"`
}

func (field NameField) updateIfNeeded(existing NameField, upd *Mods) *Mods {
	if field.Name != existing.Name {
		upd.Set("name", field.Name)
	}
	return upd
}

type NutrientsField struct {
	Nutrients []nutrientMeasurement `bson:"nutrients,omitempty" json:"nutrients,omitempty"`
}

func (field NutrientsField) Validate() error {
	// TODO: anything else in here?
	for i, nute := range field.Nutrients {
		if !slices.Contains(nutrients, nute.Nutrient) {
			return fmt.Errorf(`invalid nutrient at position %d: %s`, i, nute.Nutrient)
		}
	}
	return nil
}

type ParentTypeField struct {
	ParentType *string `bson:"parentType,omitempty" json:"parentType,omitempty"`
}

type CreationDateField struct {
	CreationDate unixTime `bson:"creationDate" json:"creationDate"`
}

type DisposedField struct {
	Disposed *unixTime `bson:"disposed,omitempty" json:"disposed,omitempty"`
}

type GenerationsFields struct {
	GenSporeField        `bson:"inline"`
	GenSinceFruitOrSpore *Generation `bson:"genFruitOrSpore,omitempty" json:"genFruitOrSpore,omitempty"`
}

func GenerationsFieldFor(gen *Generation) GenerationsFields {
	return GenerationsFields{
		GenSporeField:        GenSporeField{gen},
		GenSinceFruitOrSpore: gen,
	}
}

type GenSporeField struct { // TODO: only used on fruit and embedded in GenerationsFields?
	GenSinceSpore *Generation `bson:"genSpore,omitempty" json:"genSpore,omitempty"`
}

type LastUpdatedField struct {
	LastUpdated unixTime `bson:"lastUpdated" json:"lastUpdated"`
}

func LastUpdatedFieldForNow() LastUpdatedField {
	return LastUpdatedField{unixTimeForNow()}
}

type NotesField struct {
	Notes []Note `bson:"notes,omitempty" json:"notes,omitempty"`
}

func (field NotesField) withAllTimesSetTo(t time.Time) NotesField {
	return NotesField{sliceutils.Map(field.Notes, func(n Note) Note {
		return Note{
			Time: unixTimeFor(t),
			Note: n.Note,
		}
	})}
}

type ConfirmedCleanField struct {
	ConfirmedClean *bool `bson:"confirmedClean,omitempty" json:"confirmedClean,omitempty"`
}

type PressureCookedTouchingWaterOptionalField struct {
	PressureCookedTouchingWater *bool `bson:"pressureCookedTouchingWater,omitempty" json:"pressureCookedTouchingWater,omitempty"`
}

type KnownFruitableField struct { // TODO: what about fields that already know? Like on fruits?
	KnownFruitable *bool `bson:"knownFruitable,omitempty" json:"knownFruitable,omitempty"` // set on transfer in, or once fruited
}

func (existing KnownFruitableField) UpdateKnownFruitableIfNeeded(upd *Mods, future *bool) {
	updatePointerIfNeededNew(upd, "knownFruitable", future, existing.KnownFruitable)
}

type StandardField struct {
	Standard bool `bson:"standard" json:"standard"` // If this is a standard recipe
}

func (field StandardField) updateIfDifferent(existing StandardField, upd *Mods) *Mods { // TODO: remove???
	if field.Standard != existing.Standard {
		upd.Set("standard", field.Standard)
	}
	return upd
}

type SugarsField struct {
	Sugars []sugarMeasurement `bson:"sugars,omitempty" json:"sugars,omitempty"`
}

func (field SugarsField) Validate() error {
	// TODO: anything else in here?
	for i, item := range field.Sugars {
		if !slices.Contains(sugars, item.Type) {
			return fmt.Errorf(`invalid sugar at position %d: %s`, i, item.Type)
		}
	}
	return nil
}

type WriteTagToField struct {
	WriteTagTo *string `json:"writeTagTo,omitempty"`
}
