package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// config

const (
	username = "root"
	password = "example"
	network  = "tcp"
	server   = "localhost"
	port     = 13306
	database = "test"
)

// NewMysqlHandler create mysql handler
func NewMysqlHandler() *MysqlHandler {
	handler := &MysqlHandler{}
	if err := handler.buildDB(); err != nil {
		log.Fatalln("Build db failed, error:", err)
	}
	return handler
}

// MysqlHandler mysql handler
type MysqlHandler struct {
	db *sql.DB
}

func (handler *MysqlHandler) buildDB() error {
	if handler.db != nil {
		return nil
	}

	dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s", username, password, network, server, port, database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("Open mysql failed, err: %v", err)
	}

	db.SetConnMaxLifetime(time.Duration(60) * time.Second)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(3)
	handler.db = db
	return nil
}

// ExecSelect run sql elect statement.
func (handler *MysqlHandler) ExecSelect(sql string) ([]string, error) {
	if handler.db == nil {
		log.Fatalln("Pls build db first.")
	}
	log.Println("Run sql:", sql)

	rows, err := handler.db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	row := make([]interface{}, len(cols))
	rowRef := make([]interface{}, len(cols)) // Scan()函数入参为指针类型
	for i := range row {
		rowRef[i] = &row[i]
	}

	var res []string
	for rows.Next() {
		if err := rows.Scan(rowRef...); err != nil {
			return nil, err
		}
		tmp := make([]string, len(cols))
		for i := range row {
			if v, ok := row[i].([]byte); ok {
				if len(v) > 0 {
					tmp[i] = string(v)
				} else {
					tmp[i] = "null"
				}
			}
		}
		res = append(res, strings.Join(tmp, ","))
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
