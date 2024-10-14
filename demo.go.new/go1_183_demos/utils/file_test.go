package utils_test

import (
	"encoding/csv"
	"os"
	"testing"

	"demo.apps/utils"
	"github.com/stretchr/testify/assert"
)

func TestOSGetwd(t *testing.T) {
	path, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("cur path:", path)
}

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

func TestIsSymlinkFile(t *testing.T) {
	for _, fpath := range []string{
		"/tmp/test/test.csv",
		"/tmp/test/test_link.csv", // ln -s test.csv test_link.csv
	} {
		result, err := utils.IsSymlinkFile(fpath)
		assert.NoError(t, err)
		t.Logf("is symlink file (%s): %v", fpath, result)
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

func TestWriteCSV(t *testing.T) {
	fpath := "/tmp/test/output.csv"
	f, err := os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY, 0644)
	assert.NoError(t, err)
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	for _, row := range [][]string{
		{"Name", "Age", "City"},
		{"Foo", "25", "New York"},
		{"Bob", "30", "London"},
		{"Bar", "20", "Paris"},
	} {
		err := w.Write(row)
		assert.NoError(t, err)
	}

	t.Log("write csv finish")
}
