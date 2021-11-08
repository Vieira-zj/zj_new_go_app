package main

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
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
//

type NewCompany struct {
	ID        string        `gorm:"primary_key;column:ID"`
	Name      string        `gorm:"column:Name"`
	Address   NewAddress    `gorm:"foreignkey:CompanyID"` // one-to-one
	Employees []NewEmployee `gorm:"foreignkey:CompanyID"` // one-to-many
}

func (NewCompany) TableName() string {
	return "new_company"
}

func (c NewCompany) String() string {
	return fmt.Sprintf("ID: %s | Name: %s | Address: %v | Employees:\n%v\n", c.ID, c.Name, c.Address, c.Employees)
}

func (c *NewCompany) BeforeCreate(tx *gorm.DB) (err error) {
	c.ID = uuid.New().String()
	return
}

func (c *NewCompany) AfterDelete(tx *gorm.DB) (err error) {
	fmt.Println("AfterDelete:", c.ID)
	return
}

type NewEmployee struct {
	ID               string    `gorm:"primary_key;column:ID"`
	FirstName        string    `gorm:"column:FirstName"`
	LastName         string    `gorm:"column:LastName"`
	SocialSecurityNo string    `gorm:"column:SocialSecurityNo"`
	DateOfBirth      time.Time `gorm:"column:DateOfBirth"`
	Deleted          bool      `gorm:"column:Deleted;default:false"`
	CompanyID        string    `gorm:"column:Company_ID"`
}

func (NewEmployee) TableName() string {
	return "new_employee"
}

func (e NewEmployee) String() string {
	return fmt.Sprintf("ID: %s | FirstName: %s\n", e.ID, e.FirstName)
}

func (e *NewEmployee) BeforeCreate(tx *gorm.DB) (err error) {
	e.ID = uuid.New().String()
	return
}

type NewAddress struct {
	ID        uint
	Country   string `gorm:"column:Country"`
	City      string `gorm:"column:City"`
	PostCode  string `gorm:"column:PostCode"`
	Line1     string `gorm:"column:Line1"`
	Line2     string `gorm:"column:Line2"`
	CompanyID string `gorm:"column:Company_ID"`
}

func (NewAddress) TableName() string {
	return "new_address"
}

func TestExampleCreateTables(t *testing.T) {
	db := mysqlDB
	fmt.Println("Create Tables")
	if err := db.AutoMigrate(&NewEmployee{}); err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&NewAddress{}); err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&NewCompany{}); err != nil {
		t.Fatal(err)
	}

	// TOFIX: add Foreign Key Constraint
}

func TestExampleInitCompany(t *testing.T) {
	company := NewCompany{
		Name: "Google",
		Address: NewAddress{
			Country:  "USA",
			City:     "Moutain View",
			PostCode: "1600",
			Line1:    "Cloverfield Lane, 32",
			Line2:    "Apt 64",
		},
		Employees: []NewEmployee{
			{
				FirstName:        "John",
				LastName:         "Doe",
				SocialSecurityNo: "00-000-0000",
				DateOfBirth:      time.Now(),
			},
			{
				FirstName:        "James",
				LastName:         "Dean",
				SocialSecurityNo: "00-000-0001",
				DateOfBirth:      time.Now().AddDate(1, 0, 0),
			},
			{
				FirstName:        "Joan",
				LastName:         "Dutsch",
				SocialSecurityNo: "00-000-0002",
				DateOfBirth:      time.Now().AddDate(2, 0, 0),
			},
		},
	}

	fmt.Println("Create Company, Address and Employees")
	if result := mysqlDB.Create(&company); result.Error != nil {
		t.Fatal(result.Error)
	}
}

func TestExampleRetrieveCompany(t *testing.T) {
	db := mysqlDB
	company := NewCompany{}
	db.First(&company)

	fmt.Println("Find Only Company (Lazy load)...")
	var firstComp NewCompany
	db.Where("ID = ?", company.ID).First(&firstComp)
	fmt.Println("First Company:", firstComp)
	fmt.Println()

	fmt.Println("Find Only Company (Eager load)...")
	var fullComp NewCompany
	db.Preload("Employees").Joins("Address").First(&fullComp)
	fmt.Println("Full Company: ", fullComp)
}

func TestExampleUpdateCompany(t *testing.T) {
	db := mysqlDB
	company := NewCompany{}
	db.Joins("Address").First(&company)

	fmt.Println("Update Company and Address")
	company.Name = "Google Plus"
	if len(company.Address.Country) > 0 {
		company.Address.Country = "France"
	}
	db.Save(&company)
	db.Save(&company.Address)
}

func TestExampleDeleteCompany(t *testing.T) {
	db := mysqlDB
	company := NewCompany{}
	db.First(&company)

	fmt.Println("Delete Company")
	transaction := db.Begin()
	if result := transaction.Delete(&company); result.Error != nil {
		transaction.Rollback()
		t.Fatal(result.Error)
	}
	transaction.Commit()
	fmt.Println("Delete Company Done")
}
