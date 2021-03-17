package main

import (
	"flag"
	"fmt"

	funcdiff "demo.hello/apps/funcdiff/pkg"
)

var (
	help       bool
	srcPath    string
	targetPath string
	isextend   bool
)

func main() {
	flag.BoolVar(&help, "h", false, "help.")
	flag.StringVar(&srcPath, "s", "", "source: go file path to diff.")
	flag.StringVar(&targetPath, "t", "", "target: go file path to diff.")
	flag.BoolVar(&isextend, "e", false, "show extend func diff info.")

	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	if isextend {
		fmt.Printf("func diff extend between [%s] and [%s]:\n", srcPath, targetPath)
		diffResults, err := funcdiff.GoFileDiffByFuncExt(srcPath, targetPath)
		if err != nil {
			panic(err)
		}
		for _, entry := range diffResults {
			fmt.Printf("[%s]:%s\n", entry.FuncName, entry.DiffType)
		}
		return
	}
	fmt.Printf("func diff between [%s] and [%s]:\n", srcPath, targetPath)
	diffEntries, err := funcdiff.GoFileDiffByFunc(srcPath, targetPath)
	if err != nil {
		panic(err)
	}
	for _, entry := range diffEntries {
		msg := "same"
		if entry.IsDiff {
			msg = "diff"
		}
		fmt.Printf("[%s]:%s\n", entry.FuncName, msg)
	}
}

/*
func diff, output (TO FIX):
[helloWorld4]:diff
[helloWorld2]:diff
[helloWorld1]:same

func diff extend, output:
[helloWorld4]:del
[helloWorld3]:add
[helloWorld1]:same
[helloWorld2]:change
*/
