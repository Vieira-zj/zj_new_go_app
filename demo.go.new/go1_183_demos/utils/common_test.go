package utils_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"demo.apps/utils"
)

func TestFormatDateTime(t *testing.T) {
	result := utils.FormatDateTime(time.Now())
	t.Log("now:", result)
}

func TestMyString(t *testing.T) {
	s := utils.NewMyString()
	s.SetValue("hello")
	t.Log("value:", s.GetValue())
}

func TestMultiSplitString(t *testing.T) {
	fields := utils.MultiSplitString("a,b.c|d.e|f,g", []rune{',', '.', '|'})
	for _, field := range fields {
		t.Log("field:", field)
	}
}

func TestIsDirExist(t *testing.T) {
	for _, path := range []string{
		"/tmp/test",
		"/tmp/test/mock",
		"/tmp/test/test.json",
	} {
		result := utils.IsExist(path)
		t.Logf("%s is exist: %v", path, result)
		result = utils.IsDirExist(path)
		t.Logf("%s is dir exist: %v\n", path, result)
	}
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

func TestGetLocalIPAddr(t *testing.T) {
	addr, err := utils.GetLocalIPAddr()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("local ip addr:", addr)

	addr, err = utils.GetLocalIPAddrByDial()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("local ip addr:", addr)
}

func TestGetCallerDetails(t *testing.T) {
	details := utils.GetCallerDetails(1)
	t.Log("details:\n", details)
}

func TestGetGoroutineID(t *testing.T) {
	ch := make(chan int)
	for i := 0; i < 3; i++ {
		idx := i
		go func() {
			id, err := utils.GetGoroutineID()
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Printf("[%d] goroutine id: %d, start\n", idx, id)
			for val := range ch {
				fmt.Printf("[%d] goroutine id: %d, get value: %d\n", idx, id, val)
			}
			fmt.Printf("[%d] goroutine id: %d, exit\n", idx, id)
		}()
	}

	for i := 0; i < 10; i++ {
		ch <- i
		time.Sleep(10 * time.Millisecond)
	}
	close(ch)

	time.Sleep(100 * time.Millisecond)
	t.Log("test goroutine id done")
}
