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
	AppConfig.GocHost = testGocLocalHost
	if err := removeUnhealthSrvInGocTask(); err != nil {
		t.Fatal(err)
	}
	fmt.Println("done")
}

func TestGetSrvCoverTaskDeprecated(t *testing.T) {
	AppConfig.GocHost = testGocLocalHost
	savedDir := "/tmp/test/apa_goc_echoserver"
	param := SyncSrvCoverParam{
		SrvName:   "staging_th_apa_goc_echoserver_master_518e0a570c",
		Addresses: []string{"http://127.0.0.1:51007"},
	}
	savedPath, isUpdate, err := getSrvCoverTaskDeprecated(savedDir, param)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("saved=%s, updated=%v\n", savedPath, isUpdate)
}

func TestGetAndSaveSrvCover(t *testing.T) {
	AppConfig.GocHost = testGocLocalHost
	savedDir := "/tmp/test/apa_goc_echoserver"
	param := SyncSrvCoverParam{
		SrvName:   "staging_th_apa_goc_echoserver_master_518e0a570c",
		Addresses: []string{"http://127.0.0.1:51007"},
	}
	savedPath, err := getAndSaveSrvCover(savedDir, param)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("saved cover file to:", savedPath)
}

func TestGetSrvCoverTask(t *testing.T) {
	AppConfig.GocHost = testGocLocalHost
	param := SyncSrvCoverParam{
		SrvName:   "staging_th_apa_goc_echoserver_master_845820727e",
		Addresses: []string{"http://127.0.0.1:51007"},
	}
	savePath, isUpdate, err := getSrvCoverTask(param)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("save_path=%s, is_update=%v\n", savePath, isUpdate)
}

func TestSyncSrvRepo(t *testing.T) {
	// run: go test -timeout 300s -run ^TestSyncSrvRepo$ demo.hello/gocwatchdog/pkg -v -count=1
	if err := mockLoadConfig("/tmp/test"); err != nil {
		t.Fatal(err)
	}

	workingDir := "/tmp/test/apa_goc_echoserver/repo"
	param := SyncSrvCoverParam{
		SrvName:   "staging_th_apa_goc_echoserver_master_518e0a570c",
		Addresses: []string{"http://127.0.0.1:51007"},
	}
	if err := syncSrvRepo(workingDir, param); err != nil {
		t.Fatal(err)
	}
}

func TestCreateSrvCoverReportTask(t *testing.T) {
	// run: go test -timeout 300s -run ^TestCreateSrvCoverReportTask$ demo.hello/gocwatchdog/pkg -v -count=1
	AppConfig.RootDir = "/tmp/test"
	instance := NewGocSrvCoverDBInstance()

	srvName := "staging_th_apa_goc_echoserver_master_845820727e"
	meta := GetSrvMetaFromName(srvName)
	row, err := instance.GetLatestSrvCoverRow(meta)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("is_latest=%v, path=%s\n", row.IsLatest, row.CovFilePath)

	param := SyncSrvCoverParam{
		SrvName:   "staging_th_apa_goc_echoserver_master_845820727e",
		Addresses: []string{"http://127.0.0.1:51007"},
	}
	total, err := createSrvCoverReportTask(row.CovFilePath, param)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("last cover total: %s\n", total)
}

func TestGetSrvCoverAndCreateReportTask(t *testing.T) {
	// run: go test -timeout 300s -run ^TestGetSrvCoverAndCreateReportTask$ demo.hello/gocwatchdog/pkg -v -count=1
	if err := mockLoadConfig("/tmp/test"); err != nil {
		t.Fatal(err)
	}

	param := SyncSrvCoverParam{
		SrvName:   "staging_th_apa_goc_echoserver_master_845820727e",
		Addresses: []string{"http://127.0.0.1:51007"},
	}
	if err := GetSrvCoverAndCreateReportTask(param); err != nil {
		t.Fatal(err)
	}
}

//
// Helper Test
//

func TestGetSavedCovFileName(t *testing.T) {
	param := SyncSrvCoverParam{
		SrvName:   "staging_th_apa_goc_echoserver_master_518e0a570c",
		Addresses: []string{"http://127.0.0.1:51007"},
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
	meta := GetSrvMetaFromName(name)
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
