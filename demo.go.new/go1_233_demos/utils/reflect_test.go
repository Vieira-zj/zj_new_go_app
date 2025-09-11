package utils_test

import (
	"testing"

	"zjin.goapp.demo/utils"
)

type TestPerson struct {
	ID   int
	Name string
	Age  int
	Role string
}

func TestDiffStruct(t *testing.T) {
	p1 := TestPerson{
		ID:   1,
		Name: "Foo",
		Age:  30,
		Role: "QA",
	}
	p2 := TestPerson{
		ID:   2,
		Name: "Bar",
		Age:  41,
		Role: "QA",
	}

	results := utils.DiffStruct(p1, p2)
	for _, item := range results {
		t.Logf("diff: %s", item)
	}
}
