package utils_test

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"zjin.goapp.demo/utils"
)

func TestIsImplement(t *testing.T) {
	// The background logic for noCopy checking in "go vet".

	const src = `
package demo

type Greeter interface {
	Greet() string
}

type Human struct {}
func (p Human) Greet() string {
	return "Hello!"
}

type Robot struct {}
func (r *Robot) Charge() {}
`

	fset := token.NewFileSet()

	// Parse the source code and create an AST
	f, err := parser.ParseFile(fset, "demo.go", src, parser.AllErrors)
	assert.NoError(t, err, "Parse source code failed")

	// Configure the type checking
	conf := types.Config{Importer: importer.Default()}
	info := &types.Info{
		Defs: make(map[*ast.Ident]types.Object),
	}

	// Type check the package
	pkg, err := conf.Check("demo", fset, []*ast.File{f}, info)
	assert.NoError(t, err, "Type checking failed")

	// Extract the target Greeter interface object
	greeterObj := pkg.Scope().Lookup("Greeter")
	assert.NotNil(t, greeterObj, "Greeter interface not found")
	greeterIface, ok := greeterObj.Type().Underlying().(*types.Interface)
	assert.True(t, ok, "Greeter is not an interface")

	// Extract the Human and Robot struct objects
	humanType := pkg.Scope().Lookup("Human").Type()
	assert.NotNil(t, humanType, "Human struct not found")
	robotType := pkg.Scope().Lookup("Robot").Type()
	assert.NotNil(t, robotType, "Robot struct not found")

	// Check if the structs implement the interface
	assert.True(t, utils.IsImplement(humanType, greeterIface), "Human should implement Greeter")
	assert.False(t, utils.IsImplement(robotType, greeterIface), "Robot should not implement Greeter")
}
