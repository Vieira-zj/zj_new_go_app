package structs_test

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"demo.apps/structs"
)

func TestOrderedJsonDict(t *testing.T) {
	t.Run("1-level json", func(t *testing.T) {
		b := []byte(`{"d":"four", "c":"three", "b":"two"}`)
		jd := structs.OrderedJsonDict{}
		if err := jd.UnmarshalJSON(b); err != nil {
			t.Fatal(err)
		}

		jd.Set("a", json.RawMessage("one"))

		orderedBs, err := jd.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}
		t.Log("json with sorted key:", string(orderedBs))
	})

	t.Run("nested 2-level json", func(t *testing.T) {
		b := []byte(`{"d":"four", "c":"three", "b":"two", "sub":{"3":"three", "5":"five", "1":"one"}}`)
		jd := structs.OrderedJsonDict{}
		if err := jd.UnmarshalJSON(b); err != nil {
			t.Fatal(err)
		}

		sub, ok := jd.Get("sub")
		if !ok {
			t.Fatal("not found")
		}

		// handle nest json
		subJd := structs.OrderedJsonDict{}
		if err := subJd.UnmarshalJSON(sub); err != nil {
			t.Fatal(err)
		}
		subOrderedBs, err := subJd.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}

		jd.Set("sub", subOrderedBs)

		orderedBs, err := jd.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("results: %s", orderedBs)
	})
}

func TestStreamJsonArray(t *testing.T) {
	t.Run("parse full json array", func(t *testing.T) {
		b := []byte(`["one", "three", "two", "six"]`)
		sj := structs.StreamJsonArray{}
		if err := sj.UnmarshalJSON(b); err != nil {
			t.Fatal(err)
		}

		resultBs, err := sj.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}
		t.Log("results:", string(resultBs))
	})

	t.Run("stream handle json array", func(t *testing.T) {
		sb := strings.Builder{}
		sb.WriteByte('[')
		for i := 0; i < 14; i++ {
			if i != 0 {
				sb.WriteByte(',')
			}
			sb.WriteString(`"key-` + strconv.Itoa(i) + `"`)
		}
		sb.WriteByte(']')

		hanler := func(items []json.RawMessage) error {
			results := make([]string, 0, len(items))
			for _, item := range items {
				results = append(results, string(item))
			}
			t.Log("results:", strings.Join(results, "||"))
			return nil
		}

		sj := structs.StreamJsonArray{}
		if err := sj.StreamUnmarshalJSON([]byte(sb.String()), 3, hanler); err != nil {
			t.Fatal(err)
		}
		t.Log("stream handle json array finish")
	})
}
