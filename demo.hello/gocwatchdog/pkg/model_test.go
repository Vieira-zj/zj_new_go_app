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

	fmt.Println("rows count:", len(rows))
	for _, row := range rows {
		fmt.Printf("row: id=%d, is_latest=%v, cover_total=%.2f\n", row.ID, row.IsLatest, row.CoverTotal.Float64)
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

func TestAddLatestSrvCoverRow(t *testing.T) {
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
	if err := instance.AddLatestSrvCoverRow(row); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateCovFileOfLatestSrvCoverRow(t *testing.T) {
	srvName := "staging_th_apa_goc_echoserver_master_518e0a570c"
	meta := getSrvMetaFromName(srvName)
	instance := NewGocSrvCoverDBInstance("/tmp/test")
	covFilePath := "/tmp/test/profile3x.cov"
	if err := instance.UpdateCovFileOfLatestSrvCoverRow(meta, covFilePath); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateSrvCoverTotal(t *testing.T) {
	srvName := "staging_th_apa_goc_echoserver_master_518e0a570c"
	meta := getSrvMetaFromName(srvName)
	instance := NewGocSrvCoverDBInstance("/tmp/test")
	if err := instance.UpdateSrvCoverTotalByCommit(61.59, meta.GitCommit); err != nil {
		t.Fatal(err)
	}
}
