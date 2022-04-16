package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/creack/pty"
)

func main() {
	c := exec.Command("grep", "--color=auto", "bar")
	f, err := pty.Start(c)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	go func() {
		f.Write([]byte("foo\n"))
		f.Write([]byte("bar\n"))
		f.Write([]byte("baz\n"))
		f.Write([]byte{4}) // EOT
	}()

	if _, err := io.Copy(os.Stdout, f); err != nil {
		panic(err)
	}
	fmt.Println("pty command demo done.")
}
