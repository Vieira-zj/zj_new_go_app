package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/build"
	"log"
	"strings"

	"demo.hello/apps/reversecall/pkg"
)

/*
Refer:
- https://github.com/baixiaoustc/go_code_analysis
- https://github.com/ofabry/go-callvis
*/

var (
	fullPackage string
	goFilePath  string
	packageName string
	funcName    string

	nointer bool
	include string
	ignore  string

	help bool
)

func getItemsFromString(input string) []string {
	if len(input) == 0 {
		return []string{}
	}
	return strings.Split(input, ",")
}

func main() {
	flag.StringVar(&fullPackage, "fullpackage", "", "Full package import path.")
	flag.StringVar(&goFilePath, "gofile", "", "Go file path.")
	flag.StringVar(&packageName, "package", "", "Package name.")
	flag.StringVar(&funcName, "func", "", "Function name.")

	flag.BoolVar(&nointer, "nointer", false, "Whether include internal (private) functions.")
	flag.StringVar(&include, "include", "", "Include packages with matched prefix.")
	flag.StringVar(&ignore, "ignore", "", "Ignore packages with matched prefix.")

	flag.BoolVar(&help, "help", false, "Help.")

	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	args := []string{fullPackage}
	a, err := pkg.RunAnalysis(&build.Default, false, args)
	if err != nil {
		panic(fmt.Sprintln("analysis error:", err))
	}

	opts := &pkg.RenderOpts{
		Nointer: nointer,
		Nostd:   true,
		Include: getItemsFromString(include),
		Ignore:  getItemsFromString(ignore),
	}
	b, err := json.Marshal(opts)
	if err != nil {
		panic(fmt.Sprintln("opts marshal error:", err))
	}
	log.Println("render options:", string(b))

	callMap, err := a.Render(fullPackage, opts)
	if err != nil {
		panic(fmt.Sprintln("render process error:", err))
	}

	tree := &pkg.MWTree{}
	param := pkg.MWTreeBuildParam{
		GoFilePath: goFilePath,
		PkgName:    packageName,
		FnName:     funcName,
	}
	tree.BuildReverseCallTreeFromCallMap(param, callMap)

	relationList := tree.GetReverseCallRelations()
	for _, relation := range relationList {
		funcs := make([]string, len(relation.Callees), len(relation.Callees)+1)
		for i, funcDesc := range relation.Callees {
			funcs[i] = fmt.Sprintf("%s.%s", funcDesc.Package, funcDesc.Name)
		}

		if len(funcs) == 1 {
			funcs = append(funcs, "null")
		}
		log.Printf("反向调用链:%s", strings.Join(funcs, "<-"))
	}
}
