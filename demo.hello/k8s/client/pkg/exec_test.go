package pkg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	dockerclient "github.com/docker/docker/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	execClient   *kubernetes.Clientset
	execResource *Resource
	executor     Executor
)

func init() {
	var err error
	kubeConfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	execClient, err = CreateK8sClientLocal(kubeConfig)
	if err != nil {
		panic("build k8s client error: " + err.Error())
	}
	execResource = NewResource(context.Background(), execClient)
	executor = NewExecutor(execClient, execResource)
}

func TestDockerClient(t *testing.T) {
	// refer: https://github.com/moby/moby/tree/master/client
	// start a test pod:
	// kubectl run test-pod -it --rm --restart=Never -n k8s-test --image=busybox:1.30 sh
	namespace := "k8s-test"
	pod, err := execResource.GetPod(namespace, "test-pod")
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
	containers, err := client.ContainerList(context.Background(), types.ContainerListOptions{})
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
	ctx := context.Background()
	daemonSet, err := execClient.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}

	targetPod, err := execResource.GetPod(namespace, "test-pod")
	if err != nil {
		t.Fatal(err)
	}

	daemonPod, err := executor.getPodInDaemonSetByAddr(ctx, daemonSet, targetPod.Status.HostIP)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("daemon pod: name=%s, host_ip=%s\n", daemonPod.GetObjectMeta().GetName(), daemonPod.Status.HostIP)
}

func TestExecCmd(t *testing.T) {
	namespace := "k8s-test"
	name := "test-pod"
	pod, err := execResource.GetPod(namespace, name)
	if err != nil {
		t.Fatal(err)
	}

	cmd := "hostname;printenv"
	res, err := executor.RunCmd(context.Background(), pod, cmd)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("resutls for cmd [%s]: %s\n", cmd, res)
}

func TestExecCmdBypass(t *testing.T) {
	namespace := "k8s-test"
	name := "test-pod"
	targetPod, err := execResource.GetPod(namespace, name)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	daemonSet, err := execClient.AppsV1().DaemonSets(namespace).Get(ctx, "debug-daemon", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	daemonPod, err := executor.getPodInDaemonSetByAddr(ctx, daemonSet, targetPod.Status.HostIP)
	if err != nil {
		t.Fatal(err)
	}

	// run "ps" within pid namespace
	cmd := "ps"
	res, err := executor.RunCmdBypass(ctx, targetPod, daemonPod, cmd)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res)
}
