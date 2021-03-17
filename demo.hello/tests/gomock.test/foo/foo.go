package foo

import "fmt"

// Ref: https://github.com/golang/mock
//
// mock gen cmd:
// mockgen -source=foo.go -destination=mock_foo/mock_foo.go

// Foo interface for mock impl.
type Foo interface {
	Bar(x int) int
}

// PrintFoo prints Foo results.
func PrintFoo(f Foo, x int) {
	fmt.Println(f.Bar(x))
}
