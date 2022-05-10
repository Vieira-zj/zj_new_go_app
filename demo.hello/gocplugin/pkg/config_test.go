package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"demo.hello/utils"
)

func mockInitConfig(workingDir string) error {
	subDir := "Workspaces/zj_repos/zj_go2_project/demo.hello/gocplugin/config"
	srcDir := filepath.Join(os.Getenv("HOME"), subDir)
	for _, file := range []string{"gocplugin.json", "module_repo_map.json"} {
		dstPath := filepath.Join(workingDir, file)
		if utils.IsExist(dstPath) {
			continue
		}
		srcPath := filepath.Join(srcDir, file)
		if err := utils.CopyFile(srcPath, dstPath); err != nil {
			return err
		}
	}

	if err := InitConfig(workingDir); err != nil {
		return err
	}
	return nil
}

func TestInitConfig(t *testing.T) {
	workingDir := "/tmp/test/goc_plugin_space"
	if err := mockInitConfig(workingDir); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("config: %+v\n", AppConfig)
	fmt.Printf("repo map: %+v\n", ModuleToRepoMap)
}

func TestSliceSpaceGrowth(t *testing.T) {
	s := make([]int, 0, 1)
	fmt.Printf("init: len=%d,cap=%d\n", len(s), cap(s))
	for i := 0; i < 3; i++ {
		s = append(s, i)
		fmt.Printf("put %d: len=%d,cap=%d\n", i, len(s), cap(s))
		fmt.Println(s)
	}
}

func TestDeleteMapItem(t *testing.T) {
	m := map[int]string{
		1: "one",
		2: "two",
	}

	delete(m, 1)
	delete(m, 3)
	fmt.Printf("map: %+v\n", m)
}
