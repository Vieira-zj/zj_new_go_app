package demos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"text/template"
	"time"
	"unsafe"

	"github.com/bep/debounce"
	"gopkg.in/yaml.v2"
)

func TestDemo01(t *testing.T) {
	want := "demo01"
	if got := demo01(); got != want {
		t.Errorf("demo01() = %q, want %q", got, want)
	}
}

func TestDemo02(t *testing.T) {
	want := 3
	if got := demo02(); got != want {
		t.Errorf("demo02() = %d, want %d", got, want)
	}
}

func TestDemo03(t *testing.T) {
	want := 5
	if got := demo03(); got != want {
		t.Errorf("demo03() = %d, want %d", got, want)
	}
}

func TestDemo04(t *testing.T) {
	demo04()
}

func TestDemo05(t *testing.T) {
	demo0502()
}

type person struct {
	ID   int
	Name string
}

func TestDemo06(t *testing.T) {
	// slice copy
	src := make([]*person, 0)
	src = append(src, &person{ID: 1, Name: "foo"})
	src = append(src, &person{ID: 2, Name: "bar"})

	dst := make([]*person, len(src))
	fmt.Println("total copied:", copy(dst, src))
	for _, p := range dst {
		fmt.Printf("%+v\n", *p)
	}
}

func TestDemo07(t *testing.T) {
	// slice deep copy
	src := make([]person, 0)
	src = append(src, person{ID: 1, Name: "foo"})
	src = append(src, person{ID: 2, Name: "bar"})

	dst := make([]person, len(src))
	fmt.Println("total copied:", copy(dst, src))
	src[0].Name = "fooNew"

	fmt.Println("src:", src)
	fmt.Println("dst:", dst)
}

func TestDemo08(t *testing.T) {
	// exclude:
	// 1. matched, remove item from src. res = append(res[:i], res[i+1:]...)
	// 2. not matched, append item to out.
	// prefer to use 2, because slice append op perf is better than remove op.
	names := []string{"foo", "bar", "hello", "world"}
	exclude := []string{"foo", "world"}

	res := make([]string, 0)
	for _, name := range names {
		matched := false
		for _, item := range exclude {
			if name == item {
				matched = true
				break
			}
		}
		if !matched {
			res = append(res, name)
		}
	}
	fmt.Println("results:", res)
}

func TestDemo09(t *testing.T) {
	// channels are completely thread safe.
	// They are the official way to communicate between goroutines.
	produce := func(wg *sync.WaitGroup, ch chan int, num int) {
		defer wg.Done()
		for i := 0; i < num; i++ {
			ch <- i
		}
	}

	ch := make(chan int)
	wg := sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go produce(&wg, ch, 10000)
	}

	go func() {
		wg.Wait()
		t.Log("done and close data channel.")
		close(ch)
	}()

	idx := 0
	for n := range ch {
		fmt.Fprint(ioutil.Discard, n)
		idx++
	}
	if idx != 50000 {
		t.Fatalf("want 5000, and go %d\n", idx)
	}
}

func TestDemo10(t *testing.T) {
	// slice size and cap
	s := make([]int16, 0, 10)
	s = append(s, 1)
	s = append(s, 2)
	fmt.Printf("len=%d, cap=%d, value:%v\n", len(s), cap(s), s)

	s = make([]int16, 10)
	s = append(s, 10)
	s = append(s, 20)
	fmt.Printf("len=%d, cap=%d, value:%v\n", len(s), cap(s), s)

	s = make([]int16, 10, 10)
	s[0] = 10
	s[1] = 20
	fmt.Printf("len=%d, cap=%d, value:%v\n", len(s), cap(s), s)
}

func TestDemo11(t *testing.T) {
	// get map value
	m := make(map[string]string, 3)
	m["1"] = "one"
	m["2"] = "two"

	fmt.Println("map:", m)
	if m["3"] != "" {
		t.Fatalf("want empty, got %v\n", m["3"])
	}

	_, ok := m["3"]
	if ok {
		t.Fatalf("want false, got %v\n", ok)
	}
}

/*
yaml load
*/

type data struct {
	Students []student `yaml:"students"`
}

type student struct {
	ID       int    `yaml:"id"`
	Name     string `yaml:"name"`
	internal string `yaml:"internal"`
}

func TestDemo12(t *testing.T) {
	input := `
students:
- id: 1010
  name: tester_a
  interval: private desc
- id: 1011
  name: tester_b
  interval: private desc
`

	// field "internal" will not be exported
	output := data{}
	if err := yaml.Unmarshal([]byte(input), &output); err != nil {
		t.Fatal(err)
	} else {
		fmt.Printf("%+v", output)
	}
}

/*
json dump bytes
*/

