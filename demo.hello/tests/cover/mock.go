package cover

import (
	"fmt"
	"strings"
)

func sayHello(name string) {
	// mock comments1
	fmt.Print("hello ")
	fmt.Println("world")
	fmt.Printf("hello, %s!\n", strings.Title(name))
}

func isOk(flag bool) string {
	fmt.Println("isok test")
	if flag {
		// mock comments2
		fmt.Println("pass")
		return "pass"
	}
	return "fail"
}

// lineCoverage returns "ax", "ay", "bx", "by"
func lineCoverage(flagA bool, flagX bool) string {
	ret := ""
	if flagA {
		ret += "a"
	} else {
		ret += "b"
	}

	if flagX {
		ret += "x"
	} else {
		ret += "y"
	}
	return ret
}

// StartLine,StartCol,EndLine,EndCol,NumStmt,Count
// block1: 8.28,13.2 3 2
// block2: 15.29,17.10 2 1
// block3: 22.2,22.15 1 0
// block4: 17.10,21.3 2 1
//
// block5: 26.50,28.11 2 2
// block6: 34.2,34.11 1 2
//
// 以 block 为单位来统计
// tab => 2 spaces
