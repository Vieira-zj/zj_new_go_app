package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	k8spkg "demo.hello/k8s/client/pkg"
	localpkg "demo.hello/k8s/informer/pkg"
	"k8s.io/client-go/informers"
)

const (
	typePod        = "pod"
	typeReplicaSet = "rs"
)

var (
	interval  int
	namespace string
	wType     string
	help      bool
)

func init() {
	flag.IntVar(&interval, "i", 15, "watch interval")
	flag.StringVar(&namespace, "n", "k8s-test", "watch namespace")
	flag.StringVar(&wType, "t", typePod, "k8s resource informer type: pod, rs. default: pod")
	flag.BoolVar(&help, "h", false, "help")
}

func main() {
	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	client, err := k8spkg.CreateK8sClientLocalDefault()
	if err != nil {
		panic(fmt.Sprintf("init k8s client error: %v\n", err))
	}

	var watcher localpkg.Watcher
	factory := informers.NewSharedInformerFactoryWithOptions(
		client, time.Duration(interval)*time.Second, informers.WithNamespace(namespace))
	podInformer := localpkg.NewMyPodInformer(factory)
	switch wType {
	case typePod:
		watcher = localpkg.NewPodWatcher(client, podInformer)
	case typeReplicaSet:
		rsInformer := localpkg.NewMyReplicaSetInformer(factory)
		dInformer := localpkg.NewMyDeploymentInformer(factory)
		watcher = localpkg.NewReplicaSetWatcher(client, rsInformer, dInformer, podInformer)
	default:
		panic("invalid informer type: " + wType)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM)
	factory.Start(ctx.Done())
	watcher.Run(ctx)

	<-ctx.Done()
	stop()
	log.Println("k8s informer exit")
}
