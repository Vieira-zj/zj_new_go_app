package main

import (
	"flag"
	"fmt"
)

var (
	help    bool
	srcPath string
	dstPath string
)

func main() {
	flag.BoolVar(&help, "h", false, "help.")
	flag.StringVar(&srcPath, "s", "", "source: go file path to diff.")
	flag.StringVar(&dstPath, "d", "", "target: go file path to diff.")

	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	fmt.Println("Func Diff")
}
