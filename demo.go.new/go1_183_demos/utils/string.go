package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func ConcatString(sl ...string) string {
	n := 0
	for i := 0; i < len(sl); i++ {
		n += len(sl[i])
	}

	b := strings.Builder{}
	b.Grow(n)
	for _, s := range sl {
		b.WriteString(s)
	}
	return b.String()
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

func StrMultiSplit(str string, splits []rune) []string {
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
