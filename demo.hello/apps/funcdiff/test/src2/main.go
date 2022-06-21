package main

import (
	"fmt"
	"log"
)

var aFunc = func() {
	log.Println("anonymous func")
}

func fnHello(name /* user name */, msg string) {
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

func fnAdd() {
	log.Println("func to add1")
	log.Println("func to add2")
	log.Println("func to add3")
}

func fnChange() {
	log.Println("func to changed")
}

func fnConditional(cond bool) {
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

type person struct {
	name string
	age  int
}

func (p person) fnToString() string {
	return fmt.Sprintf("name=%s,age=%d\n", p.name, p.age)
}

func main() {

	fnHello("foo", "test")
}
