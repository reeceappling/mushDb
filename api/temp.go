package api

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"github.com/reeceappling/goUtils/v2/utils"
	"github.com/reeceappling/mushDb/api/pics"
	"go.mongodb.org/mongo-driver/mongo"
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
	// TODO: DO THIS!!!!! IDEAS IN test/main
	return errors.New("not yet implemented") // TODO: write tag
}

func StandardizeMainCollectionId(id string) (*MainCollectionId, error) {
	println(id)    // TODO: delete
	if id == "1" { // TODO: DO THIS ELSEWHERE!
		println("making ID 1!")
		return utils.Pointer(MainCollectionId([]byte{0, 0, 0, 0, 0, 0, 0, 0})), nil // TODO: not sure we actually want this....
	}
	//var out MainCollectionId
	//idBytes := []byte(id)
	//if len(idBytes) == 8 { // TODO: unsure if base58 bytes should ever be len 8 in this case... Will likely cause bugs since we shouldnt be expecting base2 to come in from anywhere except the db...
	//	out = MainCollectionId(idBytes)
	//	return &out, nil
	//}
	println("ID BYTES NOT LENGTH 8! CONVERTING!")
	realId, err := Base58Str(id).toMainCollectionId()
	if err != nil {
		return nil, err
	}
	println("CONVERTED", id, " TO", string(realId[:]))
	return &realId, nil
}

func StandardizeAltCollectionId(id string) (*AlternateCollectionId, error) {
	//idBytes := []byte(id)
	//var out AlternateCollectionId
	//if len(idBytes) == 12 { // TODO: unsure if base58 bytes should ever be len 12 in this case...
	//	out = [12]byte(idBytes)
	//	return &out, nil
	//}
	realId, err := Base58Str(id).toAltCollectionId()
	if err != nil {
		return nil, err
	}
	println("CONVERTED", id, " TO", string(realId[:]))
	return &realId, nil
}

// Perms have not been checked yet
func GetMainCollectionItem[T MainCollectionItem](ctx context.Context, id MainCollectionId, resultItemType T) (out MainCollectionItem, err error) {
	println("reading mcitem from " + resultItemType.CollectionName())
	encodedResult := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(resultItemType.CollectionName()).FindOne(ctx, BsonFindFilter("_id", id))
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

//func GetCollectionItemInTxn(ctx context.Context, id MainCollectionId, sourceType string) (out MainCollectionItem, err error) {
//	db := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName)
//	out, err = typeForSource(sourceType) // TODO: this should be sourceType instead
//	if err != nil {
//		return out, err
//	}
//	err = db.Collection(out.CollectionName()).FindOne(ctx, BsonFindFilter("_id", id)).Decode(&out)
//	if err != nil {
//		return nil, err // mongo.ErrNoDocuments if 404
//	}
//	// TODO: auth info????
//	//authinfo, err := GetAuthInfo(ctx)
//	//if err != nil {
//	//	return nil, err
//	//}
//	//if out.Permissions().PermissionFor(authinfo) == perms.None {
//	//	err = errors.New("no perms on getMainCollItemInTxn")
//	//}
//	return
//}

func GetAltCollectionItem[T AltCollectionItem[U], U AltCollectionIdType](ctx context.Context, id U, item T) (out T, err error) {
	err = ctx.Value(mongoClientContextKey).(*mongo.Client).
		Database(dbName).Collection(item.CollectionName()).
		FindOne(ctx, BsonFindFilter("_id", id)).Decode(item)
	return item, err
}

// TODO: used to be in txn!
func GetAltCollectionItemOutsideTxn[T AltCollectionItem[U], U AltCollectionIdType](ctx context.Context, id AlternateCollectionId, item T) (out T, err error) {
	out = item
	encodedResult := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).
		Collection(item.CollectionName()).
		FindOne(ctx, BsonFindFilter("_id", id))
	if encodedResult.Err() != nil {
		return out, encodedResult.Err() // mongo.ErrNoDocuments if 404
	}
	err = encodedResult.Decode(&out)
	if err != nil {
		return out, err
	}
	//authInfo, err := GetAuthInfo(ctx)
	//if err != nil {
	//	return out, err
	//}
	//if reflect.TypeOf(out).Implements(reflect.TypeOf((*Permissioned)(nil)).Elem()) {
	//	temp, ok := interface{}(out).(Permissioned)
	//	if !ok {
	//		return out, errors.New("this should never happen, but a thing implements a thing but does not implement the thing")
	//	}
	//
	//	if temp.Permissions().PermissionFor(authInfo) == perms.None {
	//		return out, errors.New("no permission")
	//	}
	//}
	return out, nil
}

