package main

import (
	"cmp"
	"fmt"
	"os"
)

func main() {
	v := cmp.Or(os.Getenv("SOME_VAR"), "null")
	fmt.Printf("SOME_VAR=%s\n", v)
	fmt.Println("go app template.")
}
