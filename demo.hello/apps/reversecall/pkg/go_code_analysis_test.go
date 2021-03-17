package pkg

import (
	"fmt"
	"go/build"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnalysisCallGraphy(t *testing.T) {
	pkg := "demo.hello/apps/reversecall/pkg/test"
	args := []string{pkg}
	a, err := RunAnalysis(&build.Default, false, args)
	if err != nil {
		t.Fatal(err)
	}

	opts := &RenderOpts{
		// Focus: "build.Package.ImportPath",
		// Ignore: []string{"third", "backend/common", fmt.Sprintf("%s/vendor", project)},
		// Include: []string{"backend/code_inspector/testing_bai"},
		Nointer: false,
		Nostd:   true,
	}
	callMap, err := a.Render(pkg, opts)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range callMap {
		log.Printf("正向调用关系:%s %+v", k, v)
	}
}

func TestAnalysisReverseCallGraphy(t *testing.T) {
	pkg := "demo.hello/apps/reversecall/pkg/test"
	args := []string{pkg}
	a, err := RunAnalysis(&build.Default, false, args)
	if err != nil {
		t.Fatal(err)
	}

	opts := &RenderOpts{
		Nointer: false,
		Nostd:   true,
	}
	callMap, err := a.Render(pkg, opts)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range callMap {
		log.Printf("正向调用关系:%s %+v", k, v)
	}

	paramsList := make([]MWTreeBuildParam, 2)
	rootPath := filepath.Join(os.Getenv("HOME") + "Workspaces/zj_repos/zj_go2_project/demo.hello/apps/reversecall/pkg/test/example/")
	paramsList[0] = MWTreeBuildParam{
		GoFilePath: filepath.Join(rootPath, "test4.go"),
		PkgName:    "example",
		FnName:     "Test4a",
	}
	paramsList[1] = MWTreeBuildParam{
		GoFilePath: filepath.Join(rootPath, "test3.go"),
		PkgName:    "example",
		FnName:     "XYZ@print", // 类方法
	}

	tree := &MWTree{}
	for _, param := range paramsList {
		tree.BuildReverseCallTreeFromCallMap(param, callMap)
		relationList := tree.GetReverseCallRelations()

		for _, relation := range relationList {
			fns := make([]string, len(relation.Callees))
			for i, funcDesc := range relation.Callees {
				fns[i] = fmt.Sprintf("%s.%s", funcDesc.Package, funcDesc.Name)
			}
			log.Printf("反向调用链:%s", strings.Join(fns, "<-"))
		}
	}
}
