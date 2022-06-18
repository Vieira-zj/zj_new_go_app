package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunGoFmt(t *testing.T) {
	path := filepath.Join(testRootDir, "src1/main.go")
	if err := runGoFmt(path); err != nil {
		t.Fatal(err)
	}
	fmt.Println("gofmt done")
}

func TestGetGoPackage(t *testing.T) {
	path := filepath.Join(testRootDir, "src1")
	res, err := getGoPackage(path)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("package:", res)
}

func TestGetComments(t *testing.T) {
	srcPath := filepath.Join(testRootDir, "src1/main.go")
	comments, err := GetComments(srcPath)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("comments:\n" + strings.Join(comments, "\n"))
}

func TestFormatGoFile(t *testing.T) {
	srcPath := filepath.Join(testRootDir, "src1/main.go")
	dstPath := filepath.Join("/tmp/test/go_project", "main_format.go")
	if err := FormatGoFile(srcPath, dstPath); err != nil {
		t.Fatal(err)
	}
}

func TestGetSpecifiedFuncInfo(t *testing.T) {
	path := filepath.Join(testRootDir, "src1/main_format.go")
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
		fmt.Printf("\nFunc [%s]: %+v\n", name, *info)

		res := GetFuncSrc(src, info)
		fmt.Printf("Source code:\n%s\n", res)
	}
}

func TestGetFuncInfos(t *testing.T) {
	// NOTE: go file must be in a go project for ast parser
	// path := filepath.Join(testRootDir, "src1/main.go")
	path := filepath.Join("/tmp/test/go_project", "main_format.go")
	infos, err := GetFuncInfos(path)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("all Func info:\n")
	for _, info := range infos {
		fmt.Printf("%+v\n", *info)
	}
}
