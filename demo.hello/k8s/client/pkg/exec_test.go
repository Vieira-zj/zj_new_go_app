package pkg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	dockerclient "github.com/docker/docker/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	executor *Executor
	execCtx  context.Context
)

func init() {
	var err error
	kubeConfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	executor, err = NewExecutor(kubeConfig)
	if err != nil {
		panic(err)
	}
	execCtx = context.Background()
}

func TestDockerClient(t *testing.T) {
	// refer: https://github.com/moby/moby/tree/master/client
	// start a test pod:
	// kubectl run test-pod -it --rm --restart=Never -n k8s-test --image=busybox:1.30 sh
	namespace := "k8s-test"
	pod, err := executor.resource.GetPod(execCtx, namespace, "test-pod")
	if err != nil {
		t.Fatal(err)
	}

	os.Setenv("DOCKER_CERT_PATH", filepath.Join(os.Getenv("HOME"), ".minikube/certs"))
	os.Setenv("DOCKER_HOST", fmt.Sprintf("tcp://%s:2376", pod.Status.HostIP))
	client, err := dockerclient.NewClientWithOpts(
		dockerclient.FromEnv,
		dockerclient.WithVersion("1.40"),
	)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("containers info:")
	containers, err := client.ContainerList(execCtx, types.ContainerListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, container := range containers {
		if !strings.HasPrefix(container.Image, "k8s.gcr.io") {
			fmt.Printf("%s %s\n", container.ID[:8], container.Image)
		}
	}
}

func TestGetPodInDaemonSetByAddr(t *testing.T) {
	namespace := "k8s-test"
	name := "debug-daemon"
	daemonSet, err := executor.client.AppsV1().DaemonSets(namespace).Get(execCtx, name, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}

	targetPod, err := executor.resource.GetPod(execCtx, namespace, "test-pod")
	if err != nil {
		t.Fatal(err)
	}

	daemonPod, err := executor.getPodInDaemonSetByAddr(execCtx, daemonSet, targetPod.Status.HostIP)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("daemon pod: name=%s, host_ip=%s\n", daemonPod.GetObjectMeta().GetName(), daemonPod.Status.HostIP)
}

func TestExecCmd(t *testing.T) {
	namespace := "k8s-test"
	name := "test-pod"
	pod, err := executor.resource.GetPod(execCtx, namespace, name)
	if err != nil {
		t.Fatal(err)
	}

	cmd := "hostname;printenv"
	res, err := executor.RunCmd(execCtx, pod, cmd)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("resutls for cmd [%s]:\n%s", cmd, res)
}

func TestExecCmdBypass(t *testing.T) {
	namespace := "k8s-test"
	name := "test-pod"
	targetPod, err := executor.resource.GetPod(execCtx, namespace, name)
	if err != nil {
		t.Fatal(err)
	}

	daemonSet, err := executor.client.AppsV1().DaemonSets(namespace).Get(execCtx, "debug-daemon", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	daemonPod, err := executor.getPodInDaemonSetByAddr(execCtx, daemonSet, targetPod.Status.HostIP)
	if err != nil {
		t.Fatal(err)
	}

	// run "ps" within pid namespace
	cmd := "ps"
	res, err := executor.RunCmdBypass(execCtx, targetPod, daemonPod, cmd)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res)
}

func TestRemoteExec(t *testing.T) {
	kubeConfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	exec, err := NewRemoteExec(kubeConfig)
	if err != nil {
		t.Fatal(err)
	}

	namespace := "k8s-test"
	podName := "test-pod"
	timeoutCtx, cancel := context.WithTimeout(execCtx, time.Duration(5)*time.Second)
	defer cancel()
	stdout, stderr, err := exec.RunCommandInPod(timeoutCtx, namespace, podName, "ps -ef")
	if err != nil {
		t.Fatal(err)
	}
	if len(stderr) > 0 {
		fmt.Println("stderr output:", stderr)
	}
	fmt.Println("std output:", stdout)
}
