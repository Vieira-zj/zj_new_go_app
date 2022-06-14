package main

import (
	"fmt"
	"log"
)

// test struct
type person struct {
	name string
	age  int
}

func (p person) fnHello() string {
	return fmt.Sprintf("name=%s,age=%d\n", p.name, p.age)
}

func fnHello(name /* user name */, msg /* display message */ string) {
	// hello
	log.Println(fmt.Sprintf("hello %s: %s", name, msg))
}

func fnChange() {
	log.Println("func to change")
}

func fnDel() {
	log.Println("func to del") // test func delete
}

func fnConditional(cond bool /* test condition */) {
	// test for cond
	if cond { log.Println("cond: true")
	} else { log.Println("cond: false")
	}
}

func main() {
	fnHello("foo", "test")
}
