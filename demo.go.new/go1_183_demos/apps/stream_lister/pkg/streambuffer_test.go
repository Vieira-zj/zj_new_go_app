package streamlister

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"
)

func TestBinaryCodec01(t *testing.T) {
	buf := make([]byte, 10)
	t.Run("binary encode", func(t *testing.T) {
		ts := uint32(time.Now().Unix())
		binary.BigEndian.PutUint16(buf[0:], 0xa20c)
		binary.BigEndian.PutUint16(buf[2:], 0x04af)
		binary.BigEndian.PutUint32(buf[4:], ts)
		binary.BigEndian.PutUint16(buf[8:], 479)
		fmt.Printf("bytes: %x\n", buf)
	})

	t.Run("binary decode", func(t *testing.T) {
		sensorID := binary.BigEndian.Uint16(buf[0:])
		locID := binary.BigEndian.Uint16(buf[2:])
		tstamp := binary.BigEndian.Uint32(buf[4:])
		temp := binary.BigEndian.Uint16(buf[8:])
		fmt.Printf("sid: %0#x, locID %0#x ts: %0#x, temp:%d\n", sensorID, locID, tstamp, temp)
	})
}

func TestBinaryCodec02(t *testing.T) {
	tmp := make([]byte, 10)
	var buf []byte

	t.Run("binary encode int", func(t *testing.T) {
		n := binary.PutUvarint(tmp, 3)
		n += binary.PutUvarint(tmp[n:], 6)
		n += binary.PutUvarint(tmp[n:], 9)
		buf = tmp[:n]
		t.Log("buf len:", len(buf))
	})

	t.Run("binary decode int", func(t *testing.T) {
		offset := 0
		for {
			num, n := binary.Uvarint(buf[offset:])
			if n <= 0 {
				t.Log("decode failed")
				break
			}

			t.Log("number:", num)
			offset += n
		}
	})

	t.Run("binary decode in by reader", func(t *testing.T) {
		r := bytes.NewReader(buf)
		if _, err := r.Seek(0, io.SeekStart); err != nil {
			t.Fatal(err)
		}

		for {
			num, err := binary.ReadUvarint(r)
			if err != nil {
				if errors.Is(err, io.EOF) {
					t.Log("eof")
					break
				}
				t.Fatal(err)
			}
			t.Log("number:", num)
		}
	})
}

func TestBinaryCodec03(t *testing.T) {
	var tmp [10]byte // use array instead of slice
	buf := make([]byte, 0)

	t.Run("write to buffer", func(t *testing.T) {
		for k, v := range map[string]string{
			"one":   "bar",
			"two":   "hello",
			"three": "foo",
		} {
			// format: key-len | value-len | key | value
			n := binary.PutUvarint(tmp[0:], uint64(len(k)))
			n += binary.PutUvarint(tmp[n:], uint64(len(v)))
			buf = append(buf, tmp[:n]...)
			buf = append(buf, []byte(k)...)
			buf = append(buf, []byte(v)...)
		}

		t.Log("buf len:", len(buf))
	})

	t.Run("read kv from buffer", func(t *testing.T) {
		r := bytes.NewReader(buf)
		if _, err := r.Seek(0, io.SeekStart); err != nil {
			t.Fatal(err)
		}

		for {
			klen, err := binary.ReadUvarint(r)
			if err != nil {
				if errors.Is(err, io.EOF) {
					t.Log("eof")
					break
				}
				t.Fatal(err)
			}

			vlen, _ := binary.ReadUvarint(r)

			kbuf := make([]byte, klen)
			if _, err := io.ReadFull(r, kbuf); err != nil {
				t.Fatal(err)
			}
			vbuf := make([]byte, vlen)
			if _, err := io.ReadFull(r, vbuf); err != nil {
				t.Fatal(err)
			}

			t.Logf("read: key=%s, value=%s", kbuf, vbuf)
		}
	})
}

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
			b := make([]byte, 2) // 2 bytes for unit16
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
