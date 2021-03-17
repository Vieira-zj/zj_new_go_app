package ast

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/scanner"
	"go/token"
	"os"
	"strings"
	"testing"
)

func TestScanner(t *testing.T) {
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

func TestParserAST(t *testing.T) {
	src := []byte(`/*comment0*/
package main
import "fmt"
//comment1
/*comment2*/
func main() {
	fmt.Println("Hello, world!")
}
`)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	ast.Print(fset, f)
}

func TestInspectAST(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "./source/source1.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("package name:", f.Name.Name)

	// 遍历AST树 寻找return返回
	ast.Inspect(f, func(n ast.Node) bool {
		if ret, ok := n.(*ast.ReturnStmt); ok {
			fmt.Printf("return statement found on line %v:\n", fset.Position(ret.Pos()))
			printer.Fprint(os.Stdout, fset, ret)
			fmt.Println()
		}
		return true
	})
}

type Visitor int

func (v Visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}
	fmt.Printf("%s%T\n", strings.Repeat("\t", int(v)), n)
	return v + 1
}

func TestASTWalk(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", "package main; var a = 3", parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	// 使用visitor, 递归地打印所有的token节点
	v := new(Visitor)
	ast.Walk(v, f)
}

func init() {
	GFixedFunc = make(map[string]Fixed)
}

func TestFindAllMethod(t *testing.T) {
	// 查找方法
	file := "./source/source1.go"
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

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
	err := printer.Fprint(&fType, fset, field.Type)
	if err != nil {
		return "", err
	}
	return fType.String(), nil
}

func TestFindAllCase(t *testing.T) {
	// 查找调用了 context.WithCancel 函数，并且入参为 nil
	file := "./source/source2.go"
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
	fmt.Printf("GFixedFunc: %+v", GFixedFunc)
}
