package pics

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/reeceappling/mushDb/api/cache"
	"os"
	"path/filepath"
	"strings"
)

var (
	lruCache, lruForCacheability *cache.LRU
)

func init() {
	lruCache = cache.NewLRU(maxImagesInCache)
	lruForCacheability = cache.NewLRU(cacheabilityhistory) // TODO: utilize?
}

const cacheabilityhistory = 20
const maxImagesInCache = 10 // TODO: 10 ok?
const filePathCtxKey = "dbImageFilePath"

func SetFilePath(ctx context.Context, filePath string) context.Context {
	return context.WithValue(ctx, filePathCtxKey, filePath)
}

func GetFilePath(ctx context.Context) string {
	return ctx.Value(filePathCtxKey).(string)
}
func AddCacheableEntry(ctx context.Context, entryType string, entryId string) { // TODO: use?
	key := entryType + "/" + entryId // TODO: ensure ok!
	_, found, _ := lruForCacheability.Get(key)
	if !found {
		lruForCacheability.Add(key, nil, true)
	}
}
func SavePicFile(ctx context.Context, bs []byte, externallyCacheable bool, prefixPath ...string) (string, error) {
	// TODO: needs to see whether public or not for external cacheability...
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
					lruCache.Add(whereToWrite, bs, externallyCacheable)
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

func getFile(pathPrefix, path string) (bytes []byte, cacheableExternally bool, err error) {
	// TODO: if getting a nonstatic file, check whether or not it is externally cacheable!
	fullPath := filepath.Join(pathPrefix, path)
	// Try to get from LRU cache first
	if bs, found, externallyCacheable := lruCache.Get(fullPath); found {
		// TODO: use externallyCacheable
		return bs, externallyCacheable, nil
	}
	// Cache miss
	bytes, err = os.ReadFile(fullPath)
	if err != nil {
		return bytes, false, err
	}
	externallyCacheable := false // TODO: false ok here?  Requires finalPerms?
	pathParts := strings.Split(path, "/")
	if len(pathParts) > 2 {
		_, externallyCacheable, _ = lruForCacheability.Get(pathParts[0] + "/" + pathParts[1]) // TODO: ensure ok
	}
	_ = lruCache.Add(fullPath, bytes, externallyCacheable) // TODO: should only ever pull from cache?!
	return bytes, externallyCacheable, err
}
func GetStaticFile(path string) (bytes []byte, externallyCacheable bool, err error) {
	return getFile("/staticFiles", path)
}

func GetPic(ctx context.Context, imgSubPath string) (bytes []byte, externallyCacheable bool, err error) {
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
