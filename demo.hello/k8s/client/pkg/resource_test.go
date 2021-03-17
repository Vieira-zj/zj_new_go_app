package pkg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/client-go/kubernetes"
)

var (
	client     *kubernetes.Clientset
	kubeConfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
)

func init() {
	var err error
	client, err = CreateK8sClientLocal(kubeConfig)
	if err != nil {
		panic("build k8s client error: " + err.Error())
	}
}

func TestGetSpecifiedPod(t *testing.T) {
	ns := "mini-test-ns"
	name := "hello-minikube-865c7f68f4-dgwcx"
	resource := NewResource(context.TODO(), client)
	pod, err := resource.GetSpecifiedPod(ns, name)
	if err != nil {
		t.Fatal(err)
	}
	PrintPodInfo(pod)
}

func TestGetPodsByNamespace(t *testing.T) {
	ns := "mini-test-ns"
	resource := NewResource(context.TODO(), client)
	pods, err := resource.GetPodsByNamespace(ns)
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
	resource := NewResource(context.TODO(), client)
	namespace, err := resource.GetNamespace(ns)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Namespace:", namespace.ObjectMeta.Name)
}

func TestGetPodNamespace(t *testing.T) {
	name := "hello-minikube-865c7f68f4-dgwcx"
	resource := NewResource(context.TODO(), client)
	namespace, err := resource.GetPodNamespace(name)
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
	resource := NewResource(context.TODO(), client)
	pods, err := resource.GetPodsByService(namespace, service)
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
	resource := NewResource(context.TODO(), client)
	pods, err := resource.GetPodsByServiceV2(namespace, service)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("pods in service [%s]:\n", service)
	for _, pod := range pods {
		PrintPodStatus(&pod)
		fmt.Println()
	}
}
