package pkg

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
)

func TestGetSrvModuleDir(t *testing.T) {
	AppConfig.RootDir = "/tmp/test"
	srvName := "staging_th_apa_goc_echoserver_master_b63d82705a"
	fmt.Println(GetSrvModuleDir(srvName))
}

func TestIsAttachSrvOK(t *testing.T) {
	host := "http://127.0.0.1:53075"
	if ok, err := isAttachSrvOK(host); ok {
		fmt.Println("service ok:", ok)
	} else {
		fmt.Println("service is not ok:", err)
	}
}

func TestIsPodOK(t *testing.T) {
	if err := InitConfig("/tmp/test/goc_staging_space"); err != nil {
		t.Fatal(err)
	}

	ok, err := isPodOK("http://100.79.49.20:32911")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("pod ok:", ok)

	ok, err = isPodOK("100.79.56.6")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("pod ok:", ok)
}

func TestIsSrvOK(t *testing.T) {
	if err := InitConfig("/tmp/test/goc_staging_space"); err != nil {
		t.Fatal(err)
	}

	addr := "http://100.79.49.20:32911"
	ok, err := isSrvOK(addr)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("service ok:", ok)
}

func TestRemoveUnhealthSrvInGocTask(t *testing.T) {
	if err := InitConfig("/tmp/test/goc_staging_space"); err != nil {
		t.Fatal(err)
	}

	if err := RemoveUnhealthSrvInGocTask(); err != nil {
		t.Fatal(err)
	}
	fmt.Println("unhealth check done")
}

func TestSyncAndListRegisterSrvsTask(t *testing.T) {
	if err := InitConfig("/tmp/test/goc_staging_space"); err != nil {
		t.Fatal(err)
	}

	srvs, err := SyncAndListRegisterSrvsTask()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("register services info: %+v\n", srvs)
}

func TestFetchAndSaveSrvCover(t *testing.T) {
	AppConfig.GocCenterIngHost = testGocLocalHost
	savedDir := "/tmp/test/goc_space"
	param := SyncSrvCoverParam{
		SrvName: "staging_th_apa_goc_echoserver_master_845820727e",
	}
	savedPath, err := FetchAndSaveSrvCover(savedDir, param.SrvName)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("saved cover file to:", savedPath)
}

func TestGetSrvCoverTask(t *testing.T) {
	AppConfig.GocCenterIngHost = testGocLocalHost
	srvName := "staging_th_apa_goc_echoserver_master_845820727e"
	savePath, isUpdate, err := getSrvCoverTask(srvName)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("save_path=%s, is_update=%v\n", savePath, isUpdate)
}

func TestCheckoutSrvRepo(t *testing.T) {
	// run: go test -timeout 300s -run ^TestCheckoutSrvRepo$ demo.hello/gocplugin/pkg -v -count=1
	if err := InitConfig("/tmp/test/goc_plugin_space"); err != nil {
		t.Fatal(err)
	}

	workingDir := "/tmp/test/goc_plugin_space/apa_echoserver_goc/repo"
	srvName := "staging_th_apa_echoserver_goc_master_b63d82705a"
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
	if err := InitConfig("/tmp/test"); err != nil {
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

func TestRenameLastSrvCoverResults(t *testing.T) {
	if err := InitConfig("/tmp/test/goc_staging_space"); err != nil {
		t.Fatal(err)
	}

	dir := "/tmp/test/goc_staging_space/apa_echoserver/cover_data"
	src := filepath.Join(dir, "staging_th_apa_echoserver_master_c32684d0b1_20220517_121525.cov")
	dst := filepath.Join(dir, "staging_th_apa_echoserver_master_c32684d0b1_20220517_121525x.cov")
	if err := renameLastSrvCoverResults(src, dst); err != nil {
		t.Fatal(err)
	}

	if err := renameLastSrvHTMLCoverResults("apa_echoserver", src, dst); err != nil {
		t.Fatal(err)
	}
}

//
// Helper Test
//

func TestGetSavedCovFileNameWithSuffix(t *testing.T) {
	if err := InitConfig("/tmp/test"); err != nil {
		t.Fatal(err)
	}

	srvName := "staging_th_apa_goc_echoserver_master_518e0a570c"
	name := getSavedCovFileNameWithSuffix(srvName, "")
	fmt.Println("saved file name:", name)
}

func TestGetSrvMetaFromName(t *testing.T) {
	for _, name := range [2]string{
		"staging_th_apa_goc_echoserver_master_518e0a570c",
		"staging_th_is_is-giro_staging_for_cover_8e370087c3b"} {
		meta := GetSrvMetaFromName(name)
		fmt.Printf("service meta data: %+v\n", meta)
	}
}

func TestPodStatusRespUnmarshal(t *testing.T) {
	body := `
	{
		"total": 1,
		"data": [
		  {
			"namespace": "test",
			"name": "echoserver-557d6f987f-r8r9s",
			"ip": "100.79.56.4",
			"value": "Running"
		  }
		]
	}`

	status := podStatusResp{}
	if err := json.Unmarshal([]byte(body), &status); err != nil {
		t.Fatal(err)
	}
	fmt.Println(status.Total)
	for _, item := range status.Data {
		fmt.Printf("pod status: %+v\n", item)
	}
}
