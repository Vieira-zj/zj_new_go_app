package demos

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"expvar"
	"fmt"
	"io"
	"io/ioutil"
	"net"
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
	"golang.org/x/tools/imports"
	"gopkg.in/yaml.v3"
)

func TestDemo01(t *testing.T) {
	t.Skip("it's a test, and skipped")
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
	t.Run("time ticker case1", func(t *testing.T) {
		demo0501()
	})

	t.Run("time ticker case2", func(t *testing.T) {
		demo0502()
	})
}

func TestDemoMain(t *testing.T) {
	t.Run("test rpc call case", func(t *testing.T) {
		demo06()
	})

	t.Run("test channel case01", func(t *testing.T) {
		demo0701()
	})

	t.Run("test channel case02", func(t *testing.T) {
		demo0702()
	})

	t.Run("handle panic in goroutine case", func(t *testing.T) {
		demo08()
	})

	t.Run("func deco case", func(t *testing.T) {
		demo10()
	})

	t.Run("atomic int op case", func(t *testing.T) {
		demo11()
	})
}

type person struct {
	ID   int
	Name string
}

func TestDemo06(t *testing.T) {
	t.Run("bytes copy", func(t *testing.T) {
		text := "hello world"
		b := make([]byte, 16)
		n := copy(b, text)
		fmt.Printf("%d copy bytes: %s\n", n, b[:n])
		fmt.Println()
	})

	t.Run("multi bytes copies", func(t *testing.T) {
		// 字符串较长, 多次 copy 的情况
		text := "abcdefgh"
		b := make([]byte, 4)
		res := ""
		for {
			if len(text) < len(b) {
				n := copy(b, text)
				res += string(b[:n])
				text = text[n:]
				break
			}
			n := copy(b, text)
			res += string(b[:n])
			text = text[n:]
		}
		fmt.Println("text size:", len(text))
		fmt.Printf("%d copy bytes: %s\n", len(res), res)
		fmt.Println()
	})
}

