package utils

import (
	"reflect"
	"unsafe"
)

type flag uintptr

const (
	flagKindWidth        = 5 // there are 27 kinds
	flagKindMask    flag = 1<<flagKindWidth - 1
	flagStickyRO    flag = 1 << 5
	flagEmbedRO     flag = 1 << 6
	flagIndir       flag = 1 << 7
	flagAddr        flag = 1 << 8
	flagMethod      flag = 1 << 9
	flagMethodShift      = 10
	flagRO          flag = flagStickyRO | flagEmbedRO
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

//
// Assignable if addressable and not read-only.
//
// flagAddr => panic: using unaddressable value.
// flagRO => panic: value obtained using unexported field.
//

func SetForflagUnAddrValue(val *reflect.Value) {
	flagField := reflect.ValueOf(val).Elem().FieldByName("flag")
	flagPtr := (*uintptr)(unsafe.Pointer(flagField.UnsafeAddr()))
	*flagPtr |= uintptr(flagAddr) // 设置可寻址 Addressable 标志位
}

func SetForflagROValue(val *reflect.Value) {
	flagField := reflect.ValueOf(val).Elem().FieldByName("flag")
	flagPtr := (*uintptr)(unsafe.Pointer(flagField.UnsafeAddr()))
	*flagPtr &= ^uintptr(flagRO) // 去掉 flagRO 标志位
}
