package api

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"github.com/reeceappling/goUtils/v2/logging"
	"github.com/reeceappling/goUtils/v2/utils"
	"github.com/reeceappling/mushDb/api/env"
	"github.com/reeceappling/mushDb/api/pics"
	"github.com/reeceappling/pi-pn532-i2c-Ntag21x-ws/v2/websocketSessions"
	"github.com/reeceappling/pi-pn532-i2c-Ntag21x-ws/v2/websocketSessions/shared"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
)

func randomRFID(bytes int) []byte {
	b := make([]byte, bytes)
	_, _ = rand.Read(b)
	return b
}

func writeRfidTagIfNecessary(ctx context.Context, writeTagTo *string, id MainCollectionId) error {
	if writeTagTo == nil {
		return nil // Don't write
	}
	err := GetService().WriteRfid(ctx, shared.RfidReaderName(*writeTagTo), id)
	if err != nil {
		logging.GetSugaredLogger(ctx).Errorw("failed to write tag", "error", err.Error())
		return err
	}
	return nil
}

var globalRfidReadWriter ReadWriteService = readWriteSvc{}

func SetService(rw ReadWriteService) { // TODO: USE THIS IN TESTS!
	globalRfidReadWriter = rw
}
func GetService() ReadWriteService { // TODO: USE THIS IN TESTS!
	return globalRfidReadWriter
}

//TODO: for other mocks? mockery:structname: ReadWriteService

//mockery:generate: true
type ReadWriteService interface {
	WriteRfid(ctx context.Context, readerName shared.RfidReaderName, id MainCollectionId) error
	ReadRfid(ctx context.Context, readerName shared.RfidReaderName) ([8]byte, error)
}
type readWriteSvc struct{} // TODO: TEST THIS!

func (rw readWriteSvc) WriteRfid(ctx context.Context, readerName shared.RfidReaderName, id MainCollectionId) (err error) {
	mgr := websocketSessions.GetSessionManager(ctx)
	if mgr == nil {
		err = websocketSessions.ErrNoSessionManager
		logging.GetSugaredLogger(ctx).Error("no session manager found")
		return
	}
	return mgr.WriteRfid(ctx, readerName, id) // TODO: handle manager nil check within this!
}
func (rw readWriteSvc) ReadRfid(ctx context.Context, readerName shared.RfidReaderName) (out [8]byte, err error) {
	mgr := websocketSessions.GetSessionManager(ctx)
	if mgr == nil {
		err = websocketSessions.ErrNoSessionManager
		logging.GetSugaredLogger(ctx).Error("no session manager found")
		return
	}
	return mgr.ReadRfid(ctx, readerName) // TODO: handle manager nil check within this!
}

type MockRfidSvc struct {
	writeResults map[shared.RfidReaderName]map[string]error
	readResults  map[shared.RfidReaderName]mockRfidReadResult
}

type mockRfidReadResult struct {
	Out [8]byte
	Err error
}

func NewMockRfidService() *MockRfidSvc { // TODO: USE IN TESTING!
	return &MockRfidSvc{}
}

func (rw *MockRfidSvc) WithTestWrite(readerName shared.RfidReaderName, id MainCollectionId, result error) {
	if rw.writeResults == nil {
		rw.writeResults = map[shared.RfidReaderName]map[string]error{}
	}
	if _, exists := rw.writeResults[readerName]; !exists {
		rw.writeResults[readerName] = make(map[string]error)
	}
	rw.writeResults[readerName][string(id[:])] = result
}

func (rw *MockRfidSvc) WithTestRead(readerName shared.RfidReaderName, result MainCollectionId, err error) {
	if rw.readResults == nil {
		rw.readResults = map[shared.RfidReaderName]mockRfidReadResult{}
	}
	if _, exists := rw.readResults[readerName]; !exists {
		rw.writeResults[readerName] = make(map[string]error)
	}
	rw.readResults[readerName] = mockRfidReadResult{
		Out: result,
		Err: err,
	}
}
func (rw *MockRfidSvc) WriteRfid(ctx context.Context, readerName shared.RfidReaderName, id MainCollectionId) error {
	writerResults, ok := rw.writeResults[readerName]
	if !ok {
		return errors.New("writer not found")
	}
	finalErr, ok := writerResults[string(id[:])]
	if !ok {
		return errors.New("id not found for mock writer")
	}
	return finalErr
}
func (rw *MockRfidSvc) ReadRfid(ctx context.Context, readerName shared.RfidReaderName) ([8]byte, error) {
	results, ok := rw.readResults[readerName]
	if !ok {
		return [8]byte{}, errors.New("writer not found")
	}
	return results.Out, results.Err
}

