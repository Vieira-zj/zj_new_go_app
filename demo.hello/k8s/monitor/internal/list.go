package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	k8spkg "demo.hello/k8s/client/pkg"
	v1 "k8s.io/api/core/v1"
)

// PodStatus contains pod name, status, readiness and message.
type PodStatus struct {
	PodName   string          `json:"name"`
	Status    string          `json:"status"`
	Readiness bool            `json:"readiness"`
	Message   json.RawMessage `json:"message"`
}

// GetAllPodInfos returns all pod status.
func GetAllPodInfos(ctx context.Context, resource *k8spkg.Resource, namespace string) ([]*PodStatus, error) {
	var (
		pods []v1.Pod
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
		podInfo := &PodStatus{
			PodName: pod.ObjectMeta.Name,
		}

		var b []byte
		containerStatus := pod.Status.ContainerStatuses[0]
		if containerStatus.State.Running != nil {
			podInfo.Status = "Running"
			b, err = json.Marshal(containerStatus.State.Running)
			if err != nil {
				return nil, err
			}
		} else if containerStatus.State.Terminated != nil {
			podInfo.Status = "Terminated"
			b, err = json.Marshal(containerStatus.State.Terminated)
			if err != nil {
				return nil, err
			}
		} else if containerStatus.State.Waiting != nil {
			podInfo.Status = "Waiting"
			b, err = json.Marshal(containerStatus.State.Waiting)
			if err != nil {
				return nil, err
			}
		}
		podInfo.Message = b

		if pod.Spec.Containers[0].ReadinessProbe != nil {
			podIP := pod.Status.PodIP
			podPort := pod.Spec.Containers[0].ReadinessProbe.TCPSocket.Port.String()
			result, err := pingSocket(podIP, podPort)
			if err != nil {
				fmt.Println(err.Error())
			}
			podInfo.Readiness = result
		}
		podInfos = append(podInfos, podInfo)
	}

	return podInfos, nil
}

func pingSocket(host string, port string) (bool, error) {
	addr := net.JoinHostPort(host, port)
	timeout := time.Duration(3) * time.Second
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return false, err
	}

	if conn != nil {
		defer conn.Close()
		return true, nil
	}
	return false, fmt.Errorf("open tcp connection failed: %s:%s", host, port)
}
