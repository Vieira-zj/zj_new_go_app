package pkg

import (
	"context"
	"encoding/json"
	"fmt"

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
		queue:           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "pod"),
	}

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    watcher.addPod,
		UpdateFunc: watcher.updatePod,
		DeleteFunc: watcher.deletePod,
	})
	return watcher
}

// Run .
func (watcher *PodWatcher) Run(ctx context.Context) {
	defer func() {
		watcher.queue.ShutDown()
		fmt.Println("pod watcher shutdown")
	}()
	fmt.Println("pod watcher started")

	if !cache.WaitForNamedCacheSync("watcher-pod-test", ctx.Done(), watcher.podListerSynced) {
		fmt.Println("pod cache is not sync, and exit")
		return
	}

	go watcher.work(ctx)
	<-ctx.Done()
}

func (watcher *PodWatcher) work(ctx context.Context) {
	for watcher.processNextItem(ctx) {
		select {
		case <-ctx.Done():
			fmt.Println("watcher exit:", ctx.Err())
			return
		default:
		}
	}
}

func (watcher *PodWatcher) processNextItem(ctx context.Context) bool {
	key, quit := watcher.queue.Get()
	if quit {
		return false
	}
	defer watcher.queue.Done(key)

	if err := watcher.syncHandler(ctx, key.(string)); err != nil {
		fmt.Println("process work item (pod) error:", err)
		watcher.queue.Forget(key)
		return false
	}
	return true
}

func (watcher *PodWatcher) syncHandler(ctx context.Context, key string) error {
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		fmt.Println("Failed to split meta namespace cache key:", err)
		return err
	}

	pod, err := watcher.podLister.Pods(ns).Get(name)
	if errors.IsNotFound(err) {
		return fmt.Errorf("pod [%s/%s] has been deleted", ns, name)
	}
	if err != nil {
		return err
	}

	fmt.Printf("pod:%s, labels:%+v, phase:%s\n", name, pod.Labels, pod.Status.Phase)
	if pod.Status.Phase == "Pending" {
		return nil
	}

	for _, container := range pod.Status.ContainerStatuses {
		if !container.Ready {
			b, err := json.MarshalIndent(&container, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(b))
		}
	}
	fmt.Println()
	return nil
}

//
// Pod Informer Callbacks
//

func (watcher *PodWatcher) addPod(obj interface{}) {
	pod := obj.(*corev1.Pod)
	fmt.Println("[cb] add pod:", pod.GetObjectMeta().GetName())
}

func (watcher *PodWatcher) updatePod(old, new interface{}) {
	newPod := new.(*corev1.Pod)
	fmt.Println("[cb] update pod:", newPod.GetObjectMeta().GetName())
	watcher.enqueue(newPod)
}

func (watcher *PodWatcher) deletePod(obj interface{}) {
	pod := obj.(*corev1.Pod)
	fmt.Println("[cb] delete pod:", pod.GetObjectMeta().GetName())
}

func (watcher *PodWatcher) enqueue(pod *corev1.Pod) {
	key, err := controller.KeyFunc(pod)
	if err != nil {
		fmt.Printf("couldn't get key for object %#v: %v\n", pod, err)
		return
	}
	watcher.queue.Add(key)
}
