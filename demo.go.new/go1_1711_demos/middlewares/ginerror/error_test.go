package main

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestErrorLoad(t *testing.T) {
	errStr := `{"status_code":"404", "code":200404, "msg":"not found"}`
	err := Error{}
	if err := json.Unmarshal([]byte(errStr), &err); err != nil {
		t.Fatal(err)
	}
	t.Logf("error: %+v", err)
}

func TestErrorDump(t *testing.T) {
	myErr := Error{
		StatusCode: 404,
		Code:       200404,
		Msg:        "not found",
	}
	b, err := json.Marshal(&myErr)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("error:", string(b))
}
