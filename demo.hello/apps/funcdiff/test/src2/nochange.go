package main

import "fmt"

// noChange .
func noChange() {
	fmt.Println("nochange func")
}

// noChangeCond .
func noChangeCond(flagA, flagB bool) {
	if flagA {
		fmt.Println("one")
	} else {
		fmt.Println("two")
	}

	if flagB {
		fmt.Println("three")
	} else {
		fmt.Println("four")
	}
}
