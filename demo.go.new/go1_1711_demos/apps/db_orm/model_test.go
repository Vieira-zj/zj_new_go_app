package db_orm

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestFindUsersByRawSql(t *testing.T) {
	db, err := createLocalMysqlDb()
	if err != nil {
		t.Fatal(err)
	}

	users, err := FindUsersByRawSql(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}

	for _, u := range users {
		t.Logf("name=%s, age=%d", u.Name, u.Age)
	}
}

func TestFindUserByReflect(t *testing.T) {
	db, err := createLocalMysqlDb()
	if err != nil {
		t.Fatal(err)
	}

	users, err := FindUserByReflect(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}

	for _, u := range users {
		t.Logf("name=%s, age=%d", u.Name, u.Age)
	}
}

func TestFindUser(t *testing.T) {
	db, err := createLocalMysqlDb()
	if err != nil {
		t.Fatal(err)
	}

	for _, cols := range [][]string{
		{"id", "name", "age", "ctime", "mtime"},
		{"name", "age"},
	} {
		b := &SelectBuilder{builder: &strings.Builder{}}
		b = b.Select(cols...).From("user").
			Where(GT("id", 0), GT("age", 29)).OrderBy("id").Limit(0, 20)

		users, err := FindUser(context.Background(), db, b)
		if err != nil {
			t.Fatal(err)
		}

		for _, u := range users {
			t.Logf("id=%d, name=%s, age=%d", u.Id, u.Name, u.Age)
		}
	}
}

func createLocalMysqlDb() (*sql.DB, error) {
	// refer: https://stackoverflow.com/questions/45040319/unsupported-scan-storing-driver-value-type-uint8-into-type-time-time
	uri := "root:@tcp(127.0.0.1:3306)/test?parseTime=true"
	return sql.Open("mysql", uri)
}
