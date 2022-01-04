package pkg

import (
	"context"
	"encoding/json"
	"fmt"

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
	Namespace string
	Name      string
	Expect    int32
	Ready     int32
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
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "replicaset")

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
		w.queue.ShutDown()
		fmt.Println("replicaset watcher shutdown")
	}()
	fmt.Println("replicaset watcher started")

	if !cache.WaitForNamedCacheSync("watch-replicaset-test", ctx.Done(), w.rsListerSynced, w.dListerSynced) {
		fmt.Println("rs/deployment cache is not sync, and exit")
		return
	}

	go w.work(ctx)
	<-ctx.Done()
}

func (w *ReplicaSetWatcher) work(ctx context.Context) {
	for w.processNextItem(ctx) {
		select {
		case <-ctx.Done():
			fmt.Println("watcher exit:", ctx.Err())
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
		fmt.Println("process work item (rs) error:", err)
		w.queue.Forget(key)
		return false
	}
	return true
}

func (w *ReplicaSetWatcher) syncHandler(ctx context.Context, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		fmt.Println("failed to split meta namespace cache key:", key)
		return err
	}

	rs, err := w.rsList.ReplicaSets(namespace).Get(name)
	if errors.IsNotFound(err) {
		return fmt.Errorf("replicaset [%s/%s] has been deleted", namespace, name)
	}
	if err != nil {
		return err
	}

	if rs.Status.Replicas != rs.Status.ReadyReplicas {
		state := ReplicaSetState{
			Namespace: namespace,
			Name:      name,
			Expect:    rs.Status.Replicas,
			Ready:     rs.Status.ReadyReplicas,
		}
		b, err := json.MarshalIndent(&state, "", "  ")
		if err != nil {
			return err
		}
		fmt.Printf("replicaset is not as desired:\n%s\n\n", b)
	}
	return nil
}

//
// ReplicaSet Informer Callbacks
//

func (w *ReplicaSetWatcher) addReplicaSet(obj interface{}) {
	rs := obj.(*apps.ReplicaSet)
	fmt.Println("[cb] add replicaset:", rs.GetObjectMeta().GetName())
}

func (w *ReplicaSetWatcher) updateReplicaSet(old, new interface{}) {
	newRs := new.(*apps.ReplicaSet)
	fmt.Println("[cb] update replicaset:", newRs.GetObjectMeta().GetName())

	deployments := w.getDeploymentsForReplicaSet(newRs)
	for _, deploy := range deployments {
		fmt.Println("[cb] related deployment:", deploy.GetObjectMeta().GetName())
	}
	w.enqueue(newRs)
}

func (w *ReplicaSetWatcher) deleteReplicaSet(obj interface{}) {
	rs, ok := obj.(*apps.ReplicaSet)
	if !ok {
		fmt.Println("[deleteReplicaSet] convert replicaset object failed.")
		return
	}
	fmt.Println("[cb] delete replicaset:", rs.GetObjectMeta().GetName())
}

func (w *ReplicaSetWatcher) enqueue(rs *apps.ReplicaSet) {
	key, err := controller.KeyFunc(rs)
	if err != nil {
		fmt.Printf("couldn't get key for object %#v: %v\n", rs, err)
		return
	}
	w.queue.Add(key)
}

func (w *ReplicaSetWatcher) getDeploymentsForReplicaSet(rs *apps.ReplicaSet) []*apps.Deployment {
	deployments, err := util.GetDeploymentsForReplicaSet(w.dList, rs)
	if len(deployments) == 0 {
		if err != nil {
			fmt.Println("GetDeploymentsForReplicaSet error:", err)
		}
		return nil
	}

	if len(deployments) > 1 {
		fmt.Printf("more than one deployment is selecting replica set [%s]", rs.GetLabels())
	}
	return deployments
}
