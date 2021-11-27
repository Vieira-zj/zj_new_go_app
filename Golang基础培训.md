# Golang 培训

## 概述

### 培训目的

了解 Golang 语言。

### 优点

1. 高并发
2. 部署简单

------

## Go 环境

### 安装

下载安装包 <https://golang.org/>

设置环境变量：

```sh
GOROOT=/usr/local/go
GOPATH=/Users/jinzheng/Workspaces/.go
GOBIN=/Users/jinzheng/Workspaces/.go/bin
```

### `go mod` 包管理

类似 `pip`, `mvn`。

1. 创建项目

```sh
go mod init [go_module_name]
```

查看 `go.mod` 文件。

2. 安装依赖库

```sh
go mod download
go mod tidy
```

### 项目结构

- 程序入口必须是 `main()` 函数。
- 目录为一个 `package`。

### 编译与执行

例子：diskusage 计算一个目录下文件总数及所占用总的磁盘空间。

1. 编译

```sh
go build .
```

2. 执行

```sh
# 本地调试 编译+执行
go run main.go

./diskusage -v ~/Workspaces/.go
./diskusage -v -p 8 ~/Workspaces/.go
```

优点：

- 不依赖任务第三方库
- 支持平台包括 linux, andorid, ios, window
  - 测试工具 atx

------

## 基本语法

### 基础类型

```text
int
uint
int8
int16
int32
int64

float32
float64

byte
string
rune

nil
```

注：int和uint是可变大小类型，如果是32位CPU就是4个字节，如果是64位就是8个字节。

#### rune 类型

用于表示unicode字符。

```golang
for _, c := range "abcd" {
	fmt.Printf("%c\n", c)
}

for _, c := range "中文" {
	fmt.Printf("%c\n", c)
}
```

### 变量声明与赋值

```golang
var i int
var s = "hello"

func print(s string) {
	local := "hello"
	fmt.Println(local)
}
```

任何类型在未初始化时都对应一个零值：bool是false、int是0、string串是""；而指针、slice、channel和map的零值都是nil。

- 常量

```golang
const c = "hello"
```

- byte 与数组转换

```golang
s := "hello"
b := []byte(s)
s1 := string(b)
```

### 控制语句

```text
if [cond_1] {
	...
} else if [cond_2] {
	...
} else {
	...
}

i := "python"
switch i {
case "python":
	fmt.Println("1")
case "java":
	fmt.Println("2")
default:
	fmt.Println("default")
}

# 没有 while
for {
	...
}

for i := 0; i < 10; i++ {
	// ...
}

goto
```

### 函数

```golang
func add(x, y int) int {
	return x + y
}
```

没有默认参数、不支持函数重载。

### 错误处理

```golang
func div(x, y int) (int, error) {
	if y == 0 {
		return -1, fmt.Errorf("y is zero.")
	}
	return x/y, nil
}

res, err := div(x, y)
if err != nil {
	// ...
	panic(err)
}
```

### defer 函数

```golang
func myFunc() {
	defer func() {
		if err := recover(); err != nil {
			// ...
		}
		conn.Close()
	}()
	// ...
	panic("error")
}
```

### 常用库

- `fmt.Printf()`

```text
%s    字符串
%d    数字
%.2f  浮点型数字
%v    默认输出
%c    一个字符，参数对应ASCII码
%p    十六进制指针
```

- "math"

函数定义为 float64 类型，需要类型转换。

- "strings"
- "os"
- "path/filepath"
- "encoding/json"
- "log"
- "text/template"
- "net/http"

------

## 基本数据结构

### Slice

```golang
// 数组
s := [2]string{"a", "b"}
s[0] = "c"
// 切片
s := []string{"a", "b"}

s := make([]string, 0, 10)
s = append(s, "a")

for idx, item := range s {
	// ...
}
```

没有 Set 数据结构。

### Map

```golang
m := make(map[int]string, 10)

val, ok := m[key]
if !ok {
	// ...
}

for k, v := range m {
	// ...
}
```

