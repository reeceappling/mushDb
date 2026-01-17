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

type SporePrint struct {
	MainCollectionIdField `bson:"inline"` // TODO: was alt
	// Parent is always either fruit, or purchased
	MainCollectionOptionalParentField `bson:"inline"` // TODO: handle now a pointer // TODO: used to be an altCollId       // TODO: likely won't exist for pre-existing
	CreationDateField                 `bson:"inline"` // Print or receive date
	SpeciesField                      `bson:"inline"`
	SubspeciesOptionalField           `bson:"inline"`
	PicsField                         `bson:"inline"`
	SaleField                         `bson:"inline"`
	DisposedField                     `bson:"inline"`
	MostRecentImageField              `bson:"inline"`
	NotesField                        `bson:"inline"`
	LastUpdatedField                  `bson:"inline"`
	AclField                          `bson:"inline"` // TODO: handle EVERYWHERE
}

func (sp SporePrint) Innoculatable() bool {
	return false
}

func (sp SporePrint) CanTransferTo(dst geneticSource) error {
	return errors.New("sporePrints cannot transfer. Only be made into mss or swab")
}

func (sp SporePrint) setTransferParent(ctx context.Context, xfer Transfer) (error, func() error) {
	// TODO: can this even occur?
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(sp.CollectionName())
	upd, err := NewMods().addTransferOut(xfer.Id).Finalized()
	if err != nil {
		return err, nil
	}
	res, err := coll.UpdateByID(ctx, sp.Id, upd)
	if err != nil {
		return err, nil
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer, nil
	}
	return nil, func() error {
		return coll.FindOneAndReplace(ctx, bson.D{{"_id", sp.Id}}, sp).Err()
	}
}

func (sp SporePrint) setTransferChild(ctx context.Context, xfer Transfer, from geneticSource) error {
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
		errors.New("only fruits are supported as a transfer source type into sporePrints")
	}
	upd, err := xfer.
		PicsModsForChild().
		withInnoc(xfer).
		withParent(utils.Pointer(from.DbId())).
		withSpecies(parentInfo.Species).
		withSubspecies(parentInfo.SubSpecies).
		updateLastUpdatedIfNeeded().
		Finalized()
	res, err := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(sp.CollectionName()).UpdateByID(ctx, sp.Id, upd)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNoParentModifiedForTransfer
	}
	return nil
}

func (sp SporePrint) GeneticInfoAsParent() (GeneticParentInfo, error) {
	return GeneticParentInfo{
		SpeciesOptionalField:    sp.SpeciesField.AsOptional(),
		SubspeciesOptionalField: sp.SubspeciesOptionalField,
		GenerationsFields:       GenerationsFieldFor(utils.Pointer(Generation(0))),
	}, nil
}

func (sp SporePrint) generation() (sinceSpore *Generation, sinceSporeOrClone *Generation) {
	return utils.Pointer(Generation(0)), utils.Pointer(Generation(0))
}

func (sp SporePrint) EntryTypeField() *string {
	return nil
}

func (sp SporePrint) id() []byte {
	return []byte(sp.Id.dbIdStr())
}

func initializeSporePrints(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(SporePrintCollectionName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		//newSimpleIndex("parent", "parent", false, false, false),
		//newSimpleIndex("printDate", "creationDate", true, false, false), // TODO: INDEX CREATION DATES EVERYWHERE!
		//newSimpleIndex("species", "species", false, false, false),
		//newSimpleIndex("subSpecies", "subSpecies", false, true, false),
		// Pics
		// TODO: projectsIndexModel,
		//saleIndexModel,
		//disposedIndexModel,
		// MostRecentImage
		//Notes (no index unless tags)
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	existingEntry := SporePrint{}
	testItem := SporePrint{
		MainCollectionIdField:             MainCollectionIdField{exSporePrint},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&exFruitId},
		CreationDateField:                 exampleTime.asCreationDate(),
		SpeciesField:                      SpeciesField{testEntryStringId},
		SubspeciesOptionalField:           SubspeciesOptionalField{&testEntryStringId},
		PicsField:                         PicsField{exPics},
		SaleField:                         SaleField{&exAltId},
		DisposedField:                     DisposedField{&exampleTime},
		MostRecentImageField:              MostRecentImageField{utils.Pointer(exPics[0])},
		NotesField:                        NotesField{exampleNotes()},
		LastUpdatedField:                  LastUpdatedField{exampleTime},
	}
	err = coll.FindOneAndReplace(ctx, bson.D{{"_id", exAltId}}, testItem).Decode(&existingEntry)
	if err == nil {
		if reflect.DeepEqual(existingEntry, testItem) {
			return nil
		}
	}
	return testExistingEntry(ctx, coll, exAltId, testItem, existingEntry)
}

