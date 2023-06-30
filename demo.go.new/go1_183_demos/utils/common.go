package utils

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

// String

func MultiSplitString(str string, splits []rune) []string {
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

// MyString: string demo to reuse bytes.
type MyString struct {
	data []byte
}

func NewMyString() MyString {
	return MyString{
		data: make([]byte, 0),
	}
}

func (s *MyString) SetValue(value string) {
	s.data = append(s.data[:0], value...)
}

func (s *MyString) SetValueBytes(value []byte) {
	s.data = append(s.data[:0], value...)
}

func (s MyString) GetValue() string {
	return Bytes2string(s.data)
}

// Bytes2string: unsafe convert bytes to string.
func Bytes2string(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// String2bytes unsafe convert string to bytes.
func String2bytes(s string) (b []byte) {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Cap = sh.Len
	bh.Len = sh.Len
	return b
}

// Datetime

const timeLayout = "2006-01-02 15:04:05"

func FormatDateTime(ti time.Time) string {
	return ti.Format(timeLayout)
}

// IO

func IsExist(path string) bool {
	_, err := os.Stat(path)
	return os.IsExist(err)
}

// Runtime

func GetCallerDetails(level int) string {
	pc, _, _, _ := runtime.Caller(level)
	details := runtime.FuncForPC(pc)
	f, line := details.FileLine(pc)
	return fmt.Sprintf("%s:%d %s", f, line, details.Name())
}

func GetGoroutineID() (int, error) {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	stk := strings.TrimPrefix(string(buf[:n]), "goroutine")

	idField := strings.Fields(stk)[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		return -1, fmt.Errorf("cannot get goroutine id: %v", err)
	}

	return id, nil
}
