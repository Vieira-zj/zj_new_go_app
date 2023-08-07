package utils

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"
)

func RunShellCmd(command string) (string, error) {
	tag := fmt.Sprintf("RunShellCmd [%s]:", command)
	log.Println("exec sh command:", command)

	sh := getShellPath()
	cmd := exec.Command(sh, "-c", command)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("%s get stdout error: %v", tag, err)
	}
	defer stdout.Close()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("%s get stderr error: %v", tag, err)
	}
	defer stderr.Close()

	if err = cmd.Start(); err != nil {
		return "", fmt.Errorf("%s start cmd error: %v", tag, err)
	}

	// here, blocked until cmd exec finish
	output, err := ioutil.ReadAll(stdout)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("%s read stdout error: %v", tag, err)
	}

	errOutput, err := ioutil.ReadAll(stderr)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("%s read stderr error: %v", tag, err)
	}

	timer := time.AfterFunc(time.Second, func() {
		if err := cmd.Process.Kill(); err != nil {
			log.Printf("%s process kill error: %v", tag, err)
			return
		}
	})
	defer timer.Stop()

	if err := cmd.Wait(); err != nil {
		errMsg := fmt.Sprintf("%s wait cmd error: %v", tag, err)
		if len(errOutput) > 0 {
			errMsg += "\nerror output: " + string(errOutput)
		}
		return "", fmt.Errorf(errMsg)
	}

	return string(output), nil
}

func RunShellCmdInDir(command, dir string) (string, error) {
	log.Println("exec sh command:", command)
	sh := getShellPath()
	cmd := exec.Command(sh, "-c", command)
	cmd.Dir = dir
	b, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("run cmd [%s] error: exit_code: %d, stderr: %s", command, ee.ExitCode(), ee.Stderr)
		}
		return "", err
	}

	return string(b), nil
}

func getShellPath() string {
	path, ok := os.LookupEnv("SHELL")
	if !ok {
		if path, err := exec.LookPath("sh"); err == nil {
			return path
		}
		path = "/bin/sh"
	}
	return path
}
