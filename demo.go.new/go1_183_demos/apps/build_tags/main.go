package main

import "fmt"

/*
空格分隔的元素 => // +build pro enterprise => pro OR enterprise
逗号分隔的元素 => // +build pro,enterprise => pro AND enterprise
感叹号元素     => // +build !pro           => NOT pro
*/

var features = []string{
	"Free Feature #1",
	"Free Feature #2",
}

func main() {
	for _, f := range features {
		fmt.Println(">", f)
	}
}
