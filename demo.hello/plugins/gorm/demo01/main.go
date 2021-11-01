package main

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

/*
refer: https://gorm.io/docs/models.html
*/

// Product .
type Product struct {
	gorm.Model
	Code  string
	Price uint
}

func main() {
	// Quick Start
	//
	// test:
	// sqlite3 test.db
	// > .tables
	// > select * from products;
	// > .quit

	fmt.Println("Create db")
	db, err := gorm.Open(sqlite.Open("/tmp/test/test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	fmt.Println("Migrate the schema")
	db.AutoMigrate(&Product{})

	fmt.Println("Create")
	db.Create(&Product{Code: "D42", Price: 100})
	db.Create(&Product{Code: "D43", Price: 200})

	fmt.Println("Read")
	var productFoo Product
	// find product with integer primary key
	db.First(&productFoo, 1)
	fmt.Println("product id=1:", productFoo)
	// find product with code D43
	productBar := Product{}
	db.First(&productBar, "code = ?", "D43")
	fmt.Println("product code=D43:", productBar)

	fmt.Println("Update")
	db.Model(&productFoo).Updates(Product{Price: 200, Code: "F42"})
	db.Model(&productBar).Updates(map[string]interface{}{"Price": 300, "Code": "F43"})

	fmt.Println("Delete")
	var product Product
	db.Delete(&product, 1)
}
