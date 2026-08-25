package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/reeceappling/mushDb/api/env"
	"github.com/reeceappling/mushDb/api/pics"
	"github.com/reeceappling/mushDb/api/request"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"io"
	"mime/multipart"
	"net/http"
)

type TransfersOutField struct {
	TransfersOut []AlternateCollectionId `bson:"transfersOut,omitempty" json:"transfersOut,omitempty"`
}

type InnocField struct { // TODO: multi-innocs?
	Innoc *AlternateCollectionId `bson:"innoc,omitempty" json:"innoc,omitempty"`
}

func (f InnocField) RequireNoInnoculation() error {
	if f.Innoc != nil {
		return errors.New("must not be innoculated")
	}
	return nil
}

type transferReason string

const xferReasonColonized transferReason = "colonized"
const xferReasonReady transferReason = "ready"
const xferReasonHarvest transferReason = "harvest"

var transferReasons = map[transferReason]string{ // TODO: move these to autogenned?
	"outgrew":           "outgrew plate", // TODO: is colonized just this?
	"contaminated":      "parent was contaminated",
	"sectoring":         "transferring a specific sector",
	"age":               "sample is very old",
	xferReasonColonized: "fully colonized",
	xferReasonReady:     "ready",   // TODO: ?????
	xferReasonHarvest:   "harvest", // TODO: ONLY FOR HARVESTING!!!! NOT FOR NORMAL TRANSFERS!
}

var sporePrintColors = []SporePrintColor{ // TODO: move these to autogenned?
	SpColorBlack,
	SpColorBrown,
	SpColorRed,
	SpColorOrange,
	SpColorYellow,
	SpColorTanLight,
	SpColorWhite,
	SpColorClear,
}

var sporePrintDensities = []SporePrintDensity{ // TODO: move these to autogenned?
	SpDensityHeavy,
	SpDensityAvg,
	SpDensitySparse,
	spDensityNoneMinimal,
}

type Transfer struct { // TODO: does not include multi-jar transfers from jars to monotubs
	AlternateCollectionIdField `bson:"inline"`
	From                       MainCollectionId `bson:"from" json:"from"`
	To                         MainCollectionId `bson:"to" json:"to"` // fruit is mainCollectionId
	FromType                   string           `json:"fromType"`     //sourceType     //fruit, sporePrint, mss, plate, jar, stasis, lc, slant, bag, box
	ToType                     string           `json:"toType"`       //sourceType
	CreationDateField          `bson:"inline"`
	Reason                     transferReason `bson:"reason" json:"reason"`
	FromImage                  *ImageLocation `bson:"fromImage,omitempty" json:"fromImage,omitempty"`
	ToImage                    *ImageLocation `bson:"toImage,omitempty" json:"toImage,omitempty"`
	NotesField                 `bson:"inline"`
	LastUpdatedField           `bson:"inline"`
	AclField                   `bson:"inline"`
}

func (t Transfer) Blank() CollectionItem {
	return &Transfer{}
}

func (t Transfer) PicsModsForChild(child HasPicsField) *Mods {
	if t.ToImage == nil {
		return NewMods()
	}
	pic := newPicWithNotes(t.CreationDate, []Note{}, *t.ToImage) // TODO: note?
	return NewMods().
		withMostRecentImage(&pic).
		withPics(child.currentPics().addPic(pic).Pics) // TODO: ensure works as expected
}
func (t Transfer) PicsModsForParent(parent HasPicsField) *Mods {
	if t.FromImage == nil {
		return NewMods()
	}
	pic := newPicWithNotes(t.CreationDate, []Note{}, *t.FromImage) // TODO: note?
	return NewMods().
		withMostRecentImage(&pic).
		withPics(parent.currentPics().addPic(pic).Pics) // TODO: ensure works as expected
}

// Perms have not been checked when this runs yet
func getGeneticItem(ctx context.Context, entryType string, id MainCollectionId) (MainCollectionItem, error) {
	tempItem, exists := mainCollMap(entryType)
	if !exists {
		return nil, errors.New("invalid entry type: " + entryType)
	}
	out, err := GetMainCollectionItem(ctx, id, tempItem)
	if err != nil {
		err = errors.Join(errors.New("failed to get main coll item genetic item"), err)
	}
	return out, err
}

