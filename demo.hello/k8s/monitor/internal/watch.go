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
	namespace       string
	resource        *k8spkg.Resource
	factory         informers.SharedInformerFactory
	ErrorPodStateCh chan *k8spkg.PodState
}

// NewWatcher creates a list watcher instance.
func NewWatcher(client *kubernetes.Clientset, namespace string, interval uint) *Watcher {
	factory := informers.NewSharedInformerFactoryWithOptions(
		client, time.Duration(interval)*time.Second, informers.WithNamespace(namespace))
	resource := k8spkg.NewResource(client)
	return &Watcher{
		namespace:       namespace,
		resource:        resource,
		factory:         factory,
		ErrorPodStateCh: make(chan *k8spkg.PodState, interval),
	}
}

// Run exec watcher informer.
func (w *Watcher) Run(ctx context.Context, isWatchDeploy bool) {
	log.Println("watcher started")
	go w.podInformer().Run(ctx.Done())

	if isWatchDeploy {
		go w.deploymentInformer().Run(ctx.Done())
	}
}

func (w *Watcher) podInformer() cache.SharedIndexInformer {
	onAdd := func(obj interface{}) {
		if pod, ok := obj.(*corev1.Pod); ok {
			log.Println("onAdd pod:", pod.ObjectMeta.GetName())
		}
	}

	onUpdate := func(oldObj, newObj interface{}) {
		if pod, ok := newObj.(*corev1.Pod); ok {
			log.Println("onUpdate pod:", pod.ObjectMeta.GetName())
			state, err := w.resource.GetPodStateRaw(pod, "")
			if err != nil {
				log.Printf("get pod [%s/%s] state error: %v\n",
					pod.GetObjectMeta().GetNamespace(), pod.GetObjectMeta().GetName(), err)
				k8spkg.PrintPodInfo(pod)
				return
			}
			if state.Value != "Running" {
				w.ErrorPodStateCh <- state
			}
		}
	}

	onDelete := func(obj interface{}) {
		if pod, ok := obj.(*corev1.Pod); ok {
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
				log.Println("onAdd deployment:", deploy.ObjectMeta.GetName())
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if deploy, ok := newObj.(*appsv1.Deployment); ok {
				log.Println("onUpdate deployment:", deploy.ObjectMeta.GetName())
			}
		},
		DeleteFunc: func(obj interface{}) {
			if deploy, ok := obj.(*appsv1.Deployment); ok {
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
	return podLister.List(labels.Everything())
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
