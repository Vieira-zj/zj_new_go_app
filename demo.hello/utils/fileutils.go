package utils

import (
	"bytes"
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

/*
Common
*/

// GetGoFileAbsPath returns .go file absolute path.
func GetGoFileAbsPath(path string) (string, error) {
	dir, file := filepath.Split(path)
	pkg, err := build.Import(dir, ".", build.FindOnly)
	if err != nil {
		return "", fmt.Errorf("can't find %q: %v", file, err)
	}
	return filepath.Join(pkg.Dir, file), nil
}

// ReadFileWithExpandEnv returns file content with expand env.
func ReadFileWithExpandEnv(path string) (string, error) {
	// 替换原文件
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return os.ExpandEnv(string(bytes)), nil
}

// WalkDir 获取指定目录及所有子目录下的所有文件, 根据后缀过滤
func WalkDir(dirPath, suffix string) (files []string, err error) {
	files = make([]string, 0, 30)
	suffix = strings.ToLower(suffix)

	onWalk := func(fullFilename string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi.IsDir() {
			return nil
		}

		if strings.HasSuffix(strings.ToLower(fi.Name()), suffix) {
			files = append(files, fullFilename)
		}
		return nil
	}

	err = filepath.Walk(dirPath, onWalk)
	return files, err
}

/*
File IO
*/

// IsExist .
func IsExist(filePath string) bool {
	if _, err := os.Stat(filePath); err != nil {
		return !os.IsNotExist(err)
	}
	return true
}

// CreateFile create a file with buf content.
func CreateFile(filePath string, buf *bytes.Buffer) error {
	newFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("os.Create %s error: %v", filePath, err)
	}

	if _, err = newFile.Write(buf.Bytes()); err != nil {
		return err
	}
	return nil
}
