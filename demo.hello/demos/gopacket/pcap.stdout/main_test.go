package main

import (
	"fmt"
	"net"
	"testing"

	"github.com/google/gopacket/pcap"
)

func TestFindAllDevs(t *testing.T) {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		t.Fatal(err)
	}

	for _, device := range devices {
		if device.Name == "en0" || device.Name == "lo0" {
			for _, addr := range device.Addresses {
				fmt.Println(device.Name, addr.IP)
			}
		}
	}
}

func TestNetInterfaces(t *testing.T) {
	netIfs, err := net.Interfaces()
	if err != nil {
		t.Fatal(err)
	}

	for _, netIf := range netIfs {
		if netIf.Name == "en0" || netIf.Name == "lo0" {
			addrs, err := netIf.Addrs()
			if err != nil {
				t.Fatal(err)
			}
			for _, addr := range addrs {
				fmt.Println(netIf.Name, addr.String())
			}
		}
	}
}
