package jsondiff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasherForMap(t *testing.T) {
	src := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
	}
	hasher := hasher{}
	srcHash := hasher.digest(src)
	t.Log("src hash:", srcHash)

	dst := map[string]int{
		"three": 3,
		"one":   1,
		"two":   2,
	}
	dstHash := hasher.digest(dst)
	t.Log("dst hash:", dstHash)
	assert.True(t, srcHash == dstHash)
}
