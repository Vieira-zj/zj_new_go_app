package utils

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strings"
	"unicode/utf8"
)

func JsonLoad(b []byte, s any) error {
	if reflect.TypeOf(s).Kind() != reflect.Ptr {
		return fmt.Errorf("input should be pointer")
	}

	decoder := json.NewDecoder(bytes.NewBuffer(b))
	decoder.UseNumber()
	return decoder.Decode(s)
}

// Codec: 10<=>62

const chars62 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func Encode62(num int) string {
	b := []byte{}
	for num > 0 {
		b = append(b, chars62[num%62])
		num /= 62
	}
	reverseBytes(b)
	return string(b)
}

func Decode62(s string) int {
	var num int
	for i := 0; i < len(s); i++ {
		idx := strings.IndexByte(chars62, s[i])
		num += int(math.Pow(62, float64(len(s)-i-1)) * float64(idx))
	}
	return num
}

func reverseBytes(b []byte) {
	start, end := 0, len(b)-1
	for start < end {
		b[end], b[start] = b[start], b[end]
		start += 1
		end -= 1
	}
}

// Codec: Hex / Base64 / MD5

func HexEncode(b []byte) string {
	return hex.EncodeToString(b)
}

func Base64Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func Base64Decode(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

func Md5Sum(b []byte) (string, error) {
	hash := md5.New()
	if _, err := hash.Write(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func Md5SumV2(b []byte) string {
	sum := md5.Sum(b)
	// return fmt.Sprintf("%x", sum)
	// sum[:] convert [16]byte to []byte
	return hex.EncodeToString(sum[:])
}

// Gzip Compress / Decompress

func Gzip(data []byte) ([]byte, error) {
	b := bytes.Buffer{}
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}

	if err := gz.Flush(); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func Ungzip(data []byte) ([]byte, error) {
	b := bytes.NewBuffer(data)
	r, err := gzip.NewReader(b)
	if r != nil {
		defer r.Close()
	}
	if err != nil {
		return nil, err
	}

	ret := bytes.Buffer{}
	if _, err = ret.ReadFrom(r); err != nil {
		return nil, err
	}

	return ret.Bytes(), nil
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
