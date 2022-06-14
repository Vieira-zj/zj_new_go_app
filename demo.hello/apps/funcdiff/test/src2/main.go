package main

import (
	"fmt"
	"log"
)

func fnHello(name /* user name */, msg /* display message */ string) {
	// hello
	log.Println(fmt.Sprintf("hello %s: %s", name, msg))
}

func fnAdd() {
	log.Println("func to add1")
	log.Println("func to add2")
	log.Println("func to add3")
}

func fnChange() {
	log.Println("func to changed")
}

func fnConditional(cond bool /* test condition */) {
	// test for cond
	if cond {
		log.Println("cond: true")
	} else {
		log.Println("cond: false")
	}
}

type person struct {
	name string
	age int
}

func (p person) fnHello() string {
	return fmt.Sprintf("name=%s,age=%d\n", p.name, p.age)
}

func main() {
	fnHello("foo", "test")
}
