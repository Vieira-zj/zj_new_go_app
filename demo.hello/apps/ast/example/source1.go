package example

import (
	"fmt"
	"strings"
)

func source1() {
	hello := "Hello"
	world := "World"
	words := []string{hello, world}
	SayHello(words)
}

// SayHello says Hello
func SayHello(words []string) bool {
	fmt.Println(joinStrings(words))
	return true
}

// joinStrings joins strings
func joinStrings(words []string) string {
	return strings.Join(words, ", ")
}

type person struct {
	name string
}

func (p *person) GetName() string {
	return p.name
}

func (p *person) Say(greet string) {
	fmt.Printf("%s, my name is %s\n", greet, p.name)
}
