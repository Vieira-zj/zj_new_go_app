package pkg

import (
	"k8s.io/client-go/informers"
	appsv1 "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
)

// MyReplicaSetInformer .
type MyReplicaSetInformer struct {
	factory informers.SharedInformerFactory
}

// NewMyReplicaSetInformer .
func NewMyReplicaSetInformer(factory informers.SharedInformerFactory) *MyReplicaSetInformer {
	return &MyReplicaSetInformer{
		factory: factory,
	}
}

// Informer .
func (informer *MyReplicaSetInformer) Informer() cache.SharedIndexInformer {
	return informer.factory.Apps().V1().ReplicaSets().Informer()
}

// Lister .
func (informer *MyReplicaSetInformer) Lister() appsv1.ReplicaSetLister {
	return informer.factory.Apps().V1().ReplicaSets().Lister()
}
