package pkg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var k8sResource *Resource

func init() {
	kubeConfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	client, err := CreateK8sClientLocal(kubeConfig)
	if err != nil {
		panic("build k8s client error: " + err.Error())
	}
	k8sResource = NewResource(context.TODO(), client)
}

func TestGetSpecifiedPod(t *testing.T) {
	ns := "mini-test-ns"
	name := "hello-minikube-865c7f68f4-dgwcx"
	pod, err := k8sResource.GetSpecifiedPod(ns, name)
	if err != nil {
		t.Fatal(err)
	}
	PrintPodInfo(pod)
}

func TestGetPodsByNamespace(t *testing.T) {
	ns := "mini-test-ns"
	pods, err := k8sResource.GetPodsByNamespace(ns)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Pods count:", len(pods))

	for _, pod := range pods {
		PrintPodInfo(&pod)
	}
}

func TestGetNamespace(t *testing.T) {
	ns := "k8s-test-ns"
	namespace, err := k8sResource.GetNamespace(ns)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Namespace:", namespace.ObjectMeta.Name)
}

func TestGetPodNamespace(t *testing.T) {
	name := "hello-minikube-865c7f68f4-dgwcx"
	namespace, err := k8sResource.GetPodNamespace(name)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Pod [%s] in namespace: [%s]\n", name, namespace)

	name = "pod-not-exist"
	namespace, err = resource.GetPodNamespace(name)
	if err == nil {
		t.Fatal("want pod NotFoundError, got specified pod")
	}
	fmt.Println("Error:", err)
}

func TestGetPodsByService(t *testing.T) {
	namespace := "mini-test-ns"
	service := "hello-minikube"
	pods, err := k8sResource.GetPodsByService(namespace, service)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("pods in service [%s]:\n", service)
	for _, pod := range pods {
		PrintPodStatus(&pod)
		fmt.Println()
	}
}

func TestGetPodsByServiceV2(t *testing.T) {
	namespace := "mini-test-ns"
	service := "hello-minikube"
	pods, err := k8sResource.GetPodsByServiceV2(namespace, service)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("pods in service [%s]:\n", service)
	for _, pod := range pods {
		PrintPodStatus(&pod)
		fmt.Println()
	}
}

func TestGetNonSystemPods(t *testing.T) {
	context, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
	defer cancel()
	k8sResource.SetContext(context)

	pods, err := k8sResource.GetNonSystemPods()
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("all non kube system pods:")
	for _, pod := range pods {
		fmt.Println(pod.ObjectMeta.Name)
	}
}
