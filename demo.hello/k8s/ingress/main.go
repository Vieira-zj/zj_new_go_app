package main

import (
	"context"
	"flag"
	"fmt"

	"demo.hello/k8s/ingress/server"
	"demo.hello/k8s/ingress/watcher"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	host          string
	port, tlsPort int
)

func main() {
	flag.StringVar(&host, "host", "0.0.0.0", "the host to bind")
	flag.IntVar(&port, "port", 80, "the insecure http port")
	flag.IntVar(&tlsPort, "tls-port", 443, "the secure https port")
	flag.Parse()

	runtime.ErrorHandlers = []func(error){
		func(err error) {
			fmt.Println("[k8s]", err)
		},
	}

	// 从集群内的token和ca.crt获取 Config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(fmt.Sprintln("get kubernetes configuration failed:", err))
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(fmt.Sprintln("create kubernetes client failed:", err))
	}

	s := server.New(server.WithHost(host), server.WithPort(port), server.WithTLSPort(tlsPort))
	w := watcher.New(client, func(payload *watcher.Payload) {
		s.Update(payload)
	})

	// run
	var eg errgroup.Group
	eg.Go(func() error {
		return s.Run(context.TODO())
	})
	eg.Go(func() error {
		return w.Run(context.TODO())
	})

	if err := eg.Wait(); err != nil {
		panic(err)
	}
}
