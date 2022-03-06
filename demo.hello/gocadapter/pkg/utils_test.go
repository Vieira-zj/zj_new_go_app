package pkg

import (
	"fmt"
	"testing"
)

func TestGetSimpleNowDatetime(t *testing.T) {
	fmt.Println("now:", getSimpleNowDatetime())
}
