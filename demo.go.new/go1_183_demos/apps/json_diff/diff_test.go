package jsondiff

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDifferCompare01(t *testing.T) {
	src := make(map[string]any)
	srcb := []byte(`{"slice":[1,2,3], "map":{"name":"foo","age":41}}`)
	err := json.Unmarshal(srcb, &src)
	assert.NoError(t, err)

	dst := make(map[string]any)
	dstb := []byte(`{"slice":[1,3,2], "map":{"age":41,"name":"foo"}}`)
	err = json.Unmarshal(dstb, &dst)
	assert.NoError(t, err)

	t.Run("diff slice order", func(t *testing.T) {
		differ := NewDiffer()
		differ.Compare(src, dst)

		t.Log("differ results:\n", differ.Patches().string())
	})

	t.Run("diff with slice order", func(t *testing.T) {
		differ := NewDiffer(WithSliceOrders([]string{"/slice"}))
		differ.Compare(src, dst)

		l := differ.Patches().len()
		assert.True(t, l == 0)
		t.Log("differ pass")
	})
}

func TestDifferCompare02(t *testing.T) {
	src := make(map[string]any)
	srcb := []byte(`{"slice":[1,2,3], "map":{"name":"foo","age":41}}`)
	err := json.Unmarshal(srcb, &src)
	assert.NoError(t, err)

	dst := make(map[string]any)
	dstb := []byte(`{"slice":[1,3,2,4], "map":{"age":41,"name":"bar"}}`)
	err = json.Unmarshal(dstb, &dst)
	assert.NoError(t, err)

	t.Run("silce and map value diff", func(t *testing.T) {
		differ := NewDiffer()
		differ.Compare(src, dst)
		t.Log("differ results:\n", differ.Patches().string())
	})

	t.Run("diff with ignores", func(t *testing.T) {
		differ := NewDiffer(WithIgnores([]string{"/slice/3"}))
		differ.Compare(src, dst)
		t.Log("differ results:\n", differ.Patches().string())
	})
}

func TestDifferCompare03(t *testing.T) {
	src := make(map[string]any)
	srcb := []byte(`{"map1":{"key1":"value1","key2":"value2"}, "map2":{"keya":"valuea","keyb":"valueb"}}`)
	err := json.Unmarshal(srcb, &src)
	assert.NoError(t, err)

	dst := make(map[string]any)
	dstb := []byte(`{"map1":{"key1":"value1","key2":"valueb"}, "map2":{"keya":"value1","keyb":"valueb"}}`)
	err = json.Unmarshal(dstb, &dst)
	assert.NoError(t, err)

	t.Run("diff maps value", func(t *testing.T) {
		differ := NewDiffer()
		differ.Compare(src, dst)
		t.Log("differ results:\n", differ.Patches().string())
	})
}

// Compare Slice

func TestCompareSliceOfFloat(t *testing.T) {
	src := make(map[string]any)
	srcb := []byte(`{"slice":[1,7,2,3,4]}`)
	err := json.Unmarshal(srcb, &src)
	assert.NoError(t, err)

	dst := make(map[string]any)
	dstb := []byte(`{"slice":[1,3,11,2,9,4]}`)
	err = json.Unmarshal(dstb, &dst)
	assert.NoError(t, err)

	t.Run("failed, diff with slice order", func(t *testing.T) {
		differ := NewDiffer(WithSliceOrders([]string{"/slice"}))
		differ.Compare(src, dst)
		t.Log("differ results:\n", differ.Patches().string())
	})

	t.Run("pass, order and diff slice", func(t *testing.T) {
		srcSlice, srcOk := src["slice"].([]any)
		dstSlice, dstOk := dst["slice"].([]any)
		assert.True(t, srcOk && dstOk)

		results := CompareSliceOfFloat(srcSlice, dstSlice)
		t.Log("compare results:\n", strings.Join(results, " | "))
	})
}

func CompareSliceOfFloat(src, dst []any) []string {
	// make sure no duplicated items in slice.
	m := make(map[float64]uint8, len(src))
	for _, num := range src {
		m[num.(float64)] += 1
	}
	for _, num := range dst {
		m[num.(float64)] += 10
	}

	results := make([]string, 0, len(src))
	for key, val := range m {
		switch val {
		case 1: // in old
			results = append(results, fmt.Sprintf("del: %.0f", key))
		case 10: // in new
			results = append(results, fmt.Sprintf("add: %.0f", key))
		case 11: // in old and new
			results = append(results, fmt.Sprintf("exist: %.0f", key))
		default:
			panic("not happen")
		}
	}

	sort.Strings(results)
	return results
}
