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
