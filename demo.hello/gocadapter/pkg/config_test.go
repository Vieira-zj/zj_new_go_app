package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"demo.hello/utils"
)

func mockLoadConfig(workingDir string) error {
	subDir := "Workspaces/zj_repos/zj_go2_project/demo.hello/gocadapter/config"
	srcDir := filepath.Join(os.Getenv("HOME"), subDir)
	for _, file := range []string{"adapter.json", "module_repo_map.json"} {
		dstPath := filepath.Join(workingDir, file)
		if utils.IsExist(dstPath) {
			continue
		}
		srcPath := filepath.Join(srcDir, file)
		if err := utils.CopyFile(srcPath, dstPath); err != nil {
			return err
		}
	}

	if err := LoadConfig(filepath.Join(workingDir, "adapter.json")); err != nil {
		return err
	}
	if err := LoadModuleToRepoMap(); err != nil {
		return err
	}
	return nil
}

func TestLoadConfig(t *testing.T) {
	workingDir := "/tmp/test"
	if err := mockLoadConfig(workingDir); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("config: %+v\n", AppConfig)
}

func TestLoadModuleToRepoMap(t *testing.T) {
	workingDir := "/tmp/test"
	if err := mockLoadConfig(workingDir); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("module to repo map: %+v\n", ModuleToRepoMap)
}
