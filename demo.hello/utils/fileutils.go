package utils

import (
	"archive/zip"
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"go/build"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

//
// IO utils
//

// GetCurRunPath returns the current run abs path.
func GetCurRunPath() string {
	dir, _ := filepath.Split(os.Args[0])
	return dir
}

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
		return os.ErrExist
	}
	return os.MkdirAll(dirPath, os.ModePerm)
}

// MoveFile .
func MoveFile(src, dst string) error {
	if !IsExist(src) {
		return fmt.Errorf("src file not found: %s", src)
	}

	dstDir := filepath.Dir(dst)
	if !IsDirExist(dstDir) {
		if err := MakeDir(dstDir); err != nil {
			return err
		}
	}
	return os.Rename(src, dst)
}

// ListFilesInDir returns file names with specified ext in a dir.
func ListFilesInDir(dirPath, ext string) ([]string, error) {
	if len(ext) > 0 && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	retFileNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if len(ext) > 0 && filepath.Ext(entry.Name()) != ext {
			continue
		}
		retFileNames = append(retFileNames, entry.Name())
	}
	return retFileNames, nil
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
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return os.ExpandEnv(string(bytes)), nil
}

// WalkDir 获取指定目录及所有子目录下的所有文件，根据后缀过滤
func WalkDir(dirPath, suffix string) (files []string, err error) {
	if suffix[0] != '.' {
		suffix = "." + suffix
	}

	files = make([]string, 0, 16)
	onWalk := func(fullFilename string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}

		if filepath.Ext(fi.Name()) == suffix {
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

// RemoveExpiredFiles removes files in spec dir by expired time.
func RemoveExpiredFiles(dir string, expired float64, unit int) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	deletedFiles := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		since := time.Since(info.ModTime())
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
			absPath := filepath.Join(dir, entry.Name())
			if err := os.Remove(absPath); err != nil {
				return nil, err
			}
			deletedFiles = append(deletedFiles, absPath)
		}
	}
	return deletedFiles, nil
}

var (
	// ErrNoFilesExistInDir .
	ErrNoFilesExistInDir = fmt.Errorf("ErrNoFilesExistInDir")
)

// GetLatestFileInDir returns latest modify file name with specified ext in a dir.
func GetLatestFileInDir(dirPath, ext string) (string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", nil
	}

	if len(ext) > 0 && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	type fileItem struct {
		name    string
		modTime float64
	}

	fileItems := make([]fileItem, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if len(ext) > 0 && filepath.Ext(entry.Name()) != ext {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return "", err
		}
		item := fileItem{
			name:    entry.Name(),
			modTime: time.Since(info.ModTime()).Seconds(),
		}
		fileItems = append(fileItems, item)
	}

	if len(fileItems) == 0 {
		return "", ErrNoFilesExistInDir
	}

	sort.Slice(fileItems, func(x, y int) bool {
		xModtime := fileItems[x].modTime
		yModtime := fileItems[y].modTime
		return xModtime < yModtime
	})
	return fileItems[0].name, nil
}

//
// Common IO
//

// ReadFile .
func ReadFile(filePath string) ([]byte, error) {
	// equal to: return os.ReadFile(filePath)
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

// CreateFile creates a file by bytes content.
func CreateFile(filePath string, b []byte) error {
	if IsExist(filePath) {
		return fmt.Errorf("[%s]: %w", filePath, os.ErrExist)
	}

	newFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("Create file error: %v", err)
	}
	defer newFile.Close()

	buf := bytes.NewBuffer(b)
	if _, err = newFile.Write(buf.Bytes()); err != nil {
		return err
	}
	return nil
}

// ReadLinesFile read and return file content lines.
func ReadLinesFile(filePath string) ([]string, error) {
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

// CopyFile .
func CopyFile(srcPath, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("CopyFile error: %w", err)
	}
	defer srcFile.Close()

	outFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("CopyFile error: %w", err)
	}

	readerBuf := bufio.NewReader(srcFile)
	writerBuf := bufio.NewWriter(outFile)
	if _, err := io.Copy(writerBuf, readerBuf); err != nil {
		return fmt.Errorf("CopyFile error: %w", err)
	}
	defer writerBuf.Flush()
	return nil
}

// MergeFiles .
func MergeFiles(inPaths []string, outPath string) error {
	if err := os.Remove(outPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	// NOTE: os.O_APPEND append text to existing file content. Do not use here.
	outFile, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer outFile.Close()

	writerBuf := bufio.NewWriter(outFile)
	for _, path := range inPaths {
		if err := func(path string) error {
			inFile, err := os.OpenFile(path, os.O_RDONLY, 0644)
			if err != nil {
				return err
			}
			defer inFile.Close()

			if _, err = io.Copy(writerBuf, inFile); err != nil {
				return err
			}
			return nil
		}(path); err != nil {
			return err
		}
	}
	writerBuf.Flush()
	return nil
}

// IsFileSizeEqual .
func IsFileSizeEqual(srcPath, dstPath string) (bool, error) {
	srcStat, err := os.Stat(srcPath)
	if err != nil {
		return false, err
	}
	dstStat, err := os.Stat(dstPath)
	if err != nil {
		return false, err
	}
	return srcStat.Size() == dstStat.Size(), nil
}

// IsFilesEqual .
func IsFilesEqual(srcPath, dstPath string) (bool, error) {
	isFileSzeEqual, err := IsFileSizeEqual(srcPath, dstPath)
	if err != nil {
		return false, err
	}
	if !isFileSzeEqual {
		return false, nil
	}

	srcBytes, err := ReadFile(srcPath)
	if err != nil {
		return false, err
	}
	dstBytes, err := ReadFile(dstPath)
	if err != nil {
		return false, err
	}

	// compare md5 instead of bytes equal.
	// bytes.Equal(srcBytes, dstBytes)
	srcMD5 := GetMd5HexText(srcBytes)
	dstMD5 := GetMd5HexText(dstBytes)
	return srcMD5 == dstMD5, nil
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

//
// Zip
//

// Zip 压缩
func Zip(srcDir, dstZipFile string) error {
	zipFile, err := os.Create(dstZipFile)
	if err != nil {
		return fmt.Errorf("Zip create zip file error: %w", err)
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("Zip get file [%s] header error: %w", info.Name(), err)
		}

		header.Name = strings.Replace(path, srcDir, "", 1)
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("Zip create header error: %w", err)
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("Zip open file [%s] error: %w", path, err)
			}
			defer file.Close()

			if _, err = io.Copy(writer, file); err != nil {
				return fmt.Errorf("Zip copy error: %w", err)
			}
		}
		return nil
	})

	return nil
}

// Unzip 解压
func Unzip(zipFile, dstDir string) error {
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return fmt.Errorf("Unzip open zip file error: %w", err)
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		filePath := filepath.Join(dstDir, f.Name)
		if f.FileInfo().IsDir() {
			if err = MakeDir(filePath); err != nil {
				return fmt.Errorf("Unzip make dir [%s] error: %w", filePath, err)
			}
			continue
		}

		inFile, err := f.Open()
		if err != nil {
			return fmt.Errorf("Unzip open file error: %w", err)
		}
		defer inFile.Close()

		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("Unzip open file [%s] error: %w", filePath, err)
		}
		defer outFile.Close()

		if _, err = io.Copy(outFile, inFile); err != nil {
			return fmt.Errorf("Unzip copy error: %w", err)
		}
	}

	return nil
}
