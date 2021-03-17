package pkg

import (
	"context"
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Injector injects components into k8s resources.
type Injector struct {
	ctx    context.Context
	client *kubernetes.Clientset
}

var injector *Injector

// NewInjector returns an instance of injector.
func NewInjector(ctx context.Context, client *kubernetes.Clientset) *Injector {
	if injector != nil {
		return injector
	}
	injector = &Injector{
		ctx:    ctx,
		client: client,
	}
	return injector
}

// InjectBusyboxSidecar inject a busybox sidecar into a pod.
func (injector *Injector) InjectBusyboxSidecar(nsName string, deployName string) error {
	deploy, err := injector.client.AppsV1().Deployments(nsName).Get(injector.ctx, deployName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	containerName := "inject-container-test"
	containers := deploy.Spec.Template.Spec.Containers
	for _, c := range containers {
		if c.Name == containerName {
			fmt.Printf("sidecar container [%s] already exist.\n", containerName)
			return nil
		}
	}

	injectContainer := apiv1.Container{
		Image:           "busybox:1.30",
		ImagePullPolicy: apiv1.PullIfNotPresent,
		Name:            containerName,
		Env: []apiv1.EnvVar{
			{
				Name:  "Author",
				Value: "ZhengJin",
			},
		},
		Command: []string{
			"sh",
			"-c",
			"while true; do echo $(date +'%Y-%m-%d_%H:%M:%S') 'inject busybox is running ...'; sleep 5; done;",
		},
	}

	containers = append(containers, injectContainer)
	deploy.Spec.Template.Spec.Containers = containers
	_, err = injector.client.AppsV1().Deployments(nsName).Update(injector.ctx, deploy, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

// InjectInitTCPForword injects a init container to update tcp forward chain by iptables.
func (injector *Injector) InjectInitTCPForword(nsName string, deployName string) error {
	deploy, err := injector.client.AppsV1().Deployments(nsName).Get(injector.ctx, deployName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	containerName := "init-container-tcp-forword"
	initContainers := deploy.Spec.Template.Spec.InitContainers
	for _, c := range initContainers {
		if c.Name == containerName {
			fmt.Printf("init container [%s] already exist.\n", containerName)
			return nil
		}
	}

	securityCtx := &apiv1.SecurityContext{
		Capabilities: &apiv1.Capabilities{
			Add: []apiv1.Capability{"NET_ADMIN"},
		},
	}
	// iptables -dport: 数据要到达的目的端口 -sport: 是数据来源的端口
	initContainer := apiv1.Container{
		Image:           "biarca/iptables",
		ImagePullPolicy: apiv1.PullIfNotPresent,
		Name:            containerName,
		Command: []string{
			"sh",
			"-c",
			"iptables -t nat -A PREROUTING -p tcp --dport 8080 -j REDIRECT --to-port 8088",
		},
		SecurityContext: securityCtx,
	}

	initContainers = append(initContainers, initContainer)
	deploy.Spec.Template.Spec.InitContainers = initContainers
	_, err = injector.client.AppsV1().Deployments(nsName).Update(injector.ctx, deploy, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}
