package structs_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"zjin.goapp.demo/structs"
)

func TestMarshalMap(t *testing.T) {
	t.Run("custom marshal and omit empty", func(t *testing.T) {
		m := structs.MarshalMap{
			"int":          1,
			"empty_string": "",
			"zero_int":     0,
			"string":       "hello",
			"bool_false":   false,
		}

		b, err := json.Marshal(&m)
		assert.NoError(t, err)
		t.Log("marshal:", string(b))
	})

	t.Run("custom unmarshal and omit empty", func(t *testing.T) {
		b := `{"int":1,"empty_string":"","zero_int":0,"string":"hello","bool_false":false}`
		m := structs.MarshalMap{}
		err := json.Unmarshal([]byte(b), &m)
		assert.NoError(t, err)
		for k, v := range m {
			t.Logf("key=%s, value=%v", k, v)
		}
	})
}
