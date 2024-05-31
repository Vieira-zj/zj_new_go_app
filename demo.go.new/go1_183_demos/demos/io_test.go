package demos_test

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"demo.apps/utils"
)

func TestBufIOScan(t *testing.T) {
	lines := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		lines = append(lines, fmt.Sprintf("mock line: %d", i))
	}
	content := strings.Join(lines, "\n")

	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Split(bufio.ScanLines)

	count := 0
	for scanner.Scan() {
		line := scanner.Text()
		if err := scanner.Err(); err != nil {
			t.Fatal(err)
		}
		count += 1
		t.Log("read line:", line)
	}

	t.Log("lines count:", count)
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
