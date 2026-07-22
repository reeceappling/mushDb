package api

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"github.com/reeceappling/goUtils/v2/utils"
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
	mgr := websocketSessions.GetSessionManager(ctx)
	if mgr == nil {
		println("no session mgr found?") // TODO: del
		return websocketSessions.ErrNoSessionManager
	}
	err := mgr.WriteRfid(ctx, shared.RfidReaderName(*writeTagTo), id)
	if err != nil {
		println("failed to write tag! " + err.Error())
		return err
	}
	return nil
}

func StandardizeMainCollectionId(id string) (*MainCollectionId, error) {
	if id == "1" { // TODO: DO THIS ELSEWHERE!
		println("making ID 1!")
		return utils.Pointer(MainCollectionId([]byte{0, 0, 0, 0, 0, 0, 0, 0})), nil // TODO: not sure we actually want this....
	}
	//println("ID BYTES NOT LENGTH 8! CONVERTING!") // TODO: del
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
	//r.Body = http.MaxBytesReader(w, r.Body, maxMultipartRequestSize)
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
	newPics = map[int]string{}
	newContams = map[int]string{}
	newFlushes = map[int]string{}
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
		println("checking parts")
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
	defer r.Body.Close()
	prefixPath := r.PathValue("variant")
	if prefixPath == "" {
		return nil, nil, nil, errors.New("no prefix path 'variant' provided")
	}
	reader, err := multipartReaderForRequest(r, w, data)
	if err != nil {
		// Already wrote
		return nil, nil, nil, err
	}
	return getMultipartImages(r.Context(), prefixPath, w, reader, b58id)
}
