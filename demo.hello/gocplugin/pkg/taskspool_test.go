package pkg

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"demo.hello/utils"
)

func TestSubmitTaskToGoPool(t *testing.T) {
	pool := utils.NewGoPool(3, 4, 3*time.Second)
	pool.Start()
	defer pool.Cancel()

	wg := &sync.WaitGroup{}
	for i := 1; i < 4; i++ {
		wg.Add(1)
		local := i
		pool.Submit(func() {
			defer wg.Done()
			fmt.Println("task start:", local)
			time.Sleep(time.Second)
			fmt.Println("task finish:", local)
		})
		time.Sleep(time.Second)
	}

	wg.Wait()
	fmt.Println(pool.Usage())
	fmt.Println("stop pool")
}

func TestSubmitSrvCoverSyncTask(t *testing.T) {
	// run: go test -timeout 300s -run ^TestSubmitSrvCoverSyncTask$ demo.hello/gocplugin/pkg -v -count=1
	if err := mockLoadConfig("/tmp/test"); err != nil {
		t.Fatal(err)
	}

	InitSrvCoverSyncTasksPool()
	defer CloseSrvCoverSyncTasksPool()

	param := SyncSrvCoverParam{
		SrvName:   "staging_th_apa_goc_echoserver_master_845820727e",
		Addresses: []string{"http://127.0.0.1:51007"},
	}
	retCh := SubmitSrvCoverSyncTask(param)
	select {
	case res := <-retCh:
		switch res.(type) {
		case error:
			t.Fatal(res)
		case string:
			fmt.Println("cover total:", res)
		}
	case <-time.After(time.Minute):
		t.Fatal("run test timeout")
	}
	fmt.Println("srv cover sync task done")
}
