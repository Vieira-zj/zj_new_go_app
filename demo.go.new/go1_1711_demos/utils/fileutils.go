package utils

import (
	"archive/zip"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

/*
Truncate a file to 100 bytes.
If file is less than 100 bytes the original contents will remain at the beginning,
and the rest of the space is filled will null bytes.
If it is over 100 bytes, Everything past 100 bytes will be lost.
Either way we will end up with exactly 100 bytes.
Pass in 0 to truncate to a completely empty file.
=> os.Truncate("test.txt", 100)

Change Permissions, Ownership, and Timestamps
=> os.Chmod("test.txt", 0777)
=> os.Chown("test.txt", os.Getuid(), os.Getgid())
=> os.Chtimes("test.txt", lastAccessTime, lastModifyTime)
*/

func IsExist(filePath string) bool {
	if _, err := os.Stat(filePath); err != nil {
		return os.IsExist(err)
	}
	return true
}

func IsDirExist(dirPath string) bool {
	f, err := os.Stat(dirPath)
	if err != nil {
		return false
	}
	return f.IsDir()
}

func CheckReadPermission(fpath string) bool {
	return CheckPermission(fpath, os.O_RDONLY)
}

func CheckWritePermission(fpath string) bool {
	return CheckPermission(fpath, os.O_WRONLY)
}

func CheckPermission(fpath string, op int) bool {
	file, err := os.OpenFile(fpath, op, 0666)
	if err != nil {
		if os.IsPermission(err) {
			log.Println("permission denied")
			return false
		}
	}
	file.Close()
	return true
}

// scan file

func ReadFileLines(fpath string) ([]string, error) {
	file, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	lines := make([]string, 0, 16)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}
	return lines, nil
}

func ReadFileWords(fpath string) ([]string, error) {
	file, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)

	words := make([]string, 0, 16)
	for scanner.Scan() {
		words = append(words, scanner.Text())
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}
	return words, nil
}

// copy file

func CopyFile(srcPath, dstPath string) error {
	const tag = "CopyFile"
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("%s error: %v", tag, err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("%s error: %v", tag, err)
	}
	defer dstFile.Close()

	if _, err = io.Copy(srcFile, dstFile); err != nil {
		return fmt.Errorf("%s error: %v", tag, err)
	}
	// commit the file contents, and flushes memory to disk
	return dstFile.Sync()
}

func BufCopyFile(srcPath, dstPath string) error {
	const tag = "BufCopyFile"
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("%s error: %v", tag, err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("%s error: %v", tag, err)
	}
	defer dstFile.Close()

	readerBuf := bufio.NewReader(srcFile)
	writerBuf := bufio.NewWriter(dstFile)
	if _, err = io.Copy(writerBuf, readerBuf); err != nil {
		return fmt.Errorf("%s error: %v", tag, err)
	}
	return writerBuf.Flush()
}

// zip 打包文件

func ZipArchiveFiles(srcPath, dstZipPath string) error {
	if !strings.HasSuffix(dstZipPath, ".zip") {
		dstZipPath += ".zip"
	}

	dstZipFile, err := os.Create(dstZipPath)
	if err != nil {
		return err
	}
	defer dstZipFile.Close()

	zipWriter := zip.NewWriter(dstZipFile)
	defer zipWriter.Close()

	filepath.Walk(srcPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = strings.TrimPrefix(path, srcPath)
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		newZipWriter, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			srcFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer srcFile.Close()

			if _, err := io.Copy(newZipWriter, srcFile); err != nil {
				return err
			}
		}
		return nil
	})
	return nil
}

func UnzipArchivedFile(zipFilePath, dstPath string) error {
	zipReader, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, file := range zipReader.File {
		var localErr error
		func() {
			extractedFilePath := filepath.Join(dstPath, file.Name)
			if file.FileInfo().IsDir() {
				os.MkdirAll(extractedFilePath, file.Mode())
				return
			}

			zippedFile, err := file.Open()
			if err != nil {
				localErr = err
				return
			}
			defer zippedFile.Close()

			outputFile, err := os.OpenFile(extractedFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
			if err != nil {
				localErr = err
				return
			}
			defer outputFile.Close()

			_, localErr = io.Copy(outputFile, zippedFile)
		}()
		if localErr != nil {
			return localErr
		}
	}
	return nil
}

// gzip 压缩文件

func GzipCompressFile(srcPath, dstGzipPath string) error {
	if !strings.HasSuffix(dstGzipPath, ".gzip") {
		dstGzipPath += ".gzip"
	}

	dstGzipFile, err := os.Create(dstGzipPath)
	if err != nil {
		return err
	}
	defer dstGzipFile.Close()

	gzipWriter := gzip.NewWriter(dstGzipFile)
	defer gzipWriter.Close()

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	_, err = io.Copy(gzipWriter, srcFile)
	return err
}

func GzipUncompressFile(gzipFilePath, dstFilePath string) error {
	gzipFile, err := os.Open(gzipFilePath)
	if err != nil {
		return err
	}
	defer gzipFile.Close()

	gzipReader, err := gzip.NewReader(gzipFile)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	outputFile, err := os.Create(dstFilePath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, gzipReader)
	return err
}
