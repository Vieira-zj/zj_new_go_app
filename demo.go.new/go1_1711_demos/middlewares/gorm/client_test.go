package gorm_test

import (
	"context"
	"encoding/json"
	"go1_1711_demo/middlewares/gorm"
	"go1_1711_demo/middlewares/gorm/model"
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

func TestSearchUser(t *testing.T) {
	users, err := gorm.SearchUser(
		context.TODO(),
		gorm.WithUserNameHasPrefixScope("jin"),
		gorm.WithUserIDGreaterThanScope(101),
	)
	if err != nil {
		t.Fatal(err)
	}
	for i := range users {
		u := users[i]
		t.Log("user:", u.Name, u.Age)
	}
}

func TestInsertBlob(t *testing.T) {
	db := gorm.NewDB()
	data := map[string]string{
		"msg": "hello world",
	}
	b, err := json.Marshal(&data)
	if err != nil {
		t.Fatal(err)
	}

	blob := model.Blob{
		Text:  string(b),
		Bytes: b,
	}
	if err := db.Create(&blob).Error; err != nil {
		t.Fatal(err)
	}
	t.Log("insert blob done")
}

func TestReadBlob(t *testing.T) {
	db := gorm.NewDB()
	blob := model.Blob{}
	if err := db.First(&blob).Error; err != nil {
		t.Fatal(err)
	}
	t.Log("text:", blob.Text)

	m := make(map[string]interface{})
	if err := json.Unmarshal(blob.Bytes, &m); err != nil {
		t.Fatal(err)
	}
	t.Log("msg:", m["msg"])
}