type hotfix struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Patch []byte `json:"patch"`
}

func TestDemo13(t *testing.T) {
	fix := hotfix{
		Op:    "add",
		Path:  "json/dump",
		Patch: []byte("fix: hello world"),
	}

	// "patch" bytes will dump as base64 string
	// echo -n "Zml4OiBoZWxsbyB3b3JsZA==" | base64 -D
	if out, err := json.Marshal(fix); err != nil {
		t.Fatal(err)
	} else {
		fmt.Println(string(out))
	}
}

/*
make slice by len and cap:
s := make([]type, len, cap)
len: slice????????????????????????????????????????????????????????????????????????????????????int?????????string??????
cap: slice????????????????????????????????????????????????append?????????????????????
*/

func TestDemo14(t *testing.T) {
	// default len=3, cap=3
	s1 := make([]int, 3)
	fmt.Println(len(s1), cap(s1), "values:")
	for _, v := range s1 {
		fmt.Printf("%d,", v)
	}
	fmt.Printf("\n\n")

	// len=0, cap=3
	s2 := make([]int, 0, 3)
	fmt.Println(len(s2), cap(s2))

	s2 = append(s2, 1)
	s2 = append(s2, 2)
	fmt.Println(len(s2), cap(s2), "values:")
	for _, v := range s2 {
		fmt.Printf("%d,", v)
	}
	fmt.Println()
}

/*
sync.Once
*/

func TestDemo15(t *testing.T) {
	onceFn := func() {
		fmt.Println("run once.")
	}

	var once sync.Once
	for i := 0; i < 3; i++ {
		fmt.Println("main...")
		once.Do(onceFn)
	}
	fmt.Println()

	once = sync.Once{}
	var wg sync.WaitGroup
	// wg := sync.WaitGroup{}
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			fmt.Printf("sub %d ...\n", i)
			once.Do(onceFn)
		}(i)
	}
	wg.Wait()
	fmt.Println("sync once test done.")
}

/*
unsafe.Pointer

1. ??????????????????????????????????????????unsafe.Pointer, unsafe.Pointer??????????????????????????????????????????
2. unsafe.Pointer???uintptr????????????????????????
3. ????????????uintptr??????????????????????????????????????????????????????
*/

func TestDemo16(t *testing.T) {
	// ????????????????????????
	str1 := "hello world"
	hdr1 := (*reflect.StringHeader)(unsafe.Pointer(&str1))
	fmt.Printf("str:%s, data addr:%d, len:%d\n", str1, hdr1.Data, hdr1.Len)

	str2 := "abc"
	hdr2 := (*reflect.StringHeader)(unsafe.Pointer(&str2))

	hdr1.Data = hdr2.Data
	hdr1.Len = hdr2.Len
	fmt.Printf("str:%s, data addr:%d, len:%d\n", str1, hdr1.Data, hdr1.Len)
}

func TestDemo17(t *testing.T) {
	// slice????????????
	data := []byte("abcd")
	for i := 0; i < len(data); i++ {
		ptr := unsafe.Pointer(uintptr(unsafe.Pointer(&data[0])) + uintptr(i)*unsafe.Sizeof(data[0]))
		fmt.Printf("%c,", *(*byte)(ptr))
	}
	fmt.Println()
}

/*
ref (pointer) and value
*/

type T struct {
	ls []int
	v  int
}

func foo(t T) {
	t.ls[0] = 88 // "ls" is ref, will change source value
	t.v = 99
}

func TestDemo18(t *testing.T) {
	st := T{ls: []int{1, 2, 3}}
	foo(st)
	fmt.Println(st)
}

func TestDemo19(t *testing.T) {
	// template with func
	const templateText = `{{.Name}} last friend : {{last .Friends}}`

	templateFunc := make(map[string]interface{})
	templateFunc["last"] = func(s []string) string {
		return s[len(s)-1]
	}

	type Recipient struct {
		Name    string
		Friends []string
	}
	recipient := Recipient{
		Name:    "Jack",
		Friends: []string{"Bob", "Json", "Tom"},
	}

	temp := template.Must(template.New("TemplateWithFuncs").Funcs(templateFunc).Parse(templateText))
	var output bytes.Buffer
	if err := temp.Execute(&output, recipient); err != nil {
		t.Fatal(err)
	}
	fmt.Println("templated string:", string(output.Bytes()))
}

func TestDemo20(t *testing.T) {
	expired := 3
	old := time.Now().Unix() + int64(expired)
	time.Sleep(time.Duration(4) * time.Second)
	if time.Now().Unix() > old {
		fmt.Println("expired")
	} else {
		fmt.Println("not expired")
	}
}

/*
map and switch
*/

