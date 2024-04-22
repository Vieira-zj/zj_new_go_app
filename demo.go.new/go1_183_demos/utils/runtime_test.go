package utils_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"demo.apps/utils"
)

func TestGetProjectRootPath(t *testing.T) {
	path := utils.GetProjectRootPath()
	t.Log("project root path:", path)

	cmd := "git rev-parse --show-toplevel"
	output, err := utils.RunShellCmdInDir(cmd, "./")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("git root path:", output)
}

func TestGetFnFullName(t *testing.T) {
	anonymousFn := func() {
		fmt.Println("anonymous fn for test")
	}

	for _, fn := range []any{
		utils.GetLocalIPAddr,
		utils.GetCallerInfo,
		anonymousFn,
	} {
		t.Log("fn full name:", utils.GetFnFullName(fn))
	}
}

func TestGetCallerInfo(t *testing.T) {
	t.Run("get package path", func(t *testing.T) {
		type S struct{}
		typeOf := reflect.TypeOf(S{})
		t.Log("pkg path:", typeOf.PkgPath())
	})

	t.Run("get caller info", func(t *testing.T) {
		callerInfo := utils.GetCallerInfo(1)
		t.Log("caller info:\n", callerInfo)
	})
}

func TestGetGoroutineID(t *testing.T) {
	ch := make(chan int)
	for i := 0; i < 3; i++ {
		idx := i
		go func() {
			id, err := utils.GetGoroutineID()
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Printf("[%d] goroutine id: %d, start\n", idx, id)
			for val := range ch {
				fmt.Printf("[%d] goroutine id: %d, get value: %d\n", idx, id, val)
			}
			fmt.Printf("[%d] goroutine id: %d, exit\n", idx, id)
		}()
	}

	for i := 0; i < 10; i++ {
		ch <- i
		time.Sleep(10 * time.Millisecond)
	}
	close(ch)

	time.Sleep(100 * time.Millisecond)
	t.Log("test goroutine id done")
}
