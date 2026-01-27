package utils

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

func CopyDir(srcDir, dstDir string) error {
	srcFS := os.DirFS(srcDir)
	// 1. CopyFS 不会覆盖目标目录中已有的文件. 如果目标文件中已存在某个文件, 函数会返回一个错误, 其中 errors.Is(err, fs.ErrExist) 为 true
	// 2. 符号链接不会被复制, 而是会返回一个 ErrInvalid 错误
	return os.CopyFS(dstDir, srcFS)
}

func StreamRead(src io.Reader, dst io.Writer) error {
	buf := make([]byte, 4*1024) // reusable 4k buffer
	_, err := io.CopyBuffer(dst, src, buf)
	return err
}

func GetMimeType(filePath string) (string, error) {
	if mimeType := GetMimeTypeByExt(filePath); len(mimeType) > 0 {
		return mimeType, nil
	}

	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// read the first 512 bytes (enough for detection)
	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil {
		return "", err
	}
	if n == 0 {
		return "", fmt.Errorf("empty file")
	}

	mimeType := http.DetectContentType(buf[:n])
	return mimeType, nil
}

func GetMimeTypeByExt(path string) string {
	if ext := filepath.Ext(path); len(ext) > 0 {
		return mime.TypeByExtension(ext)
	}
	return ""
}
