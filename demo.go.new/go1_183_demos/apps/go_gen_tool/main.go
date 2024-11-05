package main

import "fmt"

//go:generate go run golang.org/x/tools/cmd/stringer -type=Pill

type Pill int

const (
	Placebo Pill = iota
	Ibuprofen
	Paracetamol
)

func main() {
	fmt.Printf("For headaches, take %v\n", Ibuprofen)
	fmt.Printf("For a fever, take %s\n", Paracetamol.String())
}
