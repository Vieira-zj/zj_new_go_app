package utils

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

func CopyDir(srcDir, dstDir string) error {
	srcFS := os.DirFS(srcDir)
	// 1. CopyFS 不会覆盖目标目录中已有的文件. 如果目标文件中已存在某个文件, 函数会返回一个错误, 其中 errors.Is(err, fs.ErrExist) 为 true.
	// 2. 符号链接不会被复制, 而是会返回一个 ErrInvalid 错误.
	return os.CopyFS(dstDir, srcFS)
}

func StreamRead(src io.Reader, dst io.Writer) error {
	buf := make([]byte, 4*1024) // reusable 4k buffer
	_, err := io.CopyBuffer(dst, src, buf)
	return err
}

func LimitRead(r io.Reader) ([]byte, error) {
	const maxAllowed = 10 << 20 // 10 MB
	// io.LimitReader 在读满 10 MB 后会给 io.ReadAll 返回 io.EOF.
	// io.ReadAll 接收到 io.EOF 会认定数据流正常读完, 并返回 err == nil.
	lr := io.LimitReader(r, maxAllowed+1) // 多读 1 byte
	b, err := io.ReadAll(lr)
	if err != nil {
		return nil, err
	}

	if len(b) > maxAllowed {
		return nil, errors.New("payload too large")
	}
	return b, nil
}

func ReadLargeFile(path string) iter.Seq[string] {
	return func(yield func(string) bool) {
		f, err := os.Open(path)
		if err != nil {
			fmt.Printf("open file failed: %v", err)
			return
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if scanner.Err() != nil {
				fmt.Printf("scan failed: %v", scanner.Err())
				return
			}
			if !yield(scanner.Text()) {
				return
			}
		}
	}
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

type Config struct {
	Address string
}

func LoadConfig(path string) (_ Config, err error) {
	defer func() {
		if err != nil {
			// wrap the error with general info
			err = fmt.Errorf("load config %q: %w", path, err)
		}
	}()

	cfg := Config{}
	b, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}

	if err = json.Unmarshal(b, &cfg); err != nil {
		return cfg, fmt.Errorf("decode JSON: %w", err)
	}
	if cfg.Address == "" {
		return Config{}, fmt.Errorf("invalid config")
	}

	return cfg, nil
}
