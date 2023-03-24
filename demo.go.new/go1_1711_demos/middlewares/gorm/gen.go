package gorm

import (
	"gorm.io/gen"
)

// GenModel generates go model from table definition.
func GenModel(tbName string) {
	g := gen.NewGenerator(gen.Config{
		OutPath: "./model",
	})

	db := NewDB()
	g.UseDB(db)
	g.GenerateModel(tbName)
	g.Execute()
}

// GenTable creates table in db by go model.
func GenTable(tabModel interface{}) error {
	db := NewDB()
	return db.AutoMigrate(tabModel)
}
