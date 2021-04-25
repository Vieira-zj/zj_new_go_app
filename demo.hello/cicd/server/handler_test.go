package server

import (
	"fmt"
	"net/url"
	"testing"
)

func TestURLEncode(t *testing.T) {
	values := []string{
		"2021.04.v3 - AirPay",
		`key in (AIRPAY-46283,SPPAY-196)`,
	}
	for _, value := range values {
		fmt.Println(url.QueryEscape(value))
	}
}

func TestLoop(t *testing.T) {
outer:
	for i := 0; i < 10; i++ {
		for _, j := range []string{"a", "b", "c"} {
			if i%2 == 0 && j == "b" {
				fmt.Println()
				continue outer
			}
			fmt.Printf("%d:%s,", i, j)
		}
		fmt.Println()
	}
}
