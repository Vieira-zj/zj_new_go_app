package main

import (
	"fmt"
	"log"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

var (
	device      string        = "lo0"
	snapshotLen int32         = 1024
	promiscuous bool          = false
	timeout     time.Duration = 10 * time.Second
	// Will reuse these for each packet
	lb  layers.Loopback
	eth layers.Ethernet
	ip4 layers.IPv4
	tcp layers.TCP
)

func main() {
	handle, err := pcap.OpenLive(device, snapshotLen, promiscuous, timeout)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	filter := "tcp and port 8080"
	if err := handle.SetBPFFilter(filter); err != nil {
		log.Fatal(err)
	}

	// Use LayerTypeLoopback for local test
	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeLoopback, &lb, &ip4, &tcp)
	decoded := []gopacket.LayerType{}

	fmt.Println("Pcap packet:")
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		if err := parser.DecodeLayers(packet.Data(), &decoded); err != nil {
			fmt.Println("Could not decode layers:", err)
		}

		for _, layerType := range decoded {
			switch layerType {
			case layers.LayerTypeIPv4:
				fmt.Printf("IPV4: %v -> %v\n", ip4.SrcIP, ip4.DstIP)
			case layers.LayerTypeTCP:
				fmt.Printf("TCP Port: %v -> %v\n", tcp.SrcPort, tcp.DstPort)
				fmt.Printf("TCP SYN: %v | ACK: %v\n", tcp.SYN, tcp.Ack)
			}
		}
		fmt.Println()
	}
}
