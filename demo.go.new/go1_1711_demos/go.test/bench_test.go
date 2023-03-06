package gotest

import (
	"go1_1711_demo/utils"
	"testing"
)

func BenchmarkConvStr2bytes(b *testing.B) {
	s := "testString"
	var bs []byte
	for n := 0; n < b.N; n++ {
		bs = []byte(s)
	}
	_ = bs
}

func BenchmarkConvStr2bytesByUnsafe(b *testing.B) {
	s := "testString"
	var bs []byte
	for i := 0; i < b.N; i++ {
		bs = utils.Str2bytes(s)
	}
	_ = bs
}

func BenchmarkConvBytes2str(b *testing.B) {
	bs := utils.Str2bytes("testString")
	var s string
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		s = string(bs)
	}
	_ = s
}

func BenchmarkConvBytes2strByUnsafe(b *testing.B) {
	bs := utils.Str2bytes("testString")
	var s string
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		s = utils.Bytes2str(bs)
	}
	_ = s
}
