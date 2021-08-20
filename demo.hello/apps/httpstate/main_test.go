package main

import (
	"fmt"
	"testing"
)

func TestHeaderKeyValue(t *testing.T) {
	headers := []string{
		"Content-Type: text/html",
		"Date: Thu, 19 Aug 2021 12:44:19 GMT",
	}

	for _, header := range headers {
		k, v := headerKeyValue(header)
		fmt.Printf("key=%s, value=%s\n", k, v)
	}
}