func TestDemo07(t *testing.T) {
	// slice copy
	var src []person // size = 0
	src = append(src, person{ID: 1, Name: "foo"})
	src = append(src, person{ID: 2, Name: "bar"})

	dst := make([]person, len(src))
	n := copy(dst, src)
	fmt.Printf("%d copied\n", n)

	src[0].Name = "Foo"
	src[1].ID = 12
	fmt.Printf("src: %+v\n", src)
	fmt.Printf("dst: %+v\n", dst)
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
		t.Log("done and close data channel")
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

	s = make([]int16, 10)
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

	if _, ok := m["3"]; ok {
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
	internal string `yaml:"interval"`
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
	}
	for _, s := range output.Students {
		fmt.Printf("id: %d, name: %s, internal: %s\n", s.ID, s.Name, s.internal)
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

- len: slice 中元素个数, 会初始化元素的值, 值为对应类型的默认值. 如: int 为 0, string为 ""
- cap: slice的容量, 不会初始化元素的值. 通过 append 方法来添加元素
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
	anotherFn := func() {
		fmt.Println("run another.")
	}

	var once sync.Once
	for i := 0; i < 3; i++ {
		fmt.Println("main...")
		once.Do(onceFn)
	}
	once.Do(anotherFn)
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

1. 任意类型的指针值都可以转换为 unsafe.Pointer, unsafe.Pointer 也可以转换为任意类型的指针值
2. unsafe.Pointer 与 uintptr 可以实现相互转换
3. 可以通过 uintptr 可以进行加减操作, 从而实现指针的运算
*/

func TestDemo16(t *testing.T) {
	// 读写结构内部成员
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
	// slice 指针运算
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
	fmt.Println("templated string:", output.String())
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
	UnPay:    "未支付",
	HadPay:   "已支付",
	Delivery: "配送中",
	Finish:   "已完成",
}

func orderStateMap(state int) string {
	return orderState[state]
}

func orderStateSwitch(state int) string {
	var stateDesc = ""

	switch state {
	case UnPay:
		stateDesc = "未支付"
	case HadPay:
		stateDesc = "已支付"
	case Delivery:
		stateDesc = "配送中"
	case Finish:
		stateDesc = "已完成"
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
switch case

1. switch 每个 case 最后默认带有 break
2. fallthrough 强制执行后面 case 的代码, 而不考虑 expr 结果是否为 true
*/

type mockErrorType int

const (
	RunTimeError = iota
	NilError
	IndexOutOfRange
	InvalidValue
	TestFallThrough
	UnMatchedType
)

func getErrorTypeMessage(errType mockErrorType) string {
	switch errType {
	case RunTimeError, NilError, IndexOutOfRange:
		return "unexpected error"
	case TestFallThrough:
		fallthrough
	case InvalidValue, UnMatchedType:
		return "catch exception"
	default:
		return "invalid error type"
	}
}

func TestDemoSwitch(t *testing.T) {
	errTypes := [6]mockErrorType{NilError, InvalidValue, IndexOutOfRange, UnMatchedType, TestFallThrough, 10}
	for idx, errType := range errTypes {
		fmt.Printf("%d:%s\n", idx, getErrorTypeMessage(errType))
	}
}

/*
imply an interface
*/

type sub interface {
	getData() []string
}

type subOneImpl struct {
	Data []string
}

func (sub *subOneImpl) getData() []string {
	return sub.Data
}

type subTwoImpl struct {
	Data []string
}

func (sub *subTwoImpl) getData() []string {
	return sub.Data
}

func printData(s sub) {
	fmt.Println(strings.Join(s.getData(), ","))
}

func TestDemo21(t *testing.T) {
	sub1 := subOneImpl{
		Data: []string{"1", "2", "3"},
	}
	printData(&sub1)

	sub2 := subTwoImpl{
		Data: []string{"one", "two", "three"},
	}
	printData(&sub2)
}

/*
chan close
*/

func TestDemo22(t *testing.T) {
	// 从关闭的 channel 中不但可以读取出已发送的数据, 还可以不断的读取零值
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
	// 如果通过 range 读取, channel 关闭后, 读取完已发送的数据, for 循环会跳出
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
	// Debouncing (防抖动) 是一种避免事件重复的方法, 我们设置一个小的延迟, 如果在达到延迟之前发生了其他事件, 则重启计时器
	var count uint32

	addFunc := func(value uint32) {
		fmt.Println("input:", value)
		atomic.AddUint32(&count, value)
	}

	// 取消之前的事件, 返回最新的结果
	debounce := debounce.New(time.Duration(200) * time.Millisecond)
	for i := 0; i < 3; i++ {
		for j := 0; j < 10; j++ {
			debounce(func() {
				addFunc(uint32(j))
			})
			time.Sleep(time.Duration(50) * time.Millisecond)
		}
		time.Sleep(time.Duration(300) * time.Millisecond)
	}
	fmt.Println("count:", count)
}

/*
status by bit
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
	return op&Create != 0
}

func (op Op) hasWrite() bool {
	return op&Write != 0
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
	// io, multiple writer
	// #1
	output, err := os.OpenFile("/tmp/test/output.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()

	str := "file write test.\n"
	writer := io.MultiWriter(output, os.Stdout)
	n, err := io.Copy(writer, bytes.NewReader([]byte(str)))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("write total bytes:", n)

	// #2
	filePath := "/tmp/test/test.txt"
	outputFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer outputFile.Close()

	multiWriter := io.MultiWriter(os.Stdout, outputFile)
	fmt.Fprintf(multiWriter, "multi writer test: stdout and file [%s]\n", filePath)
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
	// schedule task
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
			fmt.Println("wait for schedule task ...")
		case <-ch:
			fmt.Println("schedule task done")
			break outer
		case <-time.After(time.Duration(10) * time.Second):
			t.Fatal("time out for schedule task")
		}
	}
	fmt.Println("demo done")
}

/*
custom error, and err type check
*/

type iError interface {
	Text() string
}

type customError struct {
	desc string
}

func (e customError) Text() string {
	return e.desc
}

type badInputError struct {
	customError
}

func (e badInputError) Text() string {
	return e.customError.Text()
}

type uriNotFoundError struct {
	customError
}

func (e uriNotFoundError) Text() string {
	return e.desc
}

func TestDemo31(t *testing.T) {
	errs := make([]iError, 0, 3)
	err := customError{
		desc: "custom error",
	}
	errs = append(errs, err)

	errs = append(errs, badInputError{
		customError: err,
	})
	errs = append(errs, uriNotFoundError{
		customError: err,
	})

	errTypeCheckV1 := func(err iError) {
		if target, ok := err.(badInputError); ok {
			fmt.Println(target.Text(), "for bad input (v1)")
		} else if target, ok := err.(uriNotFoundError); ok {
			fmt.Println(target.Text(), "for uri not found (v1)")
		} else {
			fmt.Println(err.Text(), "(v1)")
		}
	}

	errTypeCheckV2 := func(err iError) {
		switch err.(type) {
		case badInputError:
			fmt.Println(err.Text(), "for bad input (v2)")
		case uriNotFoundError:
			fmt.Println(err.Text(), "for uri not found (v2)")
		default:
			fmt.Println(err.Text(), "(v2)")
		}
	}

	for _, err := range errs {
		errTypeCheckV1(err)
	}
	fmt.Println()

	for _, err := range errs {
		errTypeCheckV2(err)
	}
}

func TestDemo32(t *testing.T) {
	// imports go src code with format
	src := `
package main
import "fmt"
func main() {
  // imports test
  count :=0
  fmt.Printf("hello world, %d\n",count+1)
  fmt.Printf("add results: %d\n",add(2+3))
}
func add(a int /*number a*/, b int /*number b*/) int {
return a+b
}
`

	opt := &imports.Options{
		Comments:  false,
		TabIndent: true,
		TabWidth:  2,
	}
	dst, err := imports.Process("", []byte(src), opt)
	if err != nil {
		t.Fatal(err)
	}

	filePath := "/tmp/test/imports_test.go"
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write(dst)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("imports with format done")
}

func TestDemo33(t *testing.T) {
	// code block
	codeBlock01 := func() {
		fmt.Println("code block01 start")
		defer func() {
			fmt.Println("code block01 defer func")
		}()
		time.Sleep(time.Second)
		fmt.Println("code block01 end")
	}

	fmt.Println("code block test setup")
	{
		defer func() {
			fmt.Println("code block02 defer func")
		}()
		fmt.Println("code block02 start")
		time.Sleep(time.Second)
		fmt.Println("code block02 end")
	}

	codeBlock01()
	fmt.Println("code block test clearup")
}

func TestDemo34(t *testing.T) {
	// iterator for channel
	ch := make(chan string, 3)

	go func() {
		for i := 0; i < 6; i++ {
			ch <- strconv.Itoa(i)
		}
	}()

	go func() {
		time.Sleep(time.Duration(3) * time.Second)
		for i := 10; i < 16; i++ {
			ch <- strconv.Itoa(i)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(8)*time.Second)
	defer cancel()

	c := time.Tick(time.Second)
	for {
		select {
		case <-c:
			for {
				if len(ch) == 0 {
					fmt.Println("channel empty, and exit")
					break
				}
				val := <-ch
				fmt.Println("get value:", val)
			}
		case <-ctx.Done():
			fmt.Println("timeout and exit")
			return
		}
	}
}

/*
method with struct or pointer
*/

type myText struct {
	text string
}

func (t myText) Next() {
	// NOTE: here pass a new copy of myText "t"
	fmt.Printf("%p %p\n", &t, &t.text)
	t.text = t.text[1:]
}

func (t *myText) NewNext() {
	t.text = t.text[1:]
}

func (t myText) String() string {
	return t.text
}

func TestDemo35(t *testing.T) {
	t.Run("By Struct", func(t *testing.T) {
		text := &myText{
			text: "this is a test",
		}
		for i := 0; i < 3; i++ {
			text.Next()
			fmt.Println(text.String())
		}
	})

	t.Run("By Pointer", func(t *testing.T) {
		text := &myText{
			text: "hello world",
		}
		for i := 0; i < 3; i++ {
			text.NewNext()
			fmt.Println(text.String())
		}
	})
}

func TestDemo36(t *testing.T) {
	// select for chan when chan close
	ch := make(chan string)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go func() {
		for {
			select {
			case val, ok := <-ch:
				if !ok {
					// handle when chan close
					fmt.Println("channel close")
					return
				}
				fmt.Println("get", val)
			case <-ctx.Done():
				fmt.Println("cancel")
			}
		}
	}()

	for i := 0; i < 3; i++ {
		ch <- strconv.Itoa(i)
		time.Sleep(200 * time.Millisecond)
	}
	close(ch)
	time.Sleep(3 * time.Second)
	fmt.Println("done")
}

func TestDemo37(t *testing.T) {
	// select for chan when chan close
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ch := make(chan string)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("err:", err)
			}
		}()

		for i := 0; i < 10; i++ {
			select {
			// if ch is closed, will be panic here
			case ch <- strconv.Itoa(i):
				fmt.Println("put")
				time.Sleep(100 * time.Millisecond)
			case <-ctx.Done():
				fmt.Println("cancel")
			}
		}
	}()

	for i := 0; i < 3; i++ {
		res := <-ch
		fmt.Println("get", res)
	}
	close(ch)
	time.Sleep(time.Second)
	fmt.Println("done")
}

func TestDemo38(t *testing.T) {
	// time tick reset
	ch := make(chan struct{})

	go func() {
		for i := 0; i < 10; i++ {
			ch <- struct{}{}
			time.Sleep(200 * time.Millisecond)
		}
	}()

	i := 0
	tick := time.NewTicker(500 * time.Millisecond)
	defer tick.Stop()

outer:
	for {
		select {
		case <-tick.C:
			fmt.Println("exit")
			break outer
		case <-ch:
			i++
			fmt.Println("get value", i)
			tick.Reset(500 * time.Millisecond)
		}
	}
	fmt.Println("done")
}

func TestDemo39(t *testing.T) {
	// struct copy
	type fruit struct {
		ID    int
		Name  string
		Price int
	}

	f := &fruit{
		ID:    1,
		Name:  "apple",
		Price: 32,
	}

	dstFruit := *f
	f.Price = 45
	fmt.Printf("src fruit: %p, %v\n", f, *f)
	fmt.Printf("dst fruit: %p, %v\n", &dstFruit, dstFruit)
}

func TestDemo40(t *testing.T) {
	// closure
	type fruit struct {
		Name  string
		Price int
	}

	cb := func(fn func()) {
		fmt.Print("[run cb]: ")
		go fn()
	}

	ch := make(chan *fruit)
	go func() {
		for i := 0; i < 10; i++ {
			f := &fruit{
				Name:  "apple",
				Price: i,
			}
			ch <- f
			time.Sleep(200 * time.Millisecond)
		}
		close(ch)
	}()

	for f := range ch {
		local := f
		cb(func() {
			// when wait, the "f" value will be changed, should use "local" here
			time.Sleep(200 * time.Millisecond)
			fmt.Printf("fruit: %+v\n", *local)
		})
	}

	time.Sleep(time.Second)
	fmt.Println("done")
}

// demo41, closure
type ClosurePerson struct {
	ID   int
	Name string
}

func (p *ClosurePerson) sayHello() {
	fmt.Printf("[%d]: %s say: Hello\n", p.ID, p.Name)
}

func TestDemo41(t *testing.T) {
	type callBack func()

	callBacks := make([]callBack, 0, 10)
	for i := 0; i < 10; i++ {
		p := &ClosurePerson{
			ID:   i,
			Name: fmt.Sprintf("Tester_%d", i),
		}
		callBacks = append(callBacks, callBack(func() {
			time.Sleep(200 * time.Millisecond)
			p.sayHello()
		}))
	}

	for _, cb := range callBacks {
		// no closure issue here
		cb()
	}
	fmt.Println("done")
}

func TestDemo42(t *testing.T) {
	// sync time.AfterFunc
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				fmt.Println("cancelled")
				return
			default:
				fmt.Println("running...")
				time.Sleep(300 * time.Millisecond)
			}
		}
	}()

	time.AfterFunc(2*time.Second, func() {
		fmt.Println("close chan")
		close(done)
	})

	<-done
	time.Sleep(time.Second)
	fmt.Println("done")
}

