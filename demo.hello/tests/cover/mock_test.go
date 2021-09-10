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
