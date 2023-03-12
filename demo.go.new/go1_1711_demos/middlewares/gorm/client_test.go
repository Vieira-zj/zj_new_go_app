package gorm_test

import (
	"context"
	"go1_1711_demo/middlewares/gorm"
	"testing"
)

func TestNewDB(t *testing.T) {
	db := gorm.NewDB()
	names := []string{}
	result := db.WithContext(context.Background()).Table("user").Select("name").Find(&names)
	if err := result.Error; err != nil {
		t.Fatal(err)
	}

	for i := range names {
		t.Log("name:", names[i])
	}
}

func TestGetUsers(t *testing.T) {
	users, err := gorm.GetUsers(context.Background(), 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	for i := range users {
		u := users[i]
		t.Log("user:", u.Name, u.Age)
	}
}
