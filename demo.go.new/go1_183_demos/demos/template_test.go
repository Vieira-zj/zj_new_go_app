package demos_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"text/template"
)

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

func TestTemplPipeline(t *testing.T) {
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
