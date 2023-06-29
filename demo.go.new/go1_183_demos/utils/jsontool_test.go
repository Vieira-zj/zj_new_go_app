package utils_test

import (
	"encoding/json"
	"testing"

	"demo.apps/utils"
)

func TestGetValueByJsonPath(t *testing.T) {
	str := `{"a":"bar", "b":"foo", "replace":{"c":"cha"}, "move":{"x":[1,2,3]}}`
	obj := make(map[string]any)
	if err := json.Unmarshal([]byte(str), &obj); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		"/a",
		"/replace/c",
		"/move/x/2",
	} {
		val, err := utils.GetValueByJsonPath(obj, path)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("path=%s, val=%v", path, val)
	}
}

func TestUpdateValueByJsonPath(t *testing.T) {
	str := `{"a":"bar", "b":"foo", "replace":{"c":"cha"}, "move":{"x":[1,2,3]}}`
	obj := make(map[string]any)
	if err := json.Unmarshal([]byte(str), &obj); err != nil {
		t.Fatal(err)
	}

	// "/a", "bar_modify"
	// "/replace/c", "cha_modify"
	if err := utils.UpdateValueByJsonPath(obj, "/move/x/1", float64(9)); err != nil {
		t.Fatal(err)
	}
	t.Logf("object: %+v", obj)
}
