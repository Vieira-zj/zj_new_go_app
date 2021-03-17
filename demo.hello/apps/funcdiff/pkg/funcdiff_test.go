package funcdiff

import (
	"fmt"
	"testing"
)

func TestFuncsDiff(t *testing.T) {
	srcPath := "/tmp/test/cover.go"
	targetPath := "/tmp/test/cover_new.go"
	funcName := "CoverHandler"

	isDiff, err := FuncDiff(srcPath, targetPath, funcName)
	if err != nil {
		t.Fatal(err)
	}
	res := "same"
	if isDiff {
		res = "diff"
	}
	fmt.Printf("func [%s] diff results: %s\n", funcName, res)
}

func TestGoFileDiffByFunc(t *testing.T) {
	sPath := "/tmp/test/old/message_tab.go"
	tPath := "/tmp/test/new/message_tab.go"

	fmt.Printf("Func diff info between [%s] and [%s]:\n", sPath, tPath)
	res, err := GoFileDiffByFunc(sPath, tPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range res {
		msg := "same"
		if v.IsDiff {
			msg = "diff"
		}
		fmt.Printf("[%s]:%s\n", v.FuncName, msg)
	}
}
