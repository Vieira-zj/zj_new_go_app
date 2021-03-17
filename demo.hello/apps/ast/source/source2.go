package source

import (
	"context"
	"fmt"
)

func source2(a string, b int) {
	context.WithCancel(nil) // 0

	if _, err := context.WithCancel(nil); err != nil { // 1
		context.WithCancel(nil) // 2
	} else {
		context.WithCancel(nil) // 3
	}

	_, _ = context.WithCancel(nil) // 4

	go context.WithCancel(nil) // 5

	go func() {
		context.WithCancel(nil) // 6
	}()

	defer context.WithCancel(nil) // 7

	defer func() {
		context.WithCancel(nil) // 8
	}()

	data := map[string]interface{}{
		"x2": context.WithValue(nil, "k", "v"), // 9
	}
	fmt.Println(data)

	var keys []string = []string{"c"}
	for _, k := range keys {
		fmt.Println(k)
		context.WithCancel(nil)
	}
}
