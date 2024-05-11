package mocktest

import "fmt"

// mockgen by Reflect mode

//go:generate go run github.com/golang/mock/mockgen -destination=./foo_mock.go -package=mocktest demo.apps/go.test/mock Foo
type Foo interface {
	Bar(x int) int
}

type FooImpl struct {
	Foo
}

func NewFooImpl() *FooImpl {
	return &FooImpl{}
}

func (*FooImpl) Bar(x int) int {
	return x + 2
}

func Sut(f Foo) {
	result := f.Bar(99)
	fmt.Println("sut result:", result)
}
