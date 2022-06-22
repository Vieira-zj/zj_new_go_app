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

func TestGetCommentsInGoFile(t *testing.T) {
	srcPath := filepath.Join(testRootDir, "src1/main.go")
	comments, err := getCommentsInGoFile(srcPath)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("comments:\n" + strings.Join(comments, "\n"))
}

func TestDeleteEmptyLinesInText(t *testing.T) {
	srcPath := filepath.Join(testRootDir, "src1/main.go")
	src, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatal(err)
	}

	res := deleteEmptyLinesInText(src)
	fmt.Println("results:\n", res)
}

func TestFormatGoFile(t *testing.T) {
	srcPath := filepath.Join(testRootDir, "src1/main.go")
	dstPath := filepath.Join("/tmp/test/go_project", "main_format.go")
	if err := formatGoFile(srcPath, dstPath); err != nil {
		t.Fatal(err)
	}
}

func TestGetFuncInfo01(t *testing.T) {
	src := `package main
import "fmt"
func test() {
	name:="foo"
	fmt.Println("Hello:",name)
}
`

	fnInfo, err := GetFuncInfo("", []byte(src), "test")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("func [%s] format source:\n", fnInfo.Name)
	fmt.Println(fnInfo.Source)
}

func TestGetFuncInfo02(t *testing.T) {
	path := filepath.Join(testRootDir, "src1/main.go")
	funcNames := []string{"fnToString", "fnChange", "fnConditional"}

	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range funcNames {
		info, err := GetFuncInfo(path, nil, name)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(prettySprintFuncInfo(info))
		fmt.Printf("Source code:\n%s\n", GetFuncSrc(src, info))
	}
}

func TestGetFuncInfos(t *testing.T) {
	// NOTE: go file must be in a go project for ast parser
	path := filepath.Join(testRootDir, "src1/main.go")
	// path := filepath.Join("/tmp/test/go_project", "main_format.go")
	infos, err := GetFuncInfos(path, nil)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("all Func info:\n")
	for _, info := range infos {
		fmt.Println(prettySprintFuncInfo(info))
		fmt.Println(info.Source + "\n")
	}
}
