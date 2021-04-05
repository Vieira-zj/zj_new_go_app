package funcdiff

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestGetFileAllFuncs(t *testing.T) {
	path := "/tmp/test/old/message_tab.go"
	infos, err := GetFileAllFuncInfos(path)
	if err != nil {
		t.Fatal(err)
	}

	fileName := filepath.Base(path)
	fmt.Printf("All Func info for [%s]:\n", fileName)
	for _, info := range infos {
		fmt.Printf("%+v\n", *info)
	}
}

func TestGetSpecifiedFuncInfo(t *testing.T) {
	srcPath := "/tmp/test/demo.go"
	funcNames := []string{"demo01", "demo0501", "demo11"}

	for _, name := range funcNames {
		info, err := GetFuncInfo(srcPath, name)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("func [%s]: %+v\n", name, *info)

		res, err := GetFuncSrc(srcPath, info)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("func [%s] source:\n%s\n", name, res)
	}
}
