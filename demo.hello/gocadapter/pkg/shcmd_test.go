package pkg

import (
	"testing"
)

func TestNewSHCmdOnce(t *testing.T) {
	for i := 0; i < 3; i++ {
		NewShCmd()
	}
}
