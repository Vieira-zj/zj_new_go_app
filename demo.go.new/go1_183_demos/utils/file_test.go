package utils_test

import (
	"bufio"
	"io"
	"os"
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

func TestReadFileLastBytes(t *testing.T) {
	path := "/tmp/test/raw.txt"
	writeFileForTest(t, path, "\nabcd\nefghi\njkl")

	t.Run("read file last bytes", func(t *testing.T) {
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		size := len([]byte("abc"))
		t.Log("read bytes size:", size)

		reader := bufio.NewReader(f)

		// Seek(offset, start)
		if _, err = f.Seek(int64(-size), io.SeekEnd); err != nil {
			t.Fatal(err)
		}
		reader.Reset(f)

		b := make([]byte, size)
		if _, err = io.ReadFull(reader, b); err != nil {
			t.Fatal(err)
		}
		t.Logf("read last %d bytes: %s", size, b)
	})
}

func TestFileWriteAt(t *testing.T) {
	path := "/tmp/test/raw.txt"
	writeFileForTest(t, path, "abcd\nabcd\nabcd")

	t.Run("write file at offset", func(t *testing.T) {
		f, err := os.OpenFile(path, os.O_WRONLY, 0644)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		// replace existing bytes
		if _, err = f.WriteAt([]byte("xy"), 5); err != nil {
			t.Fatal(err)
		}
		if err = f.Sync(); err != nil {
			t.Fatal(err)
		}
		t.Log("file writeAt finish")
	})
}

func writeFileForTest(t *testing.T, path, content string) {
	if utils.IsExist(path) {
		os.Remove(path)
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if _, err = f.WriteString(content); err != nil {
		t.Fatal(err)
	}
}
