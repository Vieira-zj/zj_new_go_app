package main

import (
	"fmt"
	"strings"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var mysqlDB *gorm.DB

func init() {
	var err error
	dsn := "root:@tcp(127.0.0.1:3306)/default?charset=utf8mb4&parseTime=True&loc=Local"
	if mysqlDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{}); err != nil {
		panic(err)
	}
}

//
// Association: Belong To
//

// BelongToUser belongs to `Company`, and `CompanyID` is the foreign key.
type BelongToUser struct {
	gorm.Model
	Name      string
	CompanyID int     // foreignkey
	Company   Company `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

// Company .
type Company struct {
	ID   int // primarykey
	Name string
}

func TestAssocationBelongToCreateDB(t *testing.T) {
	db := mysqlDB
	fmt.Println("Migrate schema")
	if err := db.AutoMigrate(&Company{}); err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&BelongToUser{}); err != nil {
		t.Fatal(err)
	}
}

func TestAssocationBelongToCreateRows(t *testing.T) {
	db := mysqlDB
	fmt.Println("Create user and company")
	// create user and related company.
	company := Company{Name: "IBM"}
	if result := db.Create(&BelongToUser{Name: "foo", Company: company}); result.Error != nil {
		t.Fatal(result.Error)
	}
	company = Company{Name: "Apex"}
	if result := db.Create(&BelongToUser{Name: "bar", Company: company}); result.Error != nil {
		t.Fatal(result.Error)
	}

	// if company row exist, it will not create a new one.
	existCompany := Company{}
	if result := db.Find(&existCompany, "name = ?", "Apex"); result.Error != nil {
		t.Fatal(result.Error)
	}
	db.Create(&BelongToUser{Name: "jin", Company: existCompany})
}

func TestAssocationBelongToOnDeleteConstraint(t *testing.T) {
	fmt.Println("Trigger OnDelete constraint")
	db := mysqlDB
	var company Company
	db.Delete(&company, 1)
}

func TestAssocationBelongToRetrieveRows(t *testing.T) {
	var users []BelongToUser
	result := mysqlDB.Joins("Company").Find(&users)
	if result.Error != nil {
		t.Fatal(result.Error)
	}

	fmt.Println("users detail:")
	for _, user := range users {
		fmt.Println(user.ID, user.Name, user.CompanyID, user.Company.Name)
	}
}

//
// Association: Has One
//

// HasOneUser has one `CreditCard`, `UserName` is the foreign key, `Name` is the reference key.
type HasOneUser struct {
	gorm.Model
	Name       string           `gorm:"index"`
	CreditCard HasOneCreditCard `gorm:"foreignkey:UserName;references:name"`
}

// HasOneCreditCard .
type HasOneCreditCard struct {
	gorm.Model
	Number   string
	UserName string
}

func TestAssocationHasOneCreateDB(t *testing.T) {
	db := mysqlDB
	fmt.Println("Migrate schema")
	if err := db.AutoMigrate(&HasOneCreditCard{}); err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&HasOneUser{}); err != nil {
		t.Fatal(err)
	}
}

func TestAssocationHasOneCreateRows(t *testing.T) {
	db := mysqlDB
	fmt.Println("Create creditCard")
	if result := db.Create(&HasOneCreditCard{Number: "1009"}); result.Error != nil {
		t.Fatal(result.Error)
	}

	fmt.Println("Create user and creditCard")
	userFoo := &HasOneUser{Name: "foo", CreditCard: HasOneCreditCard{Number: "1010"}}
	db.Create(userFoo)
	userBar := &HasOneUser{Name: "bar", CreditCard: HasOneCreditCard{Number: "1011"}}
	db.Create(userBar)
}

func TestAssocationHasOneRetrieveRows(t *testing.T) {
	var users []HasOneUser
	result := mysqlDB.Joins("CreditCard").Where("name = ?", "foo").Find(&users)
	if result.Error != nil {
		t.Fatal(result.Error)
	}

	fmt.Println("users detail:")
	for _, user := range users {
		fmt.Println(user.Name, user.CreditCard.Number)
	}
}

//
// Raw SQL
//

type HasOneResult struct {
	Name   string
	Number string
}

func TestAssocationHasOneRawSQL(t *testing.T) {
	db := mysqlDB
	results := []HasOneResult{}
	sqlSelect := "select user.name, card.number"
	sqlFrom := "from has_one_users as user, credit_cards as card"
	sqlWhere := `where user.name = card.user_name and user.name = "foo"`
	db.Raw(strings.Join([]string{sqlSelect, sqlFrom, sqlWhere}, " ")).Scan(&results)
	fmt.Println("results:")
	for _, result := range results {
		fmt.Println(result)
	}
	fmt.Println()

	results = []HasOneResult{}
	sqlSelect = "select user.name, card.number"
	sqlFrom = "from has_one_users as user left join credit_cards as card on user.name = card.user_name"
	db.Raw(strings.Join([]string{sqlSelect, sqlFrom}, " ")).Scan(&results)
	fmt.Println("results:")
	for _, result := range results {
		fmt.Println(result)
	}
}

//
// Association: Has Many
//

// HasManyUser has many `CreditCards`, `HasManyUserID` is the foreign key.
type HasManyUser struct {
	gorm.Model
	CreditCards []HasManyCreditCard
}

type HasManyCreditCard struct {
	gorm.Model
	Number        string
	HasManyUserID uint
}

func TestAssocationHasManyCreateDB(t *testing.T) {
	db := mysqlDB
	fmt.Println("Migrate schema")
	if err := db.AutoMigrate(&HasManyUser{}); err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&HasManyCreditCard{}); err != nil {
		t.Fatal(err)
	}
}

func TestAssocationHasManyCreateRows(t *testing.T) {
	db := mysqlDB
	fmt.Println("Create creditCard")
	if result := db.Create(&HasManyCreditCard{Number: "1009"}); result.Error != nil {
		// create failed because of a foreign key constraint:
		// foreignkey has_many_credit_cards.has_many_user_id => primarykey has_many_users.id
		fmt.Println(result.Error)
	}

	fmt.Println("Create user and creditCard")
	creditCards := []HasManyCreditCard{
		{Number: "1010"},
		{Number: "1011"},
	}
	user := HasManyUser{CreditCards: creditCards}
	db.Create(&user)
}

func TestAssocationHasManyRetrieveRows(t *testing.T) {
	// for related multi-rows `CreditCards`, use db.Preload() instead of db.Joins().
	user := HasManyUser{}
	if result := mysqlDB.Preload("CreditCards").Where("id = ?", 2).Find(&user); result.Error != nil {
		t.Fatal(result.Error)
	}

	fmt.Printf("user [id=%d] detail:\n", user.ID)
	for _, card := range user.CreditCards {
		fmt.Println(card.Number)
	}
}

func TestAssocationHasManyRetrieveRowsRaw(t *testing.T) {
	rows := []struct {
		ID     uint
		Number string
	}{}
	sqlSelect := "SELECT users.id, cards.number"
	sqlFrom := "FROM has_many_users as users, has_many_credit_cards as cards"
	sqlWhere := "WHERE users.id = cards.has_many_user_id AND users.id = 2"
	result := mysqlDB.Raw(strings.Join([]string{sqlSelect, sqlFrom, sqlWhere}, " ")).Scan(&rows)
	if result.Error != nil {
		t.Fatal(result.Error)
	}
	fmt.Println("user -> cards detail:")
	for _, row := range rows {
		fmt.Printf("%+v\n", row)
	}
}
