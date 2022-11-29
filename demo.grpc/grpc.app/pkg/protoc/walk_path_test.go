package protoc

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	subPath := "Workspaces/zj_repos/zj_new_go_project/demo.grpc"
	os.Setenv("PROJECT_ROOT", filepath.Join(os.Getenv("HOME"), subPath))
	m.Run()
}

func TestGetAllSubDirs(t *testing.T) {
	path := filepath.Join(os.Getenv("PROJECT_ROOT"), "grpc.app/proto")
	dirPaths, err := getAllSubDirs(path)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("all sub dirs:")
	for _, path := range dirPaths {
		fmt.Println(path)
	}
}

func TestGetAllProtoDirs(t *testing.T) {
	path := filepath.Join(os.Getenv("PROJECT_ROOT"), "grpc.app/proto")
	dirPaths, err := GetAllProtoDirs(path)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("all proto dirs:")
	for _, path := range dirPaths {
		fmt.Println(path)
	}
}
