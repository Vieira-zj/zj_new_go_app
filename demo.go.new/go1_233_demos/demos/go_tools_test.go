package demos

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

// Go Tools: Analysis

func NewAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "UnexportedConstantsCheck",
		Doc:      "UnexportedConstantsCheck checks if unexported constants starts with _",
		Run:      run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		fmt.Println("start analysis go file:", file.Name.Name)
		for _, decl := range file.Decls {
			// 常量, 变量, 类型, import 这类通用声明
			genDecl, isGenDecl := decl.(*ast.GenDecl)
			if !isGenDecl {
				continue
			}
			if genDecl.Tok != token.CONST {
				continue
			}

			for _, spec := range genDecl.Specs {
				valueSpec, isValueSpec := spec.(*ast.ValueSpec)
				if !isValueSpec {
					continue
				}

				for _, name := range valueSpec.Names {
					if name.IsExported() {
						continue
					}
					if strings.HasPrefix(name.Name, "err") {
						continue
					}
					if strings.HasPrefix(name.Name, "_") {
						continue
					}
					pass.Report(newUnexportedConstantsCheckDiag(name))
				}
			}
		}
	}

	return nil, nil
}

func newUnexportedConstantsCheckDiag(i *ast.Ident) analysis.Diagnostic {
	msg := fmt.Sprintf("unexported constant %q should be prefixed with _", i.Name)
	return analysis.Diagnostic{
		Pos:     i.Pos(),
		End:     i.End(),
		Message: msg,
	}
}

func TestAnalyzer(t *testing.T) {
	t.Run("simple analysis", func(t *testing.T) {
		// mkdir -p /tmp/test/src
		// go mod init jin.example.com
		a := NewAnalyzer()
		goListPattern := "./..."
		results := analysistest.Run(t, "/tmp/test/src", a, goListPattern)
		for _, r := range results {
			for _, d := range r.Diagnostics {
				t.Log("diagnostics", d.Pos, d.End, d.Message)
			}
		}
	})
}

/* /tmp/test/src/simple/main.go
package simple

const myconstant = "myconstant" // want `unexported constant "myconstant" should be prefixed with _`

const Rate = 0

const errNotFound = "not found"

const (
    group        = "group" // want `unexported constant "group" should be prefixed with _`
    Of           = "Of"
    errConstants = "error"
    _yeah        = "yeah"
)

func aFunction(input int)int {
    const m = "hello"
    output := input * 2
    return output
}
*/
