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

func TestMd5SumTest(t *testing.T) {
	s := "hello"
	result, err := utils.Md5Sum([]byte(s))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("md5 sum:", result)

	result = utils.Md5SumV2([]byte(s))
	t.Log("md5 sum:", result)
}

func TestCnStringConvert(t *testing.T) {
	str := "this is a test！right？it's end"
	result := utils.CnStringConvert([]byte(str))
	t.Log("result:", string(result))
}
