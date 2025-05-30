package demos

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"strconv"
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/decorator/resolver/gopackages"
	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/packages"
)

// Go Ast

func TestAstParseComments(t *testing.T) {
	src := `
// Calculator package provides methods 
// for basic int calculation 
package calculator

// Import of fmt package 
import "fmt"

// This is a global variable
var gtotal int

// Add adds two integers
func Add(a, b int) int {
	// calculate the result
	gtotal = a + b
	// return the result
	return gtotal
}
`

	fs := token.NewFileSet()
	root, err := parser.ParseFile(fs, "", src, parser.ParseComments)
	assert.NoError(t, err)

	t.Run("print all comments", func(t *testing.T) {
		for _, c := range root.Comments {
			t.Log("comment:", c.Text())
		}
	})

	t.Run("print func doc comment", func(t *testing.T) {
		for _, decl := range root.Decls {
			if funcDecl, ok := decl.(*ast.FuncDecl); ok {
				t.Log("function:", funcDecl.Name.String())
				assert.NotNil(t, funcDecl.Doc, "funcDecl.Doc should not be nil")
				t.Log("doc comment:", funcDecl.Doc.Text())
			}
		}
	})

	t.Run("print var doc comment", func(t *testing.T) {
		ast.Inspect(root, func(n ast.Node) bool {
			if genDecl, ok := n.(*ast.GenDecl); ok && genDecl.Tok == token.VAR {
				for _, spec := range genDecl.Specs {
					if valueSpec, ok := spec.(*ast.ValueSpec); ok {
						t.Log("variable:", valueSpec.Names[0].Name)
						assert.NotNil(t, genDecl.Doc, "genDecl.Doc should not be nil")
						t.Log("gen decl doc comment:", genDecl.Doc.Text())
						assert.Nil(t, valueSpec.Doc)
					}
				}
			}
			return true
		})
	})
}

func TestAstReverseOrder(t *testing.T) {
	code := `package a
func main(){
	var a int    // foo
	var b string // bar
}
`

	t.Run("parse by ast", func(t *testing.T) {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, "", code, parser.ParseComments)
		assert.NoError(t, err)

		list := f.Decls[0].(*ast.FuncDecl).Body.List
		list[0], list[1] = list[1], list[0]

		// output with incorrect position of comments
		err = format.Node(os.Stdout, fset, f)
		assert.NoError(t, err)
	})

	t.Run("parse by dst", func(t *testing.T) {
		f, err := decorator.Parse(code)
		assert.NoError(t, err)

		list := f.Decls[0].(*dst.FuncDecl).Body.List
		list[0], list[1] = list[1], list[0]

		err = decorator.Print(f)
		assert.NoError(t, err)
	})
}

func TestAstParseVarAlias(t *testing.T) {
	src := `
package main

type intAlias = int

func main() {
	var x intAlias
	println(x)
}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "example.go", src, 0)
	assert.NoError(t, err)

	// 创建类型检查器
	conf := types.Config{Importer: nil}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
	}

	_, err = conf.Check("example", fset, []*ast.File{file}, info)
	assert.NoError(t, err)

	for ident, obj := range info.Defs {
		if ident.Name == "x" {
			fmt.Println("var x type is:", obj.Type().String())
		}
	}
}

// Go Dst

func TestDstAddComments(t *testing.T) {
	code := `package main
func main() {
	println("Hello World!")
}`

	f, err := decorator.Parse(code)
	assert.NoError(t, err)

	call := f.Decls[0].(*dst.FuncDecl).Body.List[0].(*dst.ExprStmt).X.(*dst.CallExpr)

	call.Decs.Start.Append("// you can add comments at the start...")
	call.Decs.Fun.Append("/* ...in the middle... */")
	call.Decs.End.Append("// or at the end.")

	err = decorator.Print(f)
	assert.NoError(t, err)
}

func TestDstAddSpaces(t *testing.T) {
	code := `package main

func main() {
	println(a, b, c)
}`

	f, err := decorator.Parse(code)
	assert.NoError(t, err)

	call := f.Decls[0].(*dst.FuncDecl).Body.List[0].(*dst.ExprStmt).X.(*dst.CallExpr)

	call.Decs.Before = dst.EmptyLine
	call.Decs.After = dst.EmptyLine

	for _, v := range call.Args {
		v := v.(*dst.Ident)
		v.Decs.Before = dst.NewLine
		v.Decs.After = dst.NewLine
	}

	err = decorator.Print(f)
	assert.NoError(t, err)
}

func TestDstAddCommentsAndSpaces(t *testing.T) {
	code := `package main

func main() {
	var i int
	i++
	println(i)
}`

	f, err := decorator.Parse(code)
	assert.NoError(t, err)

	list := f.Decls[0].(*dst.FuncDecl).Body.List

	list[0].Decorations().Before = dst.NewLine
	list[0].Decorations().End.Append("// the Decorations method allows access to the common")
	list[1].Decorations().End.Append("// decoration properties (Before, Start, End and After)")
	list[2].Decorations().End.Append("// for all nodes.")
	list[2].Decorations().After = dst.EmptyLine

	err = decorator.Print(f)
	assert.NoError(t, err)
}

func TestDstCloneAndUpdate(t *testing.T) {
	code := `package main
var i /* a */ int`

	f, err := decorator.Parse(code)
	assert.NoError(t, err)

	// reuse a node with clone
	cloned := dst.Clone(f.Decls[0]).(*dst.GenDecl)

	cloned.Decs.Before = dst.NewLine
	cloned.Specs[0].(*dst.ValueSpec).Names[0].Name = "j"
	cloned.Specs[0].(*dst.ValueSpec).Names[0].Decs.End.Replace("/* b */")

	f.Decls = append(f.Decls, cloned)

	err = decorator.Print(f)
	assert.NoError(t, err)
}

func TestDstLoadPackages(t *testing.T) {
	dir := "/tmp/test/go_project"
	pkgs, err := decorator.Load(&packages.Config{
		Dir: dir,
		//nolint:deprecated
		Mode: packages.LoadSyntax,
	}, "root")
	assert.NoError(t, err)

	p := pkgs[0]
	f := p.Syntax[0]

	// add a call expression
	newStmt := &dst.ExprStmt{
		X: &dst.CallExpr{
			Fun: &dst.Ident{Path: "fmt", Name: "Println"},
			Args: []dst.Expr{
				&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote("Hello, World!")},
			},
		},
	}

	b := f.Decls[0].(*dst.FuncDecl).Body
	b.List = append(b.List, newStmt)

	// add import
	r := decorator.NewRestorerWithImports("root", gopackages.New(dir))
	err = r.Print(f)
	assert.NoError(t, err)
}
