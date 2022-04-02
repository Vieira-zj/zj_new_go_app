package main

import (
	"fmt"
	"testing"
)

func TestDecodeToken(t *testing.T) {
	// echo -n "hello_world" | base64
	token := "aGVsbG9fd29ybGQ="
	b, err := decodeToken(token)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("module:", string(b))
}
