package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

/*
test: go run main.go -i /tmp/test/input.txt
*/

var inFilePath, outFilePath string

func main() {
	flag.StringVar(&inFilePath, "i", "", "input file path")
	flag.StringVar(&outFilePath, "o", "", "output file path")
	flag.Parse()

	if len(inFilePath) == 0 {
		panic("input file path is empty")
	}
	if len(outFilePath) == 0 {
		fields := strings.Split(inFilePath, ".")
		base := fields[0]
		ext := fields[1]
		outFilePath = fmt.Sprintf("%s_expand.%s", base, ext)
	}

	bytes, err := ioutil.ReadFile(inFilePath)
	if err != nil {
		panic(err)
	}
	res := os.ExpandEnv(string(bytes))
	if err := ioutil.WriteFile(outFilePath, []byte(res), 0666); err != nil {
		panic(err)
	}
	fmt.Println("env expand to:", outFilePath)
}
