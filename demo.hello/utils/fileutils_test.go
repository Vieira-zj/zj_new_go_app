package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

func TestFilePathUtils(t *testing.T) {
	path := "/tmp//test/"
	newPath := filepath.Clean(path)
	fmt.Println("clean path:", newPath)

	path = ".//fileutils.go"
	newPath, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("abs path:", newPath)
}

func TestGetFileExt(t *testing.T) {
	ext := filepath.Ext("fileutils.go")
	fmt.Println("file ext:", ext)

	// handle suffix string
	suffix := "py"
	if suffix[0] != '.' {
		suffix = "." + suffix
	}
	fmt.Println("suffix:", suffix)
}

func TestGetCurWorkPath(t *testing.T) {
	curPath, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("cur path:", curPath)
	fmt.Println("run path:", GetCurWorkPath())
}

func TestIsExist(t *testing.T) {
	for _, filePath := range [2]string{
		"/tmp/test/results.txt",
		"/tmp/test/data.txt",
	} {
		fmt.Println("file exist:", IsExist(filePath))
	}
}

func TestHasPermission(t *testing.T) {
	for _, filePath := range [2]string{
		"/tmp/test/results.txt",
		"/tmp/test/data.txt",
	} {
		if IsExist(filePath) {
			fmt.Println("has permission:", HasPermission(filePath))
		}
	}
}

func TestMakeDir(t *testing.T) {
	for _, dirPath := range [2]string{
		"/tmp/test",
		"/tmp/test/foo/bar",
	} {
		if err := MakeDir(dirPath); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("mkdir sucess:", dirPath)
		}
	}
}

func TestMoveFile(t *testing.T) {
	src := "/tmp/test/test.txt"
	dst := "/tmp/test/move/test.txt"
	if err := MoveFile(src, dst); err != nil {
		t.Fatal(err)
	}
	fmt.Println("file moved")
}

func TestListFilesInDir(t *testing.T) {
	dir := "/tmp/test/apa_goc_echoserver/cover_data_backup"
	files, err := ListFilesInDir(dir, ".cov")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("list files:\n%s", strings.Join(files, "\n"))
	fmt.Println()
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

func TestRemoveExpiredFile(t *testing.T) {
	dir := "/tmp/test"
	files, err := RemoveExpiredFiles(dir, 10.0, Second)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) > 0 {
		fmt.Println("removed files:", files)
	}
}

func TestGetLatestFileInDir(t *testing.T) {
	dir := "/tmp/test/apa_goc_echoserver/cover_data_backup"
	for _, suffix := range []string{".cov", ""} {
		name, err := GetLatestFileInDir(dir, suffix)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("latest file:", name)
	}
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
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		t.Fatal(err)
	}

	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".json" {
			continue
		}
		filePath := filepath.Join(dirPath, e.Name())
		fmt.Println("read file:", filePath)
		b, err := os.ReadFile(filePath)
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

func TestReadLinesFile(t *testing.T) {
	filePath := "/tmp/test/output.txt"
	lines, err := ReadLinesFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("read [%d] lines:\n", len(lines))
	for _, line := range lines {
		fmt.Println(line)
	}
}

