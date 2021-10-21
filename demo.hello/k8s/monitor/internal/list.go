package internal

import (
	"context"
	"fmt"
	"log"

	k8spkg "demo.hello/k8s/client/pkg"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// PodStatus .
type PodStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Log     string `json:"log,omitempty"`
}

// Lister .
type Lister struct {
	namespace string
	resource  *k8spkg.Resource
	watcher   *Watcher
}

// NewLister creates a Lister by given namespace.
func NewLister(client *kubernetes.Clientset, watcher *Watcher, namespace string) *Lister {
	if len(namespace) == 0 {
		panic("init lister, namespace cannot be empty")
	}
	return &Lister{
		namespace: namespace,
		resource:  k8spkg.NewResource(client),
		watcher:   watcher,
	}
}

// GetAllPodInfosByRaw returns all pods status info by raw api.
func (lister *Lister) GetAllPodInfosByRaw(ctx context.Context) ([]*PodStatus, error) {
	pods, err := lister.resource.GetPodsByNamespace(ctx, lister.namespace)
	if err != nil {
		return nil, err
	}
	return lister.GetAllPodInfos(ctx, getPodRefs(pods))
}

// GetAllPodInfosByListWatch returns all pods status info by watcher.
func (lister *Lister) GetAllPodInfosByListWatch(ctx context.Context) ([]*PodStatus, error) {
	pods, err := lister.watcher.ListAllPods()
	if err != nil {
		return nil, err
	}
	return lister.GetAllPodInfos(ctx, pods)
}

// GetAllPodInfos returns all pods status info.
func (lister *Lister) GetAllPodInfos(ctx context.Context, pods []*corev1.Pod) ([]*PodStatus, error) {
	podInfos := make([]*PodStatus, 0, len(pods))
	for _, pod := range pods {
		podName := pod.GetObjectMeta().GetName()
		namespace := pod.GetObjectMeta().GetName()

		podInfo := &PodStatus{
			Name: podName,
		}
		state, err := lister.resource.GetPodStateRaw(pod, "")
		if err != nil {
			podInfo.Message = fmt.Sprintf("get pod [%s/%s] state failed: %s\n", namespace, podName, err.Error())
			podInfos = append(podInfos, podInfo)
			continue
		}

		podInfo.Status = state.Value
		if len(state.Message) > 0 {
			podInfo.Message = state.Message
		}

		if state.Value != "Running" {
			if podLog, err := lister.resource.GetPodLogs(ctx, namespace, podName); err != nil {
				log.Printf("get pod [%s/%s] state failed: %s\n", namespace, podName, err.Error())
			} else {
				podInfo.Log = podLog
			}
		}
		podInfos = append(podInfos, podInfo)
	}
	return podInfos, nil
}

func getPodRefs(pods []corev1.Pod) []*corev1.Pod {
	podRefs := make([]*corev1.Pod, 0, len(pods))
	for _, pod := range pods {
		localpod := pod // save to local
		podRefs = append(podRefs, &localpod)
	}
	return podRefs
}
