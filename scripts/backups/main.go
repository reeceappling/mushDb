package main

import (
	"archive/zip"
	"flag"
	"github.com/joho/godotenv"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	var create, load bool
	flag.BoolVar(&create, "create", false, "when true, will read a zip archive into the correct area for the environment. This is creating a zip file, and will not change the current environment's files...")
	flag.BoolVar(&load, "load", false, "when true, will read a zip archive into the correct area for the environment. This is loading a zip file, overwriting the current environment's files.")
	zipFileName := flag.String("zipfile", "", "path of the zipfile to read or write") // TODO: can be relative???
	flag.Parse()
	if create == load {
		if create {
			log.Fatal("Cannot both create and load a backup at the same time...")
		} else {
			log.Fatal("Must include either --create or --load flag to specify the operation to perform...")
		}
		return
	}
	if create {
		zipName := time.Now().Format("2006-01-02_15-04-05") + ".zip"
		if zipFileName != nil && *zipFileName != "" {
			zipName = *zipFileName
		}

		// TODO: allow a flag for naming the output file?
		//envFiles := os.Args[2:] // TODO: ensure ok
		envFiles := flag.Args() // TODO: ensure ok
		imagesDir, dbDir := getDirs(envFiles...)
		createBackup(zipName, imagesDir, dbDir)
	} else {
		//panic("load is not enabled!")
		if zipFileName == nil || *zipFileName == "" {
			log.Fatal("invalid zip file name, must be a valid existing file") // TODO; ensure file exists first
		}
		// TODO: allow a flag for naming the output file?
		//envFiles := os.Args[2:] // TODO: ensure ok
		envFiles := flag.Args() // TODO: ensure ok
		imagesDir, dbDir := getDirs(envFiles...)
		loadBackup(*zipFileName, imagesDir, dbDir)
	}
	//scriptArgs := os.Args[1:]
	//createOrLoad := scriptArgs[0]
	//switch createOrLoad {
	//case "create":
	//
	////case "load": // TODO: REENABLE LATER!
	//
	//default:
	//	panic("invalid arg in position 1, must be 'create' or 'load'")
	//}
	//// TODO: get args
}

const (
	imagesZipDir = "images"
	dbZipDir     = "db"
)

func getDirs(envFiles ...string) (imagesDir, dbDir string) {
	err := godotenv.Load(envFiles...)
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	//LOCAL_IMAGES_DIR="/opt/mush/data/pictures"
	//LOCAL_DB_DIR="/opt/mush/data/db"
	//LOCAL_DB_BACKUPS_DIR="/opt/mush/backups/db"
	//#LOCAL_PICS_BACKUPS_DIR="/opt/mush/backups/pics"

	return os.Getenv("LOCAL_IMAGES_DIR"), os.Getenv("LOCAL_DB_DIR")
	//s3Bucket := os.Getenv("LOCAL_DB_BACKUPS_DIR")
	//secretKey := os.Getenv("LOCAL_PICS_BACKUPS_DIR")
	//secretKey := os.Getenv("LOCAL_BACKUPS_DIR")
}

