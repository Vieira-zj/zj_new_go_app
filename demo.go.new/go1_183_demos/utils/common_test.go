package utils_test

import (
	"encoding/hex"
	"fmt"
	"reflect"
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
	t.Run("get package path", func(t *testing.T) {
		type S struct{}
		typeOf := reflect.TypeOf(S{})
		t.Log("pkg path:", typeOf.PkgPath())
	})

	t.Run("get caller info", func(t *testing.T) {
		callerInfo := utils.GetCallerInfo(1)
		t.Log("caller info:\n", callerInfo)
	})
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

func TestGetFuncDeclare(t *testing.T) {
	for _, fn := range []any{
		utils.GetLocalIPAddr,
		utils.GetCallerInfo,
	} {
		result, err := utils.GetFuncDeclare(fn)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("func desc:", result)
	}
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