func TestDemo43(t *testing.T) {
	// expvar
	// 1. 公共变量
	// 2. 操作都是协程安全的
	// 3. 通过 HTTP 在 /debug/vars 位置以 JSON 格式导出这些变量 (cmdline, memstats)
	kvFunc := func(kv expvar.KeyValue) {
		fmt.Println(kv.Key, kv.Value)
	}

	pubInt := expvar.NewInt("Int")
	pubInt.Set(10)
	pubInt.Add(2)

	pubFloat := expvar.NewFloat("Float")
	pubFloat.Set(1.2)
	pubFloat.Add(0.1)

	pubString := expvar.NewString("String")
	pubString.Set("hello")

	pubMap := expvar.NewMap("Map").Init()
	pubMap.Set("Int", pubInt)
	pubMap.Set("Float", pubFloat)
	pubMap.Set("String", pubString)
	pubMap.Do(kvFunc)
	fmt.Println()

	pubMap.Add("Int", 1)
	pubMap.Add("NewInt", 100)
	pubMap.AddFloat("Float", 0.5)
	pubMap.AddFloat("NewFloat", 0.9)
	pubMap.Do(kvFunc)
	fmt.Println()

	expvar.Do(kvFunc)
}

func TestDemo44(t *testing.T) {
	// context err()
	sleep := 3
	future := time.Now().Add(time.Duration(sleep) * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), future)
	defer cancel()

	go func() {
		select {
		case <-time.After(5 * time.Second):
			fmt.Println("over sleep")
		case <-ctx.Done():
			fmt.Println(ctx.Err())
		}
	}()

	time.Sleep(time.Duration(sleep+1) * time.Second)
	cancel()
	fmt.Println("done")
}

