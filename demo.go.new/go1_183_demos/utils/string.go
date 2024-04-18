package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

// Bytes2string unsafe converts bytes to string.
func Bytes2string(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// String2bytes unsafe converts string to bytes.
func String2bytes(s string) (b []byte) {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Cap = sh.Len
	bh.Len = sh.Len
	return b
}

func ToString(value any) string {
	if value == nil {
		return ""
	}

	switch val := value.(type) {
	case bool:
		return strconv.FormatBool(val)
	case int, uint, int8, uint8, int16, uint16, int32, uint32:
		return strconv.Itoa(val.(int))
	case int64:
		return strconv.FormatInt(val, 10)
	case uint64:
		return strconv.FormatUint(val, 10)
	case float32, float64:
		return strconv.FormatFloat(val.(float64), 'f', -1, 64)
	case []byte:
		return Bytes2string(val)
	case string:
		return val
	case fmt.Stringer:
		return val.String()
	case error:
		return val.Error()
	default:
	}

	if b, err := json.Marshal(value); err == nil {
		return Bytes2string(b)
	} else {
		return ""
	}
}

func StringMultiSplit(str string, splits []rune) []string {
	keysDict := make(map[rune]struct{}, len(splits))
	for _, key := range splits {
		keysDict[key] = struct{}{}
	}

	results := strings.FieldsFunc(str, func(r rune) bool {
		_, ok := keysDict[r]
		return ok
	})
	return results
}
