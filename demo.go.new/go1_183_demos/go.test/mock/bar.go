package mocktest

import (
	"encoding/json"
	"strconv"
)

// mockgen by Source mode

//go:generate go run github.com/golang/mock/mockgen -source=./bar.go -destination=./bar_mock.go -package=mocktest
type Bar interface {
	Get(key string) any
	Put(key string, value any)
}

func GetString(key string, b Bar) string {
	if len(key) == 0 {
		return "null"
	}

	val := b.Get(key)
	switch v := val.(type) {
	case int:
		return strconv.Itoa(v)
	case string:
		return v
	default:
		b, _ := json.Marshal(&val)
		return string(b)
	}
}
