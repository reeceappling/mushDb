package rfid

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/reeceappling/goUtils/v2/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"reflect"
)

const (
	SporeSwabSourceType     = "sporeSwab"
	sporeSwabCollectionName = "sporeSwabs"
	sporeSwabIdPrefix       = "ss"
)

type SporeSwab struct { // TODO: FIX EVERYTHING IN THIS FILE BELOW THIS POINT!!!!
	AlternateCollectionIdField
	// Parent is always either sporePrint, or purchased
	AlternateCollectionOptionalParentField // TODO: handle now a pointer       // TODO: likely won't exist for pre-existing
	CreationDateField                      // Swab or receive date
	SpeciesField
	SubspeciesOptionalField
	SaleField // TODO: was sales! singular now
	DisposedField
	TransfersOutField
	NotesField
	LastUpdatedField
	PermsField
}

func (sw SporeSwab) projects() []projectName {
	return sw.Perms.Projects.Ids
}

func (sw SporeSwab) setTransferParent(ctx mongo.SessionContext, xfer Transfer) error {
	// TODO: can this even occur?
	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
	if err != nil {
		return err
	}
	res, err := ctx.Client().Database(dbName).Collection(sw.CollectionName()).UpdateByID(ctx, sw.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer
	}
	return nil
}

func (sw SporeSwab) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
	// TODO: can this happen????? should always be from a fruit right?
	// This is a special case because it will always be 0-gen
	parentInfo, err := from.GeneticInfoAsParent()
	if err != nil {
		return err
	}
	if parentInfo.Species == nil {
		return errors.New("parent must have a species")
	}
	if from.SourceType() != SporePrintSourceType {
		errors.New("only fruits are supported as a transfer source type into sporeSwabs")
	}
	upd, err := xfer. // TODO: fix this whole thing
				PicsModsForChild(). // TODO: fix
				withInnoc(xfer).    // TODO: fix
				withParent(utils.Pointer(from.DbId())).
				withSpecies(parentInfo.Species).
				withSubspecies(parentInfo.SubSpecies).
				updateLastUpdatedIfNeeded().
				Finalized()
	res, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(sw.CollectionName()).UpdateByID(ctx, sw.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer
	}
	return nil
}

func (sw SporeSwab) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := sw
	err := decodeItem(&out, encoded)
	return out, err
}

func (sw SporeSwab) GeneticInfoAsParent() (GeneticParentInfo, error) {
	return GeneticParentInfo{
		SpeciesOptionalField:    sw.SpeciesField.AsOptional(),
		SubspeciesOptionalField: sw.SubspeciesOptionalField,
		GenerationsFields:       GenerationsFieldFor(utils.Pointer(Generation(0))),
	}, nil
}

func (sw SporeSwab) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return utils.Pointer(Generation(0)), utils.Pointer(Generation(0))
}

func (sw SporeSwab) SourceType() string {
	return SporeSwabSourceType
}

func (sw SporeSwab) EntryTypeField() *string {
	return nil
}

func (sw SporeSwab) altId() AlternateCollectionId {
	return AlternateCollectionId(sw.Id)
}

func (sw SporeSwab) id() []byte {
	return sw.Id[:]
}

//func (sp SporeSwab) knownFruitable() bool {
//	return false
//}

func (sw SporeSwab) prefix() string {
	return sporeSwabIdPrefix
}

func (sw SporeSwab) CollectionName() string {
	return sporeSwabCollectionName
}

