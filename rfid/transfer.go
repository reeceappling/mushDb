package rfid

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/reeceappling/mushDb/rfid/pics"
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

func (in InnocField) Innoculatable() bool {
	return in.Innoc == nil
}

type transferReason string

var transferReasons = map[transferReason]string{
	"outgrew":      "outgrew plate",
	"contaminated": "parent was contaminated",
	"sectoring":    "transferring a specific sector",
	"age":          "plate is veryold",
}

var sporePrintColors = map[SporePrintColor]string{
	SpColorBlack:    string(SpColorBlack),
	SpColorTanLight: string(SpColorTanLight),
	SpColorClear:    string(SpColorClear),
}

var sporePrintDensities = map[SporePrintDensity]string{
	SpDensityHeavy:       string(SpDensityHeavy),
	SpDensityAvg:         string(SpDensityAvg),
	SpDensitySparse:      string(SpDensitySparse),
	spDensityNoneMinimal: string(spDensityNoneMinimal),
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
	AclField                   `bson:"inline"`
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

// Perms have not been checked when this runs yet
func getGeneticItem(ctx context.Context, entryType string, id MainCollectionId) (geneticSource, error) {
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
	coll := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(TransfersCollName)
	err := createIndexes(ctx, coll, []mongo.IndexModel{
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
		projectsIndexModel,
		lastUpdatedIndexModel,
	})
	if err != nil {
		return err
	}
	// If test agar batch does not exist, then create it
	// TODO: also create many-to-one monotub test transfer
	testItem := &Transfer{
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
		AclField:                   allCanReadAcl(),
	}
	return addTestAltEntries(ctx, testItem)
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

type CtxKey string

const SessionCtxKey CtxKey = "mongoTxSession"

func newTxn(ctx context.Context, transact func(mongo.SessionContext) (any, error)) (any, error) {
	sessionOptions := options.Session() // TODO: change?
	sess, err := GetMongoClient(ctx).StartSession(sessionOptions)
	if err != nil {
		return nil, err
	}
	wc := writeconcern.Majority()
	txnOptions := options.Transaction().SetWriteConcern(wc) // TODO: ok?
	// Defers ending the session after the transaction is committed or ended
	defer sess.EndSession(ctx)
	return sess.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		if err = sess.StartTransaction(txnOptions); err != nil {
			return nil, err
		}
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
		http.Error(w, "failed to read Data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	// PARSE INTO CORRECT DATA FORMAT
	err = json.Unmarshal(bs, &data)
	if err != nil {
		http.Error(w, "failed to unmarshal Data from form: "+err.Error(), http.StatusBadRequest)
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
			err = errors.New("invalid image name!")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	ctx := r.Context()
	parentId := data.From
	childId := data.To
	now := unixTimeForNow()
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	// Get parent and child items
	var parent, child geneticSource
	parent, err = getGeneticItem(ctx, data.FromType, data.From)
	if err != nil {
		http.Error(w, "failed to get parent item: "+err.Error(), http.StatusBadRequest)
		return
	}
	child, err = getGeneticItem(ctx, data.ToType, data.To)
	if err != nil {
		http.Error(w, "failed to get child item: "+err.Error(), http.StatusBadRequest)
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
		err = errors.New("you do not have permissions to create this transfer, you likely cannot modify the parent, or the child is not eligible to be transferred to")
		http.Error(w, err.Error(), http.StatusUnauthorized)
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
	_, err = newTxn(ctx, func(sessCtx mongo.SessionContext) (any, error) {
		_, err := db.Collection(TransfersCollName).InsertOne(ctx, xfer)
		if err != nil {
			return nil, err
		}

		//err = parent.setTransferParent(ctx, xfer) // TODO: Del?
		if err = setTransferParent(sessCtx, parent, xfer); err != nil { // TODO: should be session
			return nil, errors.Join(errors.New("failed to set transfer parent"), err)
		}

		if err = child.setTransferChild(sessCtx, xfer, parent); err != nil {
			return nil, errors.Join(errors.New("failed to set transfer child"), err)
		}
		return nil, nil
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
	//ctx, sess, err := createMongoSession(ctx)
	//if err != nil {
	//	http.Error(w, "failed to create db session for transfer: "+err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//defer sess.EndSession(ctx)
	//opts := []*options.TransactionOptions{} // TODO: FIXME
	//if err = sess.StartTransaction(opts...); err != nil {
	//	http.Error(w, "failed to start transfer transaction: "+err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//
	////WithTransaction(ctx, func(ctx mongo.SessionContext)(any, error){
	////	return nil, nil
	////}, opts...)
	//
	//_, err = db.Collection(TransfersCollName).InsertOne(ctx, xfer)
	//if err != nil {
	//	err = errors.Join(err, sess.AbortTransaction(ctx))
	//	http.Error(w, "failed to create transfer: "+err.Error(), http.StatusInternalServerError)
	//	return
	//}
	////// Set rollback
	////rollbackXfer := func() error {
	////	result, errrr := db.Collection(TransfersCollName).DeleteOne(ctx, bson.D{{"_id", xfer.Id}})
	////	if errrr != nil {
	////		return errrr
	////	}
	////	if result.DeletedCount == 0 {
	////		return errors.New("transfer not deleted")
	////	}
	////	return nil
	////}
	//err = setTransferParent(ctx, parent, xfer)
	////err = parent.setTransferParent(ctx, xfer) // TODO: Del?
	//if err != nil {
	//	err = errors.Join(errors.New("failed to set transfer parent"), sess.AbortTransaction(ctx), err)
	//	//err = errors.Join(errors.New("failed to set transfer parent"), rollbackXfer(), err)
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//err = child.setTransferChild(ctx, xfer, parent)
	//if err != nil {
	//	err = errors.Join(err, sess.AbortTransaction(ctx))
	//	//parentRollbackErr := rollbackParent()
	//	//if parentRollbackErr != nil {
	//	//	// TODO: HUGE ERROR. FIGURE OUT
	//	//}
	//	//xferRollbackErr := rollbackXfer()
	//	//if xferRollbackErr != nil {
	//	//	// TODO: HUGE ERROR. FIGURE OUT
	//	//}
	//	//err = errors.Join(errors.New("failed to set transfer child"), parentRollbackErr, xferRollbackErr, err)
	//	http.Error(w, "failed to set transfer child: "+err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//err = sess.CommitTransaction(ctx)
	//if err != nil {
	//	http.Error(w, "failed to commit transaction for transfer creation: "+err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//
	//bsOut, errMarshalling := json.Marshal(xfer)
	//if errMarshalling != nil { // Not err because err != nil at end will delete all images
	//	// Do not rollback here. Data made it in successfully
	//	http.Error(w, errMarshalling.Error(), http.StatusInternalServerError)
	//	return
	//}
	//
	//_, errWriting := w.Write(bsOut)
	//if errWriting != nil {
	//	// Do not rollback here. Data made it in successfully
	//	handleWriteErr(errWriting, w)
	//}
}

type updateTransferRequest struct {
	Notes          AllEntries[Note] `json:"notes,omitempty"`
	PermsOnRequest                  // TODO: ????????? handle in typescript and handler!
}

func (mods updateTransferRequest) modsFor(existing *Transfer, aclField AclField) (bson.D, error) {
	return NewMods().
		updateNotesIfNeeded(mods.Notes, existing.Notes). // TODO: make sure this works the way we want
		updatePermsIfNeeded(aclField.ACL, existing.ACL).
		updateLastUpdatedIfNeeded().
		Finalized()
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
