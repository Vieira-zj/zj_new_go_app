package codegen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"testing"

	"demo.apps/utils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/packages"
)

func TestParseProduct(t *testing.T) {
	structType := getProductStruct(t)

	fmt.Println("\nProduct struct info:")
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		tagValue := structType.Tag(i)
		fmt.Printf("field: name=%s, tag=%s, type=%s\n", field.Name(), tagValue, field.Type().String())
	}
}

func TestParseProductPreComment(t *testing.T) {
	// 1. parse package info
	pkgPath := filepath.Join(utils.GetProjectRootPath(), "apps/code_gen")
	cfg := packages.Config{Mode: packages.NeedTypes | packages.NeedImports | packages.NeedFiles}
	pkgs, err := packages.Load(&cfg, pkgPath)
	assert.NoError(t, err)
	assert.True(t, len(pkgs) > 0)

	pkg := pkgs[0]
	assert.True(t, len(pkg.GoFiles) > 0)

	// 2. parse comments for product.go file
	fpath := pkg.GoFiles[0]
	fmt.Println("parse comments for:", fpath)

	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, fpath, nil, parser.ParseComments)
	assert.NoError(t, err)

	comments := make(map[string]*ast.CommentGroup, len(f.Comments))
	for _, c := range f.Comments {
		k := fmt.Sprintf("%s|%d", filepath.Base(fpath), fs.Position(c.Pos()).Line)
		comments[k] = c
	}

	// 3. print comment for 'Product'
	srcTypeName := "Product"
	obj := pkg.Types.Scope().Lookup(srcTypeName)
	assert.NotNil(t, obj)

	pos := pkg.Fset.Position(obj.Pos()).Line - 1
	key := fmt.Sprintf("%s|%d", filepath.Base(fpath), pos)
	comment, ok := comments[key]
	assert.True(t, ok)

	fmt.Printf("comment for '%s':\n%s\n", srcTypeName, comment.Text())
}

func getProductStruct(t *testing.T) *types.Struct {
	pkgPath := filepath.Join(utils.GetProjectRootPath(), "apps/code_gen")
	pkg, err := loadPackage(pkgPath)
	assert.NoError(t, err)

	srcTypeName := "Product"
	obj := pkg.Types.Scope().Lookup(srcTypeName)
	assert.NotNil(t, obj, fmt.Sprintf("%s not found in declared types of %s", srcTypeName, pkg))

	structType, ok := obj.Type().Underlying().((*types.Struct))
	assert.True(t, ok, fmt.Sprintf("type %v is not a struct", obj))

	return structType
}

func loadPackage(path string) (*packages.Package, error) {
	cfg := packages.Config{Mode: packages.NeedTypes | packages.NeedImports}
	pkgs, err := packages.Load(&cfg, path)
	if err != nil {
		return nil, fmt.Errorf("loading packages for inspection: %v", err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		panic("loading packages error")
	}
	if len(pkgs) < 1 {
		return nil, fmt.Errorf("no package found for path: %s", path)
	}

	pkg := pkgs[0]
	b, err := pkg.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("package marshal json error: %v", err)
	}

	fmt.Printf("package info:\n%s\n", b)
	return pkg, nil
}
