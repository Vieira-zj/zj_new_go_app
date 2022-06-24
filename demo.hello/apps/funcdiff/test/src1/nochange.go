package main

import "fmt"

func noChange() {
	fmt.Println("nochange func")
}

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
