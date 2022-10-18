package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

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

func CopyFile(srcPath, dstPath string) error {
	const tag = "CopyFile"
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("%s error: %v", tag, err)
	}
	defer srcFile.Close()

	outFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("%s error: %v", tag, err)
	}

	readerBuf := bufio.NewReader(srcFile)
	writerBuf := bufio.NewWriter(outFile)
	_, err = io.Copy(writerBuf, readerBuf)
	defer writerBuf.Flush()
	return err
}
