package lib

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	CMD_ZIP  = "zip"
	CMD_PACK = "pack"
)

func RecursiveZip(i int, pathToZip, destinationPath string) error {
	Log(i, "mk:zip", "O", "Starting Zip")
	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	myZip := zip.NewWriter(destinationFile)
	err = filepath.Walk(pathToZip, func(filePath string, info os.FileInfo, err error) error {
		Log(i, "mk:zip", "O", fmt.Sprintf("Adding file: %s", filePath))
		if info.IsDir() {
			return nil
		}
		if err != nil {
			return err
		}
		relPath := strings.TrimPrefix(filePath, filepath.Dir(pathToZip))
		zipFile, err := myZip.Create(relPath)
		if err != nil {
			return err
		}
		fsFile, err := os.Open(filePath)
		if err != nil {
			return err
		}
		_, err = io.Copy(zipFile, fsFile)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	err = myZip.Close()
	if err != nil {
		return err
	}
	Log(i, "mk:zip", "O", fmt.Sprintf("Zip finished: %s", destinationPath))
	return nil
}
func RunMkCmd(i int, line string) error {
	parts := strings.Split(line, " ")
	switch parts[0] {
	case CMD_ZIP:
		if len(parts) < 3 {
			return errors.New(fmt.Sprintf("Not enought params: must be zip <dir> <zipfile>. Got: %s", strings.Join(parts, " ")))
		}
		return RecursiveZip(i, parts[1], parts[2])
	case CMD_PACK:
		var ptz string
		if len(parts) > 1 {
			ptz = parts[1]
		} else {
			for _, f := range []string{"stage", "deploy", "pack", "release"} {
				st, err := os.Stat(f)
				if err == nil && st.IsDir() {
					ptz = f
					break
				}

			}

		}
		dir, err := os.Getwd()
		if err != nil {
			return err
		}
		dname := path.Base(dir)
		ts := time.Now().Format("060102150405")
		fname := dname + "-" + ts + ".zip"
		return RecursiveZip(i, ptz, fname)
	default:
		Log(i, parts[0], "E", fmt.Sprintf("mk:cmd %s is not known", parts[0]))
		return errors.New(fmt.Sprintf("mk:cmd %s is not known", parts[0]))
	}
}
