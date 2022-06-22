package main

import "testing"

func TestFnHello(t *testing.T) {
	fnHello("foo", "it's cov test")
}

func TestFnChange(t *testing.T) {
	fnChange()
}

func TestFnDel(t *testing.T) {
	fnDel()
}

func TestFnConditional(t *testing.T) {
	fnConditional(true)
}
