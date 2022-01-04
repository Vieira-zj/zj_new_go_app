package pkg

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestMarshalIndent(t *testing.T) {
	data := struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{
		Name: "Foo",
		Age:  31,
	}

	b, err := json.Marshal(&data)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(b))
	fmt.Println()

	b, err = json.MarshalIndent(&data, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(b))
}
