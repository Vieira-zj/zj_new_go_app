package demos

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestPopItemFromSlice(t *testing.T) {
	// pop 1st element from slice
	s1 := []string{"a", "b", "c", "d"}
	for i := 0; i < 3; i++ {
		s1 = s1[1:]
	}
	fmt.Println("by ref:", strings.Join(s1, ","))

	s2 := []string{"a", "b", "c", "d"}
	for i := 0; i < 3; i++ {
		copy(s2, s2[1:])
		s2 = s2[:len(s2)-1]
	}
	fmt.Println("by copy:", strings.Join(s2, ","))
}

func TestDataBase(t *testing.T) {
	db := newDataBase()

	wg := sync.WaitGroup{}
	wg.Add(1)
	// put conn
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			conn := strconv.Itoa(i)
			fmt.Println("put conn:", conn)
			db.PutConn(conn)
			time.Sleep(time.Duration(200) * time.Millisecond)
		}
	}()

	// get conn
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if conn, err := db.Connect(context.TODO()); err != nil {
				fmt.Println("get db conn error:", err)
			} else {
				defer db.PutConn(string(conn))
				fmt.Println("get conn:", string(conn))
				time.Sleep(time.Second)
			}
		}()
		time.Sleep(time.Duration(500) * time.Millisecond)
	}

	wg.Wait()
	db.PrintInfo()
}
