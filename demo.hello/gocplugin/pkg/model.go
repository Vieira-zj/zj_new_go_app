package pkg

import (
	"database/sql"
	"errors"
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
	Addrs     string `gorm:"column:addresses"`
	GitBranch string
	GitCommit string
}

// GocSrvCoverModel .
type GocSrvCoverModel struct {
	gorm.Model
	SrvCoverMeta
	IsLatest    bool
	CovFilePath string
	CoverTotal  sql.NullString
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
		const dbFile = "sqlite.db"
		dbPath := filepath.Join(AppConfig.RootDir, dbFile)
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
		return fmt.Errorf("insertSrvCoverRowByDB error: %w", result.Error)
	}
	return nil
}

var (
	// ErrSrvCoverLatestRowNotFound .
	ErrSrvCoverLatestRowNotFound = errors.New("ErrSrvCoverLatestRowNotFound")
	// ErrMoreThanOneSrvCoverLatestRow .
	ErrMoreThanOneSrvCoverLatestRow = errors.New("ErrMoreThanOneSrvCoverLatestRow")
)

// GetLatestSrvCoverRow .
func (in *GocSrvCoverDBInstance) GetLatestSrvCoverRow(meta SrvCoverMeta) (GocSrvCoverModel, error) {
	return in.GetLatestSrvCoverRowByDB(in.sqliteDB, meta)
}

// GetLatestSrvCoverRowByDB .
func (in *GocSrvCoverDBInstance) GetLatestSrvCoverRowByDB(db *gorm.DB, meta SrvCoverMeta) (GocSrvCoverModel, error) {
	rows, err := in.getLatestSrvCoverRowsByDB(in.sqliteDB, meta)
	if err != nil {
		return GocSrvCoverModel{}, fmt.Errorf("GetLatestSrvCoverRowByDB error: %w", err)
	}

	if len(rows) == 0 {
		return GocSrvCoverModel{}, fmt.Errorf("GetLatestSrvCoverRowByDB error: %w", ErrSrvCoverLatestRowNotFound)
	}
	if len(rows) != 1 {
		return GocSrvCoverModel{}, fmt.Errorf("GetLatestSrvCoverRowByDB error: %w", ErrMoreThanOneSrvCoverLatestRow)
	}
	return rows[0], nil
}

func (*GocSrvCoverDBInstance) getLatestSrvCoverRowsByDB(db *gorm.DB, meta SrvCoverMeta) ([]GocSrvCoverModel, error) {
	var rows []GocSrvCoverModel
	condition := "env = ? and region = ? and app_name = ? and git_commit = ? and is_latest = ?"
	if result := db.Where(condition, meta.Env, meta.Region, meta.AppName, meta.GitCommit, true).Find(&rows); result.Error != nil {
		return nil, fmt.Errorf("getLatestSrvCoverRowByDB find error: %w", result.Error)
	}
	return rows, nil
}

// UpdateLatestSrvCoverRowToFalse .
func (in *GocSrvCoverDBInstance) UpdateLatestSrvCoverRowToFalse(meta SrvCoverMeta) error {
	return in.updateLatestSrvCoverRowToFalseByDB(in.sqliteDB, meta)
}

func (in *GocSrvCoverDBInstance) updateLatestSrvCoverRowToFalseByDB(db *gorm.DB, meta SrvCoverMeta) error {
	row, err := in.GetLatestSrvCoverRow(meta)
	if err != nil {
		return fmt.Errorf("updateLatestSrvCoverRowToFalseByDB error: %w", err)
	}

	data := map[string]interface{}{
		"IsLatest": false,
	}
	if result := db.Model(&GocSrvCoverModel{}).Where("id = ?", row.ID).Updates(data); result.Error != nil {
		return fmt.Errorf("updateLatestSrvCoverRowToFalseByDB update row error: %w", result.Error)
	}
	return nil
}

// UpdateCovFileOfLatestSrvCoverRow .
func (in *GocSrvCoverDBInstance) UpdateCovFileOfLatestSrvCoverRow(meta SrvCoverMeta, covFilePath string) error {
	return in.UpdateCovFileOfLatestSrvCoverRowByDB(in.sqliteDB, meta, covFilePath)
}

// UpdateCovFileOfLatestSrvCoverRowByDB .
func (in *GocSrvCoverDBInstance) UpdateCovFileOfLatestSrvCoverRowByDB(db *gorm.DB, meta SrvCoverMeta, covFilePath string) error {
	row, err := in.GetLatestSrvCoverRowByDB(db, meta)
	if err != nil {
		return fmt.Errorf("UpdateCovFileOfLatestSrvCoverRow error: %w", err)
	}

	data := GocSrvCoverModel{
		CovFilePath: covFilePath,
	}
	if result := db.Model(&GocSrvCoverModel{}).Where("id = ?", row.ID).Updates(data); result.Error != nil {
		return fmt.Errorf("updateLatestSrvCoverRowToFalseByDB update row error: %w", result.Error)
	}
	return nil
}

// UpdateSrvCoverTotalByCommit .
func (in *GocSrvCoverDBInstance) UpdateSrvCoverTotalByCommit(total, commit string) error {
	data := GocSrvCoverModel{
		CoverTotal: sql.NullString{
			String: total,
			Valid:  true,
		},
	}
	if result := in.sqliteDB.Model(&GocSrvCoverModel{}).Where("git_commit = ? and is_latest = ?", commit, true).Updates(data); result.Error != nil {
		return fmt.Errorf("UpdateSrvCoverTotalByCommit update row error: %w", result.Error)
	}
	return nil
}

// AddLatestSrvCoverRow adds latest service cover row with transaction.
func (in *GocSrvCoverDBInstance) AddLatestSrvCoverRow(row GocSrvCoverModel) error {
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

// BeginTransaction .
func (in *GocSrvCoverDBInstance) BeginTransaction() *gorm.DB {
	return in.sqliteDB.Begin()
}
