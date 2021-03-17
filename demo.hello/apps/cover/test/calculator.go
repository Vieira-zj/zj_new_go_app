package test

func add(a, b int) int {
	return a + b
}

func multiple(a, b int) int {
	return a * b
}

func divide(a, b int) int {
	if b == 0 {
		b = 1
	}
	return a / b
}
