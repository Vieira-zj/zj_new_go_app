package gorm

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"go1_1711_demo/middlewares/gorm/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	mysqlDB     *gorm.DB
	mysqlDBOnce sync.Once
)

func NewDB() *gorm.DB {
	mysqlDBOnce.Do(func() {
		mysqlDB = initDB()
	})
	return mysqlDB
}

// Init DB

type DbConfig struct {
	User     string
	Password string
	Host     string
	Port     int
	Database string
}

func initDB() *gorm.DB {
	dbLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			LogLevel:      logger.Warn,
			SlowThreshold: 100 * time.Millisecond,
			Colorful:      false,
			// Ignore ErrRecordNotFound error for logger
			IgnoreRecordNotFoundError: true,
		},
	)

	cfg := getLocalhostDBConfig()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		log.Fatal(fmt.Errorf("init db error: %v", err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(10)

	return db
}

func getLocalhostDBConfig() DbConfig {
	return DbConfig{
		User:     "root",
		Password: "",
		Host:     "localhost",
		Port:     3306,
		Database: "test",
	}
}

// DB Query

func GetUsers(ctx context.Context, offset, limit int) ([]model.User, error) {
	db := NewDB()
	var users []model.User
	result := db.WithContext(ctx).Offset(offset).Limit(limit).Find(&users)
	return users, result.Error
}
