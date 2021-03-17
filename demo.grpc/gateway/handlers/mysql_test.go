package handlers

import (
	"fmt"
	"testing"
)

func TestMysqlHandler(t *testing.T) {
	handler := NewMysqlHandler()
	res, err := handler.ExecSelect("select * from users where id < 4")
	if err != nil {
		t.Fatal("exec select failed:", err)
	}

	for i := range res {
		fmt.Println(res[i])
	}
}
