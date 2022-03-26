package pkg

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func getParamFromEnv(key string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		panic(fmt.Sprintf("Env var [%s] not set", key))
	}
	return value
}

func getSimpleNowDatetime() string {
	return time.Now().Format("20060102_150405")
}

func getFileNameWithoutExt(fileName string) string {
	ext := filepath.Ext(fileName)
	if len(ext) == 0 {
		return fileName
	}
	return strings.Replace(fileName, ext, "", 1)
}

func getRepoNameFromURL(url string) (string, error) {
	if !strings.HasPrefix(url, "http") {
		return "", fmt.Errorf("URL is not http/https")
	}

	items := strings.Split(url, "/")
	repo := items[len(items)-1]

	if strings.HasSuffix(repo, ".git") {
		return strings.Replace(repo, ".git", "", 1), nil
	}
	return repo, nil
}

func formatIPAddress(address string) (string, error) {
	u, err := url.Parse(address)
	if err != nil {
		return "", err
	}
	ip := strings.Replace(u.Hostname(), ".", "-", -1)
	return ip, nil
}
