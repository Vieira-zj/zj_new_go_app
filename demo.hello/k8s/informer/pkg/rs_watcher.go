package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	apps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	clientset "k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1"
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
	client         clientset.Interface
	rsList         appslisters.ReplicaSetLister
	dList          appslisters.DeploymentLister
	rsListerSynced cache.InformerSynced
	dListerSynced  cache.InformerSynced
	queue          workqueue.RateLimitingInterface
}

// NewReplicaSetWatcher .
func NewReplicaSetWatcher(client clientset.Interface, rsInformer appsinformers.ReplicaSetInformer, dInformer appsinformers.DeploymentInformer) *ReplicaSetWatcher {
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "replicaset-watcher-queue")
	watcher := &ReplicaSetWatcher{
		client:         client,
		rsList:         rsInformer.Lister(),
		dList:          dInformer.Lister(),
		rsListerSynced: rsInformer.Informer().HasSynced,
		dListerSynced:  dInformer.Informer().HasSynced,
		queue:          queue,
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

	if !cache.WaitForNamedCacheSync("replicaset-cache-sync-test", ctx.Done(), w.rsListerSynced, w.dListerSynced) {
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
	// key: namespace/rsId
	key, quit := w.queue.Get()
	if quit {
		return false
	}
	defer w.queue.Done(key)

	if err := w.syncHandler(ctx, key.(string)); err != nil {
		log.Println("process work item (replicaset) error:", err)
		if w.queue.NumRequeues(key) > MaxRetries {
			log.Printf("exceed max retries [%d], and forget rs key: %s\n", MaxRetries, key)
			w.queue.Forget(key)
		} else {
			w.queue.AddRateLimited(key)
		}
	}
	return true
}

func (w *ReplicaSetWatcher) syncHandler(ctx context.Context, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		log.Println("failed to split meta namespace cache key:", key)
		return err
	}

	rs, err := w.rsList.ReplicaSets(namespace).Get(name)
	if errors.IsNotFound(err) {
		return fmt.Errorf("replicaset [%s/%s] has been deleted", namespace, name)
	}
	if err != nil {
		return err
	}

	deployments := w.getDeploymentsForReplicaSet(rs)
	if deployments == nil {
		return fmt.Errorf("not matched deployment found")
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
	log.Printf("deploy is not as desired: ns:%s, deployment:%s, rs:%s", namespace, dName, name)
	fmt.Println(string(b))
	fmt.Println()
	return nil
}

func (w *ReplicaSetWatcher) getDeploymentsForReplicaSet(rs *apps.ReplicaSet) []*apps.Deployment {
	deployments, err := util.GetDeploymentsForReplicaSet(w.dList, rs)
	if err != nil || len(deployments) == 0 {
		return nil
	}

	if len(deployments) > 1 {
		log.Printf("more than one deployment is selecting replica set [%s]", rs.GetLabels())
	}
	return deployments
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