func GetSpeciesNameInTxn(ctx context.Context, name string) (out Species, err error) { // TODO: make sure this works as intended!
	out = Species{}
	encodedResult := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).
		Collection(SpeciesCollectionName).
		FindOne(ctx, BsonFindFilter("_id", name))
	if encodedResult.Err() != nil {
		return out, encodedResult.Err() // mongo.ErrNoDocuments if 404
	}
	err = encodedResult.Decode(&out)
	if err != nil {
		return out, err
	}
	//authInfo, err := GetAuthInfo(ctx)
	//if err != nil {
	//	return out, err
	//}
	//out.Permissions()
	//if reflect.TypeOf(out).Implements(reflect.TypeOf((*Permissioned)(nil)).Elem()) {
	//	temp, ok := interface{}(out).(Permissioned)
	//	if !ok {
	//		return out, errors.New("this should never happen, but a thing implements a thing but does not implement the thing")
	//	}
	//
	//	if temp.Permissions().PermissionFor(authInfo) == perms.None {
	//		return out, errors.New("no permission")
	//	} // TODO: pull this out
	//}
	return out, nil
}

func GetSubspeciesByNameInTxn(ctx context.Context, name string) (out Subspecies, err error) { // TODO: make sure this works as intended!
	out = Subspecies{}
	encodedResult := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).
		Collection(SubspeciesCollectionName).
		FindOne(ctx, BsonFindFilter("_id", name))
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
	//println("before", string(bs)) // TODO: del
	// PARSE INTO CORRECT DATA FORMAT
	err = json.Unmarshal(bs, result)
	if err != nil {
		http.Error(w, "failed to unmarshal Data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	//bs2, err := json.Marshal(result)
	//if err != nil {
	//	http.Error(w, "failed to marshal Data from form: "+err.Error(), http.StatusBadRequest)
	//	return // TODO: del
	//}
	//println("after", string(bs2)) // TODO: del
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
		//println("getting field bytes", "Form: "+p.FormName(), "File: "+p.FileName()) // TODO: THIS
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
				println("failed to save a new flush picture", errr.Error()) // TODO: del
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

	// TODO:?????
	//// CHECK THAT ALL NEW PICS EXIST
	//// PROCESS ALL NEW PICS AND CONTAMS
	//out := Data.reform()
	//for i, _ := range Data.Images.New {
	//	loc, exists := newPics[i]
	//	if !exists {
	//		http.Error(w, fmt.Sprintf("error, location for new picture index %d not found (should never happen)", i), http.StatusInternalServerError)
	//		return
	//	}
	//	out.Images.New[i].Location = imageLocation(loc)
	//}
	//for i, _ := range Data.Contams.New {
	//	if loc, exists := newContams[i]; exists {
	//		finalLoc := imageLocation(loc)
	//		out.Contams.New[i].Location = &finalLoc
	//	}
	//}
	//for i, _ := range Data.Flushes.New {
	//	loc, exists := newFlushes[i]
	//	if !exists {
	//		http.Error(w, fmt.Sprintf("error, location for new flush index %d not found (should never happen)", i), http.StatusInternalServerError)
	//		return
	//	}
	//	out.Flushes.New[i].Location = imageLocation(loc)
	//}
	return
}

// TODO: rename
// TODO: only use when writeRFID is not between the two (on updates)
func fullMultipartWithNoBreaks[T any](w http.ResponseWriter, r *http.Request, prefixPath string, data *T, b58id Base58Str) (newPics, newContams, newFlushes map[int]string, err error) { // TODO USE THIS ALL OVER THE PLACE
	defer r.Body.Close()
	reader, err := multipartReaderForRequest(r, w, data)
	if err != nil {
		// Already wrote
		return nil, nil, nil, err
	}
	return getMultipartImages(r.Context(), prefixPath, w, reader, b58id)
}
