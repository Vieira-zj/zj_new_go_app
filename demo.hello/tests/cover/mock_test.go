package cover

import (
	"fmt"
	"testing"
)

func TestSayHello(t *testing.T) {
	for _, name := range [2]string{"foo", "bar"} {
		sayHello(name)
	}
}

func TestIsOk(t *testing.T) {
	fmt.Println("results:", isOk(true))
}

func TestLineCoverage(t *testing.T) {
	data := map[string][]interface{}{
		"case01": {true, true, "ax"},
		"case02": {false, false, "by"},
		"case03": {true, false, "ay"},
	}

	for name, item := range data {
		t.Run(name, func(t *testing.T) {
			res := lineCoverage(item[0].(bool), item[1].(bool))
			expect := item[2].(string)
			if res != expect {
				t.Fatalf("want %s, got %s", expect, res)
			}
		})
	}
}
