package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/etcdserver/api/v3rpc/rpctypes"
)

const timeout = time.Duration(3) * time.Second

var (
	help          = flag.Bool("h", false, "help")
	etcdEndpoints = flag.String("e", "localhost:2379", "etcd server endpoints")
	key           = flag.String("key", "/message", "key for op")
	op            = flag.String("op", "app", "operation type. (get, put, watch)")
)

func etcdErrorHandler(err error) {
	switch err {
	case context.Canceled:
		log.Fatalf("ctx is canceled by another routine: %v", err)
	case context.DeadlineExceeded:
		log.Fatalf("ctx is attached with a deadline is exceeded: %v", err)
	case rpctypes.ErrEmptyKey:
		log.Fatalf("client-side error: %v", err)
	default:
		log.Fatalf("bad cluster endpoints, which are not etcd servers: %v", err)
	}
}

func putTest(cli *clientv3.Client) {
	kv := clientv3.NewKV(cli)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	putResp, err := kv.Put(ctx, *key, "Hello1", clientv3.WithPrevKV())
	if err != nil {
		etcdErrorHandler(err)
	}
	log.Println(putResp.Header.Revision)
	if putResp.PrevKv != nil {
		log.Printf("prev Value: %s \n CreateRevision : %d \n ModRevision: %d \n Version: %d \n",
			string(putResp.PrevKv.Value), putResp.PrevKv.CreateRevision, putResp.PrevKv.ModRevision, putResp.PrevKv.Version)
	}
}

func getTest(cli *clientv3.Client) {
	kv := clientv3.NewKV(cli)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	getResp, err := kv.Get(ctx, *key)
	if err != nil {
		etcdErrorHandler(err)
	}

	if len(getResp.Kvs) > 0 {
		log.Printf("key is [%s], value is [%s]\n", getResp.Kvs[0].Key, getResp.Kvs[0].Value)
	} else {
		log.Println(*key + " not found!")
	}
}

func getTestPrefix(cli *clientv3.Client) {
	kv := clientv3.NewKV(cli)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	getResp, err := kv.Get(ctx, *key, clientv3.WithPrefix())
	if err != nil {
		etcdErrorHandler(err)
	}
	if len(getResp.Kvs) == 0 {
		log.Fatalf("key prefix %s not found!\n", *key)
	}

	for _, kv := range getResp.Kvs {
		log.Printf("key is [%s], value is [%s]\n", kv.Key, kv.Value)
	}
}

func watchTest(cli *clientv3.Client) {
	kv := clientv3.NewKV(cli)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(8)*time.Second)
	defer cancel()
	// ctx, cancelFunc := context.WithCancel(context.TODO())
	// time.AfterFunc(time.Duration(6)*time.Second, func() {
	// 	cancelFunc()
	// })

	getResp, err := kv.Get(ctx, *key)
	if err != nil {
		etcdErrorHandler(err)
	}
	if len(getResp.Kvs) > 0 {
		log.Println("current value:", string(getResp.Kvs[0].Value))
	}

	watchStartRevision := getResp.Header.Revision + 1
	log.Println("watch from reversion:", watchStartRevision)
	watcher := clientv3.NewWatcher(cli)
	watchRespChan := watcher.Watch(ctx, *key, clientv3.WithRev(watchStartRevision))

	go func() {
		time.Sleep(time.Second)
		for i := 0; i < 10; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			if i%2 == 0 {
				kv.Put(ctx, *key, "Hello"+strconv.Itoa(i))
			} else {
				kv.Delete(ctx, *key)
			}
			time.Sleep(time.Second)
		}
	}()

	for watchResp := range watchRespChan {
		for _, event := range watchResp.Events {
			switch event.Type {
			case clientv3.EventTypePut:
				log.Println("value updated as:", string(event.Kv.Value), "Revision:", event.Kv.CreateRevision, event.Kv.ModRevision)
			case clientv3.EventTypeDelete:
				log.Println("value deleted", "Revision:", event.Kv.ModRevision)
			}
		}
	}
}

