//line mock.go:1
package cover

import (
	"fmt"
	"strings"
)

func sayHello(name string) {GoCover.Count[0]++;
	// mock comments1
	fmt.Print("hello ")
	fmt.Println("world")
	fmt.Printf("hello, %s!\n", strings.Title(name))
}

func isOk(flag bool) string {GoCover.Count[1]++;
	fmt.Println("isok test")
	if flag {GoCover.Count[3]++;
		// mock comments2
		fmt.Println("pass")
		return "pass"
	}
	GoCover.Count[2]++;return "fail"
}

// StartLine,StartCol,EndLine,EndCol,NumStmt,Count
// block1: 8.28,13.2 3 2
// block2: 15.29,17.10 2 1
// block3: 22.2,22.15 1 0
// block4: 17.10,21.3 2 1
//
// 以 block 为单位来统计
// tab => 2 spaces

var GoCover = struct {
	Count     [4]uint32
	Pos       [3 * 4]uint32
	NumStmt   [4]uint16
} {
	Pos: [3 * 4]uint32{
		8, 13, 0x2001c, // [0]
		15, 17, 0xa001d, // [1]
		22, 22, 0xf0002, // [2]
		17, 21, 0x3000a, // [3]
	},
	NumStmt: [4]uint16{
		3, // 0
		2, // 1
		1, // 2
		2, // 3
	},
}
