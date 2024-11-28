package utils_test

import (
	"encoding/hex"
	"reflect"
	"testing"
	"time"

	"demo.apps/utils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/scrypt"
)

func TestBinaryCeil(t *testing.T) {
	result := utils.BinaryCeil(12)
	t.Log("result:", result)
}

func TestTrackTime(t *testing.T) {
	defer utils.TrackTime()()

	t.Log("start ...")
	time.Sleep(time.Second)
	t.Log("end")
}

func TestIsNil01(t *testing.T) {
	var x any
	var y *int = nil
	x = y

	typOf := reflect.TypeOf(x)
	t.Logf("kind:%v, type:%s", typOf.Kind(), typOf.String())

	t.Run("nil any var check", func(t *testing.T) {
		runCase := func(a any) {
			// 即使接口持有的值为 nil, 也并不意味着接口本身为 nil
			t.Log("x:", a)
			t.Log("x != nil", a != nil) // true
		}
		runCase(x)
	})

	t.Run("is nil check", func(t *testing.T) {
		t.Log("IsNil:", utils.IsNil(x))
	})
}

type MyInterface interface {
	apply()
}

type MyInterfaceImpl struct{}

func (*MyInterfaceImpl) apply() {}

// 在编译阶段检查 impl 接口实现
var _ MyInterface = (*MyInterfaceImpl)(nil)

func TestIsNil02(t *testing.T) {
	var x MyInterface
	var y *MyInterfaceImpl = nil
	x = y

	typOf := reflect.TypeOf(x)
	t.Logf("kind:%v, type:%s", typOf.Kind(), typOf.String())

	t.Run("nil interface var check", func(t *testing.T) {
		runCase := func(a MyInterface) {
			t.Log("x:", a)
			t.Log("x != nil", a != nil) // true
		}
		runCase(x)
	})

	t.Run("is nil check", func(t *testing.T) {
		t.Log("IsNil:", utils.IsNil(x))
	})
}

type MyPersonImpl struct {
	Name string
	Age  int
}

func TestIsEmptyStruct(t *testing.T) {
	t.Run("empty person", func(t *testing.T) {
		p := MyPersonImpl{}
		assert.True(t, utils.IsEmptyStruct(p))
	})

	t.Run("empty person pointer", func(t *testing.T) {
		p := MyPersonImpl{}
		assert.True(t, utils.IsEmptyStruct(&p))
	})

	t.Run("person with name", func(t *testing.T) {
		p := MyPersonImpl{}
		p.Name = "foo"
		assert.False(t, utils.IsEmptyStruct(p))
	})

	t.Run("person with age", func(t *testing.T) {
		p := MyPersonImpl{}
		p.Age = 40
		assert.False(t, utils.IsEmptyStruct(p))
	})
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
