package utils_test

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"testing"
	"time"

	"demo.apps/utils"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/scrypt"
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

func TestGetCallerInfo(t *testing.T) {
	info := utils.GetCallerInfo(1)
	t.Log("caller info:\n", info)
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

func TestSecurity(t *testing.T) {
	str := "test123"

	t.Run("scrypt", func(t *testing.T) {
		salt := "private"
		b, err := scrypt.Key([]byte(str), []byte(salt), 32768, 8, 1, 32)
		if err != nil {
			t.Fatal(err)
		}

		t.Log("password:", hex.EncodeToString(b))
	})

	t.Run("bcrypt", func(t *testing.T) {
		hash, err := bcrypt.GenerateFromPassword([]byte(str), bcrypt.DefaultCost)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("hash:", string(hash))

		if err = bcrypt.CompareHashAndPassword(hash, []byte(str)); err != nil {
			t.Fatal(err)
		}
		t.Log("compare success")
	})
}
