package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

/*
build:
go build -o /tmp/test/envexpand .

usage:
go run main.go -i /tmp/test/input.txt -p
go run main.go -i /tmp/test/input.txt -o /tmp/test/output.txt
*/

var (
	inFilePath  string
	outFilePath string
	isPrint     bool
	help        bool
)

func main() {
	flag.StringVar(&inFilePath, "i", "", "input file path.")
	flag.StringVar(&outFilePath, "o", "", "output file path.")
	flag.BoolVar(&isPrint, "p", false, "output to standard output.")
	flag.BoolVar(&help, "h", false, "help.")

	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	if len(inFilePath) == 0 {
		panic("input file path is empty")
	}
	bytes, err := ioutil.ReadFile(inFilePath)
	if err != nil {
		panic(err)
	}

	// output to stdout
	res := os.ExpandEnv(string(bytes))
	if isPrint {
		if _, err := fmt.Fprintln(os.Stdout, res); err != nil {
			panic(err)
		}
		return
	}

	// output to file
	if len(outFilePath) == 0 {
		fields := strings.Split(inFilePath, ".")
		base := fields[0]
		ext := fields[1]
		outFilePath = fmt.Sprintf("%s_expand.%s", base, ext)
	}

	if err := ioutil.WriteFile(outFilePath, []byte(res), 0666); err != nil {
		panic(err)
	}
	fmt.Println("env expand to:", outFilePath)
}
