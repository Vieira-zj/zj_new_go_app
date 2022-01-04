package pkg

import (
	"k8s.io/client-go/informers"
	corev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

// MyPodInformer .
type MyPodInformer struct {
	factory informers.SharedInformerFactory
}

// NewMyPodInformer .
func NewMyPodInformer(factory informers.SharedInformerFactory) *MyPodInformer {
	return &MyPodInformer{
		factory: factory,
	}
}

// Informer .
func (informer *MyPodInformer) Informer() cache.SharedIndexInformer {
	return informer.factory.Core().V1().Pods().Informer()
}

// Lister .
func (informer *MyPodInformer) Lister() corev1.PodLister {
	return informer.factory.Core().V1().Pods().Lister()
}
