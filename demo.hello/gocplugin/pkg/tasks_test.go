package pkg

import (
	"fmt"
	"testing"
)

func TestGetSrvModuleDir(t *testing.T) {
	AppConfig.RootDir = "/tmp/test"
	srvName := "staging_th_apa_goc_echoserver_master_b63d82705a"
	fmt.Println(GetSrvModuleDir(srvName))
}

func TestIsAttachSrvOK(t *testing.T) {
	host := "http://127.0.0.1:51025"
	ok := isAttachSrvOK(host)
	fmt.Println("service ok:", ok)
}

func TestRemoveUnhealthSrvInGocTask(t *testing.T) {
	AppConfig.GocHost = testGocLocalHost
	if err := RemoveUnhealthSrvInGocTask(); err != nil {
		t.Fatal(err)
	}
	fmt.Println("unhealth check done")
}

func TestFetchAndSaveSrvCover(t *testing.T) {
	AppConfig.GocHost = testGocLocalHost
	savedDir := "/tmp/test/apa_goc_echoserver"
	param := SyncSrvCoverParam{
		SrvName:   "staging_th_apa_goc_echoserver_master_518e0a570c",
		Addresses: []string{"http://127.0.0.1:51007"},
	}
	savedPath, err := FetchAndSaveSrvCover(savedDir, param.SrvName)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("saved cover file to:", savedPath)
}

func TestGetSrvCoverTask(t *testing.T) {
	AppConfig.GocHost = testGocLocalHost
	srvName := "staging_th_apa_goc_echoserver_master_845820727e"
	savePath, isUpdate, err := getSrvCoverTask(srvName)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("save_path=%s, is_update=%v\n", savePath, isUpdate)
}

func TestCheckoutSrvRepo(t *testing.T) {
	// run: go test -timeout 300s -run ^TestSyncSrvRepo$ demo.hello/gocplugin/pkg -v -count=1
	if err := mockLoadConfig("/tmp/test"); err != nil {
		t.Fatal(err)
	}

	workingDir := "/tmp/test/apa_goc_echoserver/repo"
	srvName := "staging_th_apa_goc_echoserver_master_845820727e"
	if err := checkoutSrvRepo(workingDir, srvName); err != nil {
		t.Fatal(err)
	}
}

func TestCreateSrvCoverReportTask(t *testing.T) {
	// run: go test -timeout 300s -run ^TestCreateSrvCoverReportTask$ demo.hello/gocplugin/pkg -v -count=1
	AppConfig.RootDir = "/tmp/test"
	instance := NewGocSrvCoverDBInstance()

	srvName := "staging_th_apa_goc_echoserver_master_845820727e"
	meta := GetSrvMetaFromName(srvName)
	row, err := instance.GetLatestSrvCoverRow(meta)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("is_latest=%v, path=%s\n", row.IsLatest, row.CovFilePath)

	total, err := createSrvCoverReportTask(row.CovFilePath, srvName)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("last cover total: %s\n", total)
}

func TestGetSrvCoverAndCreateReportTask(t *testing.T) {
	// run: go test -timeout 300s -run ^TestGetSrvCoverAndCreateReportTask$ demo.hello/gocplugin/pkg -v -count=1
	if err := mockLoadConfig("/tmp/test"); err != nil {
		t.Fatal(err)
	}

	param := SyncSrvCoverParam{
		SrvName:   "staging_th_apa_goc_echoserver_master_845820727e",
		Addresses: []string{"http://127.0.0.1:51007"},
	}
	total, err := GetSrvCoverAndCreateReportTask(param)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("cover total:", total)
}

//
// Helper Test
//

func TestGetSavedCovFileNameWithSuffix(t *testing.T) {
	if err := mockLoadConfig("/tmp/test"); err != nil {
		t.Fatal(err)
	}

	srvName := "staging_th_apa_goc_echoserver_master_518e0a570c"
	name := getSavedCovFileNameWithSuffix(srvName, "")
	fmt.Println("saved file name:", name)
}

func TestGetSrvMetaFromName(t *testing.T) {
	name := "staging_th_apa_goc_echoserver_master_518e0a570c"
	meta := GetSrvMetaFromName(name)
	fmt.Printf("service meta data: %+v\n", meta)
}
