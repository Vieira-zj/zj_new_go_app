package codegen

import (
	"fmt"
	"go/types"
	"path/filepath"
	"testing"

	"demo.apps/utils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/packages"
)

func TestProductParse(t *testing.T) {
	structType := getProductStruct(t)

	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		tagValue := structType.Tag(i)
		fmt.Printf("field: name=%s, tag=%s, type=%s\n", field.Name(), tagValue, field.Type().String())
	}
}

func getProductStruct(t *testing.T) *types.Struct {
	path := filepath.Join(utils.GetProjectRootPath(), "apps/code_gen")
	pkg, err := loadPackage(path)
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

	return pkgs[0], nil
}