func initializeSporeSwabs(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(sporeSwabCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		newSimpleIndex("parent", "parent", false, false, false),
		newSimpleIndex("creationDate", "creationDate", true, false, false), // TODO: INDEX CREATION DATES EVERYWHERE!
		newSimpleIndex("species", "species", false, false, false),
		newSimpleIndex("subSpecies", "subSpecies", false, true, false),
		projectsIndexModel,
		saleIndexModel,
		//Notes (no index unless tags)
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	existingEntry := SporeSwab{}
	testItem := SporeSwab{
		AlternateCollectionIdField:             AlternateCollectionIdField{exAltId},
		AlternateCollectionOptionalParentField: AlternateCollectionOptionalParentField{&exAltId},
		CreationDateField:                      exampleTime.asCreationDate(),
		SpeciesField:                           SpeciesField{testEntryStringId},
		SubspeciesOptionalField:                SubspeciesOptionalField{&testEntryStringId},
		SaleField:                              SaleField{&exAltId},
		DisposedField:                          DisposedField{&exampleTime},
		NotesField:                             NotesField{exampleNotes()},
		LastUpdatedField:                       LastUpdatedField{exampleTime},
	}
	err = coll.FindOne(ctx, bson.D{{"_id", exAltId}}).Decode(&existingEntry)
	if err == nil {
		if reflect.DeepEqual(existingEntry, testItem) {
			return nil
		}
	}
	return testExistingEntry(ctx, coll, exAltId, testItem, existingEntry)
}

type createSporeSwabRequest struct {
	SporePrintId AlternateCollectionId `bson:"fruitId" json:"fruitId"`
	NotesField
}

func createSporeSwabHandler(w http.ResponseWriter, r *http.Request) { // TODO: NO PICS
	data := createSporeSwabRequest{}
	id := newAlternateCollectionId()
	defer r.Body.Close()
	// TODO: no pictures, so use other way
	// Process text (or object)
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	// PARSE INTO CORRECT DATA FORMAT
	err = json.Unmarshal(bs, &data)
	if err != nil {
		http.Error(w, "failed to unmarshal data from form: "+err.Error(), http.StatusBadRequest)
		return
	}

	_, txErr := doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		db := ctx.Client().Database(dbName)
		parent := SporePrint{}
		err = db.Collection(sporePrintCollectionName).FindOne(ctx, bson.D{{"_id", id}}).Decode(&parent)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}

		now := unixTimeForNow()
		spid := id
		toInsert := SporeSwab{
			AlternateCollectionIdField:             AlternateCollectionIdField{spid},
			AlternateCollectionOptionalParentField: AlternateCollectionOptionalParentField{&parent.Id},
			CreationDateField:                      now.asCreationDate(),
			SpeciesField:                           parent.SpeciesField,
			SubspeciesOptionalField:                parent.SubspeciesOptionalField,
			NotesField:                             NotesField{data.Notes},
			LastUpdatedField:                       LastUpdatedField{now},
			// Do not check permissions, just pass parent perms to child
			PermsField: PermsField{parent.Perms},
		}
		_, err = db.Collection(sporeSwabCollectionName).InsertOne(ctx, toInsert)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		// Update fruit with new print id
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		bsOut, err := json.Marshal(toInsert)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		return w.Write(bsOut)
	})
	if txErr != nil {
		handleWriteErr(txErr, w)
	}
}

type updateSporeSwabRequest struct { // TODO: fix everything below this
	SaleField
	DisposedField
	Notes AllEntries[Note]
	PermsField
}

func (upr updateSporeSwabRequest) reform() resolvedUpdateSporeSwabRequest {
	return resolvedUpdateSporeSwabRequest{
		SaleField:     upr.SaleField,
		DisposedField: upr.DisposedField,
		Notes:         upr.Notes,
		PermsField:    PermsField{upr.Perms},
	}
}

type resolvedUpdateSporeSwabRequest struct {
	SaleField
	DisposedField
	Notes AllEntries[Note]
	PermsField
}

