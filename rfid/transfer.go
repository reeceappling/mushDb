package rfid

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/reeceappling/mushDb/rfid/pics"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/exp/maps"
	"io"
	"net/http"
	"reflect"
)

type TransfersOutField struct {
	TransfersOut []AlternateCollectionId `bson:"transfersOut,omitempty" json:"transfersOut,omitempty"`
}

type InnocField struct { // TODO: multi-innocs?
	Innoc *AlternateCollectionId `bson:"innoc,omitempty" json:"innoc,omitempty"`
}

func (in InnocField) Innoculatable() bool {
	return in.Innoc == nil
}

type transferReason string

var transferReasons = map[transferReason]string{
	"outgrew":      "outgrew plate",
	"contaminated": "parent was contaminated",
	"sectoring":    "transferring a specific sector",
	// TODO: more?
}

type Transfer struct { // TODO: does not include multi-jar transfers from jars to monotubs
	AlternateCollectionIdField `bson:"inline"`
	From                       []MainCollectionId `bson:"from" json:"from"` // TODO: THIS USED TO NOT BE A SLICE
	To                         MainCollectionId   `bson:"to" json:"to"`     // fruit is mainCollectionId
	FromType                   string             `json:"fromType"`         //sourceType     //fruit, sporePrint, mss, plate, jar, stasis, lc, slant, bag, box
	ToType                     string             `json:"toType"`           //sourceType
	CreationDateField          `bson:"inline"`    // TODO; changed from date to creationDate
	Reason                     transferReason     `bson:"reason" json:"reason"`
	FromImage                  *imageLocation     `bson:"fromImage,omitempty" json:"fromImage,omitempty"`
	ToImage                    *imageLocation     `bson:"toImage,omitempty" json:"toImage,omitempty"`
	NotesField                 `bson:"inline"`
	LastUpdatedField           `bson:"inline"`
	AclField                   `bson:"inline"` // TODO: handle EVERYWHERE
}

func (t Transfer) EntryTypeField() *string {
	return nil
}

func (t Transfer) PicsModsForChild() *Mods {
	if t.ToImage == nil {
		return NewMods()
	}
	pic := PicWithNotes{
		Time:       t.CreationDate,
		Location:   *t.ToImage,
		NotesField: NotesField{[]Note{}},
	}
	return NewMods().
		withMostRecentImage(&pic).
		withPics([]PicWithNotes{pic})
}

// Perms have not been checked when this runs yet // TODO: ?
func getGeneticItem(ctx context.Context, entryType string, id MainCollectionId) (geneticSource, error) {
	b58id := id.asBase58()
	if tempItem, exists := mainCollMap[entryType]; exists { // TODO: no longer need this
		out, err := GetMainCollectionItem(ctx, id, tempItem) // TODO: use multiple collections here!
		if err != nil {
			return nil, errors.Join(errors.New("failed to get main coll item genetic item"), err)
		}
		return out, nil
	}
	acId, err := b58id.toAltCollectionId()
	if err != nil {
		return nil, errors.Join(errors.New("invalid altCollectionId"), err)
	}
	outType, exists := map[string]AltCollectionItem{
		"fruit":      Fruit{},
		"sporePrint": SporePrint{},
	}[entryType]
	if !exists {
		return nil, errors.Join(errors.New("invalid entry type"), err)
	}
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(outType.CollectionName())
	err = coll.FindOne(ctx, bson.D{{"_id", acId}}).Decode(&outType)
	if err != nil {
		return nil, err
	}
	return outType.(geneticSource), nil
}

