package ast

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/scanner"
	"go/token"
	"os"
	"strings"
	"testing"
)

func TestTokenScanner(t *testing.T) {
	src := []byte(`package main
import "fmt"
//comment
func main() {
	fmt.Println("Hello, world!")
}
`)

	var s scanner.Scanner
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))
	s.Init(file, src, nil, scanner.ScanComments)

	for {
		pos, tok, lit := s.Scan()
		fmt.Printf("%-6s%-8s%q\n", fset.Position(pos), tok, lit)
		if tok == token.EOF {
			break
		}
	}
}

func TestASTParser(t *testing.T) {
	src := []byte(`/*comment0*/
package main
import "fmt"
func main() {
	//comment1
	/*comment2*/
	fmt.Println("Hello, world!")
}
`)

	fset := token.NewFileSet()
	// 0 - ignore comments
	f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	ast.Print(fset, f)
	fmt.Println()

	// inpsect one nomiated node and print source
	ast.Inspect(f, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if err := format.Node(os.Stdout, fset, fn.Body); err != nil {
				t.Fatal(err)
			}
			fmt.Println()
		}
		return true
	})

	// ast output src to file
	path := "/tmp/test/ast_output.go"
	out, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := format.Node(out, fset, f); err != nil {
		t.Fatal(err)
	}
}

func TestASTInspect(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "./example/source1.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("package name:", f.Name.Name)

	// 遍历AST树 寻找return返回
	ast.Inspect(f, func(n ast.Node) bool {
		if ret, ok := n.(*ast.ReturnStmt); ok {
			fmt.Printf("return statement found on line %v:\n", fset.Position(ret.Pos()))
			if err := printer.Fprint(os.Stdout, fset, ret); err != nil {
				t.Fatal(err)
			}
			fmt.Println()
		}
		return true
	})
}

type TestVisitor int

func (v TestVisitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}
	fmt.Printf("%s%T\n", strings.Repeat("\t", int(v)), n)
	return v + 1
}

func TestASTWalk(t *testing.T) {
	// 使用 visitor 递归打印所有的 token 节点
	src := []byte("package main; var a = 3")
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	v := new(TestVisitor)
	ast.Walk(v, f)
}

func TestAstParseMethods(t *testing.T) {
	path := "./example/source1.go"
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	// 查找方法
	ast.Inspect(f, func(n ast.Node) bool {
		if v, ok := n.(*ast.FuncDecl); ok {
			fmt.Print("method: ")
			if v.Recv != nil {
				// receiver
				rev := v.Recv.List[0]
				rName := rev.Names[0]
				if rType, err := getFieldType(fset, rev); err == nil {
					fmt.Printf("(%s %s) ", rName, rType)
				}
			}

			// method name
			mName := v.Name.Name
			if len(v.Type.Params.List) == 0 {
				fmt.Printf("%s()\n", mName)
				return true
			}

			// incoming parameters
			for _, p := range v.Type.Params.List {
				fName := p.Names[0].Name
				if fType, err := getFieldType(fset, p); err == nil {
					fmt.Printf("%s(%s %s)\n", mName, fName, fType)
				}
			}
		}
		return true
	})
}

func getFieldType(fset *token.FileSet, field *ast.Field) (string, error) {
	var fType bytes.Buffer
	if err := printer.Fprint(&fType, fset, field.Type); err != nil {
		return "", err
	}
	return fType.String(), nil
}

func TestFindAllCase(t *testing.T) {
	GFixedFunc = make(map[string]Fixed)

	// 查找调用了 context.WithCancel 函数，并且入参为 nil
	file := "./example/source2.go"
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	} else {
		GFset = fset
	}

	find := &FindContext{
		File:    file,
		Package: f.Name.Name,
	}
	ast.Walk(find, f)

	fmt.Println("\nGFixedFunc:")
	for name, fn := range GFixedFunc {
		fmt.Printf("name:%s, func:%+v", name, fn)
	}
}

/*
Generate go source code by ast.

src:
func test4a(a string) {
	context.WithCancel(nil)
}
func main() {
	test4a("hello")
}

generate dst:
func test4a(ctx context.Context, a string) {
	context.WithCancel(ctx)
}
func main() {
	ctx := context.Background()
	test4a(ctx, "hello")
}
*/

func TestASTGenerateCode(t *testing.T) {
	src := `package main
import "fmt"
func test4a(a string) {
	context.WithCancel(nil)
}
func main() {
	test4a("hello")
}
`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	v := &GeneVisit{
		fset:    fset,
		Package: f.Name.Name,
	}
	ast.Walk(v, f)

	buf := new(bytes.Buffer)
	if err := format.Node(buf, fset, f); err != nil {
		t.Fatal(err)
	}

	path := "/tmp/test/ast_genrate_src.go"
	if err := formatAndGenerateGoFile(path, buf.Bytes()); err != nil {
		t.Fatal(err)
	}
}
