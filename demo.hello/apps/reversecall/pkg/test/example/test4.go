package example

import (
	"fmt"
)

// Test4a test func.
func Test4a(a string) {
	fmt.Println("Test4a")
	fmt.Println(a)
}

func test4b(a string) {
	fmt.Println("test4b")
	fmt.Println(a)
}

// ReceiveFromKafka test func.
func ReceiveFromKafka() {
	Test4a("kafka")
	test4b("kafka")
}
