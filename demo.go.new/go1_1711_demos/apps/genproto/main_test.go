package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestMain(t *testing.T) {
	subDir := "Workspaces/zj_repos/zj_new_go_project/demo.go.new/go1_1711_demos/apps/genproto"
	t.Setenv("ROOT_DIR", filepath.Join(os.Getenv("HOME"), subDir))
}

func TestReadCmdLinesFromProto(t *testing.T) {
	path := filepath.Join(os.Getenv("ROOT_DIR"), "echo.proto")
	cmdLines, err := readCmdLinesFromProto(path)
	if err != nil {
		t.Fatal(err)
	}

	for i := range cmdLines {
		fmt.Println(cmdLines[i])
	}
}

func TestParserCommandLines(t *testing.T) {
	path := filepath.Join(os.Getenv("ROOT_DIR"), "echo.proto")
	cmdLines, err := readCmdLinesFromProto(path)
	if err != nil {
		t.Fatal(err)
	}

	service, mds, err := parserCommandLines(cmdLines)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("service:", service)
	for i := range mds {
		fmt.Printf("%+v\n", mds[i])
	}
}

func TestGenRpcMethods(t *testing.T) {
	path := filepath.Join(os.Getenv("ROOT_DIR"), "echo.proto")
	cmdLines, err := readCmdLinesFromProto(path)
	if err != nil {
		t.Fatal(err)
	}

	service, mds, err := parserCommandLines(cmdLines)
	if err != nil {
		t.Fatal(err)
	}

	content := genRpcMethodsDeclare(service, mds)
	fmt.Println(content)
}

func TestGenNewProtoFile(t *testing.T) {
	path := filepath.Join(os.Getenv("ROOT_DIR"), "echo.proto")
	cmdLines, err := readCmdLinesFromProto(path)
	if err != nil {
		t.Fatal(err)
	}

	service, mds, err := parserCommandLines(cmdLines)
	if err != nil {
		t.Fatal(err)
	}

	content := genRpcMethodsDeclare(service, mds)
	dstPath := "/tmp/test/echo.proto.gen"
	if err := genNewProtoFile(path, dstPath, content); err != nil {
		t.Fatal(err)
	}
	t.Log("gen new proto file:", dstPath)
}
