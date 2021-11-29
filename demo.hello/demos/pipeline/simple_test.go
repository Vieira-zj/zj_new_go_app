package pipeline

import (
	"fmt"
	"testing"
	"time"
)

func TestRunSimplePipeline01(t *testing.T) {
	base := 10
	strings := []string{"3", "2", "1"}

	if err := RunSimplePipeline(base, strings); err != nil {
		t.Fatal(err)
	}
}

func TestRunSimplePipeline02(t *testing.T) {
	base := 1
	strings := []string{"3", "2", "1"}

	if err := RunSimplePipeline(base, strings); err != nil {
		time.Sleep(time.Second)
		fmt.Println(err)
	} else {
		t.Fatal("want error, and got nil")
	}
}

func TestRunSimplePipeline03(t *testing.T) {
	base := 2
	strings := []string{"1010", "1100", "1000"}
	if err := RunSimplePipeline(base, strings); err != nil {
		t.Fatal(err)
	}
}

func TestRunSimplePipeline04(t *testing.T) {
	base := 10
	// strings := []string{"1", "10", "100", "1000"}
	strings := []string{"1", "10", "100", "1000", "10000"}

	if err := RunSimplePipeline(base, strings); err != nil {
		time.Sleep(time.Second)
		fmt.Println(err)
	} else {
		t.Fatal("want error, and got nil")
	}
}

func TestRunSimplePipeline05(t *testing.T) {
	base := 10
	strings := []string{"3", "2", "", "1"}

	if err := RunSimplePipeline(base, strings); err != nil {
		time.Sleep(time.Second)
		fmt.Println(err)
	} else {
		t.Fatal("want error, and got nil")
	}
}
