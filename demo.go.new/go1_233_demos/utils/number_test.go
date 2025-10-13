package utils_test

import (
	"testing"

	"zjin.goapp.demo/utils"
)

func TestFormatWithPrecision(t *testing.T) {
	f := 3.1461

	t.Run("round float with precision", func(t *testing.T) {
		t.Log("result:", utils.RoundWithPrecision(f, 1))
		t.Log("result:", utils.RoundWithPrecision(f, 2))
	})

	t.Run("floor float with precision", func(t *testing.T) {
		t.Log("result:", utils.FloorWithPrecision(f, 1))
		t.Log("result:", utils.FloorWithPrecision(f, 2))
	})
}
