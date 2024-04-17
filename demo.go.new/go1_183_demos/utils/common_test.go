package utils_test

import (
	"encoding/hex"
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

func TestIsNil(t *testing.T) {
	var x interface{}
	var y *int = nil
	x = y

	typOf := reflect.TypeOf(x)
	t.Logf("kind:%v, type:%s", typOf.Kind(), typOf.String())

	//nolint:staticcheck
	t.Run("nil interface var check", func(t *testing.T) {
		// 即使接口持有的值为 nil, 也并不意味着接口本身为 nil.
		t.Log("x != nil", x != nil) // true
		t.Log("x:", x)
	})

	t.Run("nil check by reflect", func(t *testing.T) {
		t.Log("IsNil:", utils.IsNil(x))
	})
}

func TestTrackTime(t *testing.T) {
	defer utils.TrackTime()()

	t.Log("start ...")
	time.Sleep(time.Second)
	t.Log("end")
}

func TestDelFirstNItemsOfSlice(t *testing.T) {
	makeSlice := func() []any {
		s := make([]any, 0, 10)
		for i := 0; i < 10; i++ {
			s = append(s, i)
		}
		return s
	}

	n := 4

	t.Run("case1", func(t *testing.T) {
		s := makeSlice()
		res, err := utils.DelFirstNItemsOfSlice(s, n)
		if err != nil {
			t.Fatal(err)
		}

		t.Log("results:", len(res), res)
	})

	t.Run("case2", func(t *testing.T) {
		s := makeSlice()
		s = s[n:]
		t.Log("results:", len(s), s)
	})
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
