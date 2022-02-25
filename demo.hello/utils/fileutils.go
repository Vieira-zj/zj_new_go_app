package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"go/build"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// IsExist .
func IsExist(filePath string) bool {
	if _, err := os.Stat(filePath); err != nil {
		return os.IsExist(err)
	}
	return true
}

// IsDirExist .
func IsDirExist(dirPath string) bool {
	f, err := os.Stat(dirPath)
	if err != nil {
		return false
	}
	return f.IsDir()
}

// HasPermission .
func HasPermission(filePath string) bool {
	if _, err := os.Stat(filePath); err != nil {
		return os.IsPermission(err)
	}
	return true
}

// MakeDir .
func MakeDir(dirPath string) error {
	if IsExist(dirPath) {
		return fmt.Errorf("dir path is exist: %s", dirPath)
	}
	return os.MkdirAll(dirPath, os.ModePerm)
}

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

const (
	// Hour time unit hour.
	Hour = iota
	// Minute .
	Minute
	// Second .
	Second
)

// RemoveExpiredFile removes files in spec dir by expired time.
func RemoveExpiredFile(dir string, expired float64, unit int) ([]string, error) {
	items, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	removedFiles := make([]string, 0, len(items))
	for _, item := range items {
		if item.IsDir() {
			continue
		}

		since := time.Since(item.ModTime())
		var duration float64
		switch unit {
		case Hour:
			duration = since.Hours()
		case Minute:
			duration = since.Minutes()
		case Second:
			duration = since.Seconds()
		}

		if duration >= expired {
			absPath := filepath.Join(dir, item.Name())
			if err := os.Remove(absPath); err != nil {
				return nil, err
			}
			removedFiles = append(removedFiles, item.Name())
		}
	}
	return removedFiles, nil
}

//
// Common IO
//

// ReadFileLines read and return file content lines.
func ReadFileLines(filePath string) ([]string, error) {
	if !IsExist(filePath) {
		return nil, fmt.Errorf("file [%s] not found", filePath)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file [%s] error: %v", filePath, err)
	}
	defer f.Close()

	br := bufio.NewReader(f)
	retLines := make([]string, 0, 16)
	for {
		// line, err := br.ReadString('\n')
		line, isPrefix, err := br.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return retLines, err
		}
		if isPrefix {
			return retLines, fmt.Errorf("A too long line, seems unexpected")
		}
		retLines = append(retLines, string(line))
	}
	return retLines, nil
}

// WriteLinesToFile .
func WriteLinesToFile(filePath string, outLines []string) error {
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open file [%s] error: %v", filePath, err)
	}
	defer f.Close()

	for _, line := range outLines {
		if _, err := f.WriteString(line + "\n"); err != nil {
			return err
		}
	}
	return nil
}

// CreateFile creates a file by bytes content.
func CreateFile(filePath string, b []byte) error {
	if IsExist(filePath) {
		return fmt.Errorf("file [%s] is exist", filePath)
	}

	buf := bytes.NewBuffer(b)
	newFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("os.Create [%s] error: %v", filePath, err)
	}
	defer newFile.Close()

	if _, err = newFile.Write(buf.Bytes()); err != nil {
		return err
	}
	return nil
}

// FileWordsCount gets file words count. (for test)
func FileWordsCount(filePath string) (map[string]int, error) {
	if !IsExist(filePath) {
		return nil, fmt.Errorf("file [%s] not found", filePath)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file [%s] error: %v", filePath, err)
	}
	defer f.Close()

	counts := make(map[string]int, 16)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		words := strings.Split(scanner.Text(), " ")
		trimWords := make([]string, 0, len(words))
		for _, word := range words {
			word = strings.Trim(word, ",")
			word = strings.Trim(word, ".")
			trimWords = append(trimWords, word)
		}
		for _, word := range trimWords {
			counts[word]++
		}
	}
	return counts, nil
}
