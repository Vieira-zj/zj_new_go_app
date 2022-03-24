package pkg

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//
// Common
//

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
	return repo, nil
}

//
// Task Utils
//

func formatIPfromSrvAddress(address string) (string, error) {
	u, err := url.Parse(address)
	if err != nil {
		return "", err
	}
	ip := strings.Replace(u.Hostname(), ".", "-", -1)
	return ip, nil
}

func getModuleFromSrvName(name string) (string, error) {
	for mod := range ModuleToRepoMap {
		if strings.Contains(name, mod) {
			return mod, nil
		}
	}
	return "", fmt.Errorf("Module is not found for service: %s", name)
}

func getGitURLFromSrvName(name string) string {
	// TODO:
	return ""
}

func getBranchAndCommitFromSrvName(name string) (string, string) {
	items := strings.Split(name, "_")
	commitID := items[len(items)-1]

	branch := items[len(items)-2]
	if strings.Contains(branch, "/") {
		brItems := strings.Split(branch, "/")
		branch = brItems[len(brItems)-1]
	}
	return branch, commitID
}