func initializeTransfers(ctx context.Context) error {
	// Indices
	coll := DbFrom(ctx).Collection(TransfersCollName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
		creationDateIndexModel,
		// TODO: ensure from index indexes all of the child ids (multiple only for very specific cases?)
		// newSimpleIndex("from", "from", true, false, false), // TODO: do we want to search by source?
		// newSimpleIndex("to", "to", true, false, false), // TODO: do we want to search by destination?
		// newSimpleIndex("fromType", "fromType", false, false, false),
		// newSimpleIndex("toType", "toType", false, false, false),
		// newSimpleIndex("reason", "reason", false, false, false),
		//FromImage (no index)
		//ToImage (no index)
		//Notes (no index unless tags)
		projectsIndexModel,
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	// TODO: also create many-to-one monotub test transfer
	return env.IfNotProd(ctx, func() error {
		testItem := &Transfer{
			AlternateCollectionIdField: exAltId.asIdField(),
			From:                       exPlate,
			To:                         exJar,
			FromType:                   "plate",
			ToType:                     "jar",
			CreationDateField:          CreationDateField{exampleTime},
			Reason:                     xferReasonReady,
			FromImage:                  (*ImageLocation)(&exPicLoc),
			ToImage:                    (*ImageLocation)(&exPicLoc),
			NotesField:                 NotesField{exampleNotes()},
			LastUpdatedField:           LastUpdatedField{exampleTime},
			AclField:                   allCanWriteAcl(),
		}
		return addTestAltEntries(ctx, testItem)
	})
}

type createTransferRequest struct {
	From   MainCollectionId `json:"from"`
	To     MainCollectionId `json:"to"`
	Reason string           `json:"reason"`
	// FromImage == 'picFrom'
	// ToImage == 'picTo'
	FromType *string `json:"fromType,omitempty"`
	NotesField
	DisposeParent bool `json:"disposeParent"`
}

func newTxn(ctx context.Context, transact func(mongo.SessionContext) (any, error)) (any, error) {
	sessionOptions := options.Session()
	sess, err := GetMongoClient(ctx).StartSession(sessionOptions)
	if err != nil {
		return nil, err
	}
	wc := writeconcern.Majority()
	txnOptions := options.Transaction().SetWriteConcern(wc)
	// Defers ending the session after the transaction is committed or ended
	return sess.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		defer sess.EndSession(ctx)
		out, err := transact(sessCtx)
		if err != nil {
			return nil, errors.Join(err, sess.AbortTransaction(ctx))
		}
		if err = sess.CommitTransaction(ctx); err != nil {
			return nil, errors.Join(err, sess.AbortTransaction(ctx))
		}
		return out, nil
	}, txnOptions)
}

