package pics

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/reeceappling/goUtils/v2/errorreference"
	"os"
	"path/filepath"
)

//	type Interactor interface {
//		SaveFile(ctx context.Context, bs []byte, prefixPath ...string) (string, error)
//	}
//
//	type FileReadWriter interface {
//		Read(fileSubPath string)([]byte, error)
//		Save(bs []byte, path string) error
//	}
var ErrNotFound = errors.New("file not found")

//type MultiWriterPriorityReader []FileReadWriter
//func (rws MultiWriterPriorityReader) Read(path string)(out []byte, err error){
//	for _, rw := range rws {
//		out, err = rw.Read(path)
//		if err != nil && errors.Is(err, ErrNotFound) {
//			continue
//		}
//		return out, err
//	}
//	return nil, err
//}
//
//func (rws MultiWriterPriorityReader) Save(bs []byte, subPath string) error{
//	var err error = nil
//	for _, rw := range rws {
//		err = errors.Join(err, rw.Save(bs, subPath))
//	}
//	return err // TODO: ok?
//}
//
//type LocalReadWriter struct {
//	dir string
//}
//func (rw LocalReadWriter) Read(path string)([]byte, error){
//	out, err := os.ReadFile(rw.dir+"/"+path)
//	if err != nil && errors.Is(err, os.ErrNotExist) {
//		return nil, ErrNotFound
//	}
//	return out, err
//}
//func (rw LocalReadWriter) Save(bs []byte, subPath string) error{
//	return os.WriteFile(rw.dir+"/"+subPath, bs, 0666)
//}

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
				println("filePath" + filePath)             // TODO: del
				println("writing file to " + whereToWrite) // TODO: del
				err = os.WriteFile(whereToWrite, bs, 777)  // TODO: 666 instead of 777?
				if err != nil {
					println("failed to write file", err.Error())
				}
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
	picsDir := GetFilePath(ctx)
	bytes, err = os.ReadFile(filepath.Join(picsDir, imgSubPath))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = errorreference.ErrorNotFound
			//// TODO: READ FROM s3 if needed?
			//bytes, err = s3.NewFileReader().Read(ctx, imgSubPath) // TODO: ensure ok
			//if err != nil {
			//	if errors.Is(err, errorreference.ErrorNotFound) {
			//		println("file does not exist locally or in s3!") // TODO: fix
			//		return nil, ErrNotFound
			//
			//	}
			//	return nil, err
			//}
		}
	}
	return bytes, err
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