func TestDemo45(t *testing.T) {
	// selece case priority
	// 先处理 high channel, 再处理 mid channel
	const size = 5
	chHigh := make(chan int, size)
	chMid := make(chan int, size)
	for i := 0; i < size; i++ {
		chHigh <- i
		chMid <- 10 + i
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	go func() {
		for {
			select {
			case num := <-chHigh:
				fmt.Println("from high channel, get:", num)
			case <-ctx.Done():
				fmt.Println("cancelled")
				return
			default:
				select {
				case num := <-chMid:
					fmt.Println("from mid channel, get:", num)
				default:
					fmt.Println("no task get")
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	chMid <- 18
	chMid <- 19
	// 等待处理 mid channel 中的任务, 再添加任务到 high channel 中
	for i := 0; i < size; i++ {
		chHigh <- i
	}

	time.Sleep(3 * time.Second)
	cancel()
}

func TestDemo46(t *testing.T) {
	// 优先级队列
	const size = 5
	chHigh := make(chan int, size)
	chMid := make(chan int, size)
	chLow := make(chan int, size)

	for i := 0; i < size; i++ {
		chHigh <- i
		chMid <- 10 + i
		chLow <- 100 + i
	}

	worker := func(ctx context.Context) {
		chs := []chan int{chHigh, chMid, chLow}
		for {
		outer:
			for _, ch := range chs {
				select {
				case num := <-ch:
					fmt.Println("get:", num)
					break outer
				case <-ctx.Done():
					fmt.Println("cancelled")
					return
				default:
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	go worker(ctx)

	chMid <- 19
	fmt.Println("mid channel, put 19")
	chLow <- 110
	fmt.Println("low channel, put 110")
	for i := 0; i < size; i++ {
		chHigh <- i
	}
	time.Sleep(4 * time.Second)
	cancel()
}

// demo47, struct func split into 2 files
func (p *myPerson) String() string {
	skills := strings.Join(p.Skills, "|")
	return fmt.Sprintf("name:%s, age:%d, skills:%s", p.Name, p.Age, skills)
}

func (p myPerson) State() {
	fmt.Printf("arg copied [p]: %p\n", &p)
}

func (p *myPerson) StatePtr() {
	fmt.Printf("arg [*p]: %p\n", p)
}

func TestDemo47(t *testing.T) {
	p := myPerson{
		Name:   "foo",
		Age:    31,
		Skills: []string{"java", "golang", "javascript"},
	}
	p.SayHello()
	fmt.Println("src name:", p.Name)
	fmt.Println(p.String())
	fmt.Println()

	fmt.Printf("src p: %p\n", &p)
	p.State()
	p.StatePtr()

}

func TestDemo48(t *testing.T) {
	// wrap code block in a self-run func so we can defer for catch error
	func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("catch:", err)
			}
		}()

		fmt.Println("do mock a error")
		panic(fmt.Errorf("mock error"))
	}()
}

// demo49, Functional Options Pattern
// 当需要修改已有的函数时, 为了不破坏原有的签名和行为, 可以使用 Functional Options Pattern 的形式增加可变参数, 即可以增加设置项, 又能兼容已有的代码
type UserForOptionTest struct {
	Name      string
	Role      string
	MinSalary int
	MaxSalary int
}

type UserOption func(*UserForOptionTest) error

func NewUserForOptionTest(options ...UserOption) (*UserForOptionTest, error) {
	user := new(UserForOptionTest)
	for _, option := range options {
		if err := option(user); err != nil {
			return nil, err
		}
	}
	return user, nil
}

func withName(name string) UserOption {
	return func(user *UserForOptionTest) error {
		user.Name = name
		return nil
	}
}

func withRole(role string) UserOption {
	return func(user *UserForOptionTest) error {
		if role != "manager" && role != "sales" {
			return errors.New("Invalid role")
		}
		if role == "manager" {
			user.MinSalary = 20000
			user.MaxSalary = 40000
		}
		user.Role = role
		return nil
	}
}

func TestDemo49(t *testing.T) {
	user, err := NewUserForOptionTest(
		withName("foo"),
		withRole("manager"),
	)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("user: %+v\n", *user)
}

// demo50, defer
type testSlice []int

func (s *testSlice) add(i int) *testSlice {
	fmt.Println(i)
	*s = append(*s, i)
	return s
}

func TestDemo50(t *testing.T) {
	s := make(testSlice, 0)
	func() {
		defer s.add(1).add(2).add(3)
		s.add(4)
	}()
	fmt.Println("slice:", s) // 1 2 4 3
}

// demo51, interface and string type
type tPerson interface {
	SayHello()
}

type tStudent string

// tStudent 基于 string 类型, 这里参数 s 不使用指针类型
func (s tStudent) SayHello() {
	fmt.Println("Hello, my name is", string(s))
}

func TestDemo51(t *testing.T) {
	var p tPerson = tStudent("foo")
	p.SayHello()
}

func TestDemo52(t *testing.T) {
	// array 是值传递
	a := [3]int{1, 2, 3}
	func(a [3]int) {
		a[0] = 7
		fmt.Println("inner array:", a)
	}(a)
	fmt.Println("src array:", a)

	// slice 是引用传递
	s := []int{1, 2, 3}
	func(s []int) {
		s[0] = 9
		fmt.Println("inner slice:", s)
	}(s)
	fmt.Println("src slice:", s)
}

// demo53, verify interface imply
type MyReader struct{}

func (r *MyReader) Read(p []byte) (n int, err error) {
	fmt.Println("mock")
	return 0, nil
}

func TestDemo53(t *testing.T) {
	// 在编译阶段检查接口实现
	var _ io.Reader = (*MyReader)(nil)
	fmt.Println("verify interface imply")
}

func TestDemo54(t *testing.T) {
	ch1 := make(chan int)
	ch2 := make(chan int)

	go func() {
		for val := range ch1 {
			time.Sleep(10 * time.Millisecond)
			fmt.Println("get value from ch1:", val)
		}
	}()
	go func() {
		for val := range ch2 {
			time.Sleep(10 * time.Millisecond)
			fmt.Println("get value from ch2:", val)
		}
	}()

	go func() {
		for i := 0; i < 10; i++ {
			// select chan, 当 ch1 和 ch2 都可用时, 执行顺序随机
			select {
			case ch1 <- i:
				fmt.Println("pull value to ch1")
			case ch2 <- i:
				fmt.Println("pull value to ch2")
			}
			time.Sleep(200 * time.Millisecond)
		}
	}()

	time.Sleep(5 * time.Second)
	fmt.Println("Done")
}

func TestDemo55(t *testing.T) {
	ch1 := make(chan int)
	ch2 := make(chan int)

	go func() {
		for val := range ch1 {
			time.Sleep(10 * time.Millisecond)
			fmt.Println("get value from ch1:", val)
		}
	}()
	go func() {
		for val := range ch2 {
			time.Sleep(10 * time.Millisecond)
			fmt.Println("get value from ch2:", val)
		}
	}()

	go func() {
		tick := time.NewTicker(time.Second)
		defer tick.Stop()

		for i := 0; i < 10; i++ {
			// select chan, 当 ch1 和 ch2 都可用时, 使用 tick 实现优先执行 ch1
			tick.Reset(time.Second)
			select {
			case ch1 <- i:
				fmt.Println("pull value to ch1")
			case <-tick.C:
				select {
				case ch1 <- i:
					fmt.Println("pull value to ch1")
				case ch2 <- i:
					fmt.Println("pull value to ch2")
				}
			}
			time.Sleep(300 * time.Millisecond)
		}
	}()

	time.Sleep(5 * time.Second)
	fmt.Println("Done")
}

func TestDemo56(t *testing.T) {
	// sync.Cond
	locker := new(sync.Mutex)
	cond := sync.NewCond(locker)

	for i := 0; i < 5; i++ {
		local := i
		go func() {
			cond.L.Lock()
			defer cond.L.Unlock()
			fmt.Printf("goroutine [%d] start and wait\n", local)
			cond.Wait()
			fmt.Printf("goroutine [%d] resume and run\n", local)
			time.Sleep(time.Second)
		}()
	}

	for i := 0; i < 2; i++ {
		time.Sleep(time.Second)
		fmt.Println("Signal...")
		cond.Signal()
	}
	time.Sleep(3 * time.Second)
	fmt.Println("Broadcast...")
	cond.Broadcast()

	time.Sleep(3 * time.Second)
	fmt.Println("done")
}

func TestDemo95(t *testing.T) {
	// 可变参数
	myPrint := func(args ...string) {
		fmt.Println("args:", strings.Join(args, ","))
	}
	myPrint("foo", "bar", "jim")
	fmt.Println()

	// error check
	err1 := errors.New("feof")
	err2 := errors.New("feof")
	if err1 == io.EOF {
	}
	if err1 == err2 {
		fmt.Println("equal")
	} else {
		fmt.Println("not equal")
	}
	fmt.Println(errors.Is(err1, err2))
}

func TestDemo96(t *testing.T) {
	// print bytes
	b := []byte("world")
	fmt.Printf("hello %s\n", b)
	fmt.Println()

	// print char
	fmt.Println("chars:")
	for _, c := range []byte("bar") {
		fmt.Printf("%c, %d\n", c, c)
	}
	fmt.Println()

	// 泰文 bytes 转 str
	s := "\340\271\204\340\270\241\340\271\210\340\270\252\340\270\262\340\270\241\340\270\262\340\270\243\340\270\226\340\271\203\340\270\212\340\271\211\340\270\204\340\270\271\340\270\233\340\270\255\340\270\207\340\270\231\340\270\265\340\271\211"
	b = []byte(s)
	fmt.Println(string(b))
	fmt.Println()

	// byte => 2 hex => 8 bit
	// en char => 1 byte, cn word => 3 byte
	for _, str := range [2]string{"foo", "中文"} {
		b := []byte(str)
		bStr := fmt.Sprintf("%x", b)
		fmt.Printf("hex (%d): %s\n", len(bStr), bStr)
		fmt.Printf("byte (%d): %s\n", len(b), b)
	}
}

func TestDemo97(t *testing.T) {
	// bytes equal
	fmt.Println(bytes.Equal([]byte("foo"), []byte("foo")))
	fmt.Println(bytes.Equal([]byte("foo"), []byte("bar")))
	fmt.Println()

	// bytes reader
	readBytesString := func(t *testing.T, buf *bufio.Reader) []string {
		ret := make([]string, 0, 4)
		for {
			res, err := buf.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					ret = append(ret, res)
					return ret
				}
				t.Fatal(err)
			}
			ret = append(ret, res)
		}
	}

	r := bytes.NewReader([]byte("this is bytes test\nfoo,bar\nbuffer reader test"))
	buf := bufio.NewReader(r)
	fmt.Println(readBytesString(t, buf))
	fmt.Println()

	// bytes ReadByte and UnreadByte
	r = bytes.NewReader([]byte("hello"))
	buf = bufio.NewReader(r)
	b, err := buf.ReadByte()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("read byte: %c\n", b)

	fmt.Println("unread byte")
	if err := buf.UnreadByte(); err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(readBytesString(t, buf))
}

func TestDemo98(t *testing.T) {
	// string equal
	fmt.Println(strings.EqualFold("host", "host"))
	fmt.Println(strings.EqualFold("Host", "host"))
	fmt.Println(strings.EqualFold("host", "gost"))
	fmt.Println()

	// print with padding
	for _, val := range []int{123, 1331, 131008} {
		fmt.Printf("%7dms\n", val)
	}
	for _, val := range []int{123, 1331, 131008} {
		str := strconv.Itoa(val) + "ms"
		fmt.Printf("%-9seof\n", str)
	}
	fmt.Println()

	// iota
	type langType int
	const (
		Python langType = 1 << iota
		Java
		Golang
		JavaScript
	)
	fmt.Println("const:", Python, Java, Golang, JavaScript)
}

func TestDemo99(t *testing.T) {
	// iterator slice in order
	sl := []string{"one", "two", "three", "five", "four"}
	for _, value := range sl {
		fmt.Println(value)
	}
	fmt.Println()

	// delete a item of slice
	word := "one"
	for idx, item := range sl {
		if item == word {
			sl = append(sl[:idx], sl[idx+1:]...)
			break
		}
	}
	fmt.Println(sl)
	fmt.Println()

	// map marshal to json
	data := map[string]interface{}{
		"name": "foo",
		"age":  32,
	}
	b, err := json.Marshal(&data)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("json map:", string(b))
	fmt.Println()
}

func TestDemo100(t *testing.T) {
	// fetch slice 1st item
	s := []string{"one", "two", "foo", "bar"}
	value := s[0]
	copy(s, s[1:])
	s = s[:len(s)-1]
	fmt.Println("fetch 1st value:", value)
	fmt.Println(len(s), s)

	s = s[1:]
	fmt.Println(len(s), s)

	// fetch map first k,v
	m := map[int]string{
		1: "one",
		2: "two",
		3: "three",
	}

	var (
		k int
		v string
	)
	for k, v = range m {
		break
	}
	delete(m, k)

	fmt.Printf("get map 1st item: k=%d, v=%s\n", k, v)
	fmt.Printf("map: %v\n", m)
	fmt.Println()
}

func TestDemo101(t *testing.T) {
	// common lib
	// os func
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Getwd:", dir)
	fmt.Println()

	fmt.Printf("ppid:%d, pid=%d\n", os.Getppid(), os.Getpid())
	fmt.Println()

	if _, err := os.Stdout.WriteString("write stdout test\n"); err != nil {
		t.Fatal(err)
	}
	fmt.Println()

	// net func
	uri := "127.0.0.1:8080"
	host, port, err := net.SplitHostPort(uri)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("host:%s, port:%s\n", host, port)
	fmt.Println()

	// url func
	uri = "http://release.sz.test.io:8080/"
	testURL, err := url.Parse(uri)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("uri=%s, scheme=%s, host=%s, port=%s\n",
		testURL.String(), testURL.Scheme, testURL.Host, testURL.Port())
}

func TestDemo102(t *testing.T) {
	// regexp
	testStr := "test1, hello, 11, test2,test3, 99,test4"
	r, err := regexp.Compile("hello|world")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("string matched:", r.MatchString(testStr))

	r, err = regexp.Compile(`(\d\d)`)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("1st find string:", r.FindString(testStr))

	testStr = "randint(10,20)"
	r, err = regexp.Compile(`\d+`)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("find all string:", r.FindAllString(testStr, -1))
}
