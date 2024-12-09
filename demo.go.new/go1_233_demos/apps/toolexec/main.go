package main

import (
	"log"
	"os"

	"zjin.goapp.demo/utils"
)

func main() {
	// remove the tool itself (include args) from the command line
	args := os.Args[1:]
	log.Println("run cmd:", args)
	if err := utils.RunCmd(args...); err != nil {
		log.Printf("failed to run cmd [%s]: %v", args, err)
		os.Exit(1)
	}
}
