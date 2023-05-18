package utils

import (
	"reflect"
	"time"
	"unsafe"
)

//
// String
//

// MyString: string demo to reuse bytes
type MyString struct {
	data []byte
}

func (s MyString) SetValue(value string) {
	s.data = append(s.data[:0], value...)
}

func (s MyString) SetValueBytes(value []byte) {
	s.data = append(s.data[:0], value...)
}

func (s MyString) GetValue() string {
	return Bytes2string(s.data)
}

// unsafe convert: bytes and string

func Bytes2string(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func String2bytes(s string) (b []byte) {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Cap = sh.Len
	bh.Len = sh.Len
	return b
}

//
// Datetime
//

const timeLayout = "2006-01-02_15:04:05"

func FormatDateTime(ti time.Time) string {
	return ti.Format(timeLayout)
}
