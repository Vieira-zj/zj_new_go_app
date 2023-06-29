package demos

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/wI2L/jsondiff"
)

func TestJsonDiffWithIgnore(t *testing.T) {
	source := `{"a":"bar", "b":"baz", "c":"foo", "d":[1,2,3]}`
	target := `{"a":"rab", "d":"foo", "b":"bza", "d":[1,2,4]}`

	// decode json numbers as json.Number instead of float64
	decodeOption := jsondiff.UnmarshalFunc(func(b []byte, v any) error {
		dec := json.NewDecoder(bytes.NewReader(b))
		dec.UseNumber()
		return dec.Decode(v)
	})

	patch, err := jsondiff.CompareJSON([]byte(source), []byte(target), jsondiff.Ignores("/a", "/c"), decodeOption)
	if err != nil {
		t.Fatal(err)
	}

	for _, op := range patch {
		t.Logf(op.String())
	}
}

func TestJsonDiffWithEquivalence01(t *testing.T) {
	// slice length should be the same
	source := `{"name":"bar", "ids":[1,2,3]}`
	target := `{"name":"bar", "ids":[3,2,1]}`

	patch, err := jsondiff.CompareJSON([]byte(source), []byte(target), jsondiff.Equivalent())
	if err != nil {
		t.Fatal(err)
	}

	for _, op := range patch {
		t.Logf(op.String())
	}
}

func TestJsonDiffWithEquivalence02(t *testing.T) {
	source := `{"group":"abc", "users":[{"id":1, "name":"foo"},{"id":2, "name":"bar"}]}`
	target := `{"group":"xyz", "users":[{"id":2, "name":"bar"},{"id":1, "name":"foo"}]}`

	patch, err := jsondiff.CompareJSON([]byte(source), []byte(target), jsondiff.Equivalent())
	if err != nil {
		t.Fatal(err)
	}

	for _, op := range patch {
		t.Logf(op.String())
	}
}

func TestHighlightJsonDiff(t *testing.T) {
	source := `{"a":"bar", "b":"foo", "replace":{"c":"cha"}, "move":{"x":"xyz"}}`
	target := `{"a":"bar", "replace":{"c":"chz"}, "d":"ds", "move":{"y":"xyz"}}`

	patch, err := jsondiff.CompareJSON([]byte(source), []byte(target), jsondiff.Factorize())
	if err != nil {
		t.Fatal(err)
	}

	for _, op := range patch {
		t.Logf(op.String())
	}
}
