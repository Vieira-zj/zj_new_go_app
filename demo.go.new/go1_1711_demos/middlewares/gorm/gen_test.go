package gorm_test

import (
	"go1_1711_demo/middlewares/gorm"
	"testing"
)

func TestGenModel(t *testing.T) {
	gorm.GenModel("user")
	t.Log("gen model done")
}
