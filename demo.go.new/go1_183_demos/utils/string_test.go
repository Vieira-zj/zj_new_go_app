package utils_test

import (
	"strconv"
	"testing"

	"demo.apps/utils"
)

func TestToString(t *testing.T) {
	t.Logf("str value: %s", utils.ToString(false))
	t.Logf("str value: %s", utils.ToString(11.2))
	t.Logf("str value: %s", utils.ToString(map[string]int{"one": 1, "two": 2}))
}

func TestMultiSplitString(t *testing.T) {
	fields := utils.MultiSplitString("a,b.c|d.e|f,g", []rune{',', '.', '|'})
	for _, field := range fields {
		t.Log("field:", field)
	}
}

// Demo: string with reuse bytes

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
	return utils.Bytes2string(s.data)
}

func TestMyString(t *testing.T) {
	s := NewMyString()
	s.SetValue("hello")
	t.Log("value:", s.GetValue())
}

// go test -bench=BenchmarkString -run=^$ -benchtime=5s -benchmem -v
func BenchmarkString(b *testing.B) {
	s := ""
	for n := 0; n < b.N; n++ {
		s = strconv.Itoa(n)
	}
	_ = s
}

// go test -bench=BenchmarkMyString -run=^$ -benchtime=5s -benchmem -v
func BenchmarkMyString(b *testing.B) {
	s := MyString{}
	for n := 0; n < b.N; n++ {
		s.SetValue(strconv.Itoa(n))
		s.GetValue()
	}
}
