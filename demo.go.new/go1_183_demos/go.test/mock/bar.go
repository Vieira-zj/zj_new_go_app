package mocktest

// mockgen by Source mode

//go:generate go run github.com/golang/mock/mockgen -source=./bar.go -destination=./bar_mock.go -package=mocktest
type Bar interface {
	Foo(x int) int
}
