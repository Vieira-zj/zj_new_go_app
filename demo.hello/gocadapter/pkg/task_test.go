package pkg

import (
	"fmt"
	"testing"
)

func TestIsAttachServerOK(t *testing.T) {
	host := "http://127.0.0.1:51025"
	ok := isAttachServerOK(host)
	fmt.Println("service ok:", ok)
}

func TestRemoveUnhealthSrvInGocTask(t *testing.T) {
	if err := removeUnhealthSrvInGocTask(testGocLocalHost); err != nil {
		t.Fatal(err)
	}
	fmt.Println("done")
}

func TestSyncSrvRepo(t *testing.T) {
	// run: go test -timeout 300s -run ^TestSyncSrvRepo$ demo.hello/gocadapter/pkg -v -count=1
	if err := mockLoadConfig("/tmp/test"); err != nil {
		t.Fatal(err)
	}

	workingDir := "/tmp/test/echoserver/repo"
	param := SyncSrvCoverParam{
		SrvName: "staging_th_apa_goc_echoserver_master_a6023e5e",
		Address: "http://127.0.0.1:51007",
	}
	if err := syncSrvRepo(workingDir, param); err != nil {
		t.Fatal(err)
	}
}

func TestGetSrvCoverTaskDeprecated(t *testing.T) {
	savedDir := "/tmp/test/echoserver"
	param := SyncSrvCoverParam{
		GocHost: testGocLocalHost,
		SrvName: "staging_th_apa_goc_echoserver_master_518e0a570c",
		Address: "http://127.0.0.1:51007",
	}
	savedPath, isUpdate, err := getSrvCoverTaskDeprecated(savedDir, param)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("saved=%s, updated=%v\n", savedPath, isUpdate)
}

func TestGetAndSaveSrvCover(t *testing.T) {
	savedDir := "/tmp/test/echoserver"
	param := SyncSrvCoverParam{
		GocHost: testGocLocalHost,
		SrvName: "staging_th_apa_goc_echoserver_master_518e0a570c",
		Address: "http://127.0.0.1:51007",
	}
	savedPath, err := getAndSaveSrvCover(savedDir, param)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("saved cover file to:", savedPath)
}

//
// Helper Test
//

func TestGetSavedCovFileName(t *testing.T) {
	param := SyncSrvCoverParam{
		SrvName: "staging_th_apa_goc_echoserver_master_518e0a570c",
		Address: "http://127.0.0.1:49970",
	}

	if err := mockLoadConfig("/tmp/test"); err != nil {
		t.Fatal(err)
	}

	name, err := getSavedCovFileName(param)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("saved file name:", name)
}

func TestGetModuleFromSrvName(t *testing.T) {
	if err := mockLoadConfig("/tmp/test"); err != nil {
		t.Fatal(err)
	}

	name := "staging_th_apa_goc_echoserver_master_518e0a570c"
	mod, err := getModuleFromSrvName(name)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("module:", mod)
}

func TestGetSrvMetaFromName(t *testing.T) {
	name := "staging_th_apa_goc_echoserver_master_518e0a570c"
	meta := getSrvMetaFromName(name)
	fmt.Printf("service meta data: %+v\n", meta)
}

func TestGetBranchAndCommitFromSrvName(t *testing.T) {
	name := "staging_th_apa_goc_echoserver_master_518e0a570c"
	br, commit := getBranchAndCommitFromSrvName(name)
	fmt.Printf("branch=%s, commitID=%s\n", br, commit)
}

func TestGetRepoURLFromSrvName(t *testing.T) {
	if err := mockLoadConfig("/tmp/test"); err != nil {
		t.Fatal(err)
	}

	name := "staging_th_apa_goc_echoserver_master_518e0a570c"
	url, err := getRepoURLFromSrvName(name)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("url:", url)
}