### Struct 结构体

定义：

```golang
type Product struct {
	ID   int      `json:"id"`
	Name string   `json:"name"`
	private string
}

func (p Product) string() string {
	// ...
}
```

注：大写开头的字段为 public, 小写开头的字段为 private. 同样规则适用于包中变量和方法。

结构体初始化：

```golang
var p Product
p.ID = 1
p.Name = "apple"

p := Product{
	ID:   1,
	Name: "apple",
}
p.string()
```

- `interface{}`

类似于 Java 中的 `Object`。

### 接口

```golang
// 声明interface
type Birds interface {
  Twitter() string
  Fly(high int) bool
}

// 继承
type Chicken interface {
  Bird
  Walk()
}
```

### 实体变量与指针

```golang
p := &Product{
	ID:   1,
	Name: "apple",
}
fmt.Println(*p)

p := new(Product)
p.ID = 1
p.Name = "apple"
fmt.Println(*p)
```

int, string, array, struct 为非引用类型；slice, map, chan, 指针 为引用类型。

什么时候用实体变量，什么时候用指针？

------

## 测试

### 单元测试

```golang
// 方法名 Test 开头
func TestAdd(t *testing.T) {
	expect := 3
	res := add(1, 2)
	if res != expect {
		t.Errorf("want %d, but got %d\n", expect, res)
	}
}
```

执行：

```sh
go test -timeout 10s -run ^TestAdd$ go_module_test -v -count=1
```

支持单测覆盖率统计。

------

## 应用

### json 处理

- Struct to Json

```golang
p := Product{
	ID:   1,
	Name: "food",
}

b, _ := json.Marshal(p)
fmt.Println(string(b))
```

结构体中必须是大写字母开头的成员才会被 json 处理到。

- Json to Struct

```golang
var p Product
json.Unmarshal(b, &p)
fmt.Println(p)
```

- Json to Map

```golang
m := make(map[string]interface{})
json.Unmarshal(b, &m)
fmt.Println(m)
```

### IO 文件读写

```golang
// read
b, _ := os.ReadFile("file_path")
fmt.Println(string(b))
// write
os.WriteFile("file_path", b, 0644)

// reader read() and writer write()
f, _ := os.Open("file_path")
buf := bufio.NewReader(f)
buf.ReadLine()
```

### Http api 请求

```golang
resp, _ := http.Get("url")
resp, _ := http.Post("url", "application/json", bytes.NewReader([]byte("body")))

fmt.Println(resp.StatusCode)
body, _ := ioutil.ReadAll(resp.Body)
resp.Body.Close()
fmt.Println(body)
```

### Web 开发

```golang
func main() {
	router := http.NewServeMux()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "golang web")
	})

	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	fmt.Println("http serve at :8080")
	server.ListenAndServe()
}
```

API 测试 `curl http://localhost:8080/`

### 平台开发

- python: Django
- java: SrpingBoot
- golang: gin + gorm + casbin

------

## 并发

### Goroutine (协程)

```golang
go func() {
	// ...
}()
```

非常轻量级。

### Channel (管道)

```golang
var ch chan<- int // 只接收int, 不能发送
var ch <-chan int // 只发送int, 不能接收

ch := make(chan string, 1)
res := <-ch
ch <- "test"

for val := range ch {
	// ...
}

close(ch)
```

### sync 库

#### 锁

- sync.Mutex
- sync.RWMutex

#### 原子类型

- "sync/atomic"

#### WaitGroup

```golang
var wg sync.WaitGroup

for i := 0; i < 3; i++ {
	wg.Add(1)
	go func() {
		defer wg.Done()
		// ...
	}()
}
wg.Wait()
```

#### Context

todo:

------

## 使用场景

- golang
  - 微服务
  - k8s 开发

- java
  - web server / 微服务
  - android
  - 大数据
  - 测试

- python
  - 大数据、机器学习
  - 测试

