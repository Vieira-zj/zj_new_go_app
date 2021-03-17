package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os/exec"
)

var (
	help    = flag.Bool("h", false, "help")
	command = flag.String("c", "hostname", "command to run")
)

func errHandler(err error) {
	if err != nil {
		panic(err)
	}
}

func cmdTest01(command string, args []string) {
	cmd := exec.Command(command, args...)
	stdout, err := cmd.StdoutPipe()
	errHandler(err)
	defer stdout.Close()

	err = cmd.Start()
	errHandler(err)
	outBytes, err := ioutil.ReadAll(stdout)
	errHandler(err)
	fmt.Println("output:")
	fmt.Println(string(outBytes))
}

func cmdTest02(command string, args []string) {
	cmd := exec.Command(command, args...)
	buf, err := cmd.Output()
	errHandler(err)
	fmt.Println("output:")
	fmt.Println(string(buf))
}

func main() {
	// cmdTest02("go", []string{"version"})
	// run pipeline commands
	// cmdTest02("bash", []string{"-c", "ps -ef | grep 'docker.vmnetd'"})

	flag.Parse()
	if *help {
		flag.Usage()
		return
	}

	cmdTest02("sh", []string{"-c", *command})
	fmt.Println("Done")
}