//
// App: watch for key remove, restart service, and add watcher for newly created key
//
func appWatchAndRestartService(cli *clientv3.Client) {
	watchKey := *key
	for {
		addWatcherForKey(cli, watchKey)
		watchKey = getNewKey(cli)
		fields := strings.Split(watchKey, ":")
		addr := fields[2] + ":" + fields[3]
		setRestartFlag(addr)
	}
}

func getNewKey(cli *clientv3.Client) string {
	retry := 180
	fields := strings.Split(*key, ":")
	prefixKey := fields[0] + ":" + fields[1]
	kv := clientv3.NewKV(cli)

	for i := 0; i < retry; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		getResp, err := kv.Get(ctx, prefixKey, clientv3.WithPrefix())
		if err != nil {
			etcdErrorHandler(err)
		}

		if len(getResp.Kvs) > 0 {
			return string(getResp.Kvs[0].Key)
		}
		log.Printf("key prefix [%s] not found!\n", *key)
		time.Sleep(time.Second)
	}
	log.Fatalf("key prefix [%s] not found for %d seconds, and watcher exit\n", *key, retry)
	return ""
}

func addWatcherForKey(cli *clientv3.Client, key string) {
	log.Println("watch for key:", key)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	kv := clientv3.NewKV(cli)
	getResp, err := kv.Get(ctx, key)
	if err != nil {
		etcdErrorHandler(err)
	}
	if len(getResp.Kvs) == 0 {
		log.Fatalf("key [%s] not found!", key)
	}
	log.Println("current value:", string(getResp.Kvs[0].Value))

	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	watchStartRevision := getResp.Header.Revision + 1
	log.Println("watch from reversion:", watchStartRevision)
	watcher := clientv3.NewWatcher(cli)
	watchRespChan := watcher.Watch(ctx2, key, clientv3.WithRev(watchStartRevision))

	for watchResp := range watchRespChan {
		for _, event := range watchResp.Events {
			switch event.Type {
			case clientv3.EventTypePut:
				retKey := string(event.Kv.Key)
				log.Println("value updated as:", retKey, "Revision:", event.Kv.CreateRevision, event.Kv.ModRevision)
			case clientv3.EventTypeDelete:
				log.Println("value deleted", "Revision:", event.Kv.ModRevision)
				return
			}
		}
	}
}

func setRestartFlag(addr string) {
	command := "sh"
	args := []string{"-c", fmt.Sprintf("echo -n '%s' > /tmp/restart.flag", addr)}
	runShellCommand(command, args)
}

func runShellCommand(command string, args []string) {
	log.Println("run command: " + strings.Join(args[1:], " "))
	cmd := exec.Command(command, args...)
	buf, err := cmd.Output()
	if err != nil {
		log.Fatalln("run command failed, and err:", err)
	}
	if buf != nil && len(buf) > 0 {
		log.Println("output:")
		log.Println(string(buf))
	}
}

func main() {
	// refer to: https://github.com/etcd-io/etcd/tree/master/clientv3

	// notice for grpc module version
	// go mod edit -require=google.golang.org/grpc@v1.26.0
	// go get -u -x google.golang.org/grpc@v1.26.0

	flag.Parse()
	if *help {
		flag.Usage()
		return
	}

	config := clientv3.Config{
		Endpoints:   []string{*etcdEndpoints},
		DialTimeout: 5 * time.Second,
	}
	cli, err := clientv3.New(config)
	if err != nil {
		log.Fatalln(err)
	}
	defer cli.Close()

	opType := strings.ToLower(*op)
	if opType == "put" {
		putTest(cli)
	} else if opType == "watch" {
		watchTest(cli)
	} else if opType == "get" {
		getTest(cli)
	} else if opType == "get_prefix" {
		getTestPrefix(cli)
	} else {
		log.Println(("default"))
		appWatchAndRestartService(cli)
	}

	log.Println("Done")
}
