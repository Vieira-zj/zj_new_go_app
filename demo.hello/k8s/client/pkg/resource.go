package pkg

import (
	"context"
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/watch"
	"k8s.io/kubernetes/pkg/client/conditions"
)

// Resource operator for k8s objects like namespace, pod and deploy.
type Resource struct {
	client *kubernetes.Clientset
}

var resource *Resource

// NewResource creates an instance of resource.
func NewResource(client *kubernetes.Clientset) *Resource {
	if resource != nil {
		return resource
	}
	resource = &Resource{
		client: client,
	}
	return resource
}

/*
Namespace
*/

// CreateNamespace creates a namespace.
func (r *Resource) CreateNamespace(ctx context.Context, name string) (*apiv1.Namespace, error) {
	namespaceSpec := &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Status: apiv1.NamespaceStatus{
			Phase: apiv1.NamespaceActive,
		},
	}
	return r.client.CoreV1().Namespaces().Create(ctx, namespaceSpec, metav1.CreateOptions{})
}

// GetAllNamespace returns all namespace.
func (r *Resource) GetAllNamespace(ctx context.Context) ([]apiv1.Namespace, error) {
	allNamesapce, err := r.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return allNamesapce.Items, nil
}

// GetNamespace returns a namespace by name.
func (r *Resource) GetNamespace(ctx context.Context, name string) (*apiv1.Namespace, error) {
	return r.client.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
}

// DeleteNamespace delete a namespace by name.
func (r *Resource) DeleteNamespace(ctx context.Context, name string) error {
	deletePolicy := metav1.DeletePropagationForeground
	deleteOption := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	return r.client.CoreV1().Namespaces().Delete(ctx, name, deleteOption)
}

/*
Pod
*/

// GetAllPods returns all pods in k8s cluster.
func (r *Resource) GetAllPods(ctx context.Context) ([]apiv1.Pod, error) {
	return r.GetPodsByNamespace(ctx, "")
}

// GetRunningPods .
func (r *Resource) GetRunningPods(pods []apiv1.Pod) []apiv1.Pod {
	runningPods := make([]apiv1.Pod, 0, len(pods))
	for _, pod := range pods {
		if pod.Status.Phase == apiv1.PodRunning {
			runningPods = append(runningPods, pod)
		}
	}
	return runningPods
}

// GetNonSystemPods returns all non kube system pods.
func (r *Resource) GetNonSystemPods(ctx context.Context) ([]apiv1.Pod, error) {
	allNamesapce, err := r.GetAllNamespace(ctx)
	if err != nil {
		return nil, err
	}

	retPods := make([]apiv1.Pod, 0, 50)
	for _, ns := range allNamesapce {
		if strings.HasPrefix(ns.Name, "kube") {
			continue
		}
		pods, err := r.GetPodsByNamespace(ctx, ns.Name)
		if err != nil {
			return nil, err
		}
		retPods = append(retPods, pods...)
	}
	return retPods, nil
}

// GetPodsByNamespace returns pods in specified namespace.
func (r *Resource) GetPodsByNamespace(ctx context.Context, namespace string) ([]apiv1.Pod, error) {
	pods, err := r.client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}

// GetPod returns a specified pod by namespace and name.
func (r *Resource) GetPod(ctx context.Context, namespace string, podName string) (*apiv1.Pod, error) {
	pod, err := r.client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("Pod not found: namespace=%s, name=%s", namespace, podName)
		}
		return nil, err
	}
	return pod, nil
}

// GetPodsByLabelsV2 .
func (r *Resource) GetPodsByLabelsV2(ctx context.Context, namespace string, kvLabels map[string]string) ([]apiv1.Pod, error) {
	labelSelector := labels.NewSelector()
	for key, value := range kvLabels {
		req, err := labels.NewRequirement(key, selection.Equals, []string{value})
		if err != nil {
			return nil, err
		}
		labelSelector.Add(*req)
	}

	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	}
	pods, err := r.client.CoreV1().Pods(namespace).List(ctx, listOptions)
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}

