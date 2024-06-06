package jsondiff

import (
	"encoding/json"
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
