package utils_test

import (
	"testing"

	"demo.apps/utils"
)

func TestHexEncode(t *testing.T) {
	b := []byte{1, 15, 16}
	t.Log("hex text:", utils.HexEncode(b))
	// base64: 3 bytes to 4 bytes
	t.Log("base64 text:", utils.Base64Encode(b))
}
