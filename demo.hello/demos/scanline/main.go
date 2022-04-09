package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	var input string
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

outer:
	for {
		fmt.Print("input: ")
		select {
		case <-ctx.Done():
			fmt.Println("exit:", ctx.Err())
			break outer
		default:
			if _, err := fmt.Scanln(&input); err != nil {
				if errors.Is(err, io.EOF) {
					fmt.Println("eof and exit")
					break outer
				}
				panic(fmt.Sprintln("scan error:", err))
			}
			fmt.Println("read:", input)

			input = strings.ToLower(input)
			if input == "bye" || input == "exit" {
				fmt.Println("scan exit")
				break outer
			}
		}
	}

	stop()
	os.Exit(0)
}
