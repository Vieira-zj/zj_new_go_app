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

func (p person) fnToString() string {
	return fmt.Sprintf("name=%s,age=%d\n", p.name, p.age)
}

var aFunc = func() {
	log.Println("anonymous func")
}

func fnHello(name /* user name */, msg /* display message */ string) {
	if len(msg) == 0 {
		msg = "ast test"
	}
	// nest anonymous func
	hello := func(name, msg string) {
		log.Println(fmt.Sprintf("hello %s: %s", name, msg))
	}
	hello(name, msg)

	aFunc()
}

func fnChange() {
	log.Println("func to change")
}

func fnDel() {
	log.Println("func to del") // case: func delete
}

func fnConditional(cond bool /* test bool condition */) {
	// test for cond
	if cond {
		log.Println("cond: true")
	} else {
		func() { // anonymous func
			if err := recover(); err != nil {
				panic(err)
			}
			log.Println("cond: false")
		}()
	}
}

func main() {
	fnHello("foo", "test")
}
