package rfid

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"github.com/reeceappling/mushDb/rfid/pics"
	"go.mongodb.org/mongo-driver/bson"
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
	var out MainCollectionId
	idBytes := []byte(id)
	if len(idBytes) == 8 {
		out = [8]byte(idBytes)
		return &out, nil
	}
	realId, err := Base58Str(idBytes).toMainCollectionId()
	if err != nil {
		return nil, err
	}
	return &realId, nil
}

func StandardizeAltCollectionId(id string) (*AlternateCollectionId, error) {
	var out AlternateCollectionId
	idBytes := []byte(id)
	if len(idBytes) == 12 {
		out = [12]byte(idBytes)
		return &out, nil
	}
	realId, err := Base58Str(idBytes).toAltCollectionId()
	if err != nil {
		return nil, err
	}
	return &realId, nil
}

// Perms have not been checked yet
func GetMainCollectionItem[T MainCollectionItem](ctx context.Context, id MainCollectionId, resultItemType T) (out MainCollectionItem, err error) {
	encodedResult := ctx.Value(mongoClientContextKey).(*mongo.Client).Database(dbName).Collection(mainCollectionName).FindOne(ctx, bson.D{{"_id", id}})
	if encodedResult.Err() != nil {
		return resultItemType, encodedResult.Err() // mongo.ErrNoDocuments if 404
	}
	temp, err := resultItemType.Decode(encodedResult)
	if err != nil {
		err = errors.Join(errors.New("failed to decode"), err)
		return resultItemType, err
	}
	item := temp.(MainCollectionItem)
	return item, nil
}

func GetMainCollectionItemInTxn(ctx mongo.SessionContext, id MainCollectionId, optionalItemForType *MainCollectionItem) (out MainCollectionItem, err error) {
	encodedResult := ctx.Client().Database(dbName).Collection(mainCollectionName).FindOne(ctx, bson.D{{"_id", id}})
	if encodedResult.Err() != nil {
		return nil, encodedResult.Err() // mongo.ErrNoDocuments if 404
	}
	if optionalItemForType != nil {
		out = *optionalItemForType
	} else {
		raw := bson.Raw{}
		raw, err = encodedResult.Raw()
		if err != nil {
			return nil, err
		}
		out, err = rawEntryTypeConversion(raw)
		if err != nil {
			return nil, err
		}
	}
	err = encodedResult.Decode(&out)
	//authinfo, err := GetAuthInfo(ctx)
	//if err != nil {
	//	return nil, err
	//}
	//if out.Permissions().PermissionFor(authinfo) == perms.None {
	//	err = errors.New("no perms on getMainCollItemInTxn")
	//}
	return
}

func GetAltCollectionItem[T AltCollectionItem](ctx context.Context, id AlternateCollectionId, item T) (out T, err error) {
	out = item
	encodedResult := ctx.Value(mongoClientContextKey).(*mongo.Client).
		Database(dbName).
		Collection(item.CollectionName()).
		FindOne(ctx, bson.D{{"_id", id}})
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

func GetAltCollectionItemInTxn[T AltCollectionItem](ctx mongo.SessionContext, id AlternateCollectionId, item T) (out T, err error) {
	out = item
	encodedResult := ctx.Client().
		Database(dbName).
		Collection(item.CollectionName()).
		FindOne(ctx, bson.D{{"_id", id}})
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

func GetSpeciesNameInTxn(ctx mongo.SessionContext, name string) (out Species, err error) { // TODO: make sure this works as intended!
	out = Species{}
	encodedResult := ctx.Client().
		Database(dbName).
		Collection(speciesCollectionName).
		FindOne(ctx, bson.D{{"_id", name}})
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

func multipartReaderForRequest[T any](r *http.Request, w http.ResponseWriter, result *T) (reader *multipart.Reader, err error) {
	r.Body = http.MaxBytesReader(w, r.Body, maxMultipartRequestSize)
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
		http.Error(w, "failed to read data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	// PARSE INTO CORRECT DATA FORMAT
	err = json.Unmarshal(bs, result)
	if err != nil {
		http.Error(w, "failed to unmarshal data from form: "+err.Error(), http.StatusBadRequest)
		return
	}
	return
}

// TODO: rename
func getMultipartImages(ctx context.Context, prefixPath string, w http.ResponseWriter, reader *multipart.Reader, b58id Base58Str) (newPics, newContams, newFlushes map[int]string, err error) { // TODO USE THIS ALL OVER THE PLACE
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
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
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
		switch parts[0] {
		case "newPic":
			newFileNameWithPrefixPath, errr := pics.SaveFile(ctx, fieldBytes, prefixPath, string(b58id), "img")
			if err != nil {
				err = errr
				http.Error(w, "failed to save new picture: "+err.Error(), http.StatusBadRequest)
				return
			}
			picsSaved = append(picsSaved, newFileNameWithPrefixPath)
			newPics[num] = newFileNameWithPrefixPath
		case "newContam":
			newFileNameWithPrefixPath, errr := pics.SaveFile(ctx, fieldBytes, prefixPath, string(b58id), "contam")
			if err != nil {
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
			http.Error(w, "invalid picture name. Should never occur", http.StatusInternalServerError)
			return
		}
	}

	//// CHECK THAT ALL NEW PICS EXIST
	//// PROCESS ALL NEW PICS AND CONTAMS
	//out := data.reform()
	//for i, _ := range data.Images.New {
	//	loc, exists := newPics[i]
	//	if !exists {
	//		http.Error(w, fmt.Sprintf("error, location for new picture index %d not found (should never happen)", i), http.StatusInternalServerError)
	//		return
	//	}
	//	out.Images.New[i].Location = imageLocation(loc)
	//}
	//for i, _ := range data.Contams.New {
	//	if loc, exists := newContams[i]; exists {
	//		finalLoc := imageLocation(loc)
	//		out.Contams.New[i].Location = &finalLoc
	//	}
	//}
	//for i, _ := range data.Flushes.New {
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
// TODO: only use when writeRFID is not between the two
func fullMultipartWithNoBreaks[T any](w http.ResponseWriter, r *http.Request, prefixPath string, data *T, b58id Base58Str) (newPics, newContams, newFlushes map[int]string, err error) { // TODO USE THIS ALL OVER THE PLACE
	defer r.Body.Close()
	reader, err := multipartReaderForRequest(r, w, data)
	if err != nil {
		// Already wrote
		return nil, nil, nil, err
	}
	return getMultipartImages(r.Context(), prefixPath, w, reader, b58id)
}
