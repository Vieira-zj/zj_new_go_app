package utils

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

/*
Common
*/

// GetRandNextInt .
func GetRandNextInt(number int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(number)
}

// GetRandString returns rand string by specified length.
func GetRandString(length uint) (string, error) {
	// one byte converts to 2 hex char
	length = length / 2
	rand.Seed(time.Now().UnixNano())
	randBytes := make([]byte, length)
	if _, err := rand.Read(randBytes); err != nil {
		return "", err
	}
	// TODO: use base64 instead of hex
	return hex.EncodeToString(randBytes), nil
}

// GetBase62Text converts int number to base62 string.
func GetBase62Text(number int) string {
	getChars := func(start, count int) string {
		ret := ""
		for i := 0; i < count; i++ {
			ret += fmt.Sprintf("%c", start+i)
		}
		return ret
	}

	chars := getChars(int('0'), 10) // 0-9
	chars += getChars(int('a'), 26) // a-z
	chars += getChars(int('A'), 26) // A-Z

	b := make([]byte, 0, 4)
	for {
		remained := number % 62
		b = append(b, chars[remained])
		number /= 62
		if number == 0 {
			break
		}
	}
	return string(b)
}

// GobDeepCopy .
func GobDeepCopy(dst, src interface{}) error {
	// src and dst are pointer
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

// RunFunc .
type RunFunc func() interface{}

// RunFuncWithTimeout runs function, if timeout, return nil.
func RunFuncWithTimeout(fn RunFunc, timeout time.Duration) (interface{}, error) {
	ch := make(chan interface{}, 1)
	defer close(ch)

	go func() {
		res := fn()
		ch <- res
	}()

	select {
	case res := <-ch:
		return res, nil
	case <-time.After(timeout * time.Second):
		return nil, fmt.Errorf("timeout, exceed %d seconds", timeout)
	}
}

/*
Datetime
*/

// GetSimpleNowDate .
func GetSimpleNowDate() string {
	return time.Now().Format("2006-01-02")
}

// GetSimpleNowDatetime .
func GetSimpleNowDatetime() string {
	return time.Now().Format("2006/01/02 15:04:05")
}

// FormatDateTimeAsDate .
func FormatDateTimeAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

// IsWeekDay .
func IsWeekDay(t time.Time) bool {
	switch t.Weekday() {
	case time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday:
		return true
	case time.Saturday, time.Sunday:
		return false
	}
	fmt.Println("Unrecognized day of the week:", t.Weekday().String())
	panic("Explicit Panic to avoid compiler error: missing return at end of function")
}

func nextWeekDay(loc *time.Location) time.Time {
	now := time.Now().In(loc)
	for !IsWeekDay(now) {
		now = now.AddDate(0, 0, 1)
	}
	return now
}

/*
Encoder
*/

// FprintJSONPrettyText .
func FprintJSONPrettyText(w io.Writer, value interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

// FprintBase64Text .
func FprintBase64Text(w io.Writer, value string) (int64, error) {
	encoder := base64.NewEncoder(base64.StdEncoding, w)
	defer encoder.Close()
	r := strings.NewReader(value)
	return io.Copy(encoder, r)
}

// GetBase64Text .
func GetBase64Text(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

// GetURLBase64Text .
func GetURLBase64Text(bytes []byte) string {
	return base64.URLEncoding.EncodeToString(bytes)
}

// GetMd5HexText .
func GetMd5HexText(bytes []byte) string {
	return getMd5EncodedText(bytes, "hex")
}

// GetBase64MD5Text .
func GetBase64MD5Text(bytes []byte) string {
	return getMd5EncodedText(bytes, "std64")
}

// GetURLBase64MD5Text .
func GetURLBase64MD5Text(bytes []byte) string {
	return getMd5EncodedText(bytes, "url")
}

func getMd5EncodedText(bytes []byte, md5Type string) string {
	md5hash := md5.New()
	md5hash.Write(bytes)
	b := md5hash.Sum(nil)

	if md5Type == "hex" {
		return hex.EncodeToString(b)
	}
	if md5Type == "std64" {
		return base64.StdEncoding.EncodeToString(b)
	}
	return base64.URLEncoding.EncodeToString(b)
}

// GetHashFnv32 .
func GetHashFnv32(bytes []byte) (uint32, error) {
	f := fnv.New32()
	if _, err := f.Write(bytes); err != nil {
		return 0, err
	}
	return f.Sum32(), nil
}

/*
Command
*/

// GetShellPath returns sh abs path.
func GetShellPath() string {
	path := os.Getenv("SHELL")
	if len(path) == 0 {
		if path, err := exec.LookPath("sh"); err == nil {
			return path
		}
		path = "/bin/sh"
	}
	return path
}

// RunShellCmd runs a shell command and returns output.
func RunShellCmd(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	defer stdout.Close()

	if err = cmd.Start(); err != nil {
		return "", err
	}

	// blocked until eof
	output, err := ioutil.ReadAll(stdout)
	if err != nil && err != io.EOF {
		return "", err
	}

	err = cmd.Wait()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// RunShellCmdInBg runs a shell command in background and prints output.
func RunShellCmdInBg(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	defer stdout.Close()

	if err := cmd.Start(); err != nil {
		return err
	}
	fmt.Printf("cmd process started with pid: %d\n", cmd.Process.Pid)

	go func() {
		br := bufio.NewReader(stdout)
		for {
			b, _, err := br.ReadLine()
			if err != nil {
				if err == io.EOF {
					return
				}
				fmt.Println("buffer read failed:", err)
				return
			}
			fmt.Printf("%s\n", b)
		}
	}()

	if err := cmd.Wait(); err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			if status := err.Sys().(syscall.WaitStatus); status.Signaled() && status.Signal() == syscall.SIGTERM {
				return errors.New("process stopped with SIGTERM signal")
			}
		}
		return fmt.Errorf("process exited accidentally: %v", err)
	}
	fmt.Println("process stopped")
	return nil
}
