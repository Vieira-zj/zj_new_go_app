package pkg

import (
	"io"
	"log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// PtyHandler is what remotecommand expects from a pty.
type PtyHandler interface {
	io.Reader
	io.Writer
	remotecommand.TerminalSizeQueue
	Done() chan struct{}
}

// ExecPod execs command in given pod.
func ExecPod(kubeClient kubernetes.Interface, cfg *restclient.Config,
	cmd []string, ptyHandler PtyHandler, namespace, podName, containerName string) error {
	log.Println("init pod executor with sh session:", cmd)

	req := kubeClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")
	req.VersionedParams(&corev1.PodExecOptions{
		Container: containerName,
		Command:   cmd,
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(cfg, "POST", req.URL())
	if err != nil {
		return err
	}

	return executor.Stream(remotecommand.StreamOptions{
		Stdin:             ptyHandler,
		Stdout:            ptyHandler,
		Stderr:            ptyHandler,
		TerminalSizeQueue: ptyHandler,
		Tty:               true,
	})
}
