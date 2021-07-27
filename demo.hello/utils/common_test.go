package utils

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"testing"
	"time"
)

func TestIsWeekDay(t *testing.T) {
	now := time.Now()
	fmt.Println("now weekday:", now.Weekday().String())
	fmt.Println("isweekday:", IsWeekDay(now))
}

func TestScheduleTask(t *testing.T) {
	ch := make(chan struct{})
	timer := time.AfterFunc(time.Second, func() {
		for i := 0; i < 3; i++ {
			fmt.Printf("schedule task run at %d\n", i)
			time.Sleep(time.Duration(500) * time.Millisecond)
		}
		ch <- struct{}{}
	})
	defer timer.Stop()

	c := time.Tick(time.Duration(300) * time.Millisecond)
outer:
	for {
		select {
		case <-c:
			fmt.Println("wait for schedule task...")
		case <-ch:
			fmt.Println("schedule task done")
			break outer
		case <-time.After(time.Duration(10) * time.Second):
			t.Fatal("time out for schedule task")
		}
	}
	fmt.Println("schedule task demo finished")
}

func TestRunCmd(t *testing.T) {
	cmd := exec.Command("ls", "-l", "/tmp/test")
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("cmd output: %s\n", b)
}

/*
# loop.sh
for i in {1..10}; do
	echo "this is shell loop test ${i}."
	sleep 1
done
*/

func TestRunShellCmd(t *testing.T) {
	output, err := RunShellCmd("sh", "/tmp/test/loop.sh")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("output:\n" + output)
}

func TestRunShellCmdInBg(t *testing.T) {
	if err := RunShellCmdInBg("sh", "/tmp/test/loop.sh"); err != nil {
		t.Fatal(err)
	}
}

/*
Get struct desc by reflection.
*/

type meta struct {
	ID   int
	Desc string
}

type user struct {
	Meta   meta
	Name   string
	Age    int `json:"age,omitempty"`
	Skills []string
}

func TestJsonDumpDefaultValue(t *testing.T) {
	data := user{}
	b, err := json.Marshal(&data)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(b))
}

func TestReflectDemo01(t *testing.T) {
	s := []string{"a", "b"}
	sType := reflect.TypeOf(s)
	fmt.Println(sType.Kind().String()) // slice

	ele := sType.Elem()
	fmt.Println(ele.Kind().String()) // string
}

func TestReflectDemo02(t *testing.T) {
	u := &user{}
	res, err := GetStructFields(reflect.TypeOf(u))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res)
}

func TestReflectDemo03(t *testing.T) {
	u := &user{}
	s := []*user{u}
	sType := reflect.TypeOf(s)
	fmt.Println(sType.Kind().String()) // slice

	ele := sType.Elem()
	if ele.Kind() == reflect.Ptr {
		ele = ele.Elem()
	}
	fmt.Println(ele.Kind().String()) // struct

	if ele.Kind() == reflect.Struct {
		res, err := GetStructFields(ele)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("%+v\n", res)
	}
}

func TestReflectDemo04(t *testing.T) {
	u := &user{}
	s := []*user{u}
	sType := reflect.TypeOf(s)

	res, err := GetSliceElements(sType)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res)
}

/*
Init struct by reflection.
*/

func TestInitByReflect01(t *testing.T) {
	// meta fields must be "public"
	m := &meta{}

	mType := reflect.TypeOf(m)
	mValue := reflect.ValueOf(m)
	if mType.Kind() == reflect.Ptr {
		mType = mType.Elem()
		mValue = mValue.Elem()
	}
	fmt.Println(mType.Kind().String())

	numFields := mType.NumField()
	for i := 0; i < numFields; i++ {
		field := mType.Field(i)
		value := mValue.Field(i)
		if field.Type.Kind() == reflect.Int {
			value.SetInt(2)
		}
		if field.Type.Kind() == reflect.String {
			value.SetString("modify")
		}
	}

	value := reflect.ValueOf(*m)
	fmt.Printf("%+v\n", value.Interface())
}

/*
Init struct from map by reflection.
*/

type fruit struct {
	ID    int     `key:"id"`
	Name  string  `key:"name"`
	Price float32 `key:"price"`
}

func TestInitByReflect02(t *testing.T) {
	data := map[string]interface{}{
		"id":    1001,
		"name":  "apple",
		"price": 16, // int
		"meta":  "desc",
	}
	fruit := &fruit{}

	if err := InitStructByReflect(data, fruit); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", *fruit)
}

/*
Format string to []byte:
string 是 immutable 的但是 []byte 是 mutable 的，所以这么转换的时候会拷贝一次；如果要不拷贝转换的话，就需要用到 unsafe 了。
*/

func TestStringToSliceByte(t *testing.T) {
	out := StringToSliceByte("hello world")
	for i := 0; i < len(out); i++ {
		fmt.Printf("%c", out[i])
	}
	fmt.Println()
}
