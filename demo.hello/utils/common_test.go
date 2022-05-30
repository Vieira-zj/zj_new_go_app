package utils

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
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

func TestShuffle(t *testing.T) {
	list := make([]int, 0, 10)
	for i := 0; i < 10; i++ {
		list = append(list, i)
	}
	fmt.Println("src values:", list)

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(list), func(i, j int) {
		list[i], list[j] = list[j], list[i]
	})
	fmt.Println("shuffle values:", list)
}

func TestBase62(t *testing.T) {
	fmt.Printf("char to int: %d, %d, %d\n", '0', 'a', 'A')
	fmt.Printf("int to char: %c\n", 97)

	for _, num := range []int{9, 100, 201314} {
		res := GetBase62Text(num)
		fmt.Printf("number %d base62 value: %s\n", num, res)
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

func TestRegexFindAllSubString(t *testing.T) {
	regex := "the"
	s := "the regexp, the demo."
	matches, err := RegexFindAllSubString(regex, s)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("results:", matches)

	regex = "test"
	matches, err = RegexFindAllSubString(regex, s)
	if err != nil {
		if errors.Is(err, ErrRegexNotFound) {
			fmt.Println("not found")
		} else {
			t.Fatal(err)
		}
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

type testDumpData struct {
	ID     int         `json:"id" yaml:"ident"`
	Config string      `json:"cfg" yaml:"config"`
	Data   interface{} `json:"data" yaml:"data"`
}

func TestDumpJSON(t *testing.T) {
	dumpData := testDumpData{
		ID:     1,
		Config: "env=test,type=json",
		Data:   "test dump json",
	}
	outPath := "/tmp/test/dump.json"
	if err := DumpJSON(dumpData, outPath); err != nil {
		t.Fatal(err)
	}
	fmt.Println("dump done")
}

func TestDumpYAML(t *testing.T) {
	dumpData := testDumpData{
		ID:     2,
		Config: "env=test,type=yaml",
		Data: []struct {
			Name string `yaml:"name"`
			Age  int    `yaml:"age"`
		}{
			{Name: "foo", Age: 30},
			{Name: "bar", Age: 36},
		},
	}
	outPath := "/tmp/test/dump.yaml"
	if err := DumpYAML(dumpData, outPath); err != nil {
		t.Fatal(err)
	}
	fmt.Println("dump done")
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
	res, err := GetHashFnv32([]byte("hello world"))
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
