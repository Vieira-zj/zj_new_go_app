package main

import (
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// User .
type User struct {
	ID           uint
	Name         string
	Email        *string
	Age          uint8
	Birthday     *time.Time
	MemberNumber sql.NullString
	ActivatedAt  sql.NullTime
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Created      int64 `gorm:"autoCreateTime"`
	Updated      int64 `gorm:"autoUpdateTime:milli"`
}

// TableName custom table name, default "users".
func (User) TableName() string {
	return "user"
}

// Author .
type Author struct {
	Name  string
	Email string
}

// Blog embedded struct.
type Blog struct {
	ID      int
	Author  Author `gorm:"embedded;embeddedPrefix:author_"`
	Upvotes int32
}

func main() {
	// Declaring Models
	//
	// test:
	// sqlite3 module.db
	// .tables
	// > .schema users
	// > .schema blogs

	fmt.Println("Create db")
	db, err := gorm.Open(sqlite.Open("/tmp/test/module.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	fmt.Println("Migrate the schema")
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Blog{})

	var user User
	result := db.First(&user)
	fmt.Println("rows count:", result.RowsAffected)
	if result.Error != nil {
		fmt.Println("error:", result.Error)
	}

	fmt.Println("gorm done")
}
