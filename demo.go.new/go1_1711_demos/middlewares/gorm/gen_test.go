package gorm_test

import (
	"go1_1711_demo/middlewares/gorm"
	"go1_1711_demo/middlewares/gorm/model"
	"testing"
)

func TestGenModel(t *testing.T) {
	gorm.GenModel("user")
	t.Log("gen model done")
}

func TestGenTable(t *testing.T) {
	if err := gorm.GenTable(&model.Blob{}); err != nil {
		t.Fatal(err)
	}
	t.Log("gen table done")
}
