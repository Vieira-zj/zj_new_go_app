package utils_test

import (
	"testing"

	"demo.apps/utils"
)

func TestIsDirExist(t *testing.T) {
	for _, path := range []string{
		"/tmp/test",
		"/tmp/test/mock",
		"/tmp/test/test.json",
	} {
		result := utils.IsExist(path)
		t.Logf("%s is exist: %v", path, result)
		result = utils.IsDirExist(path)
		t.Logf("%s is dir exist: %v\n", path, result)
	}
}

func TestBlockedCopy(t *testing.T) {
	src := "/tmp/test/src_copy.zip"
	dest := "/tmp/test/dest_copied.zip"
	if err := utils.BlockedCopy(src, dest); err != nil {
		t.Fatal(err)
	}
	t.Log("success copied")
}

func TestGetFileContentType(t *testing.T) {
	for _, path := range []string{
		"/tmp/test/raw.json",
		"/tmp/test/public/index.html",
		"/tmp/test/gin",
	} {
		tp, err := utils.GetFileContentType(path)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("file:%s content_type:%s", path, tp)
	}
}

func TestSearchFiles(t *testing.T) {
	root := "/Users/jinzheng/Downloads/tmps"
	results, err := utils.SearchFiles(root, "*.go", "*.txt")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("total match:", len(results))
	for _, path := range results {
		t.Log(path)
	}
}
