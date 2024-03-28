package streamlister

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
	"testing"
	"time"
)

var (
	beginB  byte = 0x12
	headerB byte = 0x6b
	eofB    byte = 0x9f
)

func TestStreamBufferRw(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)

	buf := bytes.NewBuffer([]byte{beginB})

	// write loop
	go func() {
		defer wg.Done()

		for i := 0; i < 8; i++ {
			b := make([]byte, 2)
			buf.WriteByte(headerB)
			binary.LittleEndian.PutUint16(b, uint16(i))
			buf.Write(b)
			time.Sleep(10 * time.Millisecond)
		}

		buf.WriteByte(eofB)
		fmt.Println("write eof")
	}()

	errCh := make(chan error, 1)

	// stream read
	go func() {
		defer wg.Done()

		bs := newStreamBuffer(buf, 10)

		start, err := bs.Get(0)
		if err != nil {
			errCh <- err
			return
		}
		if start != beginB {
			errCh <- fmt.Errorf("invalid begin flag")
		}

		var tmpBytes []byte
		for idx := 1; ; idx++ {
			time.Sleep(15 * time.Millisecond)
			b, err := bs.Get(idx)
			if err != nil {
				errCh <- err
				return
			}

			if b == headerB {
				if len(tmpBytes) > 0 {
					number := binary.LittleEndian.Uint16(tmpBytes)
					fmt.Println("read number:", number)
					tmpBytes = tmpBytes[:0]
				}
				continue
			}
			if b == eofB && len(tmpBytes) > 0 {
				number := binary.LittleEndian.Uint16(tmpBytes)
				fmt.Println("read number:", number)
				fmt.Println("eof")
				return
			}
			tmpBytes = append(tmpBytes, b)
		}
	}()

	go func() {
		if err, ok := <-errCh; ok {
			fmt.Println("read stream error:", err)
		}
	}()

	wg.Wait()
	close(errCh)

	t.Log("stream buffer finish")
}

func TestBinaryInt(t *testing.T) {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, uint16(12))
	t.Log("bytes:", b)

	number := binary.LittleEndian.Uint16(b)
	t.Log("number:", number)
}
