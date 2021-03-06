package server

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"demo.hello/k8s/ingress/watcher"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// A RoutingTable contains the information (service backend and cert) needed to route a request.
type RoutingTable struct {
	backendsByHost     map[string][]routingTableBackend       // host:backends
	certificatesByHost map[string]map[string]*tls.Certificate // host:cert_host:cert
}

// NewRoutingTable creates a new RoutingTable.
func NewRoutingTable(payload *watcher.Payload) *RoutingTable {
	rt := &RoutingTable{
		backendsByHost:     make(map[string][]routingTableBackend),
		certificatesByHost: make(map[string]map[string]*tls.Certificate),
	}
	rt.init(payload)

	fmt.Println("[route] table records:")
	for host, backends := range rt.backendsByHost {
		for _, backend := range backends {
			fmt.Printf("ingress_host:%s,ingress_path:%s,backend_service:%s\n", host, backend.pathRE.String(), backend.url.Host)
		}
	}
	return rt
}

func (rt *RoutingTable) init(payload *watcher.Payload) {
	if payload == nil {
		return
	}

	for _, ingressPayload := range payload.Ingresses {
		for _, rule := range ingressPayload.Ingress.Spec.Rules {
			m, ok := rt.certificatesByHost[rule.Host]
			if !ok {
				m = make(map[string]*tls.Certificate)
				rt.certificatesByHost[rule.Host] = m
			}
			// 更新路由表证书信息
			for _, t := range ingressPayload.Ingress.Spec.TLS {
				for _, h := range t.Hosts {
					if cert, ok := payload.TLSCertificates[t.SecretName]; ok {
						m[h] = cert
					}
				}
			}
			rt.addBackend(ingressPayload, rule)
		}
	}
}

func (rt *RoutingTable) addBackend(ingressPayload watcher.IngressPayload, rule extensionsv1beta1.IngressRule) {
	if rule.HTTP == nil {
		if ingressPayload.Ingress.Spec.Backend != nil {
			backend := ingressPayload.Ingress.Spec.Backend
			rtb, err := newRoutingTableBackend("", backend.ServiceName,
				rt.getServicePort(ingressPayload, backend.ServiceName, backend.ServicePort))
			if err != nil {
				// this shouldn't happen
				fmt.Println(err)
				return
			}
			rt.backendsByHost[rule.Host] = append(rt.backendsByHost[rule.Host], rtb)
		}
	} else {
		for _, path := range rule.HTTP.Paths {
			backend := path.Backend
			rtb, err := newRoutingTableBackend(path.Path, backend.ServiceName,
				rt.getServicePort(ingressPayload, backend.ServiceName, backend.ServicePort))
			if err != nil {
				fmt.Println("invalid ingress rule path regex:", err)
				continue
			}
			rt.backendsByHost[rule.Host] = append(rt.backendsByHost[rule.Host], rtb)
		}
	}
}

// GetCertificate gets a certificate.
func (rt *RoutingTable) GetCertificate(sni string) (*tls.Certificate, error) {
	if hostCerts, ok := rt.certificatesByHost[sni]; ok {
		for h, cert := range hostCerts {
			if rt.matches(sni, h) {
				return cert, nil
			}
		}
	}
	return nil, errors.New("certificate not found")
}

// GetBackend gets the backend for the given host and path.
func (rt *RoutingTable) GetBackend(host, path string) (*url.URL, error) {
	if idx := strings.IndexByte(host, ':'); idx > 0 {
		host = host[:idx]
	}
	backends := rt.backendsByHost[host]
	for _, backend := range backends {
		if backend.matches(path) {
			return backend.url, nil
		}
	}
	return nil, errors.New("backend not found")
}

// getServicePort used for ingress api version: k8s.io/api/extensions/v1beta1
func (rt *RoutingTable) getServicePort(ingressPayload watcher.IngressPayload, serviceName string, servicePort intstr.IntOrString) int {
	if servicePort.Type == intstr.Int {
		return servicePort.IntValue()
	}
	if m, ok := ingressPayload.ServicePorts[serviceName]; ok {
		return m[servicePort.String()]
	}
	return 80
}

func (rt *RoutingTable) matches(sni string, certHost string) bool {
	for strings.HasPrefix(certHost, "*.") {
		if idx := strings.IndexByte(sni, '.'); idx >= 0 {
			sni = sni[idx+1:]
		} else {
			return false
		}
		certHost = certHost[2:]
	}
	return sni == certHost
}

// path => service_name:service_port
type routingTableBackend struct {
	pathRE *regexp.Regexp
	url    *url.URL
}

func newRoutingTableBackend(path string, serviceName string, servicePort int) (routingTableBackend, error) {
	rtb := routingTableBackend{
		url: &url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%d", serviceName, servicePort),
		},
	}

	var err error
	if path != "" {
		rtb.pathRE, err = regexp.Compile(path)
	}
	return rtb, err
}

func (rtb routingTableBackend) matches(path string) bool {
	if rtb.pathRE == nil {
		return true
	}
	return rtb.pathRE.MatchString(path)
}
