package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"runtime"
	"strconv"
	"strings"
)

// Refer: https://github.com/uber-go/goleak/blob/master/internal/stack/stacks.go

const _defaultBufferSize = 64 * 1024 // 64 KiB

var filterFuncs = map[string]struct{}{
	"utils.getStackBuffer": {},
	"utils.getStacks":      {},
	"utils.Current":        {},
	"utils.All":            {},
}

// Stack represents a single Goroutine's stack.
type Stack struct {
	id        int
	state     string
	firstFunc string
	fullStack *bytes.Buffer
}

// ID returns the goroutine ID.
func (s Stack) ID() int {
	return s.id
}

// State returns the Goroutine's state.
func (s Stack) State() string {
	return s.state
}

// Full returns the full stack trace for this goroutine.
func (s Stack) Full() string {
	return s.fullStack.String()
}

// FirstFunction returns the name of the first function on the stack.
func (s Stack) FirstFunction() string {
	return s.firstFunc
}

func (s Stack) String() string {
	return fmt.Sprintf(
		"Goroutine [%v] in state [%v], with [%v] on top of the stack:\n%s",
		s.id, s.state, s.firstFunc, s.Full())
}

// All returns the stacks for all running goroutines.
func All() []Stack {
	return getStacks(true)
}

// Current returns the stack for the current goroutine.
func Current() Stack {
	return getStacks(false)[0]
}

func getStacks(all bool) []Stack {
	var stacks []Stack
	var curStack *Stack

	stackReader := bufio.NewReader(bytes.NewReader(getStackBuffer(all)))
	for {
		line, err := stackReader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			// We're reading using bytes.NewReader which should never fail.
			panic("bufio.NewReader failed on a fixed string")
		}

		isFirstLine := false
		if strings.HasPrefix(line, "goroutine ") {
			if curStack != nil {
				stacks = append(stacks, *curStack)
			}
			id, goState := parseGoStackHeader(line)
			curStack = &Stack{
				id:        id,
				state:     goState,
				fullStack: &bytes.Buffer{},
			}
			isFirstLine = true
		}

		curStack.fullStack.WriteString(line)

		if !isFirstLine && curStack.firstFunc == "" {
			if fn := parseFirstFunc(line); len(fn) > 0 {
				curStack.firstFunc = fn
			}
		}
	}

	if curStack != nil {
		stacks = append(stacks, *curStack)
	}
	return stacks
}

func parseFirstFunc(line string) string {
	line = strings.TrimSpace(line)
	if len(line) == 0 || line[0] == '/' {
		return ""
	}

	if idx := strings.LastIndex(line, "("); idx > 0 {
		fn := line[:idx]
		part := fn[strings.LastIndex(fn, "/")+1:]
		if _, ok := filterFuncs[part]; ok {
			return ""
		}
		return fn
	}
	panic(fmt.Sprintf("function calls missing parents: %q", line))
}

// parseGoStackHeader parses a stack header that looks like:
// goroutine 643 [runnable]:\n
// And returns the goroutine ID, and the state.
func parseGoStackHeader(line string) (int, string) {
	line = strings.TrimSuffix(line, ":\n")
	parts := strings.SplitN(line, " ", 3)
	if len(parts) != 3 {
		panic(fmt.Sprintf("unexpected stack header format: %q", line))
	}

	id, err := strconv.Atoi(parts[1])
	if err != nil {
		panic(fmt.Sprintf("failed to parse goroutine ID: %v in line %q", parts[1], line))
	}

	tmp := parts[2]
	state := tmp[1 : len(tmp)-1]
	return id, state
}

func getStackBuffer(all bool) []byte {
	for i := _defaultBufferSize; ; i *= 2 {
		buf := make([]byte, i)
		if n := runtime.Stack(buf, all); n < i { // stw here
			return buf[:n]
		}
	}
}
