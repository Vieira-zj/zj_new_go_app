package demos

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"reflect"
	"strings"
	"testing"
	"text/template"
)

/* Common */

func TestTemplateString01(t *testing.T) {
	tmpl := "my name is {{.}}"
	parse, err := template.New("demo").Parse(tmpl)
	if err != nil {
		t.Fatal(err)
	}
	if err = parse.Execute(os.Stdout, "foo"); err != nil {
		t.Fatal(err)
	}
	fmt.Println()
}

func TestTemplateString02(t *testing.T) {
	// {{.}} 可以展示任何数据，各种类型的数据，可以理解为接收类型是 interface{}
	tmpl := "my name is {{.}}"
	parse, err := template.New("demo").Parse(tmpl)
	if err != nil {
		t.Fatal(err)
	}
	if err = parse.Execute(os.Stdout, []string{"foo", "bar"}); err != nil {
		t.Fatal(err)
	}
	fmt.Println()
}

func TestTemplateStruct(t *testing.T) {
	tmpl := `my name is "{{.Name}}"`
	parse, err := template.New("demo").Parse(tmpl)
	if err != nil {
		t.Fatal(err)
	}
	if err = parse.Execute(os.Stdout, struct{ Name string }{Name: "foo"}); err != nil {
		t.Fatal(err)
	}
	fmt.Println()
}

func TestTemplateTrim(t *testing.T) {
	// {{-content-}} 去除前后空格
	tmpl := `    {{- . -}}    `
	parse, err := template.New("demo").Parse(tmpl)
	if err != nil {
		t.Fatal(err)
	}
	if err = parse.Execute(os.Stdout, "hello"); err != nil {
		t.Fatal(err)
	}
}

func TestTemplateIf(t *testing.T) {
	// 条件语句
	tmpl := `{{if .flag -}}
	The flag=true
	{{- else -}}
	The flag=false
	{{- end}}`
	parse, err := template.New("demo").Parse(tmpl)
	if err != nil {
		t.Fatal(err)
	}

	type _map map[string]bool
	if err = parse.Execute(os.Stdout, _map{"flag": false}); err != nil {
		t.Fatal(err)
	}
	fmt.Println()
}

func TestRange01(t *testing.T) {
	// 循环语句
	tmpl := `{{range .Array}}
	{{- . -}},
	{{- end }}`
	parse, err := template.New("demo").Parse(tmpl)
	if err != nil {
		t.Fatal(err)
	}
	if err = parse.Execute(os.Stdout, struct {
		Array []string
	}{Array: []string{"itema", "itemb", "itemc"}}); err != nil {
		t.Fatal(err)
	}
	fmt.Println()
}

func TestRangeEmpty(t *testing.T) {
	// 如果数组为空，输出else的东西
	tmpl := `{{range .Array}}
	{{- . -}},
	{{ else -}}
	array null
	{{- end}}
`
	parse, err := template.New("demo").Parse(tmpl)
	if err != nil {
		t.Fatal(err)
	}
	if err = parse.Execute(os.Stdout, struct {
		Array []string
	}{Array: []string{}}); err != nil {
		t.Fatal(err)
	}
}

func TestRange02(t *testing.T) {
	tmpl := `{{range $key, $value := .Map}}
	{{- $key}}:{{$value}},
	{{- end}}
`
	parse, err := template.New("demo").Parse(tmpl)
	if err != nil {
		t.Fatal(err)
	}

	type _map map[string]string
	if err = parse.Execute(os.Stdout, struct {
		Map _map
	}{Map: _map{"one": "1", "two": "2", "three": "3"}}); err != nil {
		t.Fatal(err)
	}
}

/*
FuncMap

- 正确：如果是一个返回类型，直接返回就行了
"ReplaceAll": func(src string, old, new string) string {
  return strings.ReplaceAll(src, old, new)
},

- 错误：不允许没有返回参数，否则直接panic
"ReplaceAll": func(src string, old, new string) {
   strings.ReplaceAll(src, old, new)
},

- 正确：参数可以不传递，但是必须有返回值的（其实可以理解，没有返回值，你渲染啥）
tem, _ := template.New("").Funcs(map[string]interface{} {
  "Echo": func() string {
    return "hello world"
  },
}).Parse(`func echo : {{Echo}}`)

- 错误：不允许有两个参数返回值，如果是两个返回值，第二个必须是error
"ReplaceAll": func(src string, old, new string) (string, string) {
  return strings.ReplaceAll(src, old, new), "111"
},

- 正确：如果两个返回类型，第二个必须是error, 顺序不能颠倒
"ReplaceAll": func(src string, old, new string) (string, error) {
  return strings.ReplaceAll(src, old, new), nil
},
*/

