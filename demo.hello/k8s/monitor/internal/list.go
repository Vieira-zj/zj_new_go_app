package internal

import (
	"context"
	"fmt"

	k8spkg "demo.hello/k8s/client/pkg"
	corev1 "k8s.io/api/core/v1"
)

// PodStatus .
type PodStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Log     string `json:"log,omitempty"`
}

// GetAllPodInfos returns all pods status info.
func GetAllPodInfos(ctx context.Context, resource *k8spkg.Resource, namespace string) ([]*PodStatus, error) {
	var (
		pods []corev1.Pod
		err  error
	)
	if len(namespace) == 0 {
		pods, err = resource.GetAllPods(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		pods, err = resource.GetPodsByNamespace(ctx, namespace)
		if err != nil {
			return nil, err
		}
	}

	podInfos := make([]*PodStatus, 0, len(pods))
	for _, pod := range pods {
		podName := pod.GetObjectMeta().GetName()
		podInfo := &PodStatus{
			Name: podName,
		}
		state, err := resource.GetPodState(ctx, namespace, podName, "")
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
			if log, err := resource.GetPodLogs(ctx, namespace, podName); err != nil {
				fmt.Printf("get pod [%s/%s] state failed: %s\n", namespace, podName, err.Error())
			} else {
				podInfo.Log = log
			}
		}
		podInfos = append(podInfos, podInfo)
	}
	return podInfos, nil
}
