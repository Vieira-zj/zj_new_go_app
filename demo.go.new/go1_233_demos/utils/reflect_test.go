package utils_test

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"zjin.goapp.demo/utils"
)

type TestPerson struct {
	ID   int    `json:"id"`
	Name string `json:"name" x_idx:"1"`
	Role string `json:"role" x_idx:"3"`
	Age  int    `json:"age" x_idx:"2"`
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

func TestGetStructFieldInfo(t *testing.T) {
	p := TestPerson{
		ID:   1,
		Name: "Foo",
		Role: "QA",
		Age:  30,
	}

	b, err := json.Marshal(&p)
	assert.NoError(t, err)
	t.Logf("json: %s", string(b))

	results, err := utils.GetStructFieldsInfo(p)
	assert.NoError(t, err)

	slices.SortFunc(results, func(a, b utils.StructFieldInfo) int {
		return a.Index - b.Index
	})
	for _, item := range results {
		t.Logf("struct field: %s, index: %d, value: %v", item.Name, item.Index, item.Value)
	}
}

func MyAddForTest(a int, b string) (bool, error) {
	_, _ = a, b
	return true, nil
}

func TestGetFuncSignature(t *testing.T) {
	result, err := utils.GetFuncSignature(MyAddForTest)
	assert.NoError(t, err)
	t.Logf("func signature: %+v", result)
}
