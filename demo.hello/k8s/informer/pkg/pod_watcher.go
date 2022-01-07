package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/kubernetes/pkg/controller"
)

// PodWatcher .
type PodWatcher struct {
	client          clientset.Interface
	podLister       corelisters.PodLister
	podListerSynced cache.InformerSynced
	queue           workqueue.RateLimitingInterface
}

// NewPodWatcher .
func NewPodWatcher(client clientset.Interface, podInformer coreinformers.PodInformer) *PodWatcher {
	watcher := &PodWatcher{
		client:          client,
		podLister:       podInformer.Lister(),
		podListerSynced: podInformer.Informer().HasSynced,
		queue:           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "pod-watcher-queue"),
	}

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    watcher.addPod,
		UpdateFunc: watcher.updatePod,
		DeleteFunc: watcher.deletePod,
	})
	return watcher
}

// Run .
func (w *PodWatcher) Run(ctx context.Context) {
	defer func() {
		w.queue.ShutDown()
		log.Println("pod watcher shutdown")
	}()

	if !cache.WaitForNamedCacheSync("pod-cache-sync-test", ctx.Done(), w.podListerSynced) {
		log.Println("pod cache is not sync, and exit")
		return
	}

	log.Println("pod watcher start")
	go w.work(ctx)
	<-ctx.Done()
}

func (w *PodWatcher) work(ctx context.Context) {
	for w.processNextItem(ctx) {
		select {
		case <-ctx.Done():
			log.Println("pod watcher exit:", ctx.Err())
			return
		default:
		}
	}
}

func (w *PodWatcher) processNextItem(ctx context.Context) bool {
	key, quit := w.queue.Get()
	if quit {
		return false
	}
	defer w.queue.Done(key)

	if err := w.syncHandler(ctx, key.(string)); err != nil {
		log.Println("process work item (pod) error:", err)
		if w.queue.NumRequeues(key) > MaxRetries {
			log.Printf("exceed max retries [%d], and forget pod key: %s\n", MaxRetries, key)
			w.queue.Forget(key)
		} else {
			w.queue.AddRateLimited(key)
		}
	}
	return true
}

func (w *PodWatcher) syncHandler(ctx context.Context, key string) error {
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.Println("failed to split meta namespace cache key:", err)
		return err
	}

	pod, err := w.podLister.Pods(ns).Get(name)
	if errors.IsNotFound(err) {
		return fmt.Errorf("pod [%s/%s] has been deleted", ns, name)
	}
	if err != nil {
		return err
	}

	if pod.Status.Phase == "Pending" {
		return nil
	}

	containersInfo := make([]string, 0)
	for _, container := range pod.Status.ContainerStatuses {
		if !container.Ready {
			b, err := json.MarshalIndent(&container, "", "  ")
			if err != nil {
				return err
			}
			containersInfo = append(containersInfo, string(b))
		}
	}
	if len(containersInfo) > 0 {
		log.Printf("pod is not as desired: ns/name:%s/%s, labels:%+v, phase:%s\n", ns, name, pod.Labels, pod.Status.Phase)
		log.Println("container info:")
		fmt.Println(strings.Join(containersInfo, "\n"))
		fmt.Println()
	}
	return nil
}

//
// Pod Informer Callbacks
//

func (w *PodWatcher) addPod(obj interface{}) {
	pod := obj.(*corev1.Pod)
	log.Println("[cb] add pod:", pod.GetObjectMeta().GetName())
}

func (w *PodWatcher) updatePod(old, new interface{}) {
	newPod := new.(*corev1.Pod)
	log.Println("[cb] update pod:", newPod.GetObjectMeta().GetName())
	w.enqueue(newPod)
}

func (w *PodWatcher) deletePod(obj interface{}) {
	pod := obj.(*corev1.Pod)
	log.Println("[cb] delete pod:", pod.GetObjectMeta().GetName())
}

func (w *PodWatcher) enqueue(pod *corev1.Pod) {
	key, err := controller.KeyFunc(pod)
	if err != nil {
		log.Printf("couldn't get key for object %#v: %v\n", pod, err)
		return
	}
	w.queue.Add(key)
}
