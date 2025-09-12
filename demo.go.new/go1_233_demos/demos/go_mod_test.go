package demos

import (
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"golang.org/x/sync/singleflight"
)

func TestSingleFlightDemo(t *testing.T) {
	callCount := 0
	mockFetchData := func(key string) (string, error) {
		callCount++
		fmt.Printf("fetching data for key '%s' from origin (call #%d)...\n", key, callCount)
		time.Sleep(500 * time.Millisecond)
		return "mock_data_for_key|" + key, nil
	}

	const testKey = "singleflight_demo01"
	g := singleflight.Group{}
	wg := sync.WaitGroup{}

	for i := range 5 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			result, err, shared := g.Do(testKey, func() (any, error) {
				return mockFetchData(testKey)
			})
			if err != nil {
				fmt.Printf("goroutine %d: error fetching data: %v\n", id, err)
				return
			}
			fmt.Printf("goroutine %d: received result: '%v' (shared: %t)\n", id, result, shared)
		}(i)
	}
	wg.Wait()
	log.Printf("total calls to fetch data: %d", callCount)
}
