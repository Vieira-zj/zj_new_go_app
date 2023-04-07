package demos

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/tabwriter"

	"github.com/stretchr/testify/assert"
)

func TestRelativePath(t *testing.T) {
	dst := filepath.Join(os.Getenv("HOME"), "Workspaces/zj_repos/zj_new_go_project/demo.go.new/go1_1711_demos")
	relPath, err := filepath.Rel(os.Getenv("HOME"), dst)
	assert.NoError(t, err)
	t.Log("rel path:", relPath)
}

func TestIOReadCloser(t *testing.T) {
	// 从 request.body reader 中读出请求数据后，使用 io.NopCloser 还原 request.body reader
	r := strings.NewReader("io read closer test")
	rc := io.NopCloser(r)
	defer rc.Close()

	s, err := io.ReadAll(rc)
	assert.NoError(t, err)
	t.Log("read:", string(s))
}

func TestReadBytes(t *testing.T) {
	r := strings.NewReader("abcde")

	b := make([]byte, 2)
	for {
		n, err := r.Read(b)
		if err != nil {
			if errors.Is(err, io.EOF) {
				if n > 0 {
					// not go here
					t.Logf("read last [%d]: %s", n, b[:n])
				}
				t.Log("eof")
				break
			}
			t.Fatal(err)
		}
		// NOTE: use b[:n] but not b
		t.Logf("read [%d]: %s", n, b[:n])
	}
}

func TestFileSeek(t *testing.T) {
	fpath := "/tmp/test/hello.txt"
	f, err := os.Open(fpath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("file size: %d", fi.Size())

	b := make([]byte, 1)
	for {
		offset, err := f.Seek(1, 1)
		if err != nil {
			t.Logf("seek error: %v", err)
		}
		t.Logf("seek to: %d", offset)

		// start from offset 0, and read byte from current offset.
		// after read, offset move forward.
		n, err := f.Read(b)
		if err != nil {
			if errors.Is(err, io.EOF) {
				t.Log("eof")
				break
			}
			t.Fatal(err)
		}
		t.Logf("read [%d]: %s", n, b[:n])
	}
}

func TestReadLastBytesOfFile(t *testing.T) {
	fpath := "/tmp/test/hello.txt"
	f, err := os.Open(fpath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	n, err := f.Seek(-7, 2)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("seek to: %d", n)

	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				t.Log("eof")
				break
			}
			t.Fatal(err)
		}
		t.Log("read line:", strings.TrimSuffix(line, "\n"))
	}
}

func TestTabWriter(t *testing.T) {
	var table = [][]string{
		{"vegetables", "fruits", "rank"},
		{"potato", "strawberry", "1"},
		{"lettuce", "raspberry", "2"},
		{"carrot", "apple", "3"},
		{"broccoli", "pomegranate", "4"},
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 4, 0, '\t', 0)
	for _, line := range table {
		fmt.Fprintln(writer, strings.Join(line, "\t")+"\t")
	}
	writer.Flush()
	t.Log("table write done")
}
