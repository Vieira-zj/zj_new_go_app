package main

import "testing"

func TestNoChange(t *testing.T) {
	noChange()
}

func TestNoChangeCond01(t *testing.T) {
	noChange()
	noChangeCond(false, false)
}
