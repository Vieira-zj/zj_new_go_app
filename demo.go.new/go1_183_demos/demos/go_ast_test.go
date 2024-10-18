package demos

import (
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/decorator/resolver/gopackages"
	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/packages"
)

func TestAstParseComments(t *testing.T) {
	src := `
// Calculator package provides methods 
// for basic int calculation 
package calculator

// Import of fmt package 
import "fmt"

// Add adds two integers
func Add(a, b int) int {
	// calculate the result
	result := a + b
	// return the result
	return result
}
`

	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, "", src, parser.ParseComments)
	assert.NoError(t, err)

	for _, c := range f.Comments {
		t.Log("comment:", c.Text())
	}
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
