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

// StartLine,StartCol,EndLine,EndCol,NumStmt,Count
// block1: 8.28,13.2 3 2
// block2: 15.29,17.10 2 1
// block3: 22.2,22.15 1 0
// block4: 17.10,21.3 2 1
//
// 以 block 为单位来统计
// tab => 2 spaces