func TestAppendToFile(t *testing.T) {
	path := "/tmp/test/merged_profile.cov"
	content := []byte("append line test\n")
	if err := AppendToFile(path, content); err != nil {
		t.Fatal(err)
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
	b, err := io.ReadAll(reader)
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

func TestIsFileContentEqual(t *testing.T) {
	root := "/tmp/test/apa_goc_echoserver/cover_data"
	src := filepath.Join(root, "staging_th_apa_goc_echoserver_master_845820727e_20220420_154143.cov")
	dst := filepath.Join(root, "staging_th_apa_goc_echoserver_master_845820727e_20220420_154307.cov")
	res, err := IsFileSizeEqual(src, dst)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("file size equal:", res)

	res, err = IsFilesEqual(src, dst)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("file content equal:", res)
}

func TestCopyFile(t *testing.T) {
	srcPath := "/tmp/test/echoserver/staging_th_apa_goc_echoserver_master_518e0a570c_127-0-0-1_20220325_181410.func"
	dstPath := "/tmp/test/echoserver/staging_th_apa_goc_echoserver_master_518e0a570c_127-0-0-1_20220325_181410_copied.func"
	if err := CopyFile(srcPath, dstPath); err != nil {
		t.Fatal(err)
	}
	fmt.Println("copy done")
}

func TestMergeFiles(t *testing.T) {
	outPath := "/tmp/test/out.txt"
	inPaths := []string{
		"/tmp/test/in_part1.txt",
		"/tmp/test/in_part2.txt",
		"/tmp/test/in_part3.txt",
	}
	if err := MergeFiles(inPaths, outPath); err != nil {
		t.Fatal(err)
	}
}

/*
io.Pipe / io.TeeReader

Copy 操作将持续地将数据复制到 Writer, 直到 Reader 读完数据。但这是一个无法控制的过程，如果你处理 writer 中数据的速度不能与复制操作一样快，那么它将很快耗尽你的缓冲区资源。
Pipe 提供一对 writer 和 reader, 并且读写操作都是同步的。利用内部缓冲机制，直到之前写入的数据被完全消耗掉才能写到一个新的 writer 数据块。
*/

func TestIoPipe(t *testing.T) {
	r, w := io.Pipe()

	go func() {
		defer w.Close()
		for i := 0; i < 10; i++ {
			fmt.Fprintf(w, "[%d]: some io.Reader stream to be read\n", i)
			time.Sleep(10 * time.Millisecond)
		}
	}()

	if _, err := io.Copy(os.Stdout, r); err != nil {
		t.Fatal(err)
	}
	fmt.Println("done")
}

func TestTeeReader(t *testing.T) {
	var r io.Reader = strings.NewReader("some io.Reader stream to be read\n")
	buf := bytes.NewBufferString("")
	r = io.TeeReader(r, buf)

	if _, err := io.Copy(io.Discard, r); err != nil {
		t.Fatal(err)
	}
	fmt.Println("buf value:", buf.String())
}

//
// File Sys: "io/fs"
//

func TestFileSysReadFile(t *testing.T) {
	// read file by fs
	dirPath := filepath.Join(os.Getenv("HOME"), "Downloads/tmps")
	fsys := os.DirFS(dirPath)
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		t.Fatal(err)
	}

	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".py" {
			fmt.Println("read file:", e.Name())
			// equal to: fs.ReadFile(fsys, e.Name())
			f, err := fsys.Open(e.Name())
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			b, err := io.ReadAll(f)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Println(string(b))
		}
	}
}

// mock fs for unit test
type Finder struct {
	fsys fs.ReadFileFS
}

func (f *Finder) containsWord(name, word string) (bool, error) {
	b, err := f.fsys.ReadFile(name)
	return bytes.Contains(b, []byte(word)), err
}

func TestFinder(t *testing.T) {
	testFs := fstest.MapFS{
		"pass.txt": &fstest.MapFile{
			Data:    []byte("hello, foo"),
			Mode:    0456,
			ModTime: time.Now(),
			Sys:     1,
		},
		"fail.txt": {
			Data:    []byte("hello, bar"),
			Mode:    0456,
			ModTime: time.Now(),
			Sys:     1,
		},
	}

	finder := &Finder{fsys: testFs}
	for _, name := range []string{"pass.txt", "fail.txt"} {
		got, err := finder.containsWord(name, "foo")
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("got foo in %s: %v\n", name, got)
	}
}

// test custom fs
type fsys struct{}

func (*fsys) Open(name string) (fs.File, error) {
	// NOTE: name 参数不能以 / 开头或者结尾
	if ok := fs.ValidPath(name); !ok {
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  os.ErrInvalid,
		}
	}
	return os.Open(name)
}

func (*fsys) ReadFile(name string) ([]byte, error) {
	if ok := fs.ValidPath(name); !ok {
		return nil, &fs.PathError{Op: "open", Path: name, Err: os.ErrInvalid}
	}
	return os.ReadFile(name)
}

func TestFs(t *testing.T) {
	existFile := "fileutils_test.go"
	if err := fstest.TestFS(new(fsys), existFile); err != nil {
		t.Fatal(err)
	}
}

//
// Zip
//

func TestZip(t *testing.T) {
	srcDir := "/tmp/test/data"
	dstFile := "/tmp/test/data.zip"
	if err := Zip(srcDir, dstFile); err != nil {
		t.Fatal(err)
	}
	fmt.Println("zip done.")
}

func TestUnzip(t *testing.T) {
	zipFile := "/tmp/test/data.zip"
	dstDir := "/tmp/test/unzip"
	if err := Unzip(zipFile, dstDir); err != nil {
		t.Fatal(err)
	}
	fmt.Println("unzip done.")
}
