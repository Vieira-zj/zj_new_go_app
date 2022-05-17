package handler

import (
	"fmt"
	"testing"
)

func TestListFileNamesFromDir(t *testing.T) {
	dir := "/tmp/test/goc_staging_space/public/report/apa_echoserver"
	ext := "html"
	filter := "6cd6e61317"
	names, err := listFileNamesFromDir(dir, ext, filter, 3)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("files:", names)
}
