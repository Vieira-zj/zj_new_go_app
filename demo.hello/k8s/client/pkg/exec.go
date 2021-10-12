package pkg

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	dockerclient "github.com/docker/docker/client"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	kubectlscheme "k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

//
// K8s Remote Executor
//

// Executor .
type Executor struct {
	client   *kubernetes.Clientset
	resource *Resource
}

// NewExecutor .
func NewExecutor(kubeConfig string) (*Executor, error) {
	client, err := CreateK8sClientLocal(kubeConfig)
	if err != nil {
		return nil, err
	}

	return &Executor{
		client:   client,
		resource: NewResource(client),
	}, nil
}

// RunCmd executes certain command and returns the result.
func (e *Executor) RunCmd(ctx context.Context, pod *corev1.Pod, cmd string) (string, error) {
	podName := pod.GetObjectMeta().GetName()
	namespace := pod.GetObjectMeta().GetNamespace()
	containerName := pod.Spec.Containers[0].Name
	fmt.Printf("[namespace=%s, pod=%s, container=%s], run cmd [%s].\n", namespace, podName, containerName, cmd)

	req := e.client.CoreV1().RESTClient().Post().
		Resource("pods").Name(podName).Namespace(namespace).SubResource("exec")
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
	pods, err := e.resource.GetPodsByLabelsV2(ctx, namespace, labels)
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

//
// K8s Remote Executor V2
//

// ExecOptions includes exec options passed to RemoteExec.
type ExecOptions struct {
	Command       []string
	Namespace     string
	PodName       string
	ContainerName string

	Stdin              io.Reader
	CaptureStdout      bool
	CaptureStderr      bool
	PreserveWhitespace bool
}

// RemoteExec runs command on given pod in namespace.
type RemoteExec struct {
	client *kubernetes.Clientset
}

// NewRemoteExec returns an instance of RemoteExec.
func NewRemoteExec(kubeConfig string) (*RemoteExec, error) {
	client, err := CreateK8sClientLocal(kubeConfig)
	if err != nil {
		return nil, err
	}
	return &RemoteExec{
		client: client,
	}, nil
}

// RunCommandInPod runs a command in given pod, and returns stdout and stderr string.
func (exec *RemoteExec) RunCommandInPod(ctx context.Context, namespace, podName string, cmd string) (string, string, error) {
	pod, err := exec.client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return "", "", fmt.Errorf("Get pod [%s/%s] error: %v", namespace, podName, err)
	}
	return exec.runCommandInContainer(namespace, podName, pod.Spec.Containers[0].Name, cmd)
}

func (exec *RemoteExec) runCommandInContainer(namespace, podName, containerName string, cmd string) (string, string, error) {
	return exec.execWithOptions(ExecOptions{
		Command:       []string{"/bin/sh", "-c", cmd},
		Namespace:     namespace,
		PodName:       podName,
		ContainerName: containerName,

		Stdin:         nil,
		CaptureStdout: true,
		CaptureStderr: true,
	})
}

func (exec *RemoteExec) execWithOptions(options ExecOptions) (string, string, error) {
	const tty = false
	req := exec.client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(options.PodName).
		Namespace(options.Namespace).
		SubResource("exec").
		Param("container", options.ContainerName)
	req.VersionedParams(&corev1.PodExecOptions{
		Container: options.ContainerName,
		Command:   options.Command,
		Stdin:     options.Stdin != nil,
		Stdout:    options.CaptureStdout,
		Stderr:    options.CaptureStderr,
		TTY:       tty,
	}, kubectlscheme.ParameterCodec)

	var stdout, stderr bytes.Buffer
	if err := execute("POST", req.URL(), config.GetConfigOrDie(), options.Stdin, &stdout, &stderr, tty); err != nil {
		if stderr.String() != "" {
			err = fmt.Errorf("error: %s\npod: %s/%s\ncommand: %v", strings.TrimSuffix(stderr.String(), "\n"), options.Namespace, options.PodName, options.Command)
			return "", "", err
		}
		err = fmt.Errorf("error in streaming remotecommand: pod: %s/%s, command: %v:\n%v", options.Namespace, options.PodName, options.Command, err)
		return "", "", err
	}
	if stderr.String() != "" {
		err := fmt.Errorf("error of command %v:\n%s", options.Command, stderr.String())
		return "", "", err
	}

	return stdout.String(), stderr.String(), nil
}

func execute(method string, url *url.URL, config *restclient.Config, stdin io.Reader, stdout, stderr io.Writer, tty bool) error {
	exec, err := remotecommand.NewSPDYExecutor(config, method, url)
	if err != nil {
		return err
	}

	return exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Tty:    tty,
	})
}
