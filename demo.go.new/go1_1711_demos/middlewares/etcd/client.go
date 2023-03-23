package etcd

import (
	"log"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

const localEtcdEndpoint = "localhost:2379"

var (
	etcd     *clientv3.Client
	etcdOnce sync.Once
)

func NewEtcdClient() *clientv3.Client {
	etcdOnce.Do(func() {
		etcd = initEtcd()
	})
	return etcd
}

func initEtcd() *clientv3.Client {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{localEtcdEndpoint},
		DialTimeout: time.Second,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
	})
	if err != nil {
		log.Fatal(err)
	}
	return etcd
}
