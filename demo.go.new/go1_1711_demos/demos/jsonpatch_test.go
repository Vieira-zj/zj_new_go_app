package demos_test

import (
	"encoding/json"
	"testing"

	jsonpatch "github.com/evanphx/json-patch"
)

type JsonPatchItem struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

type JsonPatchRawItem struct {
	Op    string          `json:"op"`
	Path  string          `json:"path"`
	Value json.RawMessage `json:"value"`
}

var jsonpatchTestData = `{
	"meta": {
		"region": "en"
	},
	"data": {
		"fruit": [
			{ "name": "apple" },
			{ "name": "banana" }
		]
	}
}`

// refer: https://jsonpatch.com/

func TestJsonPatchReplace(t *testing.T) {
	item1 := JsonPatchItem{
		Op:    "test",
		Path:  "/meta/region",
		Value: "en",
	}
	item2 := JsonPatchItem{
		Op:    "replace",
		Path:  "/meta/region",
		Value: "cn",
	}
	b, err := json.Marshal([]*JsonPatchItem{&item1, &item2})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("patch: %s", b)
	patch, err := jsonpatch.DecodePatch(b)
	if err != nil {
		t.Fatal(err)
	}

	modified, err := patch.Apply([]byte(jsonpatchTestData))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("modified: %s", modified)
}

func TestJsonPatchForArrayItem(t *testing.T) {
	item1 := JsonPatchItem{
		Op:    "replace",
		Path:  "/data/fruit/0/name",
		Value: "pear",
	}
	item2 := JsonPatchItem{
		Op:    "replace",
		Path:  "/data/fruit/1/name",
		Value: "mango",
	}
	item3 := JsonPatchRawItem{
		Op:    "add",
		Path:  "/data/fruit/2",
		Value: []byte(`{ "name": "banana" }`),
	}
	b, err := json.Marshal([]interface{}{&item1, &item2, &item3})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("patch: %s", b)
	patch, err := jsonpatch.DecodePatch(b)
	if err != nil {
		t.Fatal(err)
	}

	modified, err := patch.Apply([]byte(jsonpatchTestData))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("modified: %s", modified)
}
