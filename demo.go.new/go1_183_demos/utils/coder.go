package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
)

func JsonLoads(b []byte, s any) error {
	if reflect.TypeOf(s).Kind() != reflect.Ptr {
		return fmt.Errorf("input should be pointer")
	}

	decoder := json.NewDecoder(bytes.NewBuffer(b))
	decoder.UseNumber()
	return decoder.Decode(s)
}

func HexEncode(b []byte) string {
	return hex.EncodeToString(b)
}

func Base64Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
