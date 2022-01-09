package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/controller/deployment/util"
)

// ReplicaSetState .
type ReplicaSetState struct {
	Expect  int32  `json:"expect"`
	Actual  int32  `json:"actual"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

// ReplicaSetWatcher .
type ReplicaSetWatcher struct {
	client          clientset.Interface
	rsLister        appslisters.ReplicaSetLister
	dLister         appslisters.DeploymentLister
	podLister       corelisters.PodLister
	rsListerSynced  cache.InformerSynced
	dListerSynced   cache.InformerSynced
	podListerSynced cache.InformerSynced
	queue           workqueue.RateLimitingInterface
}

// NewReplicaSetWatcher .
func NewReplicaSetWatcher(client clientset.Interface, rsInformer appsinformers.ReplicaSetInformer, dInformer appsinformers.DeploymentInformer, podInformer coreinformers.PodInformer) *ReplicaSetWatcher {
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "replicaset-watcher-queue")
	watcher := &ReplicaSetWatcher{
		client:          client,
		rsLister:        rsInformer.Lister(),
		dLister:         dInformer.Lister(),
		podLister:       podInformer.Lister(),
		rsListerSynced:  rsInformer.Informer().HasSynced,
		dListerSynced:   dInformer.Informer().HasSynced,
		podListerSynced: podInformer.Informer().HasSynced,
		queue:           queue,
	}

	rsInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    watcher.addReplicaSet,
		UpdateFunc: watcher.updateReplicaSet,
		DeleteFunc: watcher.deleteReplicaSet,
	})
	return watcher
}

// Run .
func (w *ReplicaSetWatcher) Run(ctx context.Context) {
	defer func() {
		log.Println("replicaset watcher shutdown")
		w.queue.ShutDown()
	}()

	if !cache.WaitForNamedCacheSync("replicaset-cache-sync-test", ctx.Done(), w.rsListerSynced, w.dListerSynced, w.podListerSynced) {
		log.Println("rs/deployment cache is not sync, and exit")
		return
	}

	log.Println("replicaset watcher start")
	go w.work(ctx)
	<-ctx.Done()
}

func (w *ReplicaSetWatcher) work(ctx context.Context) {
	for w.processNextItem(ctx) {
		select {
		case <-ctx.Done():
			log.Println("replicaset watcher exit:", ctx.Err())
			return
		default:
		}
	}
}

func (w *ReplicaSetWatcher) processNextItem(ctx context.Context) bool {
	key, quit := w.queue.Get()
	if quit {
		return false
	}
	// notify finish handle key
	defer w.queue.Done(key)

	err := w.syncHandler(ctx, key.(string))
	w.handleErr(err, key)
	return true
}

func (w *ReplicaSetWatcher) handleErr(err error, key interface{}) {
	if err == nil {
		w.queue.Forget(key)
		return
	}

	log.Printf("process workitem (replicaset) error: %v\n", err)
	if w.queue.NumRequeues(key) > MaxRetries {
		log.Printf("exceed max retries [%d], and drop replicaset: %s\n", MaxRetries, key)
		w.queue.Forget(key)
	} else {
		log.Printf("re enqueue replicaset: %v\n", key)
		w.queue.AddRateLimited(key)
	}
}

func (w *ReplicaSetWatcher) syncHandler(ctx context.Context, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.Println("failed to split meta namespace cache key:", key)
		return err
	}

	rs, err := w.rsLister.ReplicaSets(namespace).Get(name)
	if errors.IsNotFound(err) {
		return fmt.Errorf("replicaset [%s/%s] has been deleted", namespace, name)
	}
	if err != nil {
		return err
	}

	deployments := w.getDeploymentsForReplicaSet(rs)
	if deployments == nil {
		return fmt.Errorf("no matched deployment found")
	}

	deploy := deployments[0]
	var (
		progressCond  apps.DeploymentCondition
		availableCond apps.DeploymentCondition
	)
	for _, condition := range deploy.Status.Conditions {
		if condition.Type == "Progressing" {
			progressCond = condition
		} else if condition.Type == "Available" {
			availableCond = condition
		} else {
			return fmt.Errorf("unhandle deployment condition: %v", condition.Type)
		}
	}

	dName := deploy.GetObjectMeta().GetName()
	expect := deploy.Status.Replicas
	actual := deploy.Status.ReadyReplicas
	// progress: NewReplicaSetCreated -> ReplicaSetUpdated -> NewReplicaSetAvailable
	if progressCond.Reason == "NewReplicaSetCreated" {
		return nil
	}
	if progressCond.Reason == "ReplicaSetUpdated" {
		log.Printf("wait for deploy [%s/%s] ready: %d/%d", namespace, dName, actual, expect)
		return nil
	}
	if availableCond.Status == "True" {
		log.Printf("deploy [%s/%s] is available: %d/%d", namespace, dName, actual, expect)
		return nil
	}

	pods, err := w.getPodsForDeployment(deploy, rs)
	if err != nil {
		return err
	}
	podNamesList := make([]string, 0, len(pods))
	for _, pod := range pods {
		podNamesList = append(podNamesList, pod.GetObjectMeta().GetName())
	}
	podNames := strings.Join(podNamesList, ",")
	log.Printf("deploy is not as desired: ns:%s, deployment:%s, rs:%s, pods:[%s]", namespace, dName, name, podNames)

	state := ReplicaSetState{
		Expect:  expect,
		Actual:  actual,
		Reason:  availableCond.Reason,
		Message: availableCond.Message,
	}
	b, err := json.MarshalIndent(&state, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	fmt.Println()
	return nil
}

func (w *ReplicaSetWatcher) getDeploymentsForReplicaSet(rs *apps.ReplicaSet) []*apps.Deployment {
	deployments, err := util.GetDeploymentsForReplicaSet(w.dLister, rs)
	if err != nil || len(deployments) == 0 {
		return nil
	}

	if len(deployments) > 1 {
		log.Printf("more than one deployment is selecting replica set [%s]", rs.GetLabels())
	}
	return deployments
}

func (w *ReplicaSetWatcher) getPodsForDeployment(d *apps.Deployment, rs *apps.ReplicaSet) ([]*core.Pod, error) {
	selector, err := metav1.LabelSelectorAsSelector(d.Spec.Selector)
	if err != nil {
		return nil, err
	}

	pods, err := w.podLister.Pods(d.Namespace).List(selector)
	if err != nil {
		return nil, err
	}

	retPods := []*core.Pod{}
	for _, pod := range pods {
		controllerRef := metav1.GetControllerOf(pod)
		if controllerRef == nil {
			continue
		}
		if rs.UID == controllerRef.UID {
			retPods = append(retPods, pod)
		}
	}
	return retPods, nil
}

//
// ReplicaSet Informer Callbacks
//

func (w *ReplicaSetWatcher) addReplicaSet(obj interface{}) {
	rs := obj.(*apps.ReplicaSet)
	log.Println("[cb] add replicaset:", rs.GetObjectMeta().GetName())
}

func (w *ReplicaSetWatcher) updateReplicaSet(old, new interface{}) {
	newRs := new.(*apps.ReplicaSet)
	log.Println("[cb] update replicaset:", newRs.GetObjectMeta().GetName())
	w.enqueue(newRs)
}

func (w *ReplicaSetWatcher) deleteReplicaSet(obj interface{}) {
	rs, ok := obj.(*apps.ReplicaSet)
	if !ok {
		log.Println("[deleteReplicaSet] convert replicaset object failed.")
		return
	}
	log.Println("[cb] delete replicaset:", rs.GetObjectMeta().GetName())
}

func (w *ReplicaSetWatcher) enqueue(rs *apps.ReplicaSet) {
	key, err := controller.KeyFunc(rs)
	if err != nil {
		log.Printf("couldn't get key for object %#v: %v\n", rs, err)
		return
	}
	w.queue.Add(key)
}
