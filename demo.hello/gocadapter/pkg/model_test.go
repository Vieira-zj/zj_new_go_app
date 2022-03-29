package pkg

import (
	"database/sql"
	"fmt"
	"testing"
)

func TestInsertSrvCoverRow(t *testing.T) {
	if err := mockLoadConfig("/tmp/test"); err != nil {
		t.Fatal(err)
	}

	srvName := "staging_th_apa_goc_echoserver_master_518e0a570c"
	meta := getSrvMetaFromName(srvName)
	row := GocSrvCoverModel{
		SrvCoverMeta: meta,
		IsLatest:     true,
		CovFilePath:  "/tmp/test/profile3.cov",
		CoverTotal: sql.NullFloat64{
			Float64: 21.01,
			Valid:   true,
		},
	}

	instance := NewGocSrvCoverDBInstance()
	if err := instance.InsertSrvCoverRow(row); err != nil {
		t.Fatal(err)
	}
}

func TestGetLatestSrvCoverRow(t *testing.T) {
	if err := mockLoadConfig("/tmp/test"); err != nil {
		t.Fatal(err)
	}

	srvName := "staging_th_apa_goc_echoserver_master_518e0a570c"
	meta := getSrvMetaFromName(srvName)
	instance := NewGocSrvCoverDBInstance()
	rows, err := instance.GetLatestSrvCoverRow(meta)
	if err != nil {
		t.Fatal(err)
	}

	for _, row := range rows {
		fmt.Printf("row: %+v\n", row)
	}
}

func TestUpdateLatestRowToFalse(t *testing.T) {
	if err := mockLoadConfig("/tmp/test"); err != nil {
		t.Fatal(err)
	}

	srvName := "staging_th_apa_goc_echoserver_master_518e0a570c"
	meta := getSrvMetaFromName(srvName)
	instance := NewGocSrvCoverDBInstance()
	if err := instance.UpdateLatestRowToFalse(meta); err != nil {
		t.Fatal(err)
	}
}
