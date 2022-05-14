package pkg

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// FormatFilePathWithNewExt .
func FormatFilePathWithNewExt(filePath, newExt string) string {
	return strings.Replace(filePath, filepath.Ext(filePath), "."+newExt, 1)
}

func getParamFromEnv(key string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		log.Fatalf("Env var [%s] not set", key)
	}
	return value
}

func getSimpleNowDatetime() string {
	return time.Now().Format("20060102_150405")
}

func getFilePathWithoutExt(filePath string) string {
	ext := filepath.Ext(filePath)
	if len(ext) == 0 {
		return filePath
	}
	return strings.Replace(filePath, ext, "", 1)
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

func formatCoverPercentage(cover string) (string, error) {
	total, err := strconv.ParseFloat(cover, 32)
	if err != nil {
		return "0", fmt.Errorf("formatCoverPercentage error: %w", err)
	}
	return fmt.Sprintf("%.2f", total*100), nil
}

func getCoverTotalFromSummary(summary string) string {
	items := strings.Split(summary, "\t")
	total := items[len(items)-1]
	return strings.Replace(total, "%", "", 1)
}
