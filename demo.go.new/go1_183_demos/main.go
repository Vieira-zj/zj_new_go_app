package main

import (
	"fmt"
	"runtime"
)

func main() {
	ver := runtime.Version()
	fmt.Println("go version:", ver)
}
