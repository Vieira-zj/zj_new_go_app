package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

// ProxyServer .
type ProxyServer struct {
	tr *http.Transport
}

// NewProxyServer .
func NewProxyServer() *ProxyServer {
	return &ProxyServer{
		tr: &http.Transport{},
	}
}

// ServeHTTP .
func (p *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "CONNECT" {
		p.TransferHTTPS(w, r)
	} else {
		p.TransferPlainText(w, r)
	}
}

// TransferPlainText .
func (p *ProxyServer) TransferPlainText(w http.ResponseWriter, r *http.Request) {
	fmt.Println("TransferPlainText:", r.URL)
	resp, err := p.tr.RoundTrip(r)
	if err != nil {
		fmt.Printf("response from %v error: %v\n", r.URL, err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	for k, vals := range resp.Header {
		for _, val := range vals {
			w.Header().Set(k, val)
		}
	}
	w.WriteHeader(resp.StatusCode)

	if _, err := io.Copy(w, resp.Body); err != nil {
		fmt.Printf("send response back to client failed: %v\n", err)
		http.Error(w, "", resp.StatusCode)
		return
	}
}

// TransferHTTPS proxy for https by tcp channel.
func (p *ProxyServer) TransferHTTPS(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	fmt.Println("TransferHTTPS:", host)
	connToRemote, err := net.DialTimeout("tcp", host, time.Duration(5)*time.Second)
	if err != nil {
		fmt.Println("fail to connect to remote:", err)
		io.WriteString(w, "HTTP/1.1 502 Bad Gateway\r\n\r\n")
		return
	}

	hi, ok := w.(http.Hijacker)
	if !ok {
		fmt.Println("the http server does not support hijacker")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	connFromClient, _, err := hi.Hijack()
	if err != nil {
		fmt.Println("fail to hijack the connection:", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	connFromClient.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	// 双向转发数据
	connToRemoteTCP, _ := connToRemote.(*net.TCPConn)
	connFromClientTCP, _ := connFromClient.(*net.TCPConn)
	var wg sync.WaitGroup
	wg.Add(2)
	go copyWithWait(connToRemoteTCP, connFromClientTCP, &wg)
	go copyWithWait(connFromClientTCP, connToRemoteTCP, &wg)
	wg.Wait()
}

func copyWithWait(dst, src *net.TCPConn, wg *sync.WaitGroup) {
	defer wg.Done()
	nb, err := io.Copy(dst, src)
	if err != nil && nb == 0 {
		fmt.Printf("transfer encountering error: %v\n", err)
	}
	dst.CloseWrite()
	src.CloseRead()
}

func main() {
	p := NewProxyServer()
	fmt.Println("proxy start and listen at :8082")
	http.ListenAndServe(":8082", p)
}
