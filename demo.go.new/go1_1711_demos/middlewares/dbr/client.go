package dbr

import (
	"fmt"
	"log"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gocraft/dbr/v2"
)

var (
	dbConn     *dbr.Connection
	dbConnOnce sync.Once
)

func NewDBConn() *dbr.Connection {
	dbConnOnce.Do(func() {
		dbConn = initDBConn()
	})
	return dbConn
}

// Init DB

type DbConfig struct {
	User     string
	Password string
	Host     string
	Port     int
	Database string
}

func initDBConn() *dbr.Connection {
	cfg := getLocalhostDBConfig()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	conn, err := dbr.Open("mysql", dsn, nil)
	if err != nil {
		log.Fatal(err)
	}
	return conn
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