func createTransferHandler(w http.ResponseWriter, r *http.Request) {
	data := createTransferRequest{}
	id := newAlternateCollectionId()
	ctx, now := request.UnixTime(r.Context())
	b58id := id.AsBase58()
	r.Body = http.MaxBytesReader(w, r.Body, maxMultipartRequestSize)
	defer r.Body.Close()
	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "unable to open multipart reader: "+err.Error(), http.StatusBadRequest)
		return
	}
	p1, err := reader.NextPart()
	if err != nil {
		println("failed to go to next part: " + err.Error()) // TODO: ok?
		http.Error(w, "failed to go to next part: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer p1.Close()
	// Process text (or object)
	bs, errr := io.ReadAll(p1)
	if errr != nil {
		err = errr
		println("failed to read Data from form: " + err.Error()) // TODO: ok?
		http.Error(w, "failed to read Data from form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// PARSE INTO CORRECT DATA FORMAT
	err = json.Unmarshal(bs, &data)
	if err != nil {
		dbErrCtx(ctx, w, errors.Join(errors.New("failed to unmarshal Data from form"), err), http.StatusBadRequest)
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
		var p *multipart.Part
		p, err = reader.NextPart()
		if err != nil {
			if err == io.EOF {
				err = nil
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			break
		}
		var fieldBytes []byte

		fieldBytes, err = multipartToImageBytes(p, w)
		if err != nil {
			// Already wrote
			return
		}
		var newFileNameWithPrefixPath string
		newFileNameWithPrefixPath, err = pics.SaveFile(r.Context(), fieldBytes, "transfer", string(b58id), "img")
		if err != nil {
			http.Error(w, "failed to save image", http.StatusBadRequest)
			return
		}
		picsSaved = append(picsSaved, newFileNameWithPrefixPath)
		switch p.FileName() {
		case "picFrom":
			if fromPic != nil {
				err = errors.New("too many from images")
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			fromPic = &newFileNameWithPrefixPath
		case "picTo":
			if toPic != nil {
				err = errors.New("too many dest images")
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			toPic = &newFileNameWithPrefixPath
		default:
			err = errors.New("invalid image name: " + p.FileName())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	parentId := data.From
	childId := data.To
	childMapEntry := idMapEntry{}
	var fromType string
	if data.FromType != nil {
		fromType = *data.FromType
	} else {
		//if fromType is nil, then we must go find the fromType
		fromType, err = getEntryTypeForId(ctx, parentId)
	}
	// Get parent and child items
	err = DbFrom(ctx).Collection(idMapCollectionName).FindOne(ctx, bson.M{IDfld: childId}).Decode(&childMapEntry)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "child not found in id db: "+err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, "error getting child from id db: "+err.Error(), http.StatusInternalServerError)
		return
	}

	parent, err := getGeneticItem(ctx, fromType, data.From)
	if err != nil {
		http.Error(w, "failed to get parent item: "+err.Error(), http.StatusBadRequest)
		return
	}
	child, err := getGeneticItem(ctx, childMapEntry.EntryType, data.To)
	if err != nil {
		http.Error(w, "failed to get child item: "+err.Error(), http.StatusBadRequest)
		return
	}
	// ensure child is innoculatable and parent is not disposed
	if err = child.Innoculatable(); err != nil {
		http.Error(w, "child is not innoculatable. "+err.Error(), http.StatusBadRequest)
		return
	}
	if parent.DisposalInfo() != nil {
		http.Error(w, "cannot transfer from disposed entries", http.StatusBadRequest)
		return
	}
	if err = parent.CanTransferTo(child); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Perms are checked earlier, and only guests cannot write
	// Create Transfer
	xfer := Transfer{
		AlternateCollectionIdField: AlternateCollectionIdField{id},
		From:                       parentId,
		To:                         childId,
		FromType:                   fromType,
		ToType:                     childMapEntry.EntryType,
		CreationDateField:          CreationDateField{now},
		Reason:                     transferReason(data.Reason),
		FromImage:                  (*ImageLocation)(fromPic),
		ToImage:                    (*ImageLocation)(toPic),
		NotesField:                 data.NotesField,
		LastUpdatedField:           LastUpdatedField{now},
		AclField:                   AclField{ACL: parent.Permissions()}, //set child perms to the parent perms!
	}
	_, err = newTxn(ctx, func(sessCtx mongo.SessionContext) (any, error) {
		return nil, createTransferInTxn(sessCtx, parent, child, xfer, data.DisposeParent)
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	bsOut, errMarshalling := json.Marshal(xfer)
	if errMarshalling != nil { // Not err because err != nil at end will delete all images
		// Do not rollback here. Data made it in successfully
		http.Error(w, errMarshalling.Error(), http.StatusInternalServerError)
		return
	}

	_, errWriting := w.Write(bsOut)
	if errWriting != nil {
		// Do not rollback here. Data made it in successfully
		handleWriteErr(errWriting, w)
	}
}

func createTransferInTxn[T MainCollectionItem, U MainCollectionItem](ctx mongo.SessionContext, parent T, child U, xfer Transfer, dispose bool) error {
	db := mongo.SessionFromContext(ctx).
		Client().Database(dbName)
	_, err := db.Collection(TransfersCollName).InsertOne(ctx, xfer)
	if err != nil {
		return err
	}
	if err = setTransferParent(ctx, parent, xfer, dispose); err != nil {
		return errors.Join(errors.New("failed to set transfer parent"), err)
	}
	if err = child.setTransferChild(ctx, xfer, parent); err != nil {
		return errors.Join(errors.New("failed to set transfer child"), err)
	}
	return nil
}

type updateTransferRequest struct {
	NotesUpdateField
	PermsOnRequest `json:"acl"`
}

func (req updateTransferRequest) modsFor(existing *Transfer, aclField AclField) (bson.D, error) {
	return NewMods().
		updateNotesIfNeeded(req, existing).
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
}

func updateTransferHandler(w http.ResponseWriter, r *http.Request) {
	_, id, err := altCollIdFromRequest(r, w)
	if err != nil {
		return
	}
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
	ctx, db := Db(r)
	coll := db.Collection(TransfersCollName)
	existing, err := GetAltCollectionItemOutsideTxn(ctx, id, Transfer{})
	if err != nil {
		stat := http.StatusInternalServerError
		if err == mongo.ErrNoDocuments {
			stat = http.StatusNotFound
		}
		dbErr(w, err.Error(), stat)
		return
	}
	finishAltCollItemUpdate(ctx, w, coll, req.modsFor, &existing, req.PermsOnRequest)
}

//func deleteTransferHandler(w http.ResponseWriter, r *http.Request) {
//	http.Error(w, "transfers cannot be deleted", http.StatusNotImplemented)
//	return
//}
