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
