package test

import "testing"

func TestGetConditions(t *testing.T) {
	items := []struct {
		cond1  bool
		cond2  bool
		expect string
	}{
		{
			cond1:  true,
			cond2:  true,
			expect: "ac",
		},
		{
			cond1:  true,
			cond2:  false,
			expect: "ad",
		},
		{
			cond1:  false,
			cond2:  true,
			expect: "bc",
		},
	}

	for _, item := range items {
		result := getConditions(item.cond1, item.cond2)
		if result != item.expect {
			t.Fatalf("want %s, and get %s", item.expect, result)
		}
	}
}
