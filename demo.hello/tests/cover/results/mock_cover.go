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

// lineCoverage returns "ax", "ay", "bx", "by"
func lineCoverage(flagA bool, flagX bool) string {GoCover.Count[4]++;
	ret := ""
	if flagA {GoCover.Count[7]++;
		ret += "a"
	} else{ GoCover.Count[8]++;{
		ret += "b"
	}}

	GoCover.Count[5]++;if flagX {GoCover.Count[9]++;
		ret += "x"
	} else{ GoCover.Count[10]++;{
		ret += "y"
	}}
	GoCover.Count[6]++;return ret
}

var GoCover = struct {
	Count     [11]uint32
	Pos       [3 * 11]uint32
	NumStmt   [11]uint16
} {
	Pos: [3 * 11]uint32{
		8, 13, 0x2001c, // [0]
		15, 17, 0xa001d, // [1]
		22, 22, 0xf0002, // [2]
		17, 21, 0x3000a, // [3]
		26, 28, 0xb0032, // [4]
		34, 34, 0xb0002, // [5]
		39, 39, 0xc0002, // [6]
		28, 30, 0x3000b, // [7]
		30, 32, 0x30008, // [8]
		34, 36, 0x3000b, // [9]
		36, 38, 0x30008, // [10]
	},
	NumStmt: [11]uint16{
		3, // 0
		2, // 1
		1, // 2
		2, // 3
		2, // 4
		1, // 5
		1, // 6
		1, // 7
		1, // 8
		1, // 9
		1, // 10
	},
}
