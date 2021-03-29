package main

import (
	"context"
	"encoding/json"
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
	isMonitorPods := flag.Bool("monitorpods", false, "monitor all pods status.")
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

	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resource := k8spkg.NewResource(context, clientset)
	if *isListNamespacePods {
		if len(*namespace) == 0 {
			panic(errors.New("namepsace name is empty"))
		}

		pods, err := resource.GetPodsByNamespace(*namespace)
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

		pods, err := resource.GetPodsByServiceV2(*namespace, *service)
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("pods in service [%s]:\n", *service)
		for _, pod := range pods {
			k8spkg.PrintPodStatus(&pod)
			fmt.Println()
		}
	}

	if *isMonitorPods {
		podInfos, err := k8spkg.GetAllPodInfos(resource, *namespace)
		if err != nil {
			panic(err.Error())
		}

		b, err := json.Marshal(podInfos)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println(string(b))
	}

	if len(*inject) > 0 {
		injector := k8spkg.NewInjector(context, clientset)
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
