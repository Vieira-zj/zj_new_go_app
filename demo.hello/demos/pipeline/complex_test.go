package pipeline

import "testing"

func TestRunComplexPipeline(t *testing.T) {
	base := 10
	strings := []string{"5", "4", "3"}
	if err := RunComplexPipeline(base, strings); err != nil {
		t.Fatal(err)
	}
}
