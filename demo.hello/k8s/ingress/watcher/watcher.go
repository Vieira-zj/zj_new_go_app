package watcher

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"github.com/bep/debounce"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

/*
Watcher 负责查询 Kubernetes 和创建 Payloads 的，Payloads 包含了满足 HTTP 请求所需要的所有的 Kubernetes 数据。
*/

// IngressPayload Ingress 加上他的服务端口。
type IngressPayload struct {
	Ingress      *extensionsv1beta1.Ingress // ingress configuration
	ServicePorts map[string]map[string]int  // service_name:port_name:port
}

// Payload a collection of Kubernetes data loaded by the watcher.
type Payload struct {
	Ingresses       []IngressPayload
	TLSCertificates map[string]*tls.Certificate // cert_name:cert
}

// Watcher watches for ingresses in the kubernetes cluster.
type Watcher struct {
	client   kubernetes.Interface
	onChange func(*Payload)
}

// New creates a new Watcher.
func New(client kubernetes.Interface, onChange func(*Payload)) *Watcher {
	return &Watcher{
		client:   client,
		onChange: onChange,
	}
}

// Run runs the watcher.
func (w *Watcher) Run(ctx context.Context) error {
	factory := informers.NewSharedInformerFactory(w.client, time.Minute)
	secretLister := factory.Core().V1().Secrets().Lister()
	serviceLister := factory.Core().V1().Services().Lister()
	ingressLister := factory.Extensions().V1beta1().Ingresses().Lister()

	addBackend := func(ingressPayload *IngressPayload, backend extensionsv1beta1.IngressBackend) {
		// 通过 Ingress 所在的 namespace 和 ServiceName 获取 Service 对象
		svc, err := serviceLister.Services(ingressPayload.Ingress.Namespace).Get(backend.ServiceName)
		if err != nil {
			fmt.Printf("unknown service: ns=%s, service=%s\n", ingressPayload.Ingress.Namespace, backend.ServiceName)
			return
		}

		// Service 端口映射
		// example: {svcname: {httpport: 80, httpsport: 443}}
		m := make(map[string]int)
		for _, port := range svc.Spec.Ports {
			m[port.Name] = int(port.Port)
		}
		ingressPayload.ServicePorts[svc.Name] = m
	}

	// 检测到 k8s 变更时，从头开始重新构建所有的数据
	onChange := func() {
		// 获得所有的 ingresses
		ingresses, err := ingressLister.List(labels.Everything())
		if err != nil {
			fmt.Println("failed to list ingresses")
			return
		}
		fmt.Println("[watcher] total ingresses:", len(ingresses))

		payload := &Payload{
			TLSCertificates: make(map[string]*tls.Certificate),
		}
		for _, ingress := range ingresses {
			// ingress 和 service 处理
			ingressPayload := IngressPayload{
				Ingress:      ingress,
				ServicePorts: make(map[string]map[string]int),
			}
			payload.Ingresses = append(payload.Ingresses, ingressPayload)

			if ingress.Spec.Backend != nil {
				addBackend(&ingressPayload, *ingress.Spec.Backend)
			}

			for _, rule := range ingress.Spec.Rules {
				if rule.HTTP == nil {
					continue
				}
				for _, path := range rule.HTTP.Paths {
					addBackend(&ingressPayload, path.Backend)
				}
			}

			// 如果有 TLS 规则，则从 secrets 对象中加载证书
			for _, rec := range ingress.Spec.TLS {
				if rec.SecretName != "" {
					secret, err := secretLister.Secrets(ingress.Namespace).Get(rec.SecretName)
					if err != nil {
						fmt.Printf("unknown secret: ns=%s, secret=%s\n", ingress.Namespace, rec.SecretName)
						continue
					}
					cert, err := tls.X509KeyPair(secret.Data["tls.crt"], secret.Data["tls.key"])
					if err != nil {
						fmt.Printf("invalid tls certificate: ns=%s, secret=%s\n", ingress.Namespace, rec.SecretName)
						continue
					}
					payload.TLSCertificates[rec.SecretName] = &cert
				}
			}
		}

		// payload includes all ingressres and certs
		// onChange(payload) => server.Update(payload) => sync route table
		w.onChange(payload)
	}

	// debouncing（防抖动）是一种避免事件重复的方法，我们设置一个小的延迟，如果在达到延迟之前发生了其他事件，则重启计时器
	debounced := debounce.New(time.Second)
	handler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			debounced(onChange)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			debounced(onChange)
		},
		DeleteFunc: func(obj interface{}) {
			debounced(onChange)
		},
	}

	// 启动 Secret,Ingress,Service 的 Informer, 用同一个事件处理器 handler
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		informer := factory.Core().V1().Secrets().Informer()
		informer.AddEventHandler(handler)
		informer.Run(ctx.Done())
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		informer := factory.Extensions().V1beta1().Ingresses().Informer()
		informer.AddEventHandler(handler)
		informer.Run(ctx.Done())
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		informer := factory.Core().V1().Services().Informer()
		informer.AddEventHandler(handler)
		informer.Run(ctx.Done())
		wg.Done()
	}()

	wg.Wait()
	return nil
}
