package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// Refer: https://github.com/ayanamist/k8s-utils/blob/3d5e35f70386d8dceebdc070cadecb9be0c518bd/cmd/streamlister/main.go

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		log.Fatal("clientcmd.BuildConfigFromFlags failed:", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal("kubernetes.NewForConfig failed", err)
	}

	logMemStat("before list")

	podCh := make(chan *corev1.Pod, 5)

	// 以 stream 的方式进行 pod list, 逐个返回结果, 降低全量 list 时的内存占用
	go func() {
		defer close(podCh)
		clientset.CoreV1().RESTClient()
		// TODO: stream list
	}()

	var i int
	for range podCh {
		i++
	}
	log.Println("pldList count:", i)
	logMemStat("after request")
}

func logMemStat(message string) {
	var memStat runtime.MemStats
	runtime.ReadMemStats(&memStat)
	content := fmt.Sprintf("HeapAlloc:%d, HeapInuse:%d, HeapSys:%d, Sys:%d", memStat.HeapAlloc, memStat.HeapInuse, memStat.HeapSys, memStat.Sys)
	log.Println(message+": ", content)
}
