package rfid

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/reeceappling/goUtils/v2/utils"
	"github.com/reeceappling/goUtils/v2/utils/slices"
	"github.com/reeceappling/mushDb/rfid/pics"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// TODO: new are this, sporeSwab, agarBottle, plugs

// Naming convention "{ParentLCID}-#"

func parseName()

const (
	lcSyringeSourceType     = "lcSyringe"
	lcSyringeCollectionName = "lcSyringes"
	lcSyringeIdPrefix       = "LCS"
)

type LcSyringeID string

func (s LcSyringeID) MarshalJSON() ([]byte, error) {
	if len(s) == RfidByteSize {
		// Purchased LC syringes
		return json.Marshal(MainCollectionId([]byte(s)[0:RfidByteSize]))
	}
	// Created LC syringes
	// TODO: validate s[9:]
	return []byte(fmt.Sprintf(`"%s-%s"`, string(MainCollectionId([]byte(s)[0:8]).base58Bytes()), s[9:])), nil
}

func (id *LcSyringeID) UnmarshalJSON(bs []byte) (err error) {
	parts := strings.Split(string(bs), "-")
	switch len(parts) {
	case 1: // Purchased
		// TODO: TEST
		// TODO: does this expect double quotes to come in with it?
		mainCollId, err := Base58Str(string(bs)).toMainCollectionId()
		if err != nil {
			return err
		}
		*id = LcSyringeID(mainCollId.dbIdStr())
	case 2: // Created
		// TODO: TEST
		// TODO: does this expect double quotes to come in with it?
		binaryParentId, err := Base58Str(parts[0]).toMainCollectionId()
		if err != nil {
			return err
		}
		// TODO: validate parts[1] is a number
		*id = LcSyringeID(string(binaryParentId[:]) + "-" + parts[1])
	default:
		return errors.New("invalid LcSyringeID")
	}
	return nil
}

func (s LcSyringeID) ID() (LcSyringeIdInMemory, error) {
	parts := strings.Split(string(s), "-")
	num, err := strconv.Atoi(parts[1])
	if err != nil {
		return LcSyringeIdInMemory{}, err
	}
	return LcSyringeIdInMemory{
		parent: AlternateCollectionId([]byte(parts[0])),
		number: num,
	}, nil
}

type LcSyringeIdInMemory struct {
	parent AlternateCollectionId
	number int
}

func (s LcSyringeIdInMemory) DbId() LcSyringeIDInDB {
	return LcSyringeIDInDB(s.parent.String() + "-" + strconv.Itoa(s.number))
}

type LcSyringe struct {
	AlternateCollectionIdField
	// Parent is always either purchased (nil), LC, or LcSyringe
	MainCollectionOptionalParentField // TODO: likely won't exist for pre-existing
	CreationDateField                 // create or receive date
	SpeciesField
	SubspeciesOptionalField
	SaleField
	// TODO: add contaminated field?
	TransfersOutField
	DisposedField
	NotesField
	LastUpdatedField
	PermsField
}

func (sw LcSyringe) projects() []projectName {
	return sw.Perms.Projects.Ids
}

