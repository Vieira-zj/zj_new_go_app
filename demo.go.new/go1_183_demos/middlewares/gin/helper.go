package main

import (
	"os"
	"path/filepath"
	"strings"

	"demo.apps/utils"
)

func isAcceptEncodingGzip(elems []string) bool {
	if len(elems) == 0 {
		return false
	}
	for _, elem := range elems {
		if strings.Contains(elem, "gzip") && strings.Contains(elem, "deflate") {
			return true
		}
	}
	return false
}

func isAssetsFilePath(url string) bool {
	return strings.HasPrefix(url, "/assets")
}

func getFileSize(relPath string) (int64, error) {
	fpath := filepath.Join(getDistPath(), relPath)
	stat, err := os.Stat(fpath)
	if err != nil {
		return 0, err
	}
	return stat.Size(), nil
}

func getDistPath() string {
	const distRePath = "Workspaces/zj_repos/zj_js_project/vue3_lessons/demo_apps/app_basic/dist"
	return filepath.Join(os.Getenv("HOME"), distRePath)
}

func getPagesDistPath() string {
	const distRePath = "Workspaces/zj_repos/zj_js_project/vue3_lessons/demo_pages"
	return filepath.Join(os.Getenv("HOME"), distRePath)
}

func verifyFileMd5hash(fpath, Md5hash string) (bool, error) {
	b, err := os.ReadFile(fpath)
	if err != nil {
		return false, err
	}

	actual := utils.Md5SumV2(b)
	return actual == Md5hash, nil
}