func updateSporeSwabHandler(w http.ResponseWriter, r *http.Request) {
	data := updateSporeSwabRequest{}
	b58Id := Base58Str(r.PathValue("id")) // TODO: ensure ok
	id, err := b58Id.toMainCollectionId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	out := data.reform()

	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(sporeSwabCollectionName)
		// go get current sporePrint
		existing := SporeSwab{}
		err = coll.FindOne(ctx, bson.D{{"_id", id}}).Decode(&existing)
		if err != nil {
			return DbTxnStdErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		}
		if err = minimalPermsBetween(existing.Perms, data.Perms).ValidateUserCanWrite(ctx); err != nil {
			return DbTxnStdErr(w, "failed to validate overlapping permissions: "+err.Error(), http.StatusBadRequest)
		}
		upd, err := NewMods().
			updateSaleIfNeeded(out.Sale, existing.Sale).
			updateDisposedIfNeeded(data.Disposed, existing.Disposed).
			updateNotesIfNeeded(data.Notes, existing.Notes).
			updatePermsIfNeeded(data.Perms, existing.Perms). // TODO: ok?
			updateLastUpdatedIfNeeded().
			Finalized()
		if err != nil {
			return DbTxnStdErr(w, "error creating txn:"+err.Error(), http.StatusInternalServerError)
		}
		if len(upd) == 0 {
			return DbTxnStdErr(w, "no changes made", http.StatusBadRequest)
		}

		// write updates to db
		bsonId := bson.D{{"_id", id}}
		err = coll.FindOneAndUpdate(ctx, bsonId, upd).Err()
		if err != nil {
			return DbTxnStdErr(w, "failed to write update to db: "+err.Error(), http.StatusInternalServerError)
		}
		err = coll.FindOne(ctx, bsonId).Decode(&existing)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		bsOut, err := json.Marshal(existing)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		return w.Write(bsOut)
	})
	if err != nil {
		HandleHttpWriteError(err)
	}
}

type importSporeSwabRequest struct {
	CreationDateField
	SpeciesField
	SubspeciesOptionalField
	NotesField
	PermsField
}

func importSporeSwabHandler(w http.ResponseWriter, r *http.Request) { // TODO: NO IMAGES
	data := importSporeSwabRequest{}
	id := newAlternateCollectionId()
	defer r.Body.Close()

	// Process text (or object)
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "unable to read data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	// PARSE INTO CORRECT DATA FORMAT
	err = json.Unmarshal(bs, &data)
	if err != nil {
		http.Error(w, "unable to unmarshal json form data: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err = data.Perms.ValidateUserCanWrite(r.Context()); err != nil {
		http.Error(w, "user cannot write with these perms: "+err.Error(), http.StatusBadRequest)
		return
	}

	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		finalPerms := data.Perms
		if data.Perms != nil {
			spec, subsp, err := getSpeciesAndSubspecies(ctx, data.Species, data.SubSpecies)
			if err != nil {
				return DbTxnStdErr(w, "failed to get species or subspecies: "+err.Error(), http.StatusInternalServerError) // TODO: ok?
			}
			finalPerms = minimalPermsBetween(spec, subsp)
			// TODO: add user perms if provided, as well as make user author?
			if !finalPerms.Valid() {
				// TODO: invalid species/subspecies perm crossover. DO THIS ELSEwHERE
				return DbTxnStdErr(w, "invalid species/subspecies perm crossover: "+err.Error(), http.StatusInternalServerError) // TODO: ok?
			}
		}

		toInsert := SporeSwab{
			AlternateCollectionIdField: AlternateCollectionIdField{id},
			CreationDateField:          data.CreationDateField,
			SpeciesField:               data.SpeciesField,
			SubspeciesOptionalField:    data.SubspeciesOptionalField,
			NotesField:                 data.NotesField,
			LastUpdatedField:           LastUpdatedFieldForNow(),
			PermsField:                 PermsField{finalPerms},
		}
		coll := ctx.Client().Database(dbName).Collection(sporeSwabCollectionName)
		_, err = coll.InsertOne(ctx, toInsert)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		bsOut, err := json.Marshal(toInsert)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		return w.Write(bsOut)
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}
