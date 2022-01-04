package main

import (
	"context"
	"flag"
	"fmt"
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
	flag.StringVar(&wType, "m", typeReplicaSet, "k8s resource watch type: pod, rs. default: pod")
	flag.BoolVar(&help, "h", false, "help")
}

func main() {
	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	// init
	client, err := k8spkg.CreateK8sClientLocalDefault()
	if err != nil {
		panic(fmt.Sprintf("init k8s client error: %v\n", err))
	}

	var watcher localpkg.Watcher
	factory := informers.NewSharedInformerFactoryWithOptions(
		client, time.Duration(interval)*time.Second, informers.WithNamespace(namespace))
	switch wType {
	case typePod:
		podInformer := localpkg.NewMyPodInformer(factory)
		watcher = localpkg.NewPodWatcher(client, podInformer)
	case typeReplicaSet:
		rsInformer := localpkg.NewMyReplicaSetInformer(factory)
		dInformer := localpkg.NewMyDeploymentInformer(factory)
		watcher = localpkg.NewReplicaSetWatcher(client, rsInformer, dInformer)
	}

	// run
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM)
	factory.Start(ctx.Done())
	watcher.Run(ctx)

	// exit
	<-ctx.Done()
	stop()
	fmt.Println("k8s informer exit")
}
