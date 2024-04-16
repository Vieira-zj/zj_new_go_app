package structs

import (
	"os"
	"strings"
	"testing"
)

func TestEditBuffer(t *testing.T) {
	path := "/tmp/test/raw.txt"
	if err := createRawTextFile(path); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("bytes size:", len(b))

	edit := NewEditBuffer(b)
	edit.Delete(0, 2) // delete 1st two chars
	edit.Insert(5, "insert\n")
	edit.Replace(11, 14, "xyz")

	path = "/tmp/test/raw_update.txt"
	b, err = edit.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, b, 0644); err != nil {
		t.Fatal(err)
	}
	t.Log("edit done")
}

func createRawTextFile(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	sb := strings.Builder{}
	for i := 0; i < 3; i++ {
		sb.WriteString("abcd\n")
	}
	_, err = f.WriteString(sb.String())
	return err
}
