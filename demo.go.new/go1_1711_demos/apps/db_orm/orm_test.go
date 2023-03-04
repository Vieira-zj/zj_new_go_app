package db_orm

import (
	"strings"
	"testing"
)

func TestSelectBuilder(t *testing.T) {
	b := SelectBuilder{
		builder: &strings.Builder{},
	}
	sql, args := b.Select("id", "name", "age", "ctime", "mtime").From("user").
		Where(GT("id", 0), GT("age", 20)).OrderBy("id").Limit(0, 20).Query()

	t.Log("sql:", sql)
	t.Log("args:", args)
}

func TestGenTableDef(t *testing.T) {
	if err := GenTableDef("user"); err != nil {
		t.Fatal(err)
	}
	t.Log("generate table ddl done")
}