type createSporePrintRequest struct {
	FruitId AlternateCollectionId `bson:"fruitId" json:"fruitId"`
	NotesField
	Pics []PicWithNotesLessLocation //"newPic-1"
	// TODO: USE PARENT PERMS?
}

func (upr createSporePrintRequest) reform() resolvedCreateSporePrintRequest {
	return resolvedCreateSporePrintRequest{
		FruitId:    upr.FruitId,
		NotesField: upr.NotesField,
		PicsField: PicsField{slices.Map(upr.Pics, func(i PicWithNotesLessLocation) PicWithNotes {
			return i.asPicWithNotes(nil)
		})},
	}
}

type resolvedCreateSporePrintRequest struct {
	FruitId AlternateCollectionId `bson:"fruitId" json:"fruitId"`
	NotesField
	PicsField
}

func createSporePrintHandler(w http.ResponseWriter, r *http.Request) {
	data := createSporePrintRequest{}
	id, err := newCollectionId(r.Context(), SporePrintCollectionName)
	if err != nil {
		http.Error(w, "unable to create new mainCollId: "+err.Error(), http.StatusInternalServerError)
		return
	}
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
			newFileNameWithPrefixPath, err := pics.SaveFile(r.Context(), fieldBytes, "sporePrint", string(b58Id), "img")
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
	ctx := r.Context()
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	// TODO: add spore print to mapCollection
	parent := Fruit{}
	err = db.Collection(FruitsCollName).FindOne(ctx, bson.D{{"_id", id}}).Decode(&parent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	now := unixTimeForNow()
	spid := id
	var mri *PicWithNotes = nil
	if len(out.Pics) > 0 {
		lastPic := out.Pics[len(out.Pics)-1]
		mri = &lastPic
	}
	toInsert := SporePrint{
		MainCollectionIdField:             MainCollectionIdField{spid},
		MainCollectionOptionalParentField: MainCollectionOptionalParentField{&parent.Id},
		CreationDateField:                 now.asCreationDate(),
		SpeciesField:                      parent.SpeciesField,
		SubspeciesOptionalField:           parent.SubspeciesOptionalField,
		PicsField:                         out.PicsField,
		MostRecentImageField:              MostRecentImageField{mri},
		NotesField:                        NotesField{out.Notes},
		LastUpdatedField:                  LastUpdatedField{now},
		// Do not check permissions, just pass parent perms to child
		AclField: parent.AclField,
	}
	_, err = db.Collection(SporePrintCollectionName).InsertOne(ctx, toInsert)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Update fruit with new print id
	err = parent.addSporePrint(ctx, spid)
	if err != nil {
		// Rollback print insert
		err = errors.Join(db.Collection(SporePrintCollectionName).FindOneAndDelete(ctx, bson.D{{"_id", toInsert.Id}}).Err(), err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	bsOut, err := json.Marshal(toInsert)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(bsOut)
	if err != nil {
		handleWriteErr(err, w)
	}
}

type updateSporePrintRequest struct {
	SaleField // TODO: validate?
	DisposedField
	Notes          AllEntries[Note]
	Pics           SplitEntries[picWithNotesForm, PicWithNotesLessLocation]
	PermsOnRequest // TODO: handle in typescript and handler!
}

func (upr updateSporePrintRequest) reform() resolvedUpdateSporePrintRequest {
	return resolvedUpdateSporePrintRequest{
		SaleField:     upr.SaleField,
		DisposedField: upr.DisposedField,
		Notes:         upr.Notes,
		Pics: SplitEntries[picWithNotesForm, PicWithNotes]{
			Existing: upr.Pics.Existing,
			New: slices.Map(upr.Pics.New, func(i PicWithNotesLessLocation) PicWithNotes {
				return i.asPicWithNotes(nil)
			}),
		},
		PermsOnRequest: upr.PermsOnRequest,
	}
}

type resolvedUpdateSporePrintRequest struct {
	SaleField
	DisposedField
	Notes          AllEntries[Note]
	Pics           SplitEntries[picWithNotesForm, PicWithNotes]
	PermsOnRequest // TODO: handle in typescript and handler!
}

func updateSporePrintHandler(w http.ResponseWriter, r *http.Request) {
	data := updateSporePrintRequest{}
	b58Id := Base58Str(r.PathValue("id")) // TODO: ensure ok
	id, err := b58Id.toMainCollectionId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	newPics, _, _, err := fullMultipartWithNoBreaks(w, r, "sporePrint", &data, b58Id)
	if err != nil {
		// Already wrote
		return
	}

	// CHECK THAT ALL NEW PICS EXIST
	// PROCESS ALL NEW PICS AND CONTAMS
	out := data.reform()
	for i, _ := range data.Pics.New {
		loc, exists := newPics[i]
		if !exists {
			http.Error(w, fmt.Sprintf("error, location for new picture index %d not found (should never happen)", i), http.StatusInternalServerError)
			return
		}
		out.Pics.New[i].Location = imageLocation(loc)
	}

	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(SporePrintCollectionName)
		// go get current sporePrint
		existing := SporePrint{}
		err = coll.FindOne(ctx, bson.D{{"_id", id}}).Decode(&existing)
		if err != nil {
			return dbErr(w, "failed to find current entry: "+err.Error(), http.StatusBadRequest)
		}
		user, err := GetAuthInfo(ctx)
		if err != nil {
			return dbErr(w, err.Error(), http.StatusInternalServerError)
		}
		if !user.HasPermissionToEdit(existing) {
			return dbErr(w, "unauthorized to edit", http.StatusForbidden)
		}
		aclField, err := out.AclFor(ctx, user) // TODO: USE IN modsFor
		if err != nil {
			return dbErr(w, err.Error(), http.StatusInternalServerError)
		}
		//if err = minimalPermsBetween(existing.Perms, data.Perms).ValidateUserCanWrite(ctx); err != nil {
		//	return dbErr(w, "failed to validate overlapping permissions: "+err.Error(), http.StatusBadRequest)
		//}
		upd, err := NewMods().
			updateSaleIfNeeded(out.Sale, existing.Sale).
			updateDisposedIfNeeded(data.Disposed, existing.Disposed).
			updateNotesIfNeeded(data.Notes, existing.Notes).
			updatePicsIfNeeded(out.Pics, existing.Pics).
			updatePermsIfNeeded(aclField.ACL, existing.ACL).
			updateLastUpdatedIfNeeded().
			Finalized()
		if err != nil {
			return dbErr(w, "error creating txn:"+err.Error(), http.StatusInternalServerError)
		}
		if len(upd) == 0 {
			return dbErr(w, "no changes made", http.StatusBadRequest)
		}

		// write updates to db
		bsonId := bson.D{{"_id", id}}
		err = coll.FindOneAndUpdate(ctx, bsonId, upd).Err()
		if err != nil {
			return dbErr(w, "failed to write update to db: "+err.Error(), http.StatusInternalServerError)
		}
		err = coll.FindOne(ctx, bsonId).Decode(&existing)
		if err != nil {
			return dbErr(w, err.Error(), http.StatusInternalServerError)
		}
		bsOut, err := json.Marshal(existing)
		if err != nil {
			return dbErr(w, err.Error(), http.StatusInternalServerError)
		}
		return w.Write(bsOut)
	})
	if err != nil {
		HandleHttpWriteError(err)
	}
}

type importSporePrintRequest struct {
	CreationDateField
	SpeciesField
	SubspeciesOptionalField
	NotesField
	// pic as "img"
	PermsOnRequest // TODO: handle in typescript and handler!
}

func importSporePrintHandler(w http.ResponseWriter, r *http.Request) {
	data := importSporePrintRequest{}
	id, err := newCollectionId(r.Context(), SporePrintCollectionName)
	if err != nil {
		http.Error(w, "unable to create new mainCollId: "+err.Error(), http.StatusInternalServerError)
		return
	}
	b58id := id.asBase58()
	r.Body = http.MaxBytesReader(w, r.Body, maxMultipartRequestSize)
	defer r.Body.Close()
	reader, err := r.MultipartReader() // TODO: do streamlined
	if err != nil {
		http.Error(w, "unable to open multipart reader: "+err.Error(), http.StatusBadRequest)
		return
	}
	p1, err := reader.NextPart()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer p1.Close()
	// Process text (or object)
	bs, errr := io.ReadAll(p1)
	if errr != nil {
		err = errr
		http.Error(w, "unable to read data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	// PARSE INTO CORRECT DATA FORMAT
	err = json.Unmarshal(bs, &data)
	if err != nil {
		http.Error(w, "unable to unmarshal json form data: "+err.Error(), http.StatusBadRequest)
		return
	}
	//if err = data.Perms.ValidateUserCanWrite(r.Context()); err != nil {
	//	http.Error(w, "email cannot write with these perms: "+err.Error(), http.StatusBadRequest)
	//	return
	//}
	// Try to get pic if exists
	picsSaved := []string{}
	defer func() {
		if err != nil {
			errDel := pics.DeleteFiles(r.Context(), picsSaved...)
			if err != nil {
				handleFileDeleteErr(errDel)
			}
		}
	}()
	now := unixTimeForNow()
	// Go to next part, if exists to get image
	var importedPic *PicWithNotes = nil
	p, err := reader.NextPart()
	if err != nil {
		if err != io.EOF {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		fileName := p.FileName()
		defer p.Close()
		if fileName != "img" {
			http.Error(w, "invalid image name", http.StatusBadRequest)
			return
		}
		// Process file
		fieldBytes, err := multipartToImageBytes(p, w)
		if err != nil {
			// Already wrote
			return
		}
		newFileNameWithPrefixPath, errr := pics.SaveFile(r.Context(), fieldBytes, "sporePrint", string(b58id), "img")
		if errr != nil {
			err = errr
			http.Error(w, "failed to save file: "+err.Error(), http.StatusBadRequest)
			return
		}
		picsSaved = append(picsSaved, newFileNameWithPrefixPath)

		importedPic = &PicWithNotes{
			Time:       now,
			Location:   imageLocation(newFileNameWithPrefixPath),
			NotesField: NotesField{[]Note{}},
		}
	}
	pix := []PicWithNotes{}
	if importedPic != nil {
		pix = []PicWithNotes{*importedPic}
	}

	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		// TODO: add print to mapColl
		//finalPerms := data.Perms
		//if data.Perms != nil {
		//	spec, subsp, err := getSpeciesAndSubspecies(ctx, data.Species, data.SubSpecies)
		//	if err != nil {
		//		return dbErr(w, "failed to get species or subspecies: "+err.Error(), http.StatusInternalServerError) // TODO: ok?
		//	}
		//	finalPerms = minimalPermsBetween(spec, subsp)
		//	// TODO: add email perms if provided, as well as make email author?
		//	if !finalPerms.Valid() {
		//		// TODO: invalid species/subspecies perm crossover. DO THIS ELSEwHERE
		//		return dbErr(w, "invalid species/subspecies perm crossover: "+err.Error(), http.StatusInternalServerError) // TODO: ok?
		//	}
		//}
		perms, err := GetAuthInfo(ctx)
		if err != nil {
			return dbErr(w, err.Error(), http.StatusInternalServerError)
		}
		acl, err := data.AclFor(ctx, perms)
		if err != nil {
			return dbErr(w, err.Error(), http.StatusInternalServerError)
		}
		toInsert := SporePrint{
			MainCollectionIdField:   MainCollectionIdField{id},
			CreationDateField:       data.CreationDateField,
			SpeciesField:            data.SpeciesField,
			SubspeciesOptionalField: data.SubspeciesOptionalField,
			PicsField:               PicsField{pix},
			MostRecentImageField:    MostRecentImageField{importedPic},
			NotesField:              data.NotesField,
			LastUpdatedField:        LastUpdatedFieldForNow(),
			AclField:                acl,
		}
		coll := ctx.Client().Database(dbName).Collection(SporePrintCollectionName)
		_, err = coll.InsertOne(ctx, toInsert)
		if err != nil {
			return dbErr(w, err.Error(), http.StatusInternalServerError)
		}
		bsOut, err := json.Marshal(toInsert)
		if err != nil {
			return dbErr(w, err.Error(), http.StatusInternalServerError)
		}
		return w.Write(bsOut)
	})
	if err != nil {
		handleWriteErr(err, w)
	}
}
