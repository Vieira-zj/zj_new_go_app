package pkg

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"
)

func TestInsertSrvCoverRow(t *testing.T) {
	srvName := "staging_th_apa_goc_echoserver_master_518e0a570d"
	meta := getSrvMetaFromName(srvName)
	meta.IPAddr = "http://127.0.0.1:51007"
	row := GocSrvCoverModel{
		SrvCoverMeta: meta,
		IsLatest:     true,
		CovFilePath:  "/tmp/test/profile2.cov",
		CoverTotal: sql.NullFloat64{
			Float64: 37.92,
			Valid:   true,
		},
	}

	AppConfig.RootDir = "/tmp/test"
	instance := NewGocSrvCoverDBInstance()
	if err := instance.InsertSrvCoverRow(row); err != nil {
		t.Fatal(err)
	}
}

func TestGetLatestSrvCoverRow(t *testing.T) {
	AppConfig.RootDir = "/tmp/test"
	instance := NewGocSrvCoverDBInstance()

	srvName := "staging_th_apa_goc_echoserver_master_518e0a570c"
	meta := getSrvMetaFromName(srvName)
	row, err := instance.GetLatestSrvCoverRow(meta)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("row: id=%d, is_latest=%v, cover_total=%.2f\n", row.ID, row.IsLatest, row.CoverTotal.Float64)
}

func TestGetLatestSrvCoverRowNotFound(t *testing.T) {
	AppConfig.RootDir = "/tmp/test"
	instance := NewGocSrvCoverDBInstance()

	srvName := "staging_vn_apa_goc_echoserver_master_518e0a570c"
	meta := getSrvMetaFromName(srvName)
	_, err := instance.GetLatestSrvCoverRow(meta)
	fmt.Println("error:", err)
	if !errors.Is(err, ErrSrvCoverLatestRowNotFound) {
		t.Fatal("assert err is ErrSrvCoverLatestRowNotFound failed")
	}
}

func TestGetLatestSrvCoverRowMoreThanOneErr(t *testing.T) {
	AppConfig.RootDir = "/tmp/test"
	instance := NewGocSrvCoverDBInstance()

	srvName := "staging_th_apa_goc_echoserver_master_518e0a570c"
	meta := getSrvMetaFromName(srvName)
	_, err := instance.GetLatestSrvCoverRow(meta)
	fmt.Println("error:", err)
	if !errors.Is(err, ErrMoreThanOneSrvCoverLatestRow) {
		t.Fatal("assert err is ErrMoreThanOneSrvCoverLatestRow failed")
	}
}

func TestUpdateLatestRowToFalse(t *testing.T) {
	AppConfig.RootDir = "/tmp/test"
	instance := NewGocSrvCoverDBInstance()

	srvName := "staging_th_apa_goc_echoserver_master_518e0a570c"
	meta := getSrvMetaFromName(srvName)
	if err := instance.UpdateLatestSrvCoverRowToFalse(meta); err != nil {
		t.Fatal(err)
	}
}

func TestAddLatestSrvCoverRow(t *testing.T) {
	AppConfig.RootDir = "/tmp/test"
	instance := NewGocSrvCoverDBInstance()

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

	if err := instance.AddLatestSrvCoverRow(row); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateCovFileOfLatestSrvCoverRow(t *testing.T) {
	AppConfig.RootDir = "/tmp/test"
	instance := NewGocSrvCoverDBInstance()

	srvName := "staging_th_apa_goc_echoserver_master_518e0a570c"
	meta := getSrvMetaFromName(srvName)
	covFilePath := "/tmp/test/profile3x.cov"
	if err := instance.UpdateCovFileOfLatestSrvCoverRow(meta, covFilePath); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateSrvCoverTotal(t *testing.T) {
	AppConfig.RootDir = "/tmp/test"
	instance := NewGocSrvCoverDBInstance()

	srvName := "staging_th_apa_goc_echoserver_master_518e0a570c"
	meta := getSrvMetaFromName(srvName)
	if err := instance.UpdateSrvCoverTotalByCommit(61.59, meta.GitCommit); err != nil {
		t.Fatal(err)
	}
}
