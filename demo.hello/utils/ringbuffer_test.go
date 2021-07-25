package utils

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestBytesCopy(t *testing.T) {
	dst := make([]byte, 3)
	copy(dst, []byte("abcde"))
	fmt.Printf("dst: %s\n", dst)
}

func TestRingBuffer01(t *testing.T) {
	size := 9
	ringBuf, err := NewRingBuffer(uint32(size))
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < size; i++ {
		_, err := ringBuf.Write([]byte(strconv.Itoa(i)))
		if err != nil {
			t.Fatal(err)
		}
	}
	fmt.Printf("ring buffer: %s\n", ringBuf.Bytes())

	ringBuf.Write([]byte("9"))
	fmt.Printf("ring buffer: %s\n", ringBuf.Bytes())
	ringBuf.Write([]byte("a"))
	fmt.Printf("ring buffer: %s\n", ringBuf.Bytes())
}

func TestRingBuffer02(t *testing.T) {
	ringBuf, err := NewRingBuffer(40)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	// write
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			ringBuf.Write([]byte(fmt.Sprintf("%d,", i)))
			time.Sleep(time.Duration(200) * time.Millisecond)
		}
	}()

	wg.Add(1)
	// read
	go func() {
		defer wg.Done()
		readSeek := 0
		length := 3
		fmt.Println("start read")
		for true {
			time.Sleep(time.Duration(500) * time.Millisecond)
			ret, err := ringBuf.Read(uint32(readSeek), uint32(length))
			if err != nil {
				if err.Error() == "eof" {
					fmt.Print(string(ret))
					fmt.Println("\nend read")
					return
				}
				panic(err)
			}
			readSeek += length
			fmt.Print(string(ret))
		}
	}()

	wg.Wait()
	fmt.Printf("ring buffer (size=%d): %s\n", ringBuf.Size(), ringBuf.Bytes())
}