const (
	UnPay = iota
	HadPay
	Delivery
	Finish
)

var orderState = map[int]string{
	UnPay:    "?????????",
	HadPay:   "?????????",
	Delivery: "?????????",
	Finish:   "?????????",
}

func orderStateMap(state int) string {
	return orderState[state]
}

func orderStateSwitch(state int) string {
	var stateDesc = ""

	switch state {
	case UnPay:
		stateDesc = "?????????"
	case HadPay:
		stateDesc = "?????????"
	case Delivery:
		stateDesc = "?????????"
	case Finish:
		stateDesc = "?????????"
	}
	return stateDesc
}

func BenchmarkMap(b *testing.B) {
	// BenchmarkMap-16  74934553  13.6 ns/op
	for n := 0; n < b.N; n++ {
		orderStateMap(0)
		orderStateMap(1)
		orderStateMap(2)
		orderStateMap(3)
	}
}

func BenchmarkSwitch(b *testing.B) {
	// BenchmarkSwitch-16  1000000000  0.226 ns/op
	for n := 0; n < b.N; n++ {
		orderStateSwitch(0)
		orderStateSwitch(1)
		orderStateSwitch(2)
		orderStateSwitch(3)
	}
}

/*
switch
*/

type mockErrorType int

const (
	RunTimeError = iota
	NilError
	IndexOutOfRange
	InvalidValue
	UnMatchedType
)

func getErrorTypeMessage(errType mockErrorType) string {
	switch errType {
	case RunTimeError, NilError, IndexOutOfRange:
		return "unexpected error"
	case InvalidValue, UnMatchedType:
		return "catch exception"
	default:
		return "invalid error type"
	}
}

func TestDemoSwitch(t *testing.T) {
	for _, errType := range []mockErrorType{NilError, InvalidValue, IndexOutOfRange, UnMatchedType, 10} {
		fmt.Println(getErrorTypeMessage(errType))
	}
}

/*
imply an interface
*/

type sub interface {
	getData() []string
}

type subOne struct {
	Data []string
}

func (sub *subOne) getData() []string {
	return sub.Data
}

type subTwo struct {
	Data []string
}

func (sub *subTwo) getData() []string {
	return sub.Data
}

func printData(s sub) {
	fmt.Println(strings.Join(s.getData(), ","))
}

func TestDemo21(t *testing.T) {
	sub1 := subOne{
		Data: []string{"1", "2", "3"},
	}
	printData(&sub1)

	sub2 := subTwo{
		Data: []string{"one", "two", "three"},
	}
	printData(&sub2)
}

/*
chan close
*/

func TestDemo22(t *testing.T) {
	// ????????????channel???????????????????????????????????????????????????????????????????????????
	ch := make(chan int, 5)
	for i := 1; i < 4; i++ {
		ch <- i
	}
	close(ch)

	for i := 0; i < 5; i++ {
		fmt.Println(<-ch)
	}
}

func TestDemo23(t *testing.T) {
	// ????????????range?????????channel??????????????????????????????????????????for???????????????
	ch := make(chan int, 5)
	for i := 1; i < 4; i++ {
		ch <- i
	}
	close(ch)
	fmt.Println("channel closed")

	for i := range ch {
		fmt.Println(i)
	}
}

func TestDemo24(t *testing.T) {
	// Debouncing??????????????? ?????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????
	var count uint32

	addFunc := func() {
		atomic.AddUint32(&count, 1)
	}

	debounce := debounce.New(time.Duration(200) * time.Millisecond)
	for i := 0; i < 3; i++ {
		for j := 0; j < 10; j++ {
			debounce(addFunc)
			time.Sleep(time.Duration(50) * time.Millisecond)
		}
		time.Sleep(time.Duration(200) * time.Millisecond)
	}
	fmt.Println("count:", count)
}

/*
status by binary
refer: github.com/fsnotify/fsnotify
*/

type Op uint32

const (
	Create Op = 1 << iota
	Write
	Remove
	Rename
)

func (op Op) String() string {
	var buffer bytes.Buffer
	if op&Create == Create {
		buffer.WriteString("|Create")
	}
	if op&Write == Write {
		buffer.WriteString("|Write")
	}
	if op&Remove == Remove {
		buffer.WriteString("|Remove")
	}
	if op&Rename == Rename {
		buffer.WriteString("|Rename")
	}

	if buffer.Len() == 0 {
		return ""
	}
	return buffer.String()[1:]
}

func (op Op) hasCreate() bool {
	if op&Create != 0 {
		return true
	}
	return false
}

func (op Op) hasWrite() bool {
	if op&Write != 0 {
		return true
	}
	return false
}

