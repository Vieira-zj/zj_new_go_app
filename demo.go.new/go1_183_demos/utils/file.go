package utils

import (
	"errors"
	"io"
	"net/http"
	"os"
)

func IsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func IsDirExist(path string) bool {
	f, err := os.Stat(path)
	if err != nil {
		return false
	}
	return f.IsDir()
}

// BlockedCopy copy file by 1m block.
func BlockedCopy(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	buf := make([]byte, 1024*1024) // 1m
	for {
		n, err := srcFile.Read(buf)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		if n == 0 { // eof
			break
		}

		if _, err = destFile.Write(buf[:n]); err != nil {
			return err
		}
	}

	return nil
}

func GetFileContentType(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}

	buf := make([]byte, 512)
	if _, err = f.Read(buf); err != nil {
		return "", err
	}

	return http.DetectContentType(buf), nil
}
