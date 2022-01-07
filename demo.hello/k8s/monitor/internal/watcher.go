package internal

import (
	"context"
	"errors"
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
	isDebug          bool
	namespaces       map[string]struct{}
	resource         *k8spkg.Resource
	factory          informers.SharedInformerFactory
	ErrorPodStatusCh chan *PodStatus
}

// NewWatcher creates a list watcher instance for given namespaces.
func NewWatcher(client *kubernetes.Clientset, namespaces []string, interval uint, isDebug bool) *Watcher {
	if len(namespaces) == 0 {
		panic("init watcher, namespace cannot be empty")
	}

	namespacesMap := make(map[string]struct{}, len(namespaces))
	for _, ns := range namespaces {
		namespacesMap[ns] = struct{}{}
	}

	var factory informers.SharedInformerFactory
	if len(namespaces) == 1 {
		// monitor given namespaces
		factory = informers.NewSharedInformerFactoryWithOptions(
			client, time.Duration(interval)*time.Second, informers.WithNamespace(namespaces[0]))
	} else {
		// monitor all namespaces
		factory = informers.NewSharedInformerFactoryWithOptions(client, time.Duration(interval)*time.Second)
	}
	return &Watcher{
		isDebug:          isDebug,
		namespaces:       namespacesMap,
		resource:         k8spkg.NewResource(client),
		factory:          factory,
		ErrorPodStatusCh: make(chan *PodStatus, interval),
	}
}

// Run starts watcher informer.
func (w *Watcher) Run(ctx context.Context, isWatchDeploy bool) error {
	w.logPrintln("start watcher")
	informer := w.podInformer(ctx)
	if isWatchDeploy {
		informer = w.deploymentInformer(ctx)
	}

	w.factory.Start(ctx.Done())
	if !cache.WaitForNamedCacheSync("pod-monitor-app", ctx.Done(), informer.HasSynced) {
		return errors.New("informer cache sync failed")
	}
	// use factory.Start() instead of informer.Run()
	// go informer.Run(ctx.Done())
	return nil
}

func (w *Watcher) podInformer(ctx context.Context) cache.SharedIndexInformer {
	onAdd := func(obj interface{}) {
		if pod, ok := obj.(*corev1.Pod); ok {
			if len(w.namespaces) > 1 {
				if _, ok := w.namespaces[pod.GetObjectMeta().GetNamespace()]; !ok {
					return
				}
			}
			w.logPrintln("onAdd pod:", pod.ObjectMeta.GetName())
		}
	}

	onUpdate := func(oldObj, newObj interface{}) {
		if pod, ok := newObj.(*corev1.Pod); ok {
			if len(w.namespaces) > 1 {
				if _, ok := w.namespaces[pod.GetObjectMeta().GetNamespace()]; !ok {
					return
				}
			}

			w.logPrintln("onUpdate pod:", pod.ObjectMeta.GetName())
			localCtx, cancel := context.WithTimeout(ctx, time.Duration(3)*time.Second)
			defer cancel()
			status := GetPodStatus(localCtx, w.resource, pod, "")
			if !isPodOK(status.Value) {
				if len(w.ErrorPodStatusCh) == cap(w.ErrorPodStatusCh) {
					w.logPrintf("error pod status channel is full (size=%d)", len(w.ErrorPodStatusCh))
				}
				w.ErrorPodStatusCh <- status
			}
		}
	}

	onDelete := func(obj interface{}) {
		if pod, ok := obj.(*corev1.Pod); ok {
			if len(w.namespaces) > 1 {
				if _, ok := w.namespaces[pod.GetObjectMeta().GetNamespace()]; !ok {
					return
				}
			}
			w.logPrintln("onDelete pod:", pod.ObjectMeta.GetName())
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

func (w *Watcher) deploymentInformer(ctx context.Context) cache.SharedIndexInformer {
	handler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if deploy, ok := obj.(*appsv1.Deployment); ok {
				if len(w.namespaces) > 1 {
					if _, ok := w.namespaces[deploy.GetObjectMeta().GetNamespace()]; !ok {
						return
					}
				}
				w.logPrintln("onAdd deployment:", deploy.ObjectMeta.GetName())
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			if deploy, ok := newObj.(*appsv1.Deployment); ok {
				if len(w.namespaces) > 1 {
					if _, ok := w.namespaces[deploy.GetObjectMeta().GetNamespace()]; !ok {
						return
					}
				}
				w.logPrintln("onUpdate deployment:", deploy.ObjectMeta.GetName())
			}
		},
		DeleteFunc: func(obj interface{}) {
			if deploy, ok := obj.(*appsv1.Deployment); ok {
				if len(w.namespaces) > 1 {
					if _, ok := w.namespaces[deploy.GetObjectMeta().GetNamespace()]; !ok {
						return
					}
				}
				w.logPrintln("onDelete deployment:", deploy.ObjectMeta.GetName())
			}
		},
	}

	informers := w.factory.Apps().V1().Deployments().Informer()
	informers.AddEventHandler(handler)
	return informers
}

// ListAllPods returns all pods in watcher namespaces.
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

// ListAllDeployments returns all deployments in watcher namespaces.
func (w *Watcher) ListAllDeployments() ([]*appsv1.Deployment, error) {
	deploymentLister := w.factory.Apps().V1().Deployments().Lister()
	return deploymentLister.List(labels.Everything())
}

// ListAllService returns all services in watcher namespaces.
func (w *Watcher) ListAllService() ([]*corev1.Service, error) {
	serviceLister := w.factory.Core().V1().Services().Lister()
	return serviceLister.List(labels.Everything())
}

// ListAllIngresses returns all ingresses in watcher namespaces.
func (w *Watcher) ListAllIngresses() ([]*v1beta1.Ingress, error) {
	ingressLister := w.factory.Extensions().V1beta1().Ingresses().Lister()
	return ingressLister.List(labels.Everything())
}

func (w *Watcher) logPrintln(v ...interface{}) {
	if w.isDebug {
		log.Println(v...)
	}
}

func (w *Watcher) logPrintf(foramt string, v ...interface{}) {
	if w.isDebug {
		log.Printf(foramt, v...)
	}
}
