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
	moduleDir := "/tmp/test/echoserver"
	workingDir := filepath.Join(moduleDir, "repo")
	covFile := "staging_th_apa_goc_echoserver_master_518e0a570c_127-0-0-1_20220325_181410.cov"
	covPath := filepath.Join(moduleDir, covFile)

	cmd := NewShCmd()
	output, err := cmd.GoToolCreateCoverHTMLReport(workingDir, covPath)
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

	param := SyncSrvCoverParam{
		SrvName: "staging_th_apa_goc_echoserver_master_845820727e",
	}
	mergeFileName := getSavedCovFileNameWithSuffix(param, "merge")
	mergeFilePath := filepath.Join(workingDir, mergeFileName)

	cmd := NewShCmd()
	if err := cmd.GocToolMergeSrvCovers(srcCovFiles, mergeFilePath); err != nil {
		t.Fatal(err)
	}
}
