package utils_test

import (
	"strconv"
	"testing"

	"demo.apps/utils"
)

func TestMultiSplitString(t *testing.T) {
	fields := utils.MultiSplitString("a,b.c|d.e|f,g", []rune{',', '.', '|'})
	for _, field := range fields {
		t.Log("field:", field)
	}
}

func TestToString(t *testing.T) {
	t.Logf("str value: %s", utils.ToString(false))
	t.Logf("str value: %s", utils.ToString(11.2))
	t.Logf("str value: %s", utils.ToString(map[string]int{"one": 1, "two": 2}))
}

func TestMyString(t *testing.T) {
	s := utils.NewMyString()
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
	s := utils.MyString{}
	for n := 0; n < b.N; n++ {
		s.SetValue(strconv.Itoa(n))
		s.GetValue()
	}
}
