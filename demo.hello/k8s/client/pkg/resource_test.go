package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	ctx         context.Context
	k8sResource *Resource
)

func init() {
	kubeConfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	client, err := CreateK8sClientLocal(kubeConfig)
	if err != nil {
		panic("build k8s client error: " + err.Error())
	}
	ctx = context.Background()
	k8sResource = NewResource(client)
}

func TestSets(t *testing.T) {
	// struct set test
	words := strings.Split("this is a sets test. this is a hello world", " ")
	s := sets.NewString(words...)

	fmt.Println("sorted list:")
	for _, w := range s.List() {
		fmt.Printf("%s|", w)
	}
	fmt.Println()

	fmt.Println("unsorted list:")
	for _, w := range s.UnsortedList() {
		fmt.Printf("%s|", w)
	}
	fmt.Println()

	fmt.Println("intersection items:")
	s2 := sets.NewString([]string{"hello", "world"}...)
	ret := s.Intersection(s2)
	for _, w := range ret.List() {
		fmt.Printf("%s|", w)
	}
	fmt.Println()
}

func TestGetAllNamespace(t *testing.T) {
	namespaces, err := k8sResource.GetAllNamespaces(ctx)
	if err != nil {
		t.Fatal(err)
	}

	names := make([]string, 0, len(namespaces))
	for _, ns := range namespaces {
		names = append(names, ns.GetName())
	}
	fmt.Println("all namespaces:\n", strings.Join(names, " | "))
}

func TestGetNamespace(t *testing.T) {
	ns := "k8s-test-ns"
	namespace, err := k8sResource.GetNamespace(ctx, ns)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Namespace:", namespace.ObjectMeta.Name)
}

func TestGetSpecifiedPod(t *testing.T) {
	ns := "k8s-test"
	name := "error-exit-test-584c9756b9-kcz7c"
	pod, err := k8sResource.GetPod(ctx, ns, name)
	if err != nil {
		t.Fatal(err)
	}
	PrintPodInfo(pod)
}

func TestGetPodState(t *testing.T) {
	ns := "k8s-test"
	names := [2]string{"test-pod", "error-exit-test-584c9756b9-kcz7c"}
	for _, name := range names {
		state, err := k8sResource.GetPodState(ctx, ns, name, "")
		if err != nil {
			t.Fatal(err)
		}
		if err := prettyPrintPodState(state); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCheckPodExec(t *testing.T) {
	ns := "k8s-test"
	names := [2]string{"test-pod", "error-exit-test-55d594b94b-kspp5"}
	for _, name := range names {
		if err := k8sResource.CheckPodExec(ctx, ns, name, ""); err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("check pass:", name)
		}
	}
}

func TestGetPodLogs(t *testing.T) {
	ns := "k8s-test"
	name := "error-exit-test-584c9756b9-k7pf8"
	if err := k8sResource.CheckPodExec(ctx, ns, name, ""); err != nil {
		fmt.Println("check pod results:\n" + err.Error())
	}

	logs, err := k8sResource.GetPodLogsByTailLines(ctx, ns, name, int64(10))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("pod [%s/%s] logs:\n%s", ns, name, logs)
}

func TestGetPodsByNamespace(t *testing.T) {
	ns := "mini-test-ns"
	pods, err := k8sResource.GetPodsByNamespace(ctx, ns)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Pods count:", len(pods))

	for _, pod := range pods {
		PrintPodInfo(&pod)
	}
}

func TestGetPodNamespace(t *testing.T) {
	name := "hello-minikube-865c7f68f4-dgwcx"
	namespace, err := k8sResource.GetPodNamespace(ctx, name)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Pod [%s] in namespace: [%s]\n", name, namespace)

	name = "pod-not-exist"
	namespace, err = resource.GetPodNamespace(ctx, name)
	if err == nil {
		t.Fatal("want pod NotFoundError, got specified pod")
	}
	fmt.Println("Error:", err)
}

func TestGetPodsByService(t *testing.T) {
	namespace := "mini-test-ns"
	service := "hello-minikube"
	pods, err := k8sResource.GetPodsByService(ctx, namespace, service)
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
	pods, err := k8sResource.GetPodsByServiceV2(ctx, namespace, service)
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
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(3)*time.Second)
	defer cancel()

	pods, err := k8sResource.GetNonSystemPods(timeoutCtx)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("all non kube system pods:")
	for _, pod := range pods {
		fmt.Println(pod.ObjectMeta.Name)
	}
}

func prettyPrintPodState(state *PodState) error {
	b, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("pods state info:\n%s\n", b)
	return nil
}
