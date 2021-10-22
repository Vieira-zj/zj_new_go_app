package internal

import (
	"context"
	"log"
	"time"

	k8spkg "demo.hello/k8s/client/pkg"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// Watcher .
type Watcher struct {
	namespaces       []string
	resource         *k8spkg.Resource
	factory          informers.SharedInformerFactory
	ErrorPodStatusCh chan *PodStatus
}

// NewWatcher creates a list watcher instance.
func NewWatcher(client *kubernetes.Clientset, namespaces []string, interval uint) *Watcher {
	if len(namespaces) == 0 {
		panic("init watcher, namespace cannot be empty")
	}

	var factory informers.SharedInformerFactory
	if len(namespaces) == 1 {
		factory = informers.NewSharedInformerFactoryWithOptions(
			client, time.Duration(interval)*time.Second, informers.WithNamespace(namespaces[0]))
	} else {
		factory = informers.NewSharedInformerFactoryWithOptions(client, time.Duration(interval)*time.Second)
	}
	return &Watcher{
		namespaces:       namespaces,
		resource:         k8spkg.NewResource(client),
		factory:          factory,
		ErrorPodStatusCh: make(chan *PodStatus, interval),
	}
}

// Run exec watcher informer.
func (w *Watcher) Run(ctx context.Context, isWatchDeploy bool) {
	log.Println("start watcher")
	go w.podInformer().Run(ctx.Done())
	if isWatchDeploy {
		go w.deploymentInformer().Run(ctx.Done())
	}
}

func (w *Watcher) podInformer() cache.SharedIndexInformer {
	onAdd := func(obj interface{}) {
		if pod, ok := obj.(*corev1.Pod); ok {
			if len(w.namespaces) > 0 {
				if !isInSlice(pod.GetObjectMeta().GetNamespace(), w.namespaces) {
					return
				}
			}
			log.Println("onAdd pod:", pod.ObjectMeta.GetName())
		}
	}

	onUpdate := func(oldObj, newObj interface{}) {
		if pod, ok := newObj.(*corev1.Pod); ok {
			if len(w.namespaces) > 0 {
				if !isInSlice(pod.GetObjectMeta().GetNamespace(), w.namespaces) {
					return
				}
			}

			log.Println("onUpdate pod:", pod.ObjectMeta.GetName())
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
			defer cancel()
			status := GetPodStatus(ctx, w.resource, pod, "")
			if status.Value != "Running" && status.Value != "ContainerCreating" {
				w.ErrorPodStatusCh <- status
			}
		}
	}

	onDelete := func(obj interface{}) {
		if pod, ok := obj.(*corev1.Pod); ok {
			if len(w.namespaces) > 0 {
				if !isInSlice(pod.GetObjectMeta().GetNamespace(), w.namespaces) {
					return
				}
			}
			log.Println("onDelete pod:", pod.ObjectMeta.GetName())
		}
	}

	handler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			onAdd(obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			onUpdate(oldObj, newObj)
		},
		DeleteFunc: func(obj interface{}) {
			onDelete(obj)
		},
	}

	informer := w.factory.Core().V1().Pods().Informer()
	informer.AddEventHandler(handler)
	return informer
}

func (w *Watcher) deploymentInformer() cache.SharedIndexInformer {
	handler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if deploy, ok := obj.(*appsv1.Deployment); ok {
				if len(w.namespaces) > 0 {
					if !isInSlice(deploy.GetObjectMeta().GetNamespace(), w.namespaces) {
						return
					}
				}
				log.Println("onAdd deployment:", deploy.ObjectMeta.GetName())
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if deploy, ok := newObj.(*appsv1.Deployment); ok {
				if len(w.namespaces) > 0 {
					if !isInSlice(deploy.GetObjectMeta().GetNamespace(), w.namespaces) {
						return
					}
				}
				log.Println("onUpdate deployment:", deploy.ObjectMeta.GetName())
			}
		},
		DeleteFunc: func(obj interface{}) {
			if deploy, ok := obj.(*appsv1.Deployment); ok {
				if len(w.namespaces) > 0 {
					if !isInSlice(deploy.GetObjectMeta().GetNamespace(), w.namespaces) {
						return
					}
				}
				log.Println("onDelete deployment:", deploy.ObjectMeta.GetName())
			}
		},
	}

	informers := w.factory.Apps().V1().Deployments().Informer()
	informers.AddEventHandler(handler)
	return informers
}

// ListAllPods returns all pods in watcher namespace.
func (w *Watcher) ListAllPods() ([]*corev1.Pod, error) {
	podLister := w.factory.Core().V1().Pods().Lister()
	retPods, err := podLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	if len(w.namespaces) > 1 {
		retPods = filterPodByNamespaces(retPods, w.namespaces)
	}
	return retPods, nil
}

// ListAllDeployments returns all deployments in watcher namespace.
func (w *Watcher) ListAllDeployments() ([]*appsv1.Deployment, error) {
	deploymentLister := w.factory.Apps().V1().Deployments().Lister()
	return deploymentLister.List(labels.Everything())
}

// ListAllService returns all services in watcher namespace.
func (w *Watcher) ListAllService() ([]*corev1.Service, error) {
	serviceLister := w.factory.Core().V1().Services().Lister()
	return serviceLister.List(labels.Everything())
}

// ListAllIngresses returns all ingresses in watcher namespace.
func (w *Watcher) ListAllIngresses() ([]*v1beta1.Ingress, error) {
	ingressLister := w.factory.Extensions().V1beta1().Ingresses().Lister()
	return ingressLister.List(labels.Everything())
}
