package pics

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"os"
)

const filePathCtxKey = "dbImageFilePath"

func SetFilePath(ctx context.Context, filePath string) context.Context {
	return context.WithValue(ctx, filePathCtxKey, filePath)
}

func GetFilePath(ctx context.Context) string {
	return ctx.Value(filePathCtxKey).(string)
}

func SaveFile(ctx context.Context, bs []byte, prefixPath ...string) (string, error) {
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
		if err = os.MkdirAll(filePath+"/"+resolvedPrefix, 777); err != nil {
			fmt.Printf("Error creating directory: %s\n", err)
			return fileNameWithPrefixPath, err
		}
		if _, err = os.Stat(whereToWrite); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				println("filePath" + filePath)
				println("writing file to " + whereToWrite) // TODO: del
				err = os.WriteFile(whereToWrite, bs, 777)
				if err != nil {
					println("failed to write file", err.Error())
				}
				return fileNameWithPrefixPath, err
			}
			println("file exists already!", err.Error()) // TODO: continue???
			return "", err
		} else {
			continue
		}
	}
	return "", errors.New("failed to find a new fileName")
}

func DeleteFiles(ctx context.Context, filenamesWithPrefixPaths ...string) error {
	var errOut error = nil
	homeDir := GetFilePath(ctx)
	for _, filenameWithPrefixPath := range filenamesWithPrefixPaths {
		err := os.Remove(homeDir + filenameWithPrefixPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			errOut = errors.Join(errOut, err)
		}
	}
	return errOut
}
