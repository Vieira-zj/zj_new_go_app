package pkg

import (
	"fmt"
	"testing"
)

func TestNewSHCmdOnce(t *testing.T) {
	for i := 0; i < 3; i++ {
		NewShCmd()
	}
}

func TestRunCmd(t *testing.T) {
	sh := NewShCmd()
	res, err := sh.Run("cd ${HOME}/donwload; ls -l")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res)
}
