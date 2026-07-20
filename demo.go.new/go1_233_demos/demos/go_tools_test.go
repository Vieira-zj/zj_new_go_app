package demos

import (
	"fmt"
	"go/ast"
	"go/token"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/singleflight"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

// Go Built-In Modules

func TestRuntimeHooks(t *testing.T) {
	// Runtime Hooks
	// - AddCleanup 负责登记对象不可达后的清理函数
	// - KeepAlive  标出对象必须保持可达的最后位置
	// - SetFinalizer 留给旧代码维护和少数特殊情况

	t.Run("runtime clearup", func(t *testing.T) {
		t.Skip()

		type MockFile struct {
			fd int
		}

		openFile := func(path string) (*MockFile, error) {
			fd, err := syscall.Open(path, syscall.O_RDONLY, 0644)
			if err != nil {
				return nil, err
			}

			f := &MockFile{fd: fd}
			runtime.AddCleanup(f, func(fd int) {
				_ = syscall.Close(fd)
			}, fd)
			return f, nil
		}

		_, err := openFile("/tmp/test/out.json")
		assert.NoError(t, err)
	})
}

// Go Module: decimal

func TestDecimalCals(t *testing.T) {
	// float64 适合科学计算, decimal/int64 适合财务计算
	t.Run("float calculation", func(t *testing.T) {
		price := 99.995
		taxRate := 0.33
		tax := price * taxRate
		total := price + tax
		t.Logf("float total: %.3f", total)

		t.Log("float equal:", 0.1+0.2 == 0.3)
	})

	t.Run("decimal calculation", func(t *testing.T) {
		price := decimal.NewFromFloat(99.995)
		taxRate := decimal.NewFromFloat(0.13)
		tax := price.Mul(taxRate)
		total := price.Add(tax)
		t.Logf("decimal total: %s", total.StringFixed(3))
	})
}

// Go Module: singleflight

func TestSingleFlight(t *testing.T) {
	callCount := 0
	mockFetchData := func(key string) (string, error) {
		if len(key) == 0 {
			return "", fmt.Errorf("key is empty")
		}
		callCount++
		fmt.Printf("fetching data for key '%s' from origin (call #%d)...\n", key, callCount)
		time.Sleep(500 * time.Millisecond)
		return "mock_data_for_key|" + key, nil
	}

	const testKey = "singleflight_demo01"
	g := singleflight.Group{}
	wg := sync.WaitGroup{}

	for i := range 5 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			result, err, shared := g.Do(testKey, func() (any, error) {
				return mockFetchData(testKey)
			})
			if err != nil {
				fmt.Printf("goroutine %d: error fetching data: %v\n", id, err)
				return
			}
			fmt.Printf("goroutine %d: received result: '%v' (shared: %t)\n", id, result, shared)
		}(i)
	}
	wg.Wait()
	t.Logf("total calls to fetch data: %d", callCount)
}

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
