package pkg

import (
	"fmt"
	"testing"
)

func TestGetSimpleNowDatetime(t *testing.T) {
	fmt.Println("now:", getSimpleNowDatetime())
}

func TestGetFileNameWithoutExt(t *testing.T) {
	for _, fileName := range []string{"test.json", "sh_output.txt", "results"} {
		fmt.Println("name:", getFileNameWithoutExt(fileName))
	}
}

func TestFormatIPfromSrvAddress(t *testing.T) {
	addr := "http://127.0.0.1:49970"
	ip, err := formatIPfromSrvAddress(addr)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("ip:", ip)
}

func TestGetModuleFromSrvName(t *testing.T) {
	name := "staging_th_apa_goc_echoserver_origin/master_518e0a570c"
	mod, err := getModuleFromSrvName(name)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("module:", mod)
}

func TestGetBranchAndCommitFromSrvName(t *testing.T) {
	names := []string{
		"staging_th_apa_goc_echoserver_origin/master_518e0a570c",
		"staging_th_apa_goc_echoserver_master_518e0a",
	}

	for _, name := range names {
		br, commit := getBranchAndCommitFromSrvName(name)
		fmt.Printf("branch=%s, commitID=%s\n", br, commit)
	}
}
