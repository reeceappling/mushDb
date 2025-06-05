package pics

import (
	"context"
	"errors"
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
	resolvedPrefix := "/"
	for _, prefix := range prefixPath {
		resolvedPrefix = resolvedPrefix + prefix + "/"
	}
	for i := 0; i < 10; i++ { // TODO: max iterations?
		name, err := uuid.NewRandom()
		if err != nil {
			return "", err
		}
		fileNameWithPrefixPath := resolvedPrefix + name.String()
		whereToWrite := filePath + fileNameWithPrefixPath
		_, err = os.Stat(whereToWrite)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return fileNameWithPrefixPath, os.WriteFile(whereToWrite, bs, 0666)
			}
			return "", err
		} else {
			continue
		}
	}
	return "", errors.New("Failed to find a new fileName")
}

func DeleteFiles(ctx context.Context, filenamesWithPrefixPaths ...string) error {
	var errOut error = nil
	homeDir := GetFilePath(ctx)
	for _, filenameWithPrefixPath := range filenamesWithPrefixPaths {
		err := os.Remove(homeDir + filenameWithPrefixPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) { // TODO: ensure this is ok
				continue
			}
			errOut = errors.Join(errOut, err)
		}
	}
	return errOut
}
