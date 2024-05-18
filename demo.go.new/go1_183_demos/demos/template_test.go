package demos_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"text/template"

	"demo.apps/utils"
)

func TestTmplCustomDelims(t *testing.T) {
	tmpl := `hello, <<.Name>>!`
	parse, err := template.New("demo").Delims("<<", ">>").Parse(tmpl)
	if err != nil {
		t.Fatal(err)
	}

	if err = parse.Execute(os.Stdout, struct{ Name string }{"Foo"}); err != nil {
		t.Fatal(err)
	}
}

func TestTmplEmptySliceRange(t *testing.T) {
	// 如果数组为空, 则走 else
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

func TestTmplMapRange(t *testing.T) {
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

// Template Func Demos

func TestTmplCustomFn(t *testing.T) {
	parse, err := template.New("demo").Funcs(template.FuncMap{
		"ReplaceAll": func(src, old, new string) string {
			return strings.ReplaceAll(src, old, new)
		},
	}).Parse(`replace: {{.content}} => {{ReplaceAll .content "a" "A"}}`)
	if err != nil {
		t.Fatal(err)
	}

	type _map map[string]interface{}
	parse.Execute(os.Stdout, _map{
		"content": "aBCdax",
	})
	fmt.Println()
}

func TestTmplCall(t *testing.T) {
	parse, err := template.New("demo").Parse(`result: {{call .fun .param}}`)
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

// Template Pipeline Demos

func TestTmplPipeline01(t *testing.T) {
	fnMap := template.FuncMap{
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
	}

	parse, err := template.New("pipeline").Funcs(fnMap).Parse(`pipeline: {{split .name | title | sayHello}}`)
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

func TestTmplPipeline02(t *testing.T) {
	fnMap := template.FuncMap{
		"uppercase": func(s string) string {
			return strings.ToUpper(s)
		},
		"repeat": func(count, s string) (string, error) {
			c, err := strconv.Atoi(count)
			if err != nil {
				return "", err
			}
			return strings.Repeat(s, c), nil
		},
	}

	tmpl, err := template.New("pipeline").Funcs(fnMap).Parse(`Hello, {{. | uppercase | repeat "3"}}!`)
	if err != nil {
		t.Fatal(err)
	}

	if err := tmpl.Execute(os.Stdout, "foo"); err != nil {
		t.Fatal(err)
	}
	fmt.Println()
}

func TestTmplPipeline03(t *testing.T) {
	fnMap := template.FuncMap{
		"base64_decode": func(s string) ([]byte, error) {
			return utils.Base64Decode(s)
		},
		"json_unmarshal": func(b []byte) (map[string]any, error) {
			m := make(map[string]any)
			err := json.Unmarshal(b, &m)
			return m, err
		},
		"get_value": func(key string, m map[string]any) any {
			value, ok := m[key]
			if !ok {
				return "not_exist"
			}
			return value
		},
	}

	tmpl, err := template.New("pipeline").Funcs(fnMap).Parse(`value: {{. | base64_decode | json_unmarshal | get_value "age"}}`)
	if err != nil {
		t.Fatal(err)
	}

	// {"name":"foo", "age":41}
	if err := tmpl.Execute(os.Stdout, "eyJuYW1lIjoiZm9vIiwiYWdlIjo0MX0="); err != nil {
		t.Fatal(err)
	}
	fmt.Println()
}
