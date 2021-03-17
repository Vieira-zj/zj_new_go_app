package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var done = make(chan struct{})

func cancelled() bool {
	select {
	// 一个已经被关闭的channel不会阻塞，会立即返回，可检查返回值 ok
	// case _, ok <-done:
	case <-done:
		return true
	default:
		return false
	}
}

// 获取目录dir下的文件大小
func walkDir(dir string, wg *sync.WaitGroup, fileSizes chan<- int64) {
	defer wg.Done()
	if cancelled() {
		return
	}

	for _, entry := range dirents(dir) {
		if entry.IsDir() {
			wg.Add(1)
			subDir := filepath.Join(dir, entry.Name())
			go walkDir(subDir, wg, fileSizes)
		} else {
			fileSizes <- entry.Size()
		}
	}
}

// 信号量
var sema = make(chan struct{}, 3)

// 读取目录dir下的文件信息
func dirents(dir string) []os.FileInfo {
	select {
	case <-done:
		return nil
	// acquire token
	case sema <- struct{}{}:
	}

	// release token
	defer func() {
		<-sema
	}()

	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Read dir failed: %v\n", err)
		return nil
	}
	return entries
}

// 提供 -v 参数会显示程序进度信息
var help = flag.Bool("h", false, "help.")
var verbose = flag.Bool("v", false, "show verbose progress messages.")

func startDiskUsage() {
	flag.Parse()
	if *help {
		flag.Usage()
		return
	}

	roots := flag.Args()
	if len(roots) == 0 {
		roots = []string{"."}
	}

	var tick <-chan time.Time
	if *verbose {
		tick = time.Tick(100 * time.Millisecond)
	}

	var wg sync.WaitGroup
	fileSizes := make(chan int64, 3)
	for _, dir := range roots {
		wg.Add(1)
		go walkDir(dir, &wg, fileSizes)
	}

	go func() {
		// 从标准输入读取一个字符，执行goroutine退出
		os.Stdin.Read(make([]byte, 1))
		close(done)
	}()

	go func() {
		wg.Wait()
		close(fileSizes)
	}()

	var nfiles, nbytes int64
loop:
	for {
		select {
		case <-done:
			for range fileSizes {
			}
			return
		case size, ok := <-fileSizes:
			if !ok {
				break loop
			}
			nfiles++
			nbytes += size
		case <-tick:
			printDiskUsage(nfiles, nbytes)
		}
	}
	printDiskUsage(nfiles, nbytes)
}

func testCloseCh() {
	go func() {
		for !cancelled() {
			fmt.Println("running...")
			time.Sleep(time.Second)
		}
		fmt.Println("exit")
	}()

	time.Sleep(time.Duration(3) * time.Second)
	close(done)
	time.Sleep(time.Duration(2) * time.Second)
}

func printDiskUsage(nfiles, nbytes int64) {
	fmt.Printf("%d files %.3f GB\n", nfiles, float64(nbytes)/1e9)
}

func main() {
	// testCloseCh()
	startDiskUsage()
	fmt.Println("disusage Done.")
}
