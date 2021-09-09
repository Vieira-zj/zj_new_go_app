package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsExist(t *testing.T) {
	for _, filePath := range [2]string{"/tmp/test/results.txt", "/tmp/test/data.txt"} {
		fmt.Println("file exist:", IsExist(filePath))
	}
}

func TestHasPermission(t *testing.T) {
	for _, filePath := range [2]string{"/tmp/test/results.txt", "/tmp/test/data.txt"} {
		if IsExist(filePath) {
			fmt.Println("has permission:", HasPermission(filePath))
		}
	}
}

func TestMakeDir(t *testing.T) {
	for _, dirPath := range [2]string{"/tmp/test", "/tmp/test/foo/bar"} {
		if err := MakeDir(dirPath); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("mkdir sucess:", dirPath)
		}
	}
}

func TestCreateFile(t *testing.T) {
	filePath := "/tmp/test/test.txt"
	buf := bytes.NewBuffer([]byte("Create file with content test."))
	if err := CreateFile(filePath, buf); err != nil {
		t.Fatal(err)
	}
}

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

/*
dir tools
*/

func TestListDirFile(t *testing.T) {
	dirPath := "/tmp/test"
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		filePath := filepath.Join(dirPath, file.Name())
		fmt.Println("read file:", filePath)
		b, err := ioutil.ReadFile(filePath)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(string(b))
	}
}

func TestWalkDir(t *testing.T) {
	demoPath := "Workspaces/zj_repos/zj_go2_project/demo.hello/demos"
	dirPath := filepath.Join(os.Getenv("HOME"), demoPath)
	files, err := WalkDir(dirPath, ".go")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("go files in [%s]:\n", dirPath)
	for _, file := range files {
		fmt.Println("\t" + file)
	}
}

func TestBufferWriter(t *testing.T) {
	filePath := "/tmp/test/buffer_out.txt"
	var out *bufio.Writer
	if !IsExist(filePath) {
		out = bufio.NewWriter(os.Stdout)
	} else {
		fd, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			t.Fatal(err)
		}
		defer fd.Close()
		out = bufio.NewWriter(fd)
	}
	defer out.Flush()

	for i := 0; i < 3; i++ {
		fmt.Fprintf(out, "this is a buffer write test: %d\n", i)
	}
}
