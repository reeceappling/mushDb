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

func SavePicFile(ctx context.Context, bs []byte, prefixPath ...string) (string, error) {
	// TODO: save file to s3 if needed (probably not)
	filePath := GetFilePath(ctx)
	if filePath == "" {
		filePath = "/images" // TODO: FIXME!!!!
	}
	resolvedPrefix := ""
	for _, prefix := range prefixPath {
		resolvedPrefix = resolvedPrefix + prefix + "/"
	}
	for range 10 { // TODO: max iterations? jitter?
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
				} else {
					lruCache.Add(whereToWrite, bs)
				}
				return fileNameWithPrefixPath, err
			}
			println("file exists already!", err.Error())
			return "", err // TODO: PROBABLY CONTINUE INSTEAD OF RETURN HERE
		} else {
			println("file exists already?", "failed to stat file") // TODO: continue???
			return "", errors.New("failed to stat file")           // TODO: PROBABLY CONTINUE INSTEAD OF RETURN HERE
		}
	}
	return "", errors.New("failed to find a new fileName")
}

func getFile(pathPrefix, path string) (bytes []byte, err error) {
	fullPath := filepath.Join(pathPrefix, path)
	// Try to get from LRU cache first
	if bs, found := lruCache.Get(fullPath); found {
		return bs, nil
	}
	// Cache miss
	bytes, err = os.ReadFile(fullPath)
	if err != nil {
		return bytes, err
	}
	_ = lruCache.Add(fullPath, bytes)
	return bytes, err
}
func GetStaticFile(path string) (bytes []byte, err error) {
	return getFile("/staticFiles", path)
}

func GetPic(ctx context.Context, imgSubPath string) (bytes []byte, err error) {
	return getFile(GetFilePath(ctx), imgSubPath)
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
