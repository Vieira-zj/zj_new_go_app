package pkg

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	dockerclient "github.com/docker/docker/client"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/remotecommand"
	kubectlscheme "k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// Executor .
type Executor struct {
	client   *kubernetes.Clientset
	resource *Resource
}

// NewExecutor .
func NewExecutor(client *kubernetes.Clientset, resource *Resource) Executor {
	return Executor{
		client:   client,
		resource: resource,
	}
}

// RunCmd executes certain command and returns the result.
func (e *Executor) RunCmd(ctx context.Context, pod *corev1.Pod, cmd string) (string, error) {
	podName := pod.GetObjectMeta().GetName()
	namespace := pod.GetObjectMeta().GetNamespace()
	containerName := pod.Spec.Containers[0].Name
	fmt.Printf("[namespace=%s, pod=%s, container=%s], run cmd [%s].\n", namespace, podName, containerName, cmd)

	req := e.client.CoreV1().RESTClient().Post().Resource("pods").Name(podName).Namespace(namespace).SubResource("exec")
	req.VersionedParams(&corev1.PodExecOptions{
		Container: containerName,
		Command:   []string{"/bin/sh", "-c", cmd},
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, kubectlscheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config.GetConfigOrDie(), "POST", req.URL())
	if err != nil {
		return "", fmt.Errorf("error in creating NewSPDYExecutor for pod %s/%s:\n%v", podName, namespace, err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		if stderr.String() != "" {
			return "", fmt.Errorf("error: %s\npod: %s/%s\ncommand: %s", strings.TrimSuffix(stderr.String(), "\n"), podName, namespace, cmd)
		}
		return "", fmt.Errorf("error in streaming remotecommand: pod: %s/%s, command: %s:\n%v", podName, namespace, cmd, err)
	}
	if stderr.String() != "" {
		return "", fmt.Errorf("error of command %s:\n%s", cmd, stderr.String())
	}

	return stdout.String(), nil
}

// RunCmdBypass uses debug daemon pod to enter namespace and execute command in target pod.
func (e *Executor) RunCmdBypass(ctx context.Context, pod *corev1.Pod, daemonPod *corev1.Pod, cmd string) (string, error) {
	containerID := pod.Status.ContainerStatuses[0].ContainerID
	cid, err := formatContainerID(containerID)
	if err != nil {
		return "", err
	}

	os.Setenv("DOCKER_CERT_PATH", filepath.Join(os.Getenv("HOME"), ".minikube/certs"))
	os.Setenv("DOCKER_HOST", fmt.Sprintf("tcp://%s:2376", pod.Status.HostIP))
	dClient, err := dockerclient.NewClientWithOpts(
		dockerclient.FromEnv,
		dockerclient.WithVersion("1.40"),
	)
	if err != nil {
		return "", err
	}

	// get container pid
	container, err := dClient.ContainerInspect(ctx, cid)
	if err != nil {
		return "", err
	}
	pid := container.State.Pid
	if container.State.Pid == 0 {
		return "", fmt.Errorf("container is not running, status: %s", container.State.Status)
	}

	// run cmd in debug daemon pod (has privilege) within target container pid namespace
	cmd = buildCmdWithPidNs(ctx, pid, cmd)
	return e.RunCmd(ctx, daemonPod, cmd)
}

func (e *Executor) getPodInDaemonSetByAddr(ctx context.Context, daemonSet *appsv1.DaemonSet, addr string) (*corev1.Pod, error) {
	namespace := daemonSet.GetObjectMeta().GetNamespace()
	labels := daemonSet.Spec.Template.GetLabels()
	pods, err := e.resource.GetPodsByLabelsV2(namespace, labels)
	if err != nil {
		return nil, err
	}

	for _, pod := range pods {
		if pod.Status.HostIP == addr {
			return &pod, nil
		}
	}
	return nil, fmt.Errorf("daemon pod [addr=%s] not found in daemonSet [%s/%s]", addr, daemonSet.GetObjectMeta().GetName(), namespace)
}

func formatContainerID(containerID string) (string, error) {
	const dockerProtocolPrefix = "docker://"
	prefixLength := len(dockerProtocolPrefix)
	if len(containerID) < prefixLength {
		return "", fmt.Errorf("container id %s is not a docker container id", containerID)
	}

	prefix := containerID[0:prefixLength]
	if prefix != dockerProtocolPrefix {
		return "", fmt.Errorf("expected %s but got %s", dockerProtocolPrefix, prefix)
	}

	return containerID[prefixLength:], nil
}

func buildCmdWithPidNs(ctx context.Context, pid int, cmd string) string {
	const nsexec = "/usr/local/bin/nsexec"
	args := []string{"--", cmd}
	args = append([]string{"-p", getPidNsPath(pid)}, args...)
	return exec.CommandContext(ctx, nsexec, args...).String()
}

func getPidNsPath(pid int) string {
	return fmt.Sprintf("/proc/%d/ns/pid", pid)
}
