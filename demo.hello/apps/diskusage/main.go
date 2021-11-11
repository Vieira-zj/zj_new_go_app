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

var (
	cancelCh    chan struct{}
	semaphoreCh chan struct{}
	// flags
	help     bool
	verbose  bool
	parallel int
)

func cancelled() bool {
	// 一个已经被关闭的channel不会阻塞，会立即返回，可检查返回值 ok
	// case _, ok <-done:
	select {
	case <-cancelCh:
		return true
	default:
		return false
	}
}

// 读取目录下的文件信息
func dirents(dir string) []os.FileInfo {
	select {
	case <-cancelCh:
		return nil
	// acquire token
	case semaphoreCh <- struct{}{}:
	}

	// release token
	defer func() {
		<-semaphoreCh
	}()

	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Read dir failed: %v\n", err)
		return nil
	}
	return entries
}

func walkDir(wg *sync.WaitGroup, dir string, fileSizes chan<- int64) {
	defer wg.Done()
	if cancelled() {
		return
	}

	for _, entry := range dirents(dir) {
		if entry.IsDir() {
			wg.Add(1)
			subDir := filepath.Join(dir, entry.Name())
			go walkDir(wg, subDir, fileSizes)
		} else {
			fileSizes <- entry.Size()
		}
	}
}

func printDiskUsage(nfiles, nbytes int64) {
	fmt.Printf("scan: %d files %.3f GB\n", nfiles, float64(nbytes)/1e9)
}

func main() {
	flag.BoolVar(&help, "h", false, "help.")
	flag.BoolVar(&verbose, "v", false, "show verbose scan progress messages.")
	flag.IntVar(&parallel, "p", 3, "number of parallel to scan dir.")
	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	cancelCh = make(chan struct{})
	semaphoreCh = make(chan struct{}, parallel)
	fileSizesCh := make(chan int64, parallel)

	roots := flag.Args()
	if len(roots) == 0 {
		roots = []string{"./"}
	}

	// scan target dir
	var wg sync.WaitGroup
	for _, dir := range roots {
		wg.Add(1)
		go walkDir(&wg, dir, fileSizesCh)
	}

	// read a byte and exit
	go func() {
		os.Stdin.Read(make([]byte, 1))
		close(cancelCh)
	}()

	go func() {
		wg.Wait()
		close(fileSizesCh)
	}()

	var (
		nfiles int64
		nbytes int64
		tick   <-chan time.Time
	)
	if verbose {
		tick = time.Tick(100 * time.Millisecond)
	}

loop:
	for {
		select {
		case <-cancelCh:
			for range fileSizesCh {
			}
			fmt.Println("cancelled")
			return
		case size, ok := <-fileSizesCh:
			if !ok { // scan done
				break loop
			}
			nfiles++
			nbytes += size
		case <-tick:
			printDiskUsage(nfiles, nbytes)
		}
	}
	printDiskUsage(nfiles, nbytes)
	fmt.Println("finished")
}
