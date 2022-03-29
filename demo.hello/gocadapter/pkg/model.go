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
	Env     string
	Region  string
	AppName string
	Branch  string
	Commit  string
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
func NewGocSrvCoverDBInstance() GocSrvCoverDBInstance {
	modelOnce.Do(func() {
		dbPath := filepath.Join(AppConfig.RootDir, "sqlite.db")
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
	if result := in.sqliteDB.Create(&row); result.Error != nil {
		return fmt.Errorf("InsertSrvCoverRow error: %w", result.Error)
	}
	return nil
}

// GetLatestSrvCoverRow .
func (in *GocSrvCoverDBInstance) GetLatestSrvCoverRow(meta SrvCoverMeta) ([]GocSrvCoverModel, error) {
	var rows []GocSrvCoverModel
	condition := "env = ? and region = ? and app_name = ? and is_latest = ?"
	if result := in.sqliteDB.Where(condition, meta.Env, meta.Region, meta.AppName, true).Find(&rows); result.Error != nil {
		return nil, fmt.Errorf("GetLatestSrvCoverRow find error: %w", result.Error)
	}
	return rows, nil
}

// UpdateLatestRowToFalse .
func (in *GocSrvCoverDBInstance) UpdateLatestRowToFalse(meta SrvCoverMeta) error {
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
	if result := in.sqliteDB.Model(&GocSrvCoverModel{}).Where("id = ?", rows[0].ID).Updates(data); result.Error != nil {
		return fmt.Errorf("UpdateLatestRowToFalse update error: %w", result.Error)
	}
	return nil
}
