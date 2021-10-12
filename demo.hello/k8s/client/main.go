package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	k8spkg "demo.hello/k8s/client/pkg"
)

func main() {
	var kubeConfig *string
	if home := os.Getenv("HOME"); home != "" {
		kubeConfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file.")
	} else {
		kubeConfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file.")
	}

	namespace := flag.String("n", "", "k8s namespace name.")
	service := flag.String("s", "", "k8s service name.")
	deploy := flag.String("d", "", "k8s deploy name.")
	inject := flag.String("i", "", "inject container into a pod, available values: sidecar, tcpforward.")
	help := flag.Bool("h", false, "help.")

	isListNamespacePods := flag.Bool("listnspods", false, "get pods in a namespace.")
	isGetServicePods := flag.Bool("getservicepods", false, "get pods in a service.")
	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	clientset, err := k8spkg.CreateK8sClientLocal(*kubeConfig)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("try to create client in k8s cluster.")
		if clientset, err = k8spkg.CreateK8sClient(); err != nil {
			panic(err.Error())
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
	defer cancel()
	resource := k8spkg.NewResource(clientset)

	if *isListNamespacePods {
		if len(*namespace) == 0 {
			panic(errors.New("namepsace name is empty"))
		}

		pods, err := resource.GetPodsByNamespace(ctx, *namespace)
		if err != nil {
			panic(err.Error())
		}
		for _, pod := range pods {
			k8spkg.PrintPodInfo(&pod)
			fmt.Println()
		}
	}

	if *isGetServicePods {
		if len(*namespace) == 0 || len(*service) == 0 {
			panic(errors.New("namespace or service name is empty"))
		}

		pods, err := resource.GetPodsByServiceV2(ctx, *namespace, *service)
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("pods in service [%s]:\n", *service)
		for _, pod := range pods {
			k8spkg.PrintPodStatus(&pod)
			fmt.Println()
		}
	}

	if len(*inject) > 0 {
		injector := k8spkg.NewInjector(ctx, clientset)
		if *inject == "sidecar" {
			if err := injector.InjectBusyboxSidecar(*namespace, *deploy); err != nil {
				panic(err.Error())
			}
		}
		if *inject == "tcpforward" {
			if err := injector.InjectInitTCPForword(*namespace, *deploy); err != nil {
				panic(err.Error())
			}
		}
	}
}
