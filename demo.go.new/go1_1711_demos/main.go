package main

import (
	"fmt"
	"runtime"
)

func main() {
	ver := runtime.Version()
	fmt.Printf("%s demo\n", ver)
}
