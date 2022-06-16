package main

import "testing"

func TestFnHello(t *testing.T) {
	fnHello("foo", "it's cov test")
}

func TestFnConditional(t *testing.T) {
	fnConditional(true)
}
