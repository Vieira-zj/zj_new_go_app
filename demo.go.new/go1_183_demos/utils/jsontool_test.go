package utils_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"demo.apps/utils"
	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/ast"
	"github.com/bytedance/sonic/decoder"
	"github.com/bytedance/sonic/encoder"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/wI2L/jsondiff"
)

func TestIsValidJson(t *testing.T) {
	for _, s := range []string{
		`{"name":"foo"}`,
		`["foo", "bar"]`,
		`"bar"`,
	} {
		b := []byte(s)
		result := json.Valid(b)
		t.Log("raw result:", result)
		result = utils.IsValidJson(b)
		t.Log("result:", result)
	}
}

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

// json diff

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

func TestJsonDiffWithFactorize(t *testing.T) {
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

// json patch

func TestJsonPatchApply(t *testing.T) {
	patchJson := []byte(`[
		{"op":"replace", "path":"/name", "value":"Jane"},
		{"op":"remove", "path":"/height"},
		{"op":"add", "path":"/ids/3", "value":4},
		{"op":"test", "path":"/age", "value":24}
	]`)

	patch, err := jsonpatch.DecodePatch(patchJson)
	if err != nil {
		t.Fatal(err)
	}

	original := []byte(`{"name":"John", "age":24, "height":3.21, "ids":[1,2,3]}`)
	result, err := patch.Apply(original)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("patch result:", string(result))
}

// json coder

type JsonArray = []any
type JsonMap = map[string]any

func TestJsonCoder(t *testing.T) {
	for _, s := range []string{
		`{"name":{"first":"foo", "last":"bar"}, "ids":[1,2,3]}`,
		`{"ids":[3,2,1], "name":{"last":"bar", "first":"foo"}}`,
	} {
		var data JsonMap
		dc := decoder.NewDecoder(s)
		dc.UseNumber()
		dc.Decode(&data)
		t.Log("name:", data["name"])
		t.Log("ids:", data["ids"].(JsonArray))

		b, err := encoder.Encode(data, encoder.SortMapKeys)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("sorted json:", string(b))
	}
}

func TestJsonFormat(t *testing.T) {
	for i, s := range []string{
		`{"name":{"first":"foo", "last":"bar"}, "ids":[1,2,3]}`,
		`{"ids":[3,2,1], "name":{"last":"bar", "first":"foo"}}`,
	} {
		root, err := sonic.Get([]byte(s))
		if err != nil {
			t.Fatal(err)
		}
		if err = root.SortKeys(true); err != nil {
			t.Fatal(err)
		}
		raw, err := root.Raw()
		if err != nil {
			t.Fatal(err)
		}
		t.Log("sorted json:", raw)

		b, err := json.MarshalIndent(&root, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(fmt.Sprintf("/tmp/test/test%d.json", i+1), b, 0644); err != nil {
			t.Fatal(err)
		}
	}
}

type MyVisitor struct{}

func (*MyVisitor) OnNull() error                            { return nil }
func (*MyVisitor) OnBool(v bool) error                      { return nil }
func (*MyVisitor) OnString(v string) error                  { return nil }
func (*MyVisitor) OnFloat64(v float64, n json.Number) error { return nil }

func (*MyVisitor) OnObjectBegin(capacity int) error { return nil }
func (*MyVisitor) OnObjectEnd() error               { return nil }
func (*MyVisitor) OnArrayEnd() error                { return nil }

func (*MyVisitor) OnInt64(v int64, n json.Number) error {
	fmt.Println("OnInt64:", v)
	return nil
}
func (*MyVisitor) OnObjectKey(key string) error {
	fmt.Println("OnObjectKey:", key)
	return nil
}
func (*MyVisitor) OnArrayBegin(capacity int) error {
	fmt.Println("OnArrayBegin:", capacity) // here is slice cap, but not len
	return nil
}

func TestWalkJsonNodes(t *testing.T) {
	s := `{"name":{"first":"foo", "last":"bar"}, "ids":[1,2,3]}`
	if err := ast.Preorder(s, &MyVisitor{}, &ast.VisitorOptions{}); err != nil {
		t.Fatal(err)
	}
	t.Log("walk json nodes done")
}
