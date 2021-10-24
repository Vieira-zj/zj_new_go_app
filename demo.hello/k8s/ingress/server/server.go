package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"sync/atomic"

	"demo.hello/k8s/ingress/watcher"
	"golang.org/x/sync/errgroup"
)

// A Server serves HTTP pages.
type Server struct {
	cfg          *config
	routingTable atomic.Value // 使用的是 atomic.Value 来存储路由表
	ready        *Event
}

// New creates a new server.
func New(options ...Option) *Server {
	cfg := defaultConfig()
	for _, opt := range options {
		opt(cfg)
	}
	s := &Server{
		cfg:   cfg,
		ready: NewEvent(),
	}
	s.routingTable.Store(NewRoutingTable(nil))
	return s
}

// Run starts server.
func (s *Server) Run(ctx context.Context) error {
	// 直到第一个 payload 数据后才开始监听
	s.ready.Wait(ctx)

	var eg errgroup.Group
	// https
	eg.Go(func() error {
		srv := http.Server{
			Addr:    fmt.Sprintf("%s:%d", s.cfg.host, s.cfg.tlsPort),
			Handler: s, // 代理逻辑 => ServeHTTP()
		}
		srv.TLSConfig = &tls.Config{
			GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
				return s.routingTable.Load().(*RoutingTable).GetCertificate(hello.ServerName)
			},
		}
		fmt.Println("starting secure HTTP server:", srv.Addr)
		if err := srv.ListenAndServeTLS("", ""); err != nil {
			return fmt.Errorf("error serving tls: %w", err)
		}
		return nil
	})

	// http
	eg.Go(func() error {
		srv := http.Server{
			Addr:    fmt.Sprintf("%s:%d", s.cfg.host, s.cfg.port),
			Handler: s, // 代理逻辑 => ServeHTTP()
		}
		fmt.Println("starting insecure HTTP server", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			return fmt.Errorf("error serving non-tls: %w", err)
		}
		return nil
	})

	return eg.Wait()
}

// ServeHTTP serves an HTTP request.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 获取后端的真实服务地址
	backendURL, err := s.routingTable.Load().(*RoutingTable).GetBackend(r.Host, r.URL.Path)
	if err != nil {
		http.Error(w, "upstream server not found", http.StatusNotFound)
	}
	fmt.Printf("proxying request: [%s:/%s] to backend: %s\n", r.Host, r.URL.Path, backendURL.String())

	// 使用 NewSingleHostReverseProxy 代理请求到 backend service
	p := httputil.NewSingleHostReverseProxy(backendURL)
	p.ErrorLog = log.New(os.Stdout, "[proxy]", 0)
	p.ServeHTTP(w, r)
}

// Update 根据新的 Ingress 规则来更新路由表。
func (s *Server) Update(payload *watcher.Payload) {
	s.routingTable.Store(NewRoutingTable(payload))
	s.ready.Set()
}
