package pkg

import (
	"fmt"
	"testing"
)

func TestNewSHCmdOnce(t *testing.T) {
	for i := 0; i < 3; i++ {
		NewShCmd("")
	}
}

func TestCheckoutToCommit(t *testing.T) {
	root := "/tmp/test/gittest"
	cmd := NewShCmd(root)
	commitID := "75a3d6b"
	if err := cmd.CheckoutToCommit(commitID); err != nil {
		t.Fatal(err)
	}
	fmt.Println("checkout done")
}