func StandardizeMainCollectionId(id string) (*MainCollectionId, error) {
	if id == "1" { // TODO: DO THIS ELSEWHERE!
		println("making ID 1!")
		return utils.Pointer(MainCollectionId([]byte{0, 0, 0, 0, 0, 0, 0, 0})), nil // TODO: not sure we actually want this....
	}
	realId, err := Base58Str(id).ToMainCollectionId()
	if err != nil {
		return nil, err
	}
	return &realId, nil
}

func StandardizeAltCollectionId(id string) (*AlternateCollectionId, error) {
	realId, err := Base58Str(id).toAltCollectionId()
	if err != nil {
		return nil, err
	}
	return &realId, nil
}

// Perms have not been checked yet
func GetMainCollectionItem[T MainCollectionItem](ctx context.Context, id MainCollectionId, resultItemType T) (out MainCollectionItem, err error) {
	println("reading mcitem from " + resultItemType.CollectionName())
	encodedResult := DbFrom(ctx).Collection(resultItemType.CollectionName()).FindOne(ctx, BsonFindFilter(IDfld, id))
	if encodedResult.Err() != nil {
		return resultItemType, encodedResult.Err() // mongo.ErrNoDocuments if 404
	}
	temp, err := resultItemType.Decode(encodedResult)
	if err != nil {
		err = errors.Join(errors.New("failed to decode"), err)
		return resultItemType, err
	}
	item, ok := temp.(MainCollectionItem)
	if !ok {
		err = errors.New("failed to decode. Item was not a mainCollection item")
		return nil, err
	}
	return item, nil
}

// Perms have not been checked yet // TODO: validate works
func GetMainCollectionItemSpecific[T MainCollectionItem](ctx context.Context, id MainCollectionId, resultItemType T) (out T, err error) {
	println("B reading mcitem from " + resultItemType.CollectionName())
	encodedResult := DbFrom(ctx).Collection(resultItemType.CollectionName()).FindOne(ctx, BsonFindFilter(IDfld, id))
	if encodedResult.Err() != nil {
		return resultItemType, encodedResult.Err() // mongo.ErrNoDocuments if 404
	}
	temp, err := resultItemType.Decode(encodedResult)
	if err != nil {
		err = errors.Join(errors.New("failed to decode"), err)
		return resultItemType, err
	}
	item, ok := temp.(T)
	if !ok {
		err = errors.New("failed to decode. Item was not a mainCollection item")
		return item, err
	}
	return item, nil
}

func GetAltCollectionItem[T AltCollectionItem[U], U AltCollectionIdType](ctx context.Context, id U, item T) (out T, err error) {
	err = DbFrom(ctx).Collection(item.CollectionName()).
		FindOne(ctx, BsonFindFilter(IDfld, id)).Decode(item)
	return item, err
}
func GetRecipeWithName[T AltCollectionItem[AlternateCollectionId]](ctx context.Context, name string, item T) (out T, err error) {
	err = DbFrom(ctx).Collection(item.CollectionName()).
		FindOne(ctx, BsonFindFilter("name", name)).Decode(item) // TODO: consider doing this as a FindMany since names could theoretically overlap
	return item, err
}

func GetAltCollectionItemOutsideTxn[T AltCollectionItem[U], U AltCollectionIdType](ctx context.Context, id AlternateCollectionId, item T) (out T, err error) {
	out = item
	encodedResult := DbFrom(ctx).
		Collection(item.CollectionName()).
		FindOne(ctx, BsonFindFilter(IDfld, id))
	if encodedResult.Err() != nil {
		return out, encodedResult.Err() // mongo.ErrNoDocuments if 404
	}
	err = encodedResult.Decode(&out)
	if err != nil {
		return out, err
	}
	return out, nil
}

