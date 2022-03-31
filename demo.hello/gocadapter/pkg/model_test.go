package pkg

import (
	"database/sql"
	"fmt"
	"testing"
)

func TestInsertSrvCoverRow(t *testing.T) {
	srvName := "staging_th_apa_goc_echoserver_master_518e0a570c"
	meta := getSrvMetaFromName(srvName)
	meta.IPAddr = "http://127.0.0.1:51007"
	row := GocSrvCoverModel{
		SrvCoverMeta: meta,
		IsLatest:     true,
		CovFilePath:  "/tmp/test/profile1.cov",
		CoverTotal: sql.NullFloat64{
			Float64: 21.92,
			Valid:   true,
		},
	}

	instance := NewGocSrvCoverDBInstance("/tmp/test")
	if err := instance.InsertSrvCoverRow(row); err != nil {
		t.Fatal(err)
	}
}

func TestGetLatestSrvCoverRow(t *testing.T) {
	srvName := "staging_th_apa_goc_echoserver_master_518e0a570c"
	meta := getSrvMetaFromName(srvName)
	instance := NewGocSrvCoverDBInstance("/tmp/test")
	rows, err := instance.GetLatestSrvCoverRow(meta)
	if err != nil {
		t.Fatal(err)
	}

	for _, row := range rows {
		value, err := row.CoverTotal.Value()
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("row: id=%d, is_latest=%v, cover_total=%v\n", row.ID, row.IsLatest, value)
	}
}

func TestUpdateLatestRowToFalse(t *testing.T) {
	srvName := "staging_th_apa_goc_echoserver_master_518e0a570c"
	meta := getSrvMetaFromName(srvName)
	instance := NewGocSrvCoverDBInstance("/tmp/test")
	if err := instance.UpdateLatestSrvCoverRowToFalse(meta); err != nil {
		t.Fatal(err)
	}
}

func TestInsertLatestSrvCoverRow(t *testing.T) {
	srvName := "staging_th_apa_goc_echoserver_master_518e0a570d"
	meta := getSrvMetaFromName(srvName)
	meta.IPAddr = "http://127.0.0.1:51007"
	row := GocSrvCoverModel{
		SrvCoverMeta: meta,
		IsLatest:     true,
		CovFilePath:  "/tmp/test/profile2.cov",
		CoverTotal: sql.NullFloat64{
			Float64: 27.33,
			Valid:   true,
		},
	}

	instance := NewGocSrvCoverDBInstance("/tmp/test")
	if err := instance.InsertLatestSrvCoverRow(row); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateSrvCoverTotal(t *testing.T) {
	srvName := "staging_th_apa_goc_echoserver_master_518e0a570c"
	meta := getSrvMetaFromName(srvName)
	instance := NewGocSrvCoverDBInstance("/tmp/test")
	if err := instance.UpdateSrvCoverTotalByCommit(31.59, meta.GitCommit); err != nil {
		t.Fatal(err)
	}
}
