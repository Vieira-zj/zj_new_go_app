package pkg

import (
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	clientOnce    sync.Once
	localClient   *kubernetes.Clientset
	clusterClient *kubernetes.Clientset
)

// CreateK8sClientLocal 在 k8s 集群外（master上）执行 client.
func CreateK8sClientLocal(kubeConfig string) (*kubernetes.Clientset, error) {
	var clientErr error
	clientOnce.Do(func() {
		config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
		if err != nil {
			clientErr = err
			return
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			clientErr = err
		}
		localClient = clientset
	})

	if clientErr != nil {
		return nil, clientErr
	}
	return localClient, nil
}

// CreateK8sClient 在集群内部（pod中）执行 client.
func CreateK8sClient() (*kubernetes.Clientset, error) {
	var clientErr error
	clientOnce.Do(func() {
		config, err := rest.InClusterConfig()
		if err != nil {
			clientErr = err
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			clientErr = err
		}
		clusterClient = clientset
	})

	if clientErr != nil {
		return nil, clientErr
	}
	return clusterClient, nil
}
