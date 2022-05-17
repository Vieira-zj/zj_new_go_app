package pkg

import (
	"testing"
)

// go test -timeout 300s -run ^TestSyncRegisterSrvsCoverReport$ demo.hello/gocplugin/pkg -v -count=1
func TestSyncRegisterSrvsCoverReport(t *testing.T) {
	if err := InitConfig("/tmp/test/goc_staging_space"); err != nil {
		t.Fatal(err)
	}

	InitSrvCoverSyncTasksPool()
	if err := syncRegisterSrvsCoverReport(); err != nil {
		t.Fatal(err)
	}
}