func createBackup(zipName, imagesDir, dbDir string) {
	println("trying to create backup zip file at " + zipName)
	zipDir := filepath.Dir(zipName) // TODO: ensure ok
	// 0755 sets standard read/write/execute permissions for directories
	err := os.MkdirAll(zipDir, 0777)
	if err != nil {
		log.Fatalf("Failed to create directory: %v\n", err)
		return
	}

	zipFile, err := os.Create(zipName)
	if err != nil {
		panic("failed to create zip file: " + err.Error())
	}
	defer func() {
		if err := zipFile.Close(); err != nil {
			panic("failed to close zip file: " + err.Error())
		}
		if err := os.Chmod(zipFile.Name(), 0666); err != nil {
			panic(err)
		}
	}()

	// Initialize the zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for zipRoot, localDir := range map[string]string{
		dbZipDir:     dbDir,
		imagesZipDir: imagesDir,
	} {
		_, err = zipWriter.Create(zipRoot)
		if err != nil {
			panic(err)
		}
		// Clean the source path to ensure consistent delimiters
		source := filepath.Clean(localDir)
		// Walk through the directory tree
		err = filepath.WalkDir(source, func(path string, d os.DirEntry, err error) error {
			// TODO: THIS IS NOT COPYING EVERYTHING!
			if err != nil {
				return err
			}

			// 5. Create a local path structure inside the zip
			relPath, err := filepath.Rel(source, path)
			if err != nil {
				return err
			}

			// Skip the root directory itself
			if relPath == "." {
				return nil
			}

			// Ensure zip paths use forward slashes for cross-platform compatibility
			zipPath := filepath.ToSlash(filepath.Join(zipRoot, relPath)) // TODO: ensure ok

			// Handle directories by appending a trailing slash
			if d.IsDir() {
				_, err = zipWriter.Create(zipPath + "/")
				return err
			}

			// Create a file header inside the zip
			writer, err := zipWriter.Create(zipPath)
			if err != nil {
				return err
			}

			// Open the source file
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			// Copy the file contents into the zip writer
			_, err = io.Copy(writer, file)
			return err
		})
		if err != nil {
			panic("failed to walk for " + zipRoot + ". " + err.Error())
		}
	}
}
func loadBackup(zipPath, imagesDir, dbDir string) {
	dbTempDir, err := os.MkdirTemp("", "dbBackup") // Renamed later to keep it as a real directory
	if err != nil {
		log.Fatalf("failed to make db backup dir, %v", err)
	}
	imgsTempDir, err := os.MkdirTemp("", "imagesBackup") // Renamed later to keep it as a real directory
	if err != nil {
		log.Fatalf("failed to make images backup dir, %v", err)
	}
	for _, tempD := range []string{dbTempDir, imgsTempDir} {
		err = os.Chmod(tempD, 0777)
		if err != nil {
			log.Fatalf("failed to set perms on temp dir %v", err)
		}
	}

	// Open the ZIP file for reading
	zipReader, err := zip.OpenReader(zipPath)
	if err != nil {
		log.Fatalf("Failed to open zip file: %v", err)
	}
	defer zipReader.Close()

	imagesZipDirWithSlash := imagesZipDir + "/"
	dbZipDirWithSlash := dbZipDir + "/"
	// Iterate through all files in the archive
	for _, file := range zipReader.File {
		isImages := strings.HasPrefix(file.Name, imagesZipDirWithSlash)
		isDb := strings.HasPrefix(file.Name, dbZipDirWithSlash)

		if !isImages && !isDb {
			if file.Name == dbZipDir || file.Name == imagesZipDir {
				continue
			}
			log.Fatalf("%s has incorrect prefix! should have had either %s or %s", file.Name, imagesZipDirWithSlash, dbZipDirWithSlash)
		}

		var zipDir string
		var parentTempDir string
		if isImages {
			zipDir, parentTempDir = imagesZipDir, imgsTempDir
		} else {
			zipDir, parentTempDir = dbZipDir, dbTempDir
		}
		relativePath, err := filepath.Rel(zipDir, file.Name) // TODO: ensure ok
		if err != nil {
			log.Fatalf("failed to get relative path, %v", err)
		}
		if relativePath == "." {
			continue
		}
		finalPath := filepath.Join(parentTempDir, relativePath) // TODO: ensure ok
		// Handle directories
		if file.FileInfo().IsDir() {
			// TODO: create this dir if it does not already exist outside of the zip!
			if err = os.Mkdir(finalPath, 0777); err != nil {
				log.Fatalf("failed to create dir %s: %v", finalPath, err.Error())
			}
			continue
		}

		// Process the target file
		//fmt.Printf("\nFound file: %s (Size: %d bytes)\n", file.Name, file.UncompressedSize64)
		outputFile, err := os.OpenFile(finalPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666) // TODO: ensure ok
		if err != nil {
			log.Fatalf("failed to open file, %v", err)
		}
		err = readZipFileContentToNewFile(file, outputFile)
		if err != nil {
			log.Fatalf("Error reading file %s: %v", file.Name, err)
		}
	}
	// once here, then we can rename the temp directories to the real ones to make them permanent
	for tempDir, finalDir := range map[string]string{
		imgsTempDir: imagesDir,
		dbTempDir:   dbDir,
	} {
		err = os.Rename(tempDir, finalDir)
		if err != nil {
			log.Fatalf("failed to replace directory %s: %v", imagesDir, err)
		}
	}
}

func readZipFileContentToNewFile(src *zip.File, dst *os.File) error { // TODO: FIX!
	// Open the specific file inside the zip
	rc, err := src.Open()
	if err != nil {
		return err
	}
	defer dst.Close()
	defer rc.Close()

	_, err = dst.ReadFrom(rc) // TODO: ensure ok!
	return err
	//
	//// Example: Read the first 100 bytes of the file content
	//buffer := make([]byte, 100)
	//n, err := rc.Read(buffer)
	//if err != nil && err != io.EOF {
	//	return err
	//}
	//
	//fmt.Printf("Content preview: %s\n", string(buffer[:n]))
	//return nil
}
