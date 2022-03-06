package utils

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestGetRandNextInt(t *testing.T) {
	for _, i := range [5]int{10, 30, 50, 80, 100} {
		fmt.Printf("random int in [0-%d): %d\n", i, GetRandNextInt(i))
	}
}

func TestGetRandString(t *testing.T) {
	for _, i := range [2]uint{8, 16} {
		res, err := GetRandString(i)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("random string (%d): %s\n", i, res)
	}
}

func TestGobDeepCopy(t *testing.T) {
	f := fruit{
		ID:    1,
		Name:  "apple",
		Price: 32,
	}

	var f2 fruit
	GobDeepCopy(&f2, &f)

	f.Price = 40
	fmt.Printf("src fruit: %+v\n", f)
	fmt.Printf("dst fruit: %+v\n", f2)
}

func TestRunFuncWithTimeout(t *testing.T) {
	timeout := 2
	addFunc := func(a int, b int) int {
		time.Sleep(time.Duration(timeout) * time.Second)
		return a + b
	}

	a := 1
	b := 2
	runFunc := func() interface{} {
		return addFunc(a, b)
	}

	res, err := RunFuncWithTimeout(runFunc, time.Duration(timeout+1))
	if err != nil {
		t.Fatal(err)
	}
	if val, ok := res.(int); !ok {
		t.Fatal("results type error.")
	} else {
		fmt.Println("results:", val)
	}
}

func TestGetSimpleNowDatetime(t *testing.T) {
	fmt.Println("current date:", GetSimpleNowDate())
	fmt.Println("current datetime:", GetSimpleNowDatetime())
}

func TestFormatDateTimeAsDate(t *testing.T) {
	fmt.Println("current date:", FormatDateTimeAsDate(time.Now()))
}

func TestIsWeekDay(t *testing.T) {
	now := time.Now()
	fmt.Println("now weekday:", now.Weekday().String())
	fmt.Println("isweekday:", IsWeekDay(now))
}

func TestFprintJSONPrettyText(t *testing.T) {
	p := struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{
		Name: "foo",
		Age:  30,
	}

	path := "/tmp/test/output.json"
	outFile, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer outFile.Close()

	multiWriter := io.MultiWriter(os.Stdout, outFile)
	if err = FprintJSONPrettyText(multiWriter, p); err != nil {
		t.Fatal(err)
	}
}

func TestFprintBase64Text(t *testing.T) {
	if _, err := FprintBase64Text(os.Stdout, "foo and bar"); err != nil {
		t.Fatal(err)
	}
}

func TestGetHashFnv32(t *testing.T) {
	res, err := GetHashFnv32("hello world")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("hash fnv result: %d\n", res)
}

func TestGetShellPath(t *testing.T) {
	fmt.Println("sh path:", GetShellPath())
}

func TestRunCmd(t *testing.T) {
	cmd := exec.Command("ls", "-l", "/tmp/test")
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("cmd output: %s\n", b)
}

/*
# loop.sh
for i in {1..10}; do
	echo "this is shell loop test ${i}."
	sleep 1
done
*/

func TestRunShellCmd(t *testing.T) {
	output, err := RunShellCmd("sh", "/tmp/test/loop.sh")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("output:\n" + output)
}

func TestRunShellCmdInBg(t *testing.T) {
	if err := RunShellCmdInBg("sh", "/tmp/test/loop.sh"); err != nil {
		t.Fatal(err)
	}
}
