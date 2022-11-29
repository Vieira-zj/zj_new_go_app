package pkg

import (
	"os"
	"path/filepath"
)

// getAllProtoDirs returns 1st level of dirs which contains .proto file.
func getAllProtoDirs(path string) ([]string, error) {
	dirs, err := getAllSubDirs(path)
	if err != nil {
		return nil, err
	}

	retDirs := make([]string, 0, 4)
	for _, dir := range dirs {
		ok, err := isProtoDir(dir)
		if err != nil {
			return nil, err
		}
		if ok {
			retDirs = append(retDirs, dir)
		}
	}
	return retDirs, nil
}

func getAllSubDirs(path string) ([]string, error) {
	dirPaths := make([]string, 0, 4)
	if err := filepath.Walk(path, func(fpath string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			dirPaths = append(dirPaths, fpath)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return dirPaths, nil
}

func isProtoDir(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, err
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".proto" {
			return true, nil
		}
	}
	return false, nil
}