func GetSpeciesNameInTxn(ctx context.Context, name string) (out Species, err error) { // TODO: make sure this works as intended!
	out = Species{}
	// TODO: validate that DbFrom properly uses txn when needed
	encodedResult := DbFrom(ctx).
		Collection(SpeciesCollectionName).
		FindOne(ctx, BsonFindFilter(IDfld, name))
	if encodedResult.Err() != nil {
		return out, encodedResult.Err() // mongo.ErrNoDocuments if 404
	}
	err = encodedResult.Decode(&out)
	if err != nil {
		return out, err
	}
	return out, nil
}

func GetSubspeciesByNameInTxn(ctx context.Context, name string) (out Subspecies, err error) { // TODO: make sure this works as intended!
	out = Subspecies{}

	// TODO: validate that DbFrom properly uses txn when needed
	encodedResult := DbFrom(ctx).
		Collection(SubspeciesCollectionName).
		FindOne(ctx, BsonFindFilter(IDfld, name))
	if encodedResult.Err() != nil {
		return out, encodedResult.Err() // mongo.ErrNoDocuments if 404
	}
	err = encodedResult.Decode(&out)
	if err != nil {
		return out, err
	}
	return out, nil
}

func multipartReaderForRequest[T any](r *http.Request, w http.ResponseWriter, result *T) (reader *multipart.Reader, err error) {
	//r.Body = http.MaxBytesReader(w, r.Body, maxMultipartRequestSize) // TODO: do we need this?
	reader, err = r.MultipartReader()
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
		http.Error(w, "failed to read Data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	// PARSE INTO CORRECT DATA FORMAT
	err = json.Unmarshal(bs, result)
	if err != nil {
		http.Error(w, "failed to unmarshal Data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	return
}

func getMultipartImages(ctx context.Context, prefixPath string, w http.ResponseWriter, reader *multipart.Reader, b58id Base58Str) (newPics, newContams, newFlushes map[int]string, err error) {
	// Get any images
	newPics, newContams, newFlushes = map[int]string{}, map[int]string{}, map[int]string{}
	picsSaved := []string{}
	defer func() {
		if err != nil {
			if errDel := pics.DeleteFiles(ctx, picsSaved...); errDel != nil {
				handleFileDeleteErr(errDel)
			}
		}
	}()
	var p *multipart.Part
	for {
		// Go to next part or break
		p, err = reader.NextPart()
		if err != nil {
			if err != io.EOF {
				http.Error(w, "non-eof err!"+err.Error(), http.StatusInternalServerError)
				return
			}
			err = nil // Ensure error is nil so it does not get returned
			break
		}
		fileName := p.FileName()
		if fileName == "" {
			err = errors.New("file name is empty for what should have been an image")
			http.Error(w, "file name is empty for what should have been an image", http.StatusBadRequest)
			return
		}
		// Process file
		parts := strings.Split(fileName, "-")
		if len(parts) != 2 {
			err = errors.New("invalid image name: " + fileName)
			http.Error(w, "invalid image name: "+fileName, http.StatusBadRequest)
			return
		}
		num, errr := strconv.Atoi(parts[1])
		if errr != nil {
			err = errr
			http.Error(w, "failed to parse image number! "+errr.Error(), http.StatusBadRequest)
			return
		}
		fieldBytes, errr := multipartToImageBytes(p, w)
		if errr != nil {
			err = errr
			// Already wrote in the above func
			return
		}
		env.LogAlways("checking parts") // TODO: del?
		switch parts[0] {
		case "newPic":
			newFileNameWithPrefixPath, errr := pics.SaveFile(ctx, fieldBytes, prefixPath, string(b58id), "img")
			if errr != nil {
				err = errr
				http.Error(w, "failed to save new picture: "+err.Error(), http.StatusBadRequest)
				return
			}
			picsSaved = append(picsSaved, newFileNameWithPrefixPath)
			newPics[num] = newFileNameWithPrefixPath
		case "newContam":
			newFileNameWithPrefixPath, errr := pics.SaveFile(ctx, fieldBytes, prefixPath, string(b58id), "contam")
			if errr != nil {
				err = errr
				http.Error(w, "failed to save new contamination: "+err.Error(), http.StatusBadRequest)
				return
			}
			picsSaved = append(picsSaved, newFileNameWithPrefixPath)
			newContams[num] = newFileNameWithPrefixPath
		case "newFlush":
			newFileNameWithPrefixPath, errr := pics.SaveFile(ctx, fieldBytes, prefixPath, string(b58id), "flush")
			if errr != nil {
				err = errr
				http.Error(w, "failed to save new flush: "+err.Error(), http.StatusBadRequest)
				return
			}
			picsSaved = append(picsSaved, newFileNameWithPrefixPath)
			newFlushes[num] = newFileNameWithPrefixPath
		default:
			err = errors.New("invalid picture name. Should never occur")
			println(err.Error() + " " + fileName)
			http.Error(w, "invalid picture name. Should never occur", http.StatusInternalServerError)
			return
		}
	}
	return
}

func fullMultipartWithNoBreaks[T any](w http.ResponseWriter, r *http.Request, data *T, b58id Base58Str) (newPics, newContams, newFlushes map[int]string, err error) {
	defer r.Body.Close() // TODO: this ok here?
	prefixPath := r.PathValue("variant")
	if prefixPath == "" {
		println("variant missing from path")
		http.Error(w, "variant missing from path", http.StatusBadRequest)
		return nil, nil, nil, errors.New("no prefix path 'variant' provided")
	}
	reader, err := multipartReaderForRequest(r, w, data) // TODO: consider swapping for multipartReaderInitialize
	if err != nil {
		return nil, nil, nil, err // Already wrote
	}
	return getMultipartImages(r.Context(), prefixPath, w, reader, b58id)
}

//// TODO: move
//func checkIdTypeWithRaw[T bson.M | bson.D](ctx context.Context, collection *mongo.Collection, filter T) {
//	var rawDoc bson.Raw
//	err := collection.FindOne(ctx, filter).Decode(&rawDoc)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Lookup looks up an element in the raw document by key
//	idElement, err := rawDoc.LookupErr(IDfld)
//	if err != nil {
//		fmt.Println("_id field does not exist in this document")
//		return
//	}
//
//	// idElement.Type returns a bsontype.Type (e.g., bsontype.ObjectID, bsontype.String)
//	fmt.Printf("Exact BSON Type: %s (Hex byte value: %x)\n", idElement.Type, idElement.Type)
//}
//
//// TODO: move
//func checkIdTypeWithRawOnCursor(cursor *mongo.Cursor) error {
//	var rawDoc bson.Raw
//	err := cursor.Decode(&rawDoc)
//	if err != nil {
//		println("failed to decode document from cursor: " + err.Error())
//		return errors.Join(err, errors.New("failed to decode document from cursor"))
//	}
//
//	// Lookup looks up an element in the raw document by key
//	idElement, err := rawDoc.LookupErr(IDfld)
//	if err != nil {
//		println("_id field does not exist in this document: " + err.Error())
//		return errors.Join(err, errors.New("_id field does not exist in this document"))
//	}
//
//	// idElement.Type returns a bsontype.Type (e.g., bsontype.ObjectID, bsontype.String)
//
//	println("item: ", rawDoc.String())
//	println("id", idElement.String(), "value", string(idElement.Value))
//	//idElType := idElement.Type
//	//println(fmt.Sprintf("Exact BSON Type: %s (Hex byte value: %x)\n", idElType, idElType))
//	//println(fmt.Sprintf("As string: %s. Value: %s\n", idElement.String(), string(idElement.Value)))
//	return nil
//}

//func TimeFromId(id AlternateCollectionId) time.Time { // TODO: USE AND MOVE
//	return primitive.ObjectID(id).Timestamp()
//}
//
//func Ternary[T any](val bool, ifTrue, ifFalse T) T {
//	if val {
//		return ifTrue
//	}
//	return ifFalse
//}
