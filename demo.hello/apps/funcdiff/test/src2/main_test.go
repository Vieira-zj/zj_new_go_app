package main

import "testing"

func TestFnAdd(t *testing.T) {
	fnAdd()
}

func TestFnToString(t *testing.T) {
	p := &person{
		name: "foo",
		age:  30,
	}
	p.fnToString()
}