func TestDemo25(t *testing.T) {
	fmt.Printf("all operations: create=%d, write=%d, remove=%d, rename=%d\n", Create, Write, Remove, Rename)
	op := Create
	op += Rename
	op += Remove
	fmt.Println("op has create:", op.hasCreate())
	fmt.Println("op has write:", op.hasWrite())
	fmt.Println("cur op:", op)
}

/*
async task, and cancel
*/

func myTask(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup, idx int) {
	defer wg.Done()
	if (idx > 0) && (idx%7 == 0) {
		time.Sleep(time.Duration(3) * time.Second)
		fmt.Printf("task %d mock failed\n", idx)
		cancel()
		return
	}

	ch := make(chan struct{})
	go func(ch chan struct{}, idx int) {
		fmt.Printf("task %d process ...\n", idx)
		time.Sleep(time.Duration(idx) * time.Second)
		ch <- struct{}{}
	}(ch, idx)

	select {
	case <-ch:
		fmt.Printf("task %d done\n", idx)
		return
	case <-ctx.Done():
		fmt.Printf("task %d cancelled\n", idx)
		return
	}
}

func TestDemo26(t *testing.T) {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go myTask(ctx, cancel, &wg, i)
		}
		fmt.Println("all tasks started")
	}()

	time.Sleep(time.Second) // wait wg.Add(1) done
	wg.Wait()
	fmt.Println("all tasks finished")
}

func TestDemo2701(t *testing.T) {
	// sort string slice
	input := "this is a string slice sort demo"
	words := strings.Split(input, " ")
	sort.Strings(words)
	fmt.Println("sorted words:", words)
}

func TestDemo2702(t *testing.T) {
	// sort slice
	printChars := func(chars []rune) {
		for _, ch := range chars {
			fmt.Printf("%c,", ch)
		}
		println()
	}

	// sort slice of char
	word := "helloworld"
	sl := make([]rune, 0, len(word))
	for _, ch := range word {
		sl = append(sl, ch)
	}
	fmt.Println("src slice:")
	printChars(sl)

	sort.Slice(sl, func(i, j int) bool {
		return sl[i] < sl[j]
	})
	fmt.Println("sorted slice:")
	printChars(sl)

	// sort slice of interface
	input := "this is a slice sort test"
	words := strings.Split(input, " ")
	s := make([]interface{}, 0, len(words))
	for _, word := range words {
		s = append(s, word)
	}
	fmt.Println("\nsrc slice:", s)

	sort.Slice(s, func(i, j int) bool {
		srcWord := s[i].(string)
		dstWord := s[j].(string)
		return srcWord[0] < dstWord[0]
	})
	fmt.Println("sorted slice:", s)
}

func TestDemo28(t *testing.T) {
	// io, write content to file
	str := "file write test.\n"
	output, err := os.OpenFile("/tmp/test/output.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()

	writer := io.MultiWriter(output, os.Stdout)
	n, err := io.Copy(writer, bytes.NewReader([]byte(str)))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("write total bytes:", n)
}

func TestDemo29(t *testing.T) {
	// regexp
	re, err := regexp.Compile(`\[(.+)\]`)
	if err != nil {
		t.Fatal(err)
	}

	str := "[AS][Android][TH]Merchant portal send noti to partner app users"
	if res := re.FindAllString(str, -1); res != nil {
		fmt.Println("results:", res)
	} else {
		fmt.Println("no matched")
	}
}

func TestDemo30(t *testing.T) {
	// print bytes
	b := []byte("world")
	fmt.Printf("hello %s\n", b)

	sayHello := func(args ...string) {
		fmt.Println("hello", strings.Join(args, ","))
	}
	sayHello("foo", "bar")

	// delete item of slice
	s := []string{"a", "b", "c"}
	c := "a"
	for idx, item := range s {
		if item == c {
			s = append(s[:idx], s[idx+1:]...)
			break
		}
	}
	fmt.Println(s)
}

func TestDemo99(t *testing.T) {
	// equal
	fmt.Println(strings.EqualFold("host", "host"))
	fmt.Println(strings.EqualFold("Host", "host"))
	fmt.Println(strings.EqualFold("host", "gost"))

	// print with padding
	fmt.Println()
	for _, val := range []int{123, 1331, 131008} {
		fmt.Printf("%7dms\n", val)
	}

	for _, val := range []int{123, 1331, 131008} {
		str := strconv.Itoa(val) + "ms"
		fmt.Printf("%-9seof\n", str)
	}

	// url parse
	fmt.Println()
	uri := "http://release.i.sppay.sz.shopee.io/"
	testURL, err := url.Parse(uri)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("uri=%s, scheme=%s, host=%s\n", testURL.String(), testURL.Scheme, testURL.Host)
}
