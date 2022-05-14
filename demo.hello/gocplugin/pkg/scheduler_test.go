package pkg

import (
	"testing"
)

func TestSyncRegisterSrvsCoverReport(t *testing.T) {
	if err := InitConfig("/tmp/test/goc_staging_space"); err != nil {
		t.Fatal(err)
	}
	// TODO:
}
