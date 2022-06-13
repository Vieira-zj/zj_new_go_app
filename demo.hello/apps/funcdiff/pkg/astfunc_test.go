package funcdiff

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetFuncInfos(t *testing.T) {
	path := filepath.Join(os.Getenv("HOME"), "Downloads/go_space/src1", "main.go")
	infos, err := GetFuncInfos(path)
	if err != nil {
		t.Fatal(err)
	}

	fileName := filepath.Base(path)
	fmt.Printf("all Func info for [%s]:\n", fileName)
	for _, info := range infos {
		fmt.Printf("%+v\n", *info)
	}
}

func TestGetSpecifiedFuncInfo(t *testing.T) {
	path := filepath.Join(os.Getenv("HOME"), "Downloads/go_space/src1", "main_format.go")
	funcNames := []string{"fnHello", "fnChange", "fnConditional"}

	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range funcNames {
		info, err := GetFuncInfo(path, name)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("func [%s]: %+v\n", name, *info)

		res := GetFuncSrc(src, info)
		fmt.Printf("func [%s] source:\n%s\n", name, res)
	}
}

func TestGetComments(t *testing.T) {
	srcPath := filepath.Join(os.Getenv("HOME"), "Downloads/go_space/src1", "main.go")
	comments, err := GetComments(srcPath)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("comments:\n" + strings.Join(comments, "\n"))
}

func TestFormatGoFile(t *testing.T) {
	srcPath := filepath.Join(os.Getenv("HOME"), "Downloads/go_space/src1", "main.go")
	dstPath := filepath.Join(filepath.Dir(srcPath), "main_format.go")
	if err := FormatGoFile(srcPath, dstPath); err != nil {
		t.Fatal(err)
	}
}
