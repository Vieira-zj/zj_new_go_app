package internal

import (
	"context"
	"fmt"
	"log"
	"strings"

	k8spkg "demo.hello/k8s/client/pkg"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// ValidStatus valid status of pod.
var ValidStatus map[string]struct{}

func init() {
	ValidStatus = make(map[string]struct{}, 4)
	for _, status := range [4]string{"", "Running", "ContainerCreating", "Completed"} {
		ValidStatus[status] = struct{}{}
	}
}

// PodStatus .
type PodStatus struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Value     string `json:"value"`
	ExitCode  int32  `json:"exitcode,omitempty"`
	Message   string `json:"message,omitempty"`
	Log       string `json:"log,omitempty"`
}

// Lister .
type Lister struct {
	namespaces map[string]struct{}
	resource   *k8spkg.Resource
	watcher    *Watcher
}

// NewLister creates a Lister by given namespaces.
func NewLister(client *kubernetes.Clientset, watcher *Watcher, namespaces []string) *Lister {
	if len(namespaces) == 0 {
		panic("init lister, namespace cannot be empty")
	}

	namespacesMap := make(map[string]struct{}, len(namespaces))
	for _, ns := range namespaces {
		namespacesMap[ns] = struct{}{}
	}
	return &Lister{
		namespaces: namespacesMap,
		resource:   k8spkg.NewResource(client),
		watcher:    watcher,
	}
}

// GetAllPodInfosByRaw returns all pods status info by raw api.
func (lister *Lister) GetAllPodInfosByRaw(ctx context.Context) ([]*PodStatus, error) {
	var pods []*corev1.Pod
	if len(lister.namespaces) == 1 {
		var namespace string
		for ns := range lister.namespaces {
			namespace = ns
		}
		localPods, err := lister.resource.GetPodsByNamespace(ctx, namespace)
		if err != nil {
			return nil, err
		}
		pods = getPodRefs(localPods)
	} else {
		allPods, err := lister.resource.GetAllPods(ctx)
		log.Printf("all pods count: %d\n", len(allPods))
		if err != nil {
			return nil, err
		}
		pods = filterPodByNamespaces(getPodRefs(allPods), lister.namespaces)
	}
	return lister.GetAllPodInfos(ctx, pods)
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
		podInfo := GetPodStatus(ctx, lister.resource, pod, "")
		podInfos = append(podInfos, podInfo)
	}
	return podInfos, nil
}

// GetPodStatus returns the given pod status.
func GetPodStatus(ctx context.Context, resource *k8spkg.Resource, pod *corev1.Pod, containerName string) *PodStatus {
	podName := pod.GetObjectMeta().GetName()
	namespace := pod.GetObjectMeta().GetNamespace()
	podInfo := &PodStatus{
		Namespace: namespace,
		Name:      podName,
	}

	state, err := resource.GetPodStateRaw(pod, containerName)
	if err != nil {
		podInfo.Message = fmt.Sprintf("get pod [%s/%s] state failed: %s\n", namespace, podName, err.Error())
		return podInfo
	}

	podInfo.Value = state.Value
	if state.ExitCode != 0 {
		podInfo.ExitCode = state.ExitCode
	}
	if len(state.Message) > 0 {
		podInfo.Message = state.Message
	}

	if !isPodOK(podInfo.Value) {
		if podLog, err := resource.GetPodLogs(ctx, namespace, podName); err != nil {
			log.Printf("get pod [%s/%s] log failed: %s\n", namespace, podName, err.Error())
		} else {
			podInfo.Log = podLog
		}
	}
	return podInfo
}

func filterPodByNamespaces(pods []*corev1.Pod, namespaces map[string]struct{}) []*corev1.Pod {
	retPods := make([]*corev1.Pod, 0, len(pods))
	for _, pod := range pods {
		if _, ok := namespaces[pod.GetObjectMeta().GetNamespace()]; ok {
			retPods = append(retPods, pod)
		}
	}
	return retPods
}

func getPodRefs(pods []corev1.Pod) []*corev1.Pod {
	podRefs := make([]*corev1.Pod, 0, len(pods))
	for _, pod := range pods {
		localpod := pod // save to local
		podRefs = append(podRefs, &localpod)
	}
	return podRefs
}

func isPodOK(status string) bool {
	reason := status
	if strings.Index(status, "/") != -1 {
		reason = strings.Split(status, "/")[1]
	}
	_, ok := ValidStatus[reason]
	return ok
}
