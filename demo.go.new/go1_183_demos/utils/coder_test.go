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

func TestMd5Sum(t *testing.T) {
	s := "hello"
	result, err := utils.Md5Sum([]byte(s))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("md5 sum:", result)

	result = utils.Md5SumV2([]byte(s))
	t.Log("md5 sum:", result)
}

func TestGzipCompress(t *testing.T) {
	data := []byte("MyzYrIyMLyNqwDSTBqSwM2D6KD9sA8S/d3Vyy6ldE+oRVdWyqNQrjTxQ6uG3XBOS0P4GGaIMJEPQ=")
	t.Log("origin data size:", len(data))

	zipData, err := utils.Gzip(data)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("compress data size:", len(zipData))

	unzipData, err := utils.Ungzip(zipData)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("uncompress data size:", len(unzipData))
	t.Log("uncompress data:", string(unzipData))
}

func TestCnStringConvert(t *testing.T) {
	str := "this is a test！right？it's end"
	result := utils.CnStringConvert([]byte(str))
	t.Log("result:", string(result))
}