func TestFuncMap(t *testing.T) {
	parse, err := template.New("demo").Funcs(template.FuncMap{
		"ReplaceAll": func(src, old, new string) string {
			return strings.ReplaceAll(src, old, new)
		},
	}).Parse(`func replace: {{ReplaceAll .content "a" "A"}}`)
	if err != nil {
		t.Fatal(err)
	}

	type _map map[string]interface{}
	parse.Execute(os.Stdout, _map{
		"content": "aBC",
	})
	fmt.Println()
}

/* 内置函数 */

func TestEq(t *testing.T) {
	parse, err := template.New("demo").Parse(`{{eq .content1 .content2}}`)
	if err != nil {
		t.Fatal(err)
	}
	type _map map[string]string
	if err = parse.Execute(os.Stdout, _map{
		"content1": "a",
		"content2": "b",
	}); err != nil {
		t.Fatal(err)
	}
	fmt.Println()
}

/* Call and safeCall */

func TestCall(t *testing.T) {
	parse, err := template.New("demo").Parse(`{{call .fun .param}}`)
	if err != nil {
		t.Fatal(err)
	}
	type _map map[string]interface{}
	if err = parse.Execute(os.Stdout, _map{
		"fun":   func(str string) string { return strings.ToUpper(str) },
		"param": "abc",
	}); err != nil {
		t.Fatal(err)
	}
	fmt.Println()
}

func safeCall(fun interface{}, param ...interface{}) (val reflect.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("%v", e)
			}
		}
	}()

	funValueOf := reflect.ValueOf(fun)
	paramValueOf := reflect.ValueOf(param)

	args := make([]reflect.Value, 0, paramValueOf.Len())
	for i := 0; i < paramValueOf.Len(); i++ {
		arg := paramValueOf.Index(i).Interface()
		switch arg.(type) {
		case int:
			args = append(args, reflect.ValueOf(arg.(int)))
		case string:
			args = append(args, reflect.ValueOf(arg.(string)))
		default:
			args = append(args, reflect.ValueOf(arg))
		}
	}
	ret := funValueOf.Call(args)
	if len(ret) == 2 && !ret[1].IsNil() {
		return ret[0], ret[1].Interface().(error)
	}
	return ret[0], nil
}

func TestSafeCall(t *testing.T) {
	funcUpper := func(in string) string {
		return strings.ToUpper(in)
	}
	ret, err := safeCall(funcUpper, "abc")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(ret)

	funcAdd := func(a, b int) int {
		return a + b
	}
	ret, err = safeCall(funcAdd, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(ret)
}

/* 变量 */

func TestVariable(t *testing.T) {
	parse, err := template.New("test").Parse(`{{$a := "foo"}} hello {{$a}}`)
	if err != nil {
		t.Fatal(err)
	}
	if err = parse.Execute(os.Stdout, nil); err != nil {
		t.Fatal(err)
	}
	fmt.Println()
}

/* 模版嵌套 */

func TestTemplateInternal(t *testing.T) {
	parse, err := template.New("test").Parse(`
	{{- define "print"}}my name is {{.}} {{- end}}
	{{- template "print" .name -}}
	`)
	if err != nil {
		t.Fatal(err)
	}

	type _map map[string]string
	if err = parse.Execute(os.Stdout, _map{
		"name": "foo",
	}); err != nil {
		t.Fatal(err)
	}
	fmt.Println()
}

/* Pipeline */

func TestTemplatePipeline(t *testing.T) {
	parse, err := template.New("pipeline").Funcs(template.FuncMap{
		"split": func(name string) []string {
			return strings.Split(name, " ")
		},
		"title": func(subNames []string) string {
			retSubNames := make([]string, 0, len(subNames))
			for _, name := range subNames {
				retSubNames = append(retSubNames, strings.ToUpper(string(name[0]))+name[1:])
			}
			return strings.Join(retSubNames, " ")
		},
		"sayHello": func(name string) string {
			return "Hello, " + name
		},
	}).Parse(`pipeline: {{split .name | title | sayHello}}`)
	if err != nil {
		t.Fatal(err)
	}

	type _map map[string]string
	if err := parse.Execute(os.Stdout, _map{
		"name": "jin zheng",
	}); err != nil {
		t.Fatal(err)
	}
	fmt.Println()
}

/* Go Format */

func TestFormat(t *testing.T) {
	parse, err := template.New("format").Parse(`
	package main
	import  "fmt"


	func main(){
		fmt.Println("{{.data}}")
	}
	`)
	if err != nil {
		t.Fatal(err)
	}

	type _map map[string]string
	buf := &bytes.Buffer{}
	if err = parse.Execute(buf, _map{
		"data": "it's a test",
	}); err != nil {
		t.Fatal(err)
	}

	src, err := format.Source(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%s", src)
}
