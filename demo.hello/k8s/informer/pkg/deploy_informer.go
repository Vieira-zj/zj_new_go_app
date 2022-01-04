package pkg

import (
	"k8s.io/client-go/informers"
	apps "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
)

// MyDeploymentInformer .
type MyDeploymentInformer struct {
	factory informers.SharedInformerFactory
}

// NewMyDeploymentInformer .
func NewMyDeploymentInformer(factory informers.SharedInformerFactory) *MyDeploymentInformer {
	return &MyDeploymentInformer{
		factory: factory,
	}
}

// Informer .
func (informer *MyDeploymentInformer) Informer() cache.SharedIndexInformer {
	return informer.factory.Apps().V1().Deployments().Informer()
}

// Lister .
func (informer *MyDeploymentInformer) Lister() apps.DeploymentLister {
	return informer.factory.Apps().V1().Deployments().Lister()
}
