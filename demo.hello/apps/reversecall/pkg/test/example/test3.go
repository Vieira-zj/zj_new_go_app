package example

import (
	"fmt"

	"demo.hello/apps/reversecall/pkg/test/example/inner"
)

// Test3 test func.
func Test3() {
	fmt.Println("test3")
	fmt.Println(test3b())
}

// XYZ test struct.
type XYZ struct {
	Name string
}

func (xyz XYZ) print() {
	fmt.Println(xyz.Name)
}

// Test3a test func.
func Test3a() {
	fmt.Println("Test3a")
	xyz := XYZ{"hello"}
	xyz.print()
}

func test3b() string {
	fmt.Println("test3b")
	inner.Itest1()
	return "test3b callee"
}

// Test3c test func.
func Test3c() {
	fmt.Println("Test3c")
	go func() {
		fmt.Println("go")
	}()
	Test4a("world")
}
