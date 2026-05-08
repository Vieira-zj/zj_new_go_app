package structs

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJsonDuration(t *testing.T) {
	type Config struct {
		Timeout  Duration `json:"timeout"`
		Interval Duration `json:"interval"`
	}

	b := []byte(`{"timeout":"5s", "interval":"2m30s"}`)
	config := Config{}
	err := json.Unmarshal(b, &config)
	assert.NoError(t, err)

	t.Logf("timeout=%.2fs, interval=%.2fs", config.Timeout.Duration().Seconds(), config.Interval.Duration().Seconds())
}
