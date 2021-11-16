package utils

import (
	"log"

	"github.com/jinzhu/gorm"

	// ignore
	_ "github.com/go-sql-driver/mysql"
)

// Mysql .
var Mysql *gorm.DB

// InitDB .
func InitDB() {
	var err error
	dsn := "root:@(127.0.0.1:3306)/default?charset=utf8&parseTime=True&loc=Local"
	Mysql, err = gorm.Open("mysql", dsn)
	if err != nil {
		log.Println("connect DB error")
		panic(err)
	}
}
