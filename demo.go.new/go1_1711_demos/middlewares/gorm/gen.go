package gorm

import "gorm.io/gen"

func GenModel(tbName string) {
	g := gen.NewGenerator(gen.Config{
		OutPath: "./model",
	})

	db := NewDB()
	g.UseDB(db)
	g.GenerateModel(tbName)
	g.Execute()
}
