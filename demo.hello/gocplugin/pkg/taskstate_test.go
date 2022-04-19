package pkg

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestMapDelete(t *testing.T) {
	m := map[int]string{
		1: "one",
		2: "two",
	}
	fmt.Printf("map: %+v\n", m)

	for i := 0; i < 4; i++ {
		delete(m, i)
	}
	fmt.Printf("map: %+v\n", m)
}

func TestTimeAfter(t *testing.T) {
	// 在 for...select 循环中，time.after 会泄漏。使用 time.ticker 代替 time.after
	closeCh := make(chan struct{})
	go func() {
		time.Sleep(4 * time.Second)
		close(closeCh)
	}()

outer:
	for i := 0; i < 3; i++ {
		select {
		case <-time.After(time.Second):
			fmt.Println("run")
			time.Sleep(time.Second)
		case <-closeCh:
			fmt.Println("close")
			break outer
		}
	}
	fmt.Println("done")
}

func TestTimeTick(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	tk := time.NewTicker(time.Second)
	defer tk.Stop()

outer:
	for i := 0; i < 3; i++ {
		select {
		case <-tk.C:
			fmt.Println("run")
			time.Sleep(time.Second)
			tk.Reset(time.Second)
		case <-ctx.Done():
			fmt.Println(ctx.Err())
			break outer
		}
	}
	fmt.Println("done")
}

func TestSrvCoverSyncTasksState(t *testing.T) {
	state := NewSrvCoverSyncTasksState()
	for i := 0; i < 10; i++ {
		srvState := i % 3
		state.Put(fmt.Sprintf("srv_%d", i), srvState)
	}
	state.Usage()

	limit := make(chan struct{}, 10)
	for i := 10; i < 50; i++ {
		local := i
		go func() {
			limit <- struct{}{}
			srvState := local % 3
			state.Put(fmt.Sprintf("srv_%d", local), srvState)
			<-limit
		}()
	}

	for i := 0; i < 5; i++ {
		if len(limit) == 0 {
			close(limit)
			break
		}
		time.Sleep(time.Second)
	}
	state.Usage()
}

func TestSrvCoverSyncTasksStateExpired(t *testing.T) {
	state := NewSrvCoverSyncTasksState()
	for i := 0; i < 6; i++ {
		srvState := i % 3
		state.PutByExpired(fmt.Sprintf("srv_%d", i), srvState, 2*time.Second)
	}
	state.Usage()

	time.Sleep(3 * time.Second)
	state.Usage()
}
