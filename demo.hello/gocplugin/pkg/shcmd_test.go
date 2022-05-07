package pkg

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestNewSHCmdOnce(t *testing.T) {
	for i := 0; i < 3; i++ {
		NewShCmd()
	}
}

func TestRunCmd(t *testing.T) {
	sh := NewShCmd()
	res, err := sh.Run("cd ${HOME}/donwload; ls -l")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res)
}

func TestGetModuleNameFromFileName(t *testing.T) {
	fileName := "staging_th_apa_goc_echoserver_master_b63d82705a_20220507_173407.cov"
	fmt.Println("module name:", getModuleNameFromFileName(fileName))
}

func TestGoToolCreateCoverFuncReport(t *testing.T) {
	moduleDir := "/tmp/test/echoserver"
	workingDir := filepath.Join(moduleDir, "repo")
	covFile := "staging_th_apa_goc_echoserver_master_518e0a570c_127-0-0-1_20220325_181410.cov"
	covPath := filepath.Join(moduleDir, covFile)

	cmd := NewShCmd()
	output, err := cmd.GoToolCreateCoverFuncReport(workingDir, covPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(output) > 0 {
		fmt.Println("output:", output)
	}
}

func TestGoToolCreateCoverHTMLReport(t *testing.T) {
	if err := LoadConfig("/tmp/test/gocplugin.json"); err != nil {
		t.Fatal(err)
	}

	moduleName := "apa_goc_echoserver"
	moduleDir := filepath.Join(AppConfig.RootDir, moduleName)
	workingDir := filepath.Join(moduleDir, "repo")
	covFile := "staging_th_apa_goc_echoserver_master_b63d82705a_20220507_173407.cov"
	covPath := filepath.Join(moduleDir, "cover_data", covFile)

	cmd := NewShCmd()
	output, err := cmd.GoToolCreateCoverHTMLReport(workingDir, moduleName, covPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(output) > 0 {
		fmt.Println("output:", output)
	}
}

func TestGocToolMergeSrvCovers(t *testing.T) {
	workingDir := "/tmp/test/apa_goc_echoserver/cover_data"
	srcCovFiles := []string{
		filepath.Join(workingDir, "staging_th_apa_goc_echoserver_master_845820727e_20220420_154057.cov"),
		filepath.Join(workingDir, "staging_th_apa_goc_echoserver_master_845820727e_20220420_154143.cov"),
	}

	srvName := "staging_th_apa_goc_echoserver_master_845820727e"
	mergeFileName := getSavedCovFileNameWithSuffix(srvName, "merge")
	mergeFilePath := filepath.Join(workingDir, mergeFileName)

	cmd := NewShCmd()
	if err := cmd.GocToolMergeSrvCovers(srcCovFiles, mergeFilePath); err != nil {
		t.Fatal(err)
	}
}