func (sw LcSyringe) setTransferParent(ctx mongo.SessionContext, xfer Transfer) error {
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

func (sw LcSyringe) setTransferChild(ctx mongo.SessionContext, xfer Transfer, from geneticSource) error {
	// TODO: can this happen????? should always be from a fruit right?
	// This is a special case because it will always be 0-gen
	parentInfo, err := from.GeneticInfoAsParent()
	if err != nil {
		return err
	}
	if parentInfo.Species == nil {
		return errors.New("parent must have a species")
	}
	if from.SourceType() != FruitSourceType {
		errors.New("only fruits are supported as a transfer source type into Syringes")
	}
	upd, err := xfer.
		PicsModsForChild().
		withInnoc(xfer).
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

func (sw LcSyringe) Decode(encoded *mongo.SingleResult) (CollectionItem, error) {
	out := sw
	err := decodeItem(&out, encoded)
	return out, err
}

func (sw LcSyringe) GeneticInfoAsParent() (GeneticParentInfo, error) {
	return GeneticParentInfo{
		SpeciesOptionalField:    sw.SpeciesField.AsOptional(),
		SubspeciesOptionalField: sw.SubspeciesOptionalField,
		GenerationsFields:       GenerationsFieldFor(utils.Pointer(Generation(0))),
	}, nil
}

func (sw LcSyringe) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return utils.Pointer(Generation(0)), utils.Pointer(Generation(0))
}

func (sw LcSyringe) SourceType() string {
	return lcSyringeSourceType
}

func (sw LcSyringe) EntryTypeField() *string {
	return nil
}

func (sw LcSyringe) altId() AlternateCollectionId {
	return AlternateCollectionId(sw.Id)
}

func (sw LcSyringe) id() []byte {
	return sw.Id[:]
}

//func (sp LcSyringe) knownFruitable() bool {
//	return false
//}

func (sw LcSyringe) prefix() string {
	return lcSyringeIdPrefix
}

func (sw LcSyringe) CollectionName() string {
	return lcSyringeCollectionName
}

func initializeSyringes(ctx context.Context) error { // TODO; this
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(lcSyringeCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		//AlternateCollectionIdField
		//SubType string // spore or LC
		//// Parent is always either purchased (nil), LC, or LcSyringe
		//AlternateCollectionOptionalParentField // TODO: handle now a pointer       // TODO: likely won't exist for pre-existing
		//CreationDateField                      // create or receive date
		//SpeciesField
		//SubspeciesOptionalField
		//SaleField
		//TransfersOutField
		//DisposedField
		//// TODO: projects?
		//NotesField
		//LastUpdatedField
		//PermsField
		newSimpleIndex("subType", "subType", false, false, false),
		newSimpleIndex("parent", "parent", false, false, false),
		newSimpleIndex("creationDate", "creationDate", true, false, false), // TODO: INDEX CREATION DATES EVERYWHERE!
		newSimpleIndex("species", "species", false, false, false),
		newSimpleIndex("subSpecies", "subSpecies", false, true, false),
		projectsIndexModel,
		newSimpleIndex("sale", "sale", false, true, false),
		//Notes (no index unless tags)
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	existingEntry := LcSyringe{}
	testItem := LcSyringe{
		AlternateCollectionIdField:        AlternateCollectionIdField{exAltId},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&exLC},
		CreationDateField:                 exampleTime.asCreationDate(),
		SpeciesField:                      SpeciesField{testEntryStringId},
		SubspeciesOptionalField:           SubspeciesOptionalField{&testEntryStringId},
		SaleField:                         SaleField{&exAltId},
		DisposedField:                     DisposedField{&exampleTime},
		NotesField:                        NotesField{exampleNotes()},
		LastUpdatedField:                  LastUpdatedField{exampleTime},
		PermsField:                        PermsField{}, // TODO: fix
	}
	err = coll.FindOne(ctx, bson.D{{"_id", exAltId}}).Decode(&existingEntry)
	if err == nil {
		if reflect.DeepEqual(existingEntry, testItem) {
			return nil
		}
	}
	return testExistingEntry(ctx, coll, exAltId, testItem, existingEntry)
}

type createSyringeRequest struct {
	FruitId AlternateCollectionId `bson:"fruitId" json:"fruitId"`
	NotesField
	Pics []PicWithNotesLessLocation //"newPic-1"
}

func (upr createSyringeRequest) reform() resolvedCreateSyringeRequest {
	return resolvedCreateSyringeRequest{
		FruitId:    upr.FruitId,
		NotesField: upr.NotesField,
		PicsField: PicsField{slices.Map(upr.Pics, func(i PicWithNotesLessLocation) PicWithNotes {
			return i.asPicWithNotes(nil)
		})},
	}
}

type resolvedCreateSyringeRequest struct {
	// TODO: spore print id or lc id
	FruitId AlternateCollectionId `bson:"fruitId" json:"fruitId"`
	NotesField
	PicsField
}

func createSyringeHandler(w http.ResponseWriter, r *http.Request) {
	data := createSyringeRequest{}
	id := newAlternateCollectionId()
	b58Id := id.asBase58()
	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, maxMultipartRequestSize)
	reader, err := r.MultipartReader() // TODO: do streamlined
	if err != nil {
		http.Error(w, "unable to open multipart reader: "+err.Error(), http.StatusBadRequest)
		return
	}
	p, err := reader.NextPart()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Process text (or object)
	bs, errr := io.ReadAll(p)
	if errr != nil {
		err = errr
		http.Error(w, "failed to read data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	// PARSE INTO CORRECT DATA FORMAT
	err = json.Unmarshal(bs, &data)
	if err != nil {
		http.Error(w, "failed to unmarshal data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	// Get any images
	newPics := map[int]string{}
	picsSaved := []string{}
	defer func() {
		if err != nil {
			errDel := pics.DeleteFiles(r.Context(), picsSaved...)
			if errDel != nil {
				handleFileDeleteErr(errDel)
			}
		}
	}()
	for {
		// Go to next part or break
		p, err = reader.NextPart()
		if err != nil {
			if err != io.EOF {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			break
		}
		fileName := p.FileName()
		if fileName == "" {
			http.Error(w, "file name is empty for what should have been an image", http.StatusBadRequest)
			return
		}
		// Process file
		parts := strings.Split(fileName, "-")
		if len(parts) != 2 {
			http.Error(w, "invalid image name: "+fileName, http.StatusBadRequest)
			return
		}
		num, errr := strconv.Atoi(parts[1])
		if errr != nil {
			err = errr
			http.Error(w, "failed to parse image number! "+errr.Error(), http.StatusBadRequest)
			return
		}
		fieldBytes, err := multipartToImageBytes(p, w)
		if err != nil {
			// Already wrote
			return
		}
		switch parts[0] {
		case "newPic":
			newFileNameWithPrefixPath, err := pics.SaveFile(r.Context(), fieldBytes, "LcSyringe", string(b58Id), "img")
			if err != nil {
				http.Error(w, "failed to save new picture: "+err.Error(), http.StatusBadRequest)
				return
			}
			picsSaved = append(picsSaved, newFileNameWithPrefixPath)
			newPics[num] = newFileNameWithPrefixPath
		default:
			http.Error(w, "invalid picture name", http.StatusBadRequest)
			return
		}
	}
	// CHECK THAT ALL NEW PICS EXIST
	// PROCESS ALL NEW PICS AND CONTAMS
	out := data.reform()
	for i, _ := range data.Pics {
		loc, exists := newPics[i]
		if !exists {
			http.Error(w, fmt.Sprintf("error, location for new picture index %d not found (should never happen)", i), http.StatusInternalServerError)
			return
		}
		out.Pics[i].Location = imageLocation(loc)
	}

	_, txErr := doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		db := ctx.Client().Database(dbName)
		parent := LiquidCulture{} // TODO: not fruit
		err = db.Collection(fruitsCollName).FindOne(ctx, bson.D{{"_id", id}}).Decode(&parent)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}

		now := unixTimeForNow()
		toInsert := LcSyringe{
			AlternateCollectionIdField:        AlternateCollectionIdField{id},
			MainCollectionOptionalParentField: MainCollectionOptionalParentField{&parent.Id},
			CreationDateField:                 now.asCreationDate(),
			SpeciesField:                      SpeciesField{Species: *parent.Species},
			SubspeciesOptionalField:           parent.SubspeciesOptionalField,
			NotesField:                        NotesField{out.Notes},
			LastUpdatedField:                  LastUpdatedField{now},
			// Do not check permissions, just pass parent perms to child
			PermsField: PermsField{parent.Perms},
		}
		_, err = db.Collection(lcSyringeCollectionName).InsertOne(ctx, toInsert)
		if err != nil {
			return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		}
		// TODO: Update LC with new syringe???
		//err = parent.addSyringe(ctx, spid)
		//if err != nil {
		//	return DbTxnStdErr(w, err.Error(), http.StatusInternalServerError)
		//}
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

type updateSyringeRequest struct {
	SaleField // TODO: validate?
	DisposedField
	Notes AllEntries[Note]
	PermsField
}

func (upr updateSyringeRequest) reform() resolvedUpdateSyringeRequest {
	return resolvedUpdateSyringeRequest{
		SaleField:     upr.SaleField,
		DisposedField: upr.DisposedField,
		Notes:         upr.Notes,
		PermsField:    PermsField{upr.Perms},
	}
}

type resolvedUpdateSyringeRequest struct {
	SaleField
	DisposedField
	Notes AllEntries[Note]
	Pics  SplitEntries[picWithNotesForm, PicWithNotes]
	PermsField
}

func updateSyringeHandler(w http.ResponseWriter, r *http.Request) {
	data := updateSyringeRequest{}
	b58Id := Base58Str(r.PathValue("id")) // TODO: ensure ok
	id, err := b58Id.toMainCollectionId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	// CHECK THAT ALL NEW PICS EXIST
	// PROCESS ALL NEW PICS AND CONTAMS
	out := data.reform()

	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(lcSyringeCollectionName)
		// go get current LcSyringe
		existing := LcSyringe{}
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

type importSyringeRequest struct {
	CreationDateField
	SpeciesField
	SubspeciesOptionalField
	NotesField
	// pic as "img"
	PermsField
}

func importSyringeHandler(w http.ResponseWriter, r *http.Request) {
	data := importSyringeRequest{}
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

		toInsert := LcSyringe{
			AlternateCollectionIdField: AlternateCollectionIdField{id},
			CreationDateField:          data.CreationDateField,
			SpeciesField:               data.SpeciesField,
			SubspeciesOptionalField:    data.SubspeciesOptionalField,
			NotesField:                 data.NotesField,
			LastUpdatedField:           LastUpdatedFieldForNow(),
			PermsField:                 PermsField{finalPerms},
		}
		coll := ctx.Client().Database(dbName).Collection(lcSyringeCollectionName)
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
