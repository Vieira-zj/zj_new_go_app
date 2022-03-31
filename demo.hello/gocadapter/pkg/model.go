package pkg

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SrvCoverMeta .
type SrvCoverMeta struct {
	Env       string
	Region    string
	AppName   string
	IPAddr    string
	GitBranch string
	GitCommit string
}

// GocSrvCoverModel .
type GocSrvCoverModel struct {
	gorm.Model
	SrvCoverMeta
	IsLatest    bool
	CovFilePath string
	CoverTotal  sql.NullFloat64
}

// TableName .
func (GocSrvCoverModel) TableName() string {
	return "goc_o_staging_service_cover"
}

// GocSrvCoverDBInstance .
type GocSrvCoverDBInstance struct {
	sqliteDB *gorm.DB
}

var (
	instance  GocSrvCoverDBInstance
	modelOnce sync.Once
)

// NewGocSrvCoverDBInstance .
func NewGocSrvCoverDBInstance(moduleDir string) GocSrvCoverDBInstance {
	modelOnce.Do(func() {
		dbPath := filepath.Join(moduleDir, "sqlite.db")
		sqliteDB, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
		if err != nil {
			panic("NewGocSrvCoverDBInstance open db error: " + err.Error())
		}

		if err := sqliteDB.AutoMigrate(&GocSrvCoverModel{}); err != nil {
			panic("NewGocSrvCoverDBInstance db migrate error: " + err.Error())
		}

		instance = GocSrvCoverDBInstance{
			sqliteDB: sqliteDB,
		}
	})
	return instance
}

// InsertSrvCoverRow .
func (in *GocSrvCoverDBInstance) InsertSrvCoverRow(row GocSrvCoverModel) error {
	return in.insertSrvCoverRowByDB(in.sqliteDB, row)
}

func (*GocSrvCoverDBInstance) insertSrvCoverRowByDB(db *gorm.DB, row GocSrvCoverModel) error {
	if result := db.Create(&row); result.Error != nil {
		return fmt.Errorf("InsertSrvCoverRow error: %w", result.Error)
	}
	return nil
}

// GetLatestSrvCoverRow .
func (in *GocSrvCoverDBInstance) GetLatestSrvCoverRow(meta SrvCoverMeta) ([]GocSrvCoverModel, error) {
	return in.getLatestSrvCoverRowByDB(in.sqliteDB, meta)
}

func (*GocSrvCoverDBInstance) getLatestSrvCoverRowByDB(db *gorm.DB, meta SrvCoverMeta) ([]GocSrvCoverModel, error) {
	var rows []GocSrvCoverModel
	condition := "env = ? and region = ? and app_name = ? and is_latest = ?"
	if result := db.Where(condition, meta.Env, meta.Region, meta.AppName, true).Find(&rows); result.Error != nil {
		return nil, fmt.Errorf("GetLatestSrvCoverRow find error: %w", result.Error)
	}
	return rows, nil
}

// UpdateSrvCoverTotalByCommit .
func (in *GocSrvCoverDBInstance) UpdateSrvCoverTotalByCommit(total float64, commit string) error {
	data := GocSrvCoverModel{
		CoverTotal: sql.NullFloat64{
			Float64: total,
			Valid:   true,
		},
	}
	if result := in.sqliteDB.Model(&GocSrvCoverModel{}).Where("git_commit = ? and is_latest = ?", commit, true).Updates(data); result.Error != nil {
		return fmt.Errorf("UpdateSrvCoverTotal update row error: %w", result.Error)
	}
	return nil
}

// UpdateLatestSrvCoverRowToFalse .
func (in *GocSrvCoverDBInstance) UpdateLatestSrvCoverRowToFalse(meta SrvCoverMeta) error {
	return in.updateLatestSrvCoverRowToFalseByDB(in.sqliteDB, meta)
}

func (in *GocSrvCoverDBInstance) updateLatestSrvCoverRowToFalseByDB(db *gorm.DB, meta SrvCoverMeta) error {
	rows, err := in.GetLatestSrvCoverRow(meta)
	if err != nil {
		return fmt.Errorf("UpdateLatestRowToFalse error: %w", err)
	}
	if len(rows) != 1 {
		return fmt.Errorf("UpdateLatestRowToFalse latest rows count not equal to 1")
	}

	data := map[string]interface{}{
		"IsLatest": false,
	}
	if result := db.Model(&GocSrvCoverModel{}).Where("id = ?", rows[0].ID).Updates(data); result.Error != nil {
		return fmt.Errorf("UpdateLatestRowToFalse update row error: %w", result.Error)
	}
	return nil
}

// InsertLatestSrvCoverRow .
func (in *GocSrvCoverDBInstance) InsertLatestSrvCoverRow(row GocSrvCoverModel) error {
	transaction := in.sqliteDB.Begin()
	if err := in.updateLatestSrvCoverRowToFalseByDB(transaction, row.SrvCoverMeta); err != nil {
		transaction.Rollback()
		return fmt.Errorf("InsertLatestSrvCoverRow error: %w", err)
	}

	if err := in.insertSrvCoverRowByDB(transaction, row); err != nil {
		transaction.Rollback()
		return fmt.Errorf("InsertLatestSrvCoverRow error: %w", err)
	}

	if result := transaction.Commit(); result.Error != nil {
		return fmt.Errorf("InsertLatestSrvCoverRow transaction commit error: %w", result.Error)
	}
	return nil
}