func initializeTransfers(ctx context.Context) error {
	// Indices
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(TransfersCollName)
	_, err := coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		// TODO: UNSURE WHICH INDICES ARE NEEDED
		// TODO: ensure from index indexes all of the child ids
		// TODO: newSimpleIndex("from", "from", true, false, false),
		// TODO: newSimpleIndex("to", "to", true, false, false),
		// TODO: newSimpleIndex("fromType", "fromType", false, false, false),
		// TODO: newSimpleIndex("toType", "toType", false, false, false),
		creationDateIndexModel,
		// TODO: newSimpleIndex("reason", "reason", false, false, false),
		//FromImage (no index)
		//ToImage (no index)
		//Notes (no index unless tags)
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	existingEntry := Transfer{}
	// TODO: also create many-to-one monotub test transfer
	testItem := Transfer{
		AlternateCollectionIdField: AlternateCollectionIdField{exAltId},
		From:                       []MainCollectionId{exPlate},
		To:                         exJar,
		FromType:                   "plate",
		ToType:                     "jar",
		CreationDateField:          CreationDateField{exampleTime},
		Reason:                     "A_REASONABLE_TRANSFER_REASON",
		FromImage:                  (*imageLocation)(&exPicLoc),
		ToImage:                    (*imageLocation)(&exPicLoc),
		NotesField:                 NotesField{exampleNotes()},
		LastUpdatedField:           LastUpdatedField{exampleTime},
	}
	err = coll.FindOne(ctx, bson.D{{"_id", exAltId}}).Decode(&existingEntry)
	if err == nil {
		if reflect.DeepEqual(existingEntry, testItem) {
			return nil
		}
	}
	return testExistingEntry(ctx, coll, exAltId, testItem, existingEntry)
}

type createTransferRequest struct {
	From     MainCollectionId `json:"from,omitempty"` // TODO: check all other requests that accept bytes, will likely need to be this
	To       MainCollectionId `json:"to,omitempty"`
	FromType string           `json:"fromType,omitempty"`
	ToType   string           `json:"toType,omitempty"`
	Reason   string           `json:"reason,omitempty"`
	// FromImage == 'picFrom'
	// ToImage == 'picTo'
	NotesField
	// TODO: perms from parent?
}

func createTransferHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: CANNOT USE TRANSACTIONS!!!!!!
	data := createTransferRequest{}
	id := newAlternateCollectionId()
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
	var (
		fromPic *string
		toPic   *string
	)
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
		p, err := reader.NextPart()
		if err != nil {
			if err != io.EOF {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			break
		}
		fieldBytes, errr := multipartToImageBytes(p, w)
		if errr != nil {
			// Already wrote
			return
		}
		newFileNameWithPrefixPath, err := pics.SaveFile(r.Context(), fieldBytes, "transfer", string(b58id), "img")
		if err != nil {
			http.Error(w, "failed to save image", http.StatusBadRequest)
			return
		}
		picsSaved = append(picsSaved, newFileNameWithPrefixPath)
		switch p.FileName() {
		case "picFrom":
			if fromPic != nil {
				http.Error(w, "too many from images", http.StatusBadRequest)
				return
			}
			fromPic = &newFileNameWithPrefixPath
		case "picTo":
			if toPic != nil {
				http.Error(w, "too many dest images", http.StatusBadRequest)
				return
			}
			toPic = &newFileNameWithPrefixPath
		default:
			http.Error(w, "invalid image name!", http.StatusBadRequest)
			return
		}
	}
	ctx := r.Context()
	parentId := data.From
	childId := data.To
	now := unixTimeForNow()
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	// Get parent and child items
	parent, errr := getGeneticItem(ctx, data.FromType, data.From)
	if errr != nil {
		http.Error(w, "failed to get parent item: "+errr.Error(), http.StatusBadRequest)
		return
	}
	child, errr := getGeneticItem(ctx, data.ToType, data.To)
	if errr != nil {
		http.Error(w, "failed to get child item: "+errr.Error(), http.StatusBadRequest)
		return
	}
	if err = parent.CanTransferTo(child); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// TODO: set child perms to the parent perms!

	// TODO: ensure email has perms to make this transfer? (can edit parent)
	resolvedPerms, err := GetResolvedUserPerms(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	// to make a transfer, the email must only be able to write to the child initially
	if userChildPerm := child.Permissions().HighestPermFor(resolvedPerms); userChildPerm == nil || !(*userChildPerm) {
		http.Error(w, "you do not have permissions to create this transfer, you likely cannot modify the parent, or the child is not eligible to be transferred to", http.StatusUnauthorized)
		return
	}
	// Create Transfer
	xfer := Transfer{
		AlternateCollectionIdField: AlternateCollectionIdField{id},
		From:                       []MainCollectionId{parentId},
		To:                         childId,
		FromType:                   data.FromType,
		ToType:                     data.ToType,
		CreationDateField:          CreationDateField{now},
		Reason:                     transferReason(data.Reason),
		FromImage:                  (*imageLocation)(fromPic),
		ToImage:                    (*imageLocation)(toPic),
		NotesField:                 data.NotesField,
		LastUpdatedField:           LastUpdatedFieldForNow(),
		AclField:                   AclField{ACL: parent.Permissions()},
	}
	_, err = db.Collection(TransfersCollName).InsertOne(ctx, xfer)
	if err != nil {
		http.Error(w, "failed to create transfer: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Set rollback
	rollbackXfer := func() error {
		result, errrr := db.Collection(TransfersCollName).DeleteOne(ctx, bson.D{{"_id", xfer.Id}})
		if errrr != nil {
			return errrr
		}
		if result.DeletedCount == 0 {
			return errors.New("transfer not deleted")
		}
		return nil
	}
	err, rollbackParent := parent.setTransferParent(ctx, xfer)
	if err != nil {
		err = errors.Join(errors.New("failed to set transfer parent"), rollbackXfer(), err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = child.setTransferChild(ctx, xfer, parent)
	if err != nil {
		parentRollbackErr := rollbackParent()
		if parentRollbackErr != nil {
			// TODO: HUGE ERROR. FIGURE OUT
		}
		xferRollbackErr := rollbackXfer()
		if xferRollbackErr != nil {
			// TODO: HUGE ERROR. FIGURE OUT
		}
		err = errors.Join(errors.New("failed to set transfer child"), parentRollbackErr, xferRollbackErr, err)
		http.Error(w, "failed to set transfer child: "+err.Error(), http.StatusInternalServerError)
		return
	}

	bsOut, err := json.Marshal(xfer)
	if err != nil {
		// Do not rollback here. Data made it in successfully
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(bsOut)
	if err != nil {
		// Do not rollback here. Data made it in successfully
		handleWriteErr(err, w)
	}
}

type updateTransferRequest struct {
	Notes          AllEntries[Note] `json:"notes,omitempty"`
	PermsOnRequest                  // TODO: ????????? handle in typescript and handler!
}

func updateTransferHandler(w http.ResponseWriter, r *http.Request) {
	b58Id := Base58Str(r.PathValue("id"))
	defer r.Body.Close()
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body: "+err.Error(), http.StatusBadRequest)
		return
	}
	req := updateTransferRequest{}
	err = json.Unmarshal(bs, &req)
	if err != nil {
		http.Error(w, "failed to unmarshal body: "+err.Error(), http.StatusBadRequest)
		return
	}
	id, err := b58Id.toAltCollectionId()
	if err != nil {
		http.Error(w, "Invalid id! "+err.Error(), http.StatusBadRequest)
		return
	}
	_, err = doTxn(r.Context(), func(ctx mongo.SessionContext) (interface{}, error) {
		coll := ctx.Client().Database(dbName).Collection(TransfersCollName)
		existing, err := GetAltCollectionItemInTxn(ctx, id, Transfer{})
		if err != nil {
			stat := http.StatusInternalServerError
			if err == mongo.ErrNoDocuments {
				stat = http.StatusNotFound
			}
			return dbErr(w, err.Error(), stat)
		}
		user, err := GetAuthInfo(ctx)
		if err != nil {
			return dbErr(w, err.Error(), http.StatusInternalServerError)
		}
		if !user.HasPermissionToEdit(existing) {
			return dbErr(w, "unauthorized to edit", http.StatusForbidden)
		}
		aclField, err := req.AclFor(ctx, user)
		if err != nil {
			return dbErr(w, err.Error(), http.StatusInternalServerError)
		}
		upd, err := NewMods().
			updateNotesIfNeeded(req.Notes, existing.Notes). // TODO: make sure this works the way we want
			updatePermsIfNeeded(aclField.ACL, existing.ACL).
			updateLastUpdatedIfNeeded().
			Finalized()
		if err != nil {
			return dbErr(w, "error resolving updates list: "+err.Error(), http.StatusInternalServerError)
		}
		bsonId := bson.D{{"_id", existing.Id}}
		err = coll.FindOneAndUpdate(ctx, bsonId, upd).Err()
		if err != nil {
			return dbErr(w, err.Error(), http.StatusInternalServerError)
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
		handleWriteErr(err, w)
	}
}

func getAllTransferReasonsHandler(w http.ResponseWriter, r *http.Request) { // TODO: use
	writeAsJson(w, maps.Keys(transferReasons))
}
