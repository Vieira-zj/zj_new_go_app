package pkg

import (
	"fmt"
	"testing"
)

func TestIsAttachServerOK(t *testing.T) {
	host := "http://127.0.0.1:51025"
	ok := isAttachServerOK(host)
	fmt.Println("service ok:", ok)
}

func TestRemoveUnhealthSrvInGocTask(t *testing.T) {
	if err := removeUnhealthSrvInGocTask(localHost); err != nil {
		t.Fatal(err)
	}
	fmt.Println("done")
}
