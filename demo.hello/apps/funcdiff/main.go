package main

import (
	"flag"
	"fmt"
	"log"

	funcdiff "demo.hello/apps/funcdiff/pkg"
)

var (
	help       bool
	srcPath    string
	targetPath string
	isDebug    bool
)

func main() {
	flag.BoolVar(&help, "h", false, "help.")
	flag.StringVar(&srcPath, "s", "", "source: go file path to diff.")
	flag.StringVar(&targetPath, "t", "", "target: go file path to diff.")
	flag.BoolVar(&isDebug, "d", false, "debug for func diff.")

	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	if isDebug {
		fmt.Printf("func diff between [%s] and [%s]:\n", srcPath, targetPath)
		diffEntries, err := funcdiff.GoFileDiffByFunc(srcPath, targetPath)
		if err != nil {
			log.Fatal(err)
		}
		for _, entry := range diffEntries {
			msg := "same"
			if entry.IsDiff {
				msg = "diff"
			}
			fmt.Printf("[%s]:%s\n", entry.FuncName, msg)
		}
		return
	}

	fmt.Printf("func diff extend between [%s] and [%s]:\n", srcPath, targetPath)
	diffResults, err := funcdiff.GoFileDiffByFuncExt(srcPath, targetPath)
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range diffResults {
		fmt.Printf("[%s]:%s\n", entry.FuncName, entry.DiffType)
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
