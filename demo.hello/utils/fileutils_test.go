package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsExist(t *testing.T) {
	for _, filePath := range [2]string{"/tmp/test/results.txt", "/tmp/test/data.txt"} {
		fmt.Println("file exist:", IsExist(filePath))
	}
}

func TestHasPermission(t *testing.T) {
	for _, filePath := range [2]string{"/tmp/test/results.txt", "/tmp/test/data.txt"} {
		if IsExist(filePath) {
			fmt.Println("has permission:", HasPermission(filePath))
		}
	}
}

func TestMakeDir(t *testing.T) {
	for _, dirPath := range [2]string{"/tmp/test", "/tmp/test/foo/bar"} {
		if err := MakeDir(dirPath); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("mkdir sucess:", dirPath)
		}
	}
}

func TestGetGoFileAbsPath(t *testing.T) {
	// Get project go file abs path
	path := "demo.hello/echoserver/handlers/ping.go"
	res, err := GetGoFileAbsPath(path)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res)
}

/*
Output file content with expand env.

file content:
env expand test for $USER:
home=$HOME
go_path=${GOPATH}
cur_dir=$PWD
*/

func TestReadFileWithExpandEnv(t *testing.T) {
	path := "/tmp/test/input.txt"
	res, err := ReadFileWithExpandEnv(path)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("expand string:\n", res)
}

/*
Dir utils
*/

func TestListDirFile(t *testing.T) {
	dirPath := "/tmp/test"
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		filePath := filepath.Join(dirPath, file.Name())
		fmt.Println("read file:", filePath)
		b, err := ioutil.ReadFile(filePath)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(string(b))
	}
}

func TestWalkDir(t *testing.T) {
	demoPath := "Workspaces/zj_repos/zj_go2_project/demo.hello/demos"
	dirPath := filepath.Join(os.Getenv("HOME"), demoPath)
	files, err := WalkDir(dirPath, ".go")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("go files in [%s]:\n", dirPath)
	for _, file := range files {
		fmt.Println("\t" + file)
	}
}

/*
Common IO
*/

func TestIOReader(t *testing.T) {
	reader := strings.NewReader("Clear is better than clever")
	res := make([]byte, 0, 20)
	p := make([]byte, 4)
	for {
		n, err := reader.Read(p)
		if err != nil {
			if err == io.EOF {
				// should handle any remainding bytes
				res = append(res, p[:n]...)
				break
			}
			t.Fatal(err)
		}
		res = append(res, p[:n]...)
	}
	fmt.Println("read string:", string(res))
}

func TestIOWriter(t *testing.T) {
	proverbs := []string{
		"Channels orchestrate mutexes serialize\n",
		"Cgo is not Go\n",
		"Errors are values\n",
		"Don't panic\n",
	}

	var writer bytes.Buffer
	for _, p := range proverbs {
		n, err := writer.Write([]byte(p))
		if err != nil {
			t.Fatal(err)
		}
		if n != len(p) {
			t.Fatal("failed to write data")
		}
	}
	fmt.Printf("write string:\n%s\n", writer.String())
}

func TestBufferWriter(t *testing.T) {
	filePath := "/tmp/test/buffer_out.txt"
	var out *bufio.Writer
	if !IsExist(filePath) {
		out = bufio.NewWriter(os.Stdout)
	} else {
		fd, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			t.Fatal(err)
		}
		defer fd.Close()
		out = bufio.NewWriter(fd)
	}
	defer out.Flush()

	for i := 0; i < 3; i++ {
		fmt.Fprintf(out, "this is a buffer write test: %d\n", i)
	}
}

func TestIOWriterReader(t *testing.T) {
	var buf bytes.Buffer
	// writer
	buf.Write([]byte("writer test\n"))
	buf.WriteTo(os.Stdout)

	// reader
	fmt.Fprint(&buf, "writer test, and add buffer text\n")
	p := make([]byte, 4)
	res := make([]byte, 0, 20)
	for {
		n, err := buf.Read(p)
		if err != nil {
			if err == io.EOF {
				res = append(res, p[:n]...)
				break
			}
			t.Fatal(err)
		}
		res = append(res, p[:n]...)
	}
	fmt.Printf("write string:\n%s", string(res))
}

func TestReadFileContent(t *testing.T) {
	filePath := "/tmp/test/output.txt"
	b, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("read content:\n%s\n", b)
}

func TestReadFileLines(t *testing.T) {
	filePath := "/tmp/test/output.txt"
	lines, err := ReadFileLines(filePath)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("read [%d] lines:\n", len(lines))
	for _, line := range lines {
		fmt.Println(line)
	}
}

func TestWriteContentToFile(t *testing.T) {
	filePath := "/tmp/test/output.txt"
	content := `one, this is a test.
two, this is a hello world.`

	t.Run("create and write to file", func(t *testing.T) {
		if IsExist(filePath) {
			if err := os.Remove(filePath); err != nil {
				t.Error(err)
			}
		}
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Error(err)
		}
	})

	t.Run("overwrite file", func(t *testing.T) {
		append := `overwrite content.
three, this is a write lines test.`
		if err := os.WriteFile(filePath, []byte(append), 0644); err != nil {
			t.Error(err)
		}
	})
}

func TestWriteLinesToFile(t *testing.T) {
	lines := []string{
		"one, this is a test.",
		"two, this is a hello world.",
		"three, this is a write lines test.",
	}
	if err := WriteLinesToFile("/tmp/test/test.txt", lines); err != nil {
		t.Fatal(err)
	}
}

func TestCreateFile(t *testing.T) {
	filePath := "/tmp/test/test.txt"
	b := []byte("Create file with content test.")
	if err := CreateFile(filePath, b); err != nil {
		t.Fatal(err)
	}
}

func TestFileWordsCount(t *testing.T) {
	filePath := "/tmp/test/test.txt"
	counts, err := FileWordsCount(filePath)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("words count: %+v\n", counts)
}

/*
Custom Reader
*/

type MyReader struct {
	text string
}

func NewMyReader(text string) *MyReader {
	return &MyReader{
		text: text,
	}
}

func (r *MyReader) Read(p []byte) (int, error) {
	if len(r.text) < len(p) {
		n := copy(p, r.text)
		r.text = r.text[n:]
		return n, io.EOF
	}

	n := copy(p, r.text)
	r.text = r.text[n:]
	return n, nil
}

func TestMyReader(t *testing.T) {
	reader := NewMyReader("this is a my reader read() test.")
	res := ""
	p := make([]byte, 4)
	for {
		n, err := reader.Read(p)
		if err != nil {
			if err == io.EOF {
				res += string(p[:n])
				break
			}
			t.Fatal(err)
		}
		res += string(p[:n])
	}

	if len(reader.text) != 0 {
		t.Fatal("not all chars read.")
	}
	fmt.Println("read string:", res)
}

func TestMyReaderCopy(t *testing.T) {
	reader := NewMyReader("this is a my reader copy test.")
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("read string:", string(b))
}

func TestStringSliceCopy(t *testing.T) {
	sl := "abcdef"
	b := make([]byte, 2)
	res := ""
	for i := 0; i < 10; i++ {
		if len(sl) < len(b) {
			// 测试 sl 长度为 0 的情况
			fmt.Println("slice size:", len(sl))
			n := copy(b, sl)
			res += string(b[:n])
			break
		}
		n := copy(b, sl)
		res += string(b[:n])
		sl = sl[n:]
	}
	fmt.Println("result:", res)
}
