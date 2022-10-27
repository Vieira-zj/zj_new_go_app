package utils

import (
	"net"
	"strings"
	"testing"
	"time"
)

func TestTcpListenOnRandPort(t *testing.T) {
	ln, err := net.Listen("tcp4", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		time.Sleep(time.Second)
		ln.Close()
	}()

	tcpAddr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		t.Fatal("invalid tcp addr")
	}
	t.Logf("ip=%s, port=%d", tcpAddr.IP.String(), tcpAddr.Port)
}

func TestGetHostIpAddrs(t *testing.T) {
	localIPs, nonLocalIPs, err := GetHostIpAddrs()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("local ips:", strings.Join(localIPs, ","))
	t.Log("non local ips:", strings.Join(nonLocalIPs, ","))
}
