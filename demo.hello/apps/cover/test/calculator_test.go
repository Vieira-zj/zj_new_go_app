package test

import "testing"

func TestAdd(t *testing.T) {
	inputs := []struct {
		num1   int
		num2   int
		expect int
	}{
		{
			num1:   1,
			num2:   2,
			expect: 3,
		},
		{
			num1:   -1,
			num2:   3,
			expect: 2,
		},
	}

	for _, item := range inputs {
		result := add(item.num1, item.num2)
		if result != item.expect {
			t.Fatalf("want %d, and get %d", item.expect, result)
		}
	}
}

func TestDivide(t *testing.T) {
	items := []struct {
		num1   int
		num2   int
		expect int
	}{
		{
			num1:   1,
			num2:   -1,
			expect: -1,
		},
		{
			num1:   10,
			num2:   5,
			expect: 2,
		},
	}

	for _, item := range items {
		result := divide(item.num1, item.num2)
		if result != item.expect {
			t.Fatalf("want %d, and get %d", item.expect, result)
		}
	}
}
