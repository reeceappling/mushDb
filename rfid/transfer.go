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
	From                       MainCollectionId `bson:"from" json:"from"` // TODO: THIS USED TO NOT BE A SLICE
	To                         MainCollectionId `bson:"to" json:"to"`     // fruit is mainCollectionId
	FromType                   string           `json:"fromType"`         //sourceType     //fruit, sporePrint, mss, plate, jar, stasis, lc, slant, bag, box
	ToType                     string           `json:"toType"`           //sourceType
	CreationDateField          `bson:"inline"`
	Reason                     transferReason `bson:"reason" json:"reason"`
	FromImage                  *ImageLocation `bson:"fromImage,omitempty" json:"fromImage,omitempty"`
	ToImage                    *ImageLocation `bson:"toImage,omitempty" json:"toImage,omitempty"`
	NotesField                 `bson:"inline"`
	LastUpdatedField           `bson:"inline"`
	AclField                   `bson:"inline"`
}

func (t Transfer) PicsModsForChild() *Mods {
	if t.ToImage == nil {
		return NewMods()
	}
	pic := newPicWithNotes(t.CreationDate, []Note{}, *t.ToImage)
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
		From:                       exPlate,
		To:                         exJar,
		FromType:                   "plate",
		ToType:                     "jar",
		CreationDateField:          CreationDateField{exampleTime},
		Reason:                     "A_REASONABLE_TRANSFER_REASON",
		FromImage:                  (*ImageLocation)(&exPicLoc),
		ToImage:                    (*ImageLocation)(&exPicLoc),
		NotesField:                 NotesField{exampleNotes()},
		LastUpdatedField:           LastUpdatedField{exampleTime},
		AclField:                   allCanReadAcl(),
	}
	return addTestAltEntries(ctx, testItem)
}

type createTransferRequest struct {
	From     MainCollectionId `json:"from"` // TODO: check all other requests that accept bytes, will likely need to be this
	To       MainCollectionId `json:"to"`
	FromType *string          `json:"fromType,omitempty"`
	Reason   string           `json:"reason"`
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
	b58id := id.AsBase58()
	r.Body = http.MaxBytesReader(w, r.Body, maxMultipartRequestSize)
	defer r.Body.Close()
	reader, err := r.MultipartReader() // TODO: do streamlined
	if err != nil {
		http.Error(w, "unable to open multipart reader: "+err.Error(), http.StatusBadRequest)
		return
	}
	println("opened up multipart reader successfully")
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
		println("failed to unmarshal Data from form: " + err.Error()) // TODO: THIS!
		http.Error(w, "failed to unmarshal Data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	tempBs, err := json.MarshalIndent(data, "", " ") // TODO: del
	if err != nil {
		println("failed to re-marshal data: " + err.Error()) // TODO: THIS!
		http.Error(w, "failed to re-marshal data: "+err.Error(), http.StatusBadRequest)
		return
	}
	println(string(tempBs))
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
	ctx := r.Context()
	parentId := data.From
	childId := data.To
	now := unixTimeForNow()
	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
	childMapEntry := idMapEntry{}
	var fromType string
	if data.FromType != nil {
		fromType = *data.FromType
	} else {
		fromType, err = getEntryTypeForId(ctx, parentId)
	}
	// TODO: if fromType is nil, then we must go find the fromType
	// Get parent and child items
	err = db.Collection(idMapCollectionName).FindOne(ctx, bson.M{"_id": childId}).Decode(&childMapEntry)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			println("child not found in id db") // TODO: THIS!
			http.Error(w, "child not found in id db: "+err.Error(), http.StatusNotFound)
			return
		}
		println("error getting child from id db") // TODO: THIS!
		http.Error(w, "error getting child from id db: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var parent, child geneticSource

	parent, err = getGeneticItem(ctx, fromType, data.From)
	if err != nil {
		println("error getting parent item "+data.From.AsBase58(), err.Error()) // TODO: THIS!
		http.Error(w, "failed to get parent item: "+err.Error(), http.StatusBadRequest)
		return
	}
	child, err = getGeneticItem(ctx, childMapEntry.EntryType, data.To)
	if err != nil {
		println("error getting child item") // TODO: THIS!
		http.Error(w, "failed to get child item: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err = parent.CanTransferTo(child); err != nil {
		println("parent cannot transfer to child") // TODO: THIS!
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resolvedPerms, err := GetResolvedUserPerms(ctx)
	if err != nil {
		println("failed to get user perms") // TODO: THIS!
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	// to make a transfer, the email must only be able to write to the parent initially // TODO: is this ok?
	if userParentPerm := parent.Permissions().HighestPermFor(resolvedPerms); userParentPerm == nil || !(*userParentPerm) {
		err = errors.New("you do not have permissions to create this transfer, you likely cannot modify the parent, or the child is not eligible to be transferred to")
		println(err.Error()) // TODO: THIS!
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
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
		LastUpdatedField:           LastUpdatedFieldForNow(),
		AclField:                   AclField{ACL: parent.Permissions()}, //set child perms to the parent perms!
	}
	_, err = newTxn(ctx, func(sessCtx mongo.SessionContext) (any, error) {
		_, err := db.Collection(TransfersCollName).InsertOne(ctx, xfer)
		if err != nil {
			println("failed to insert xfer", err.Error()) // TODO: this
			return nil, err
		}

		if err = setTransferParent(sessCtx, parent, xfer); err != nil {
			println("failed to set xfer parent", err.Error()) // TODO: this
			return nil, errors.Join(errors.New("failed to set transfer parent"), err)
		}
		// TODO: set child perms to the parent perms!
		if err = child.setTransferChild(sessCtx, xfer, parent); err != nil {
			println("failed to set xfer child", err.Error()) // TODO: this
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

	println("successfully created xfer") // TODO: DEL
	_, errWriting := w.Write(bsOut)
	if errWriting != nil {
		// Do not rollback here. Data made it in successfully
		handleWriteErr(errWriting, w)
	}
}

type updateTransferRequest struct {
	NotesUpdateField
	PermsOnRequest // TODO: should transfers always keep parent or child perms? // TODO: ????????? handle in typescript and handler!
}

func (req updateTransferRequest) modsFor(existing *Transfer, aclField AclField) (bson.D, error) {
	return NewMods().
		updateNotesIfNeeded(req, existing). // TODO: make sure this works the way we want
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
