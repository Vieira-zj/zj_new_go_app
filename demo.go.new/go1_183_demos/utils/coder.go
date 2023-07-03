package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"unicode/utf8"
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

// CnStringConvert converts cn symbols in the string to en symbols.
func CnStringConvert(input []byte) []byte {
	for i := 0; i < len(input); {
		if isAscIIChar(input[i]) {
			i++
			continue
		}

		r, size := utf8.DecodeRune(input[i:])
		if size == 0 {
			break
		}
		b := convertCnChar(r)
		copy(input[i:i+size], b)
		i += size
	}

	return input
}

func convertCnChar(ch rune) []byte {
	switch ch {
	case '！':
		return []byte("!  ")
	case '？':
		return []byte("?  ")
	default:
		return []byte("nil")
	}
}

func isAscIIChar(ch byte) bool {
	return ch&0x80 != 0x80
}
