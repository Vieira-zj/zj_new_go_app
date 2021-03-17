package pkg

import (
	"context"
	"fmt"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

// Resource k8s resources object like namespace, pod and deploy.
type Resource struct {
	ctx    context.Context
	client *kubernetes.Clientset
}

var resource *Resource

// NewResource creates an instance of resource.
func NewResource(ctx context.Context, client *kubernetes.Clientset) *Resource {
	if resource != nil {
		return resource
	}
	resource = &Resource{
		ctx:    ctx,
		client: client,
	}
	return resource
}

/*
Namespace
*/

// CreateNamespace creates a namespace.
func (r *Resource) CreateNamespace(name string) (*apiv1.Namespace, error) {
	namespace := apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Status: apiv1.NamespaceStatus{
			Phase: apiv1.NamespaceActive,
		},
	}
	return r.client.CoreV1().Namespaces().Create(r.ctx, &namespace, metav1.CreateOptions{})
}

// GetNamespace returns a namespace by name.
func (r *Resource) GetNamespace(name string) (*apiv1.Namespace, error) {
	return r.client.CoreV1().Namespaces().Get(r.ctx, name, metav1.GetOptions{})
}

// DeleteNamespace delete a namespace by name.
func (r *Resource) DeleteNamespace(name string) error {
	deletePolicy := metav1.DeletePropagationForeground
	return r.client.CoreV1().Namespaces().Delete(r.ctx, name, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})
}

/*
Pod
*/

// GetAllPods returns all pods in k8s cluster.
func (r *Resource) GetAllPods() ([]apiv1.Pod, error) {
	return r.GetPodsByNamespace("")
}

// GetPodsByNamespace returns pods in specified namespace.
func (r *Resource) GetPodsByNamespace(namespace string) ([]apiv1.Pod, error) {
	pods, err := r.client.CoreV1().Pods(namespace).List(r.ctx, metav1.ListOptions{})
	if err != nil {
		return make([]apiv1.Pod, 0), err
	}
	return pods.Items, nil
}

// GetSpecifiedPod returns a specified pod by namespace and name.
func (r *Resource) GetSpecifiedPod(namespace string, podName string) (*apiv1.Pod, error) {
	pod, err := r.client.CoreV1().Pods(namespace).Get(r.ctx, podName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("Pod not found: namespace=%s, name=%s", namespace, podName)
		}
		return nil, err
	}
	return pod, nil
}

// GetPodNamespace returns a specified namespace by given pod.
func (r *Resource) GetPodNamespace(podName string) (string, error) {
	pods, err := r.GetAllPods()
	if err != nil {
		return "", err
	}

	for _, pod := range pods {
		if pod.ObjectMeta.Name == podName {
			return pod.ObjectMeta.Namespace, nil
		}
	}
	return "", fmt.Errorf("pod [%s] not found", podName)
}

/*
Get pods of a service
*/

// GetPodsByService returns pods in a specified services.
func (r *Resource) GetPodsByService(namespace, serviceName string) ([]apiv1.Pod, error) {
	service, err := r.GetSpecifiedService(namespace, serviceName)
	if err != nil {
		return nil, err
	}
	podSelector := service.Spec.Selector
	pods, err := r.GetPodsByLabels(namespace, podSelector)
	if err != nil {
		return nil, err
	}
	return pods, nil
}

// GetPodsByServiceV2 returns pods in a specified services.
func (r *Resource) GetPodsByServiceV2(namespace, serviceName string) ([]apiv1.Pod, error) {
	service, err := r.GetSpecifiedService(namespace, serviceName)
	if err != nil {
		return nil, err
	}

	podSelector := labels.Set(service.Spec.Selector)
	listOptions := metav1.ListOptions{LabelSelector: podSelector.AsSelector().String()}
	pods, err := r.client.CoreV1().Pods(namespace).List(r.ctx, listOptions)
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}

// GetSpecifiedService returns a specified service by namespace and name.
func (r *Resource) GetSpecifiedService(namespace, name string) (*apiv1.Service, error) {
	service, err := r.client.CoreV1().Services(namespace).Get(r.ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("Service not found: namespace=%s, name=%s", namespace, name)
		}
		return nil, err
	}
	return service, nil
}

// GetPodsByLabels returns pods with specified labels.
func (r *Resource) GetPodsByLabels(namespace string, labels map[string]string) ([]apiv1.Pod, error) {
	pods, err := r.GetPodsByNamespace(namespace)
	if err != nil {
		return nil, err
	}

	matchedPods := make([]apiv1.Pod, 0, len(pods))
	for _, pod := range pods {
		podLabels := pod.ObjectMeta.Labels
		isMatched := true
		// all pod labels are matched with selector labels
		for k, v := range labels {
			if podValue, ok := podLabels[k]; ok {
				if podValue != v {
					isMatched = false
				}
			} else {
				isMatched = false
			}
		}
		if isMatched {
			matchedPods = append(matchedPods, pod)
		}
	}
	return matchedPods, nil
}

/*
Printer
*/

// PrintPodInfo prints a pod information.
func PrintPodInfo(pod *apiv1.Pod) {
	maps := map[string]interface{}{
		"Name":        pod.ObjectMeta.Name,
		"Namespaces":  pod.ObjectMeta.Namespace,
		"NodeName":    pod.Spec.NodeName,
		"Annotations": pod.ObjectMeta.Annotations,
		"Labels":      pod.ObjectMeta.Labels,
		"Status":      pod.Status.Phase,
		"IP":          pod.Status.PodIP,
		"Image":       pod.Spec.Containers[0].Image,
	}
	prettyPrint(maps)
}

// PrintPodStatus pritns a pod status.
func PrintPodStatus(pod *apiv1.Pod) {
	conditions := make([]string, 0, len(pod.Status.Conditions))
	for _, cond := range pod.Status.Conditions {
		conditions = append(conditions, fmt.Sprintf("%s:%s", cond.Type, cond.Status))
	}

	maps := map[string]interface{}{
		"Name":       pod.ObjectMeta.Name,
		"IP":         pod.Status.PodIP,
		"Conditions": strings.Join(conditions, ","),
	}
	prettyPrint(maps)
}

func prettyPrint(maps map[string]interface{}) {
	length := 0
	for k := range maps {
		if len(k) > length {
			length = len(k)
		}
	}

	for k, v := range maps {
		spaces := length - len(k)
		for i := 0; i < spaces; i++ {
			k += " "
		}
		fmt.Printf("%s: %s\n", k, v)
	}
}
