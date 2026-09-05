package pics

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/reeceappling/mushDb/api/cache"
	"os"
	"path/filepath"
)

var lruCache *cache.LRU

func init() {
	lruCache = cache.NewLRU(maxImagesInCache)
}

const maxImagesInCache = 10 // TODO: 10 ok?
const filePathCtxKey = "dbImageFilePath"

func SetFilePath(ctx context.Context, filePath string) context.Context {
	return context.WithValue(ctx, filePathCtxKey, filePath)
}

func GetFilePath(ctx context.Context) string {
	return ctx.Value(filePathCtxKey).(string)
}

func SaveFile(ctx context.Context, bs []byte, prefixPath ...string) (string, error) {
	// TODO: save file to s3 if needed (probably not)
	filePath := GetFilePath(ctx)
	if filePath == "" {
		filePath = "/images" // TODO: FIXME!!!!
	}
	resolvedPrefix := ""
	for _, prefix := range prefixPath {
		resolvedPrefix = resolvedPrefix + prefix + "/"
	}
	for i := 0; i < 10; i++ { // TODO: max iterations? jitter?
		name, err := uuid.NewRandom()
		if err != nil {
			return "", err
		}
		fileNameWithPrefixPath := resolvedPrefix + name.String()
		whereToWrite := filePath + "/" + fileNameWithPrefixPath
		if err = os.MkdirAll(filePath+"/"+resolvedPrefix, 777); err != nil { // TODO: 666 instead of 777?
			fmt.Printf("Error creating directory: %s\n", err)
			return fileNameWithPrefixPath, err
		}
		if _, err = os.Stat(whereToWrite); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				err = os.WriteFile(whereToWrite, bs, 777) // TODO: 666 instead of 777?
				if err != nil {
					println("failed to write file", err.Error())
				}
				lruCache.Add(whereToWrite, bs) // TODO: ensure this is ok for the LRU cache!
				return fileNameWithPrefixPath, err
			}
			println("file exists already!", err.Error())
			return "", err // TODO: PROBABLY CONTINUE INSTEAD OF RETURN HERE
		} else {
			println("file exists already!", err.Error()) // TODO: continue???
			return "", err                               // TODO: PROBABLY CONTINUE INSTEAD OF RETURN HERE
		}
	}
	return "", errors.New("failed to find a new fileName")
}

func GetFile(ctx context.Context, imgSubPath string) (bytes []byte, err error) {
	path := filepath.Join(GetFilePath(ctx), imgSubPath)
	// Try to get from LRU cache first
	if bs, found := lruCache.Get(path); found {
		return bs, nil
	}
	// Cache miss
	return os.ReadFile(path)

	//picsDir := GetFilePath(ctx)
	//bytes, err = os.ReadFile(filepath.Join(picsDir, imgSubPath))
	//if err != nil {
	//	if errors.Is(err, os.ErrNotExist) {
	//		err = errorreference.ErrorNotFound
	//		//// TODO: READ FROM s3 if needed?
	//		//bytes, err = s3.NewFileReader().Read(ctx, imgSubPath) // TODO: ensure ok
	//		//if err != nil {
	//		//	if errors.Is(err, errorreference.ErrorNotFound) {
	//		//		println("file does not exist locally or in s3!") // TODO: fix
	//		//		return nil, ErrNotFound
	//		//
	//		//	}
	//		//	return nil, err
	//		//}
	//	}
	//}
	//return bytes, err
}

func DeleteFiles(ctx context.Context, filenamesWithPrefixPaths ...string) error {
	var errOut error = nil
	homeDir := GetFilePath(ctx)
	for _, filenameWithPrefixPath := range filenamesWithPrefixPaths {
		path := homeDir + filenameWithPrefixPath
		err := os.Remove(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			errOut = errors.Join(errOut, err)
		}
		_ = lruCache.Evict(path) // TODO: ok? do we even need this?
	}
	return errOut
}
