package main

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

/*
refer: https://gorm.io/docs/index.html
*/

// Quick Start
//
// test:
// sqlite3 test.db
// > .tables
// > select * from products;
// > .quit

// Product .
type Product struct {
	gorm.Model
	Code  string
	Price uint
}

var sqliteDB *gorm.DB

func init() {
	var err error
	fmt.Println("Open db")
	sqliteDB, err = gorm.Open(sqlite.Open("/tmp/test/app.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
}

func TestQuickStartCreateTable(t *testing.T) {
	db := sqliteDB
	fmt.Println("Migrate schema")
	if err := db.AutoMigrate(&Product{}); err != nil {
		t.Fatal(err)
	}
}

func TestQuickStartCreateRows(t *testing.T) {
	db := sqliteDB
	fmt.Println("Create")
	if result := db.Create(&Product{Code: "D42", Price: 100}); result.Error != nil {
		t.Fatal(result.Error)
	}
	if result := db.Create(&Product{Code: "D43", Price: 200}); result.Error != nil {
		t.Fatal(result.Error)
	}
	printProductsDetails()
}

func TestQuickStartUpdateRows(t *testing.T) {
	db := sqliteDB
	fmt.Println("Read")
	var productFoo Product
	// find product with integer primary key
	db.First(&productFoo, 1)
	fmt.Println("product id=1:", productFoo.ID, productFoo.Code, productFoo.Price)

	// find product with code D43
	productBar := Product{}
	result := db.First(&productBar, "code = ?", "D43")
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		t.Fatal("row [code=403] not found")
	}
	fmt.Println("product code=D43:", productBar.ID, productBar.Code, productBar.Price)

	fmt.Println("Update")
	db.Model(&productFoo).Updates(Product{Price: 200, Code: "F42"})
	db.Model(&productBar).Updates(map[string]interface{}{"Price": 300, "Code": "F43"})
	printProductsDetails()
}

func TestQuickStartDeleteRows(t *testing.T) {
	db := sqliteDB
	fmt.Println("Delete")
	var product Product
	if result := db.Delete(&product, 1); result.Error != nil {
		t.Fatal(result.Error)
	}
	printProductsDetails()
}

func printProductsDetails() {
	var products []Product
	result := sqliteDB.Find(&products)
	fmt.Println("products count:", result.RowsAffected)
	fmt.Println("products detail:")
	for _, prod := range products {
		fmt.Println(prod.ID, prod.Code, prod.Price)
	}
}

// Declaring Models
//
// test:
// sqlite3 module.db
// .tables
// > .schema users
// > .schema blogs

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

func TestDeclaringModels(t *testing.T) {
	db := sqliteDB
	fmt.Println("Migrate schema")
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Blog{})

	var user User
	result := db.First(&user)
	fmt.Println("rows count:", result.RowsAffected)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fmt.Println("no records found")
	}
}

//
// Example
// TODO:
//
