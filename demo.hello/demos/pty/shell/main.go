package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)

func main() {
	c := exec.Command("sh")
	ptmx, err := pty.Start(c)
	if err != nil {
		log.Fatal(err)
	}
	defer ptmx.Close()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
				log.Printf("error resizing pty: %s", err)
			}
		}
	}()

	// Initial resize
	ch <- syscall.SIGWINCH
	defer func() {
		signal.Stop(ch)
		close(ch)
	}()

	// Set stdin in raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
			log.Fatal(err)
		}
	}()

	// Copy stdin to the pty and the pty to stdout
	go func() {
		if _, err = io.Copy(ptmx, os.Stdin); err != nil {
			log.Fatal(err)
		}
	}()
	if _, err = io.Copy(os.Stdout, ptmx); err != nil {
		log.Fatal(err)
	}
}
