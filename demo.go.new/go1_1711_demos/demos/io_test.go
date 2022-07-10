package demos

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

func TestFileSeek(t *testing.T) {
	fpath := "/tmp/test/hello.txt"
	f, err := os.Open(fpath)
	if err != nil {
		t.Fatal(err)
	}

	_, err = f.Seek(7, 1)
	if err != nil {
		t.Fatal(err)
	}

	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				t.Log("end of file")
				break
			}
			t.Fatal(err)
		}
		t.Log(strings.Trim(line, "\n"))
	}
}
