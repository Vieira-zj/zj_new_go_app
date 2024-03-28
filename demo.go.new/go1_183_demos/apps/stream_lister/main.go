package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	streamlister "demo.apps/apps/stream_lister/pkg"
	"github.com/gogo/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		client := clientset.CoreV1().RESTClient()
		if err := streamlister.StreamList(context.TODO(), client, "pods", "default", metav1.ListOptions{ResourceVersion: "0"}, streamlister.Param{
			ObjectFactoryFunc: func() proto.Unmarshaler {
				return &corev1.Pod{}
			},
			OnListMetaFunc: func(listMeta *metav1.ListMeta) {
				log.Printf("rv=%s", listMeta.ResourceVersion)
			},
			OnObjectFunc: func(obj proto.Unmarshaler) {
				pod := obj.(*corev1.Pod)
				podCh <- pod
			},
		}); err != nil {
			log.Fatal("streamlister.StreamList failed:", err)
		}
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