// GetPodsByLabels returns pods with specified labels.
func (r *Resource) GetPodsByLabels(ctx context.Context, namespace string, labels map[string]string) ([]apiv1.Pod, error) {
	pods, err := r.GetPodsByNamespace(ctx, namespace)
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

// GetPodNamespace returns a specified namespace by given pod.
func (r *Resource) GetPodNamespace(ctx context.Context, podName string) (string, error) {
	pods, err := r.GetAllPods(ctx)
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

// StartPod starts a given pod, and wait until it's running.
func (r *Resource) StartPod(ctx context.Context, podSpec *apiv1.Pod, timeout int) (*apiv1.Pod, error) {
	pod, err := r.client.CoreV1().Pods(podSpec.GetNamespace()).Create(ctx, podSpec, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	watcher, err := r.client.CoreV1().Pods(pod.GetNamespace()).Watch(ctx, metav1.SingleObject(pod.ObjectMeta))
	if err != nil {
		return nil, err
	}
	ctx, cacnel := context.WithTimeout(ctx, time.Duration(timeout)*time.Minute)
	defer cacnel()

	fmt.Printf("Waiting for pod %s to run...\n", pod.Name)
	event, err := watch.UntilWithoutRetry(ctx, watcher, conditions.PodRunning)
	if err != nil {
		err := fmt.Errorf("Error occurred while waiting for pod to run: %v", err)
		return nil, err
	}
	pod = event.Object.(*apiv1.Pod)
	return pod, nil
}

// DeletePod .
func (r *Resource) DeletePod(ctx context.Context, namespace string, podName string) error {
	defaults := int64(5)
	gracePeriodSeconds := defaults
	deleteOptions := metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}
	return r.client.CoreV1().Pods(namespace).Delete(ctx, podName, deleteOptions)
}

/*
Pod and Service
*/

// GetService returns a specified service by namespace and name.
func (r *Resource) GetService(ctx context.Context, namespace, name string) (*apiv1.Service, error) {
	service, err := r.client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("Service not found: namespace=%s, name=%s", namespace, name)
		}
		return nil, err
	}
	return service, nil
}

// GetPodsByServiceV2 returns pods in a specified services.
func (r *Resource) GetPodsByServiceV2(ctx context.Context, namespace, serviceName string) ([]apiv1.Pod, error) {
	service, err := r.GetService(ctx, namespace, serviceName)
	if err != nil {
		return nil, err
	}

	podSelector := labels.Set(service.Spec.Selector)
	listOptions := metav1.ListOptions{
		LabelSelector: podSelector.AsSelector().String(),
	}
	pods, err := r.client.CoreV1().Pods(namespace).List(ctx, listOptions)
	if err != nil {
		return nil, err
	}
	return pods.Items, nil
}

// GetPodsByService returns pods in a specified services.
func (r *Resource) GetPodsByService(ctx context.Context, namespace, serviceName string) ([]apiv1.Pod, error) {
	service, err := r.GetService(ctx, namespace, serviceName)
	if err != nil {
		return nil, err
	}

	podSelector := service.Spec.Selector
	pods, err := r.GetPodsByLabelsV2(ctx, namespace, podSelector)
	if err != nil {
		return nil, err
	}
	return pods, nil
}

/*
Deployment
*/

// GetDeploymentsByLabels .
func (r *Resource) GetDeploymentsByLabels(ctx context.Context, namespace string, kvLabels map[string]string) ([]v1.Deployment, error) {
	labelSelector := labels.NewSelector()
	for key, value := range kvLabels {
		req, err := labels.NewRequirement(key, selection.Equals, []string{value})
		if err != nil {
			return nil, err
		}
		labelSelector.Add(*req)
	}

	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	}
	deps, err := r.client.AppsV1().Deployments(namespace).List(ctx, listOptions)
	if err != nil {
		return nil, err
	}
	return deps.Items, nil
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
