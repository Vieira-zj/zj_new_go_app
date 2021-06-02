package utils

import (
	"bytes"
	"fmt"
	"testing"
)

/*
Get project go file abs path.
*/

func TestGetGoFileAbsPath(t *testing.T) {
	path := "demo.hello/echoserver/handlers/ping.go"
	res, err := GetGoFileAbsPath(path)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res)
}

/*
Output file content with expand env.

file content:
env expand test for $USER:
home=$HOME
go_path=${GOPATH}
cur_dir=$PWD
*/

func TestReadFileWithExpandEnv(t *testing.T) {
	path := "/tmp/test/input.txt"
	res, err := ReadFileWithExpandEnv(path)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("expand string:\n", res)
}

func TestWalkDir(t *testing.T) {
	dirPath := "/Users/jinzheng/Workspaces/zj_repos/zj_go2_project/demo.hello/apps/reversecall/pkg/test"
	files, err := WalkDir(dirPath, ".go")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("go files in [%s]:\n", dirPath)
	for _, file := range files {
		fmt.Println("\t" + file)
	}
}

func TestIsExist(t *testing.T) {
	filePath := "/tmp/test/data.json"
	fmt.Println("file exist:", IsExist(filePath))
}

func TestCreateFile(t *testing.T) {
	filePath := "/tmp/test/test.txt"
	buf := bytes.NewBuffer([]byte("Create file with content test."))
	if err := CreateFile(filePath, buf); err != nil {
		t.Fatal(err)
	}
}
