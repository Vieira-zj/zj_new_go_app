package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	clientOnce    sync.Once
	localClient   *kubernetes.Clientset
	clusterClient *kubernetes.Clientset
	k8sConfig     *rest.Config
)

// CreateK8sClientLocalDefault .
func CreateK8sClientLocalDefault() (*kubernetes.Clientset, error) {
	kubeConfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	return CreateK8sClientLocal(kubeConfig)
}

// CreateK8sClientLocal 在 k8s 集群外（master上）执行 client.
func CreateK8sClientLocal(kubeConfig string) (*kubernetes.Clientset, error) {
	var err error
	clientOnce.Do(func() {
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
		if err != nil {
			return
		}

		clientset, err := kubernetes.NewForConfig(k8sConfig)
		if err != nil {
			return
		}
		localClient = clientset
	})

	if err != nil {
		return nil, err
	}
	return localClient, nil
}

// CreateK8sClient 在集群内部（pod中）执行 client.
func CreateK8sClient() (*kubernetes.Clientset, error) {
	var err error
	clientOnce.Do(func() {
		k8sConfig, err = rest.InClusterConfig()
		if err != nil {
			return
		}

		clientset, err := kubernetes.NewForConfig(k8sConfig)
		if err != nil {
			return
		}
		clusterClient = clientset
	})

	if err != nil {
		return nil, err
	}
	return clusterClient, nil
}

// GetK8sConfig returns k8s config.
func GetK8sConfig() (*rest.Config, error) {
	if k8sConfig == nil {
		return nil, fmt.Errorf("k8s config does not init")
	}
	return k8sConfig, nil
}
