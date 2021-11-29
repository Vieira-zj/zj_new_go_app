package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

var (
	device      string        = "lo0"
	snapLen     int32         = 1024
	promiscuous bool          = false
	timeout     time.Duration = 10 * time.Second
)

/*
refer:
https://colobu.com/2019/06/01/packet-capture-injection-and-analysis-gopacket/
https://pkg.go.dev/github.com/google/gopacket#section-readme
*/

func main() {
	// Open device
	handle, err := pcap.OpenLive(device, snapLen, promiscuous, timeout)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// Set filter
	filter := "tcp and port 8080"
	if err := handle.SetBPFFilter(filter); err != nil {
		log.Fatal(err)
	}

	// Use the handle as a packet source to process all packets
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	fmt.Println("Pcap packet:")
	for packet := range packetSource.Packets() {
		printPacketInfo(packet)
	}
}

func printPacketInfo(packet gopacket.Packet) {
	fmt.Println("All packet layers:")
	for _, layer := range packet.Layers() {
		fmt.Println("- ", layer.LayerType())
	}

	if loopBackLayer := packet.Layer(layers.LayerTypeLoopback); loopBackLayer != nil {
		fmt.Println("Loopback layer detected.")
		loopBackPacket, _ := loopBackLayer.(*layers.Loopback)
		fmt.Println("Bytes length:", len(loopBackPacket.LayerPayload()))
	}

	if ethernetLayer := packet.Layer(layers.LayerTypeEthernet); ethernetLayer != nil {
		fmt.Println("Ethernet layer detected.")
		ethernetPacket, _ := ethernetLayer.(*layers.Ethernet)
		fmt.Println("Source MAC: ", ethernetPacket.SrcMAC)
		fmt.Println("Destination MAC: ", ethernetPacket.DstMAC)
		fmt.Println("Ethernet type: ", ethernetPacket.EthernetType)
		fmt.Println()
	}

	if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
		fmt.Println("IPv4 layer detected.")
		ipv4, _ := ipLayer.(*layers.IPv4)
		fmt.Printf("From %s to %s\n", ipv4.SrcIP, ipv4.DstIP)
		fmt.Println("Protocol:", ipv4.Protocol)
		fmt.Println()
	}

	if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
		fmt.Println("TCP layer detected.")
		tcp, _ := tcpLayer.(*layers.TCP)
		fmt.Printf("From port %d to %d\n", tcp.SrcPort, tcp.DstPort)
		fmt.Println("Sequence number: ", tcp.Seq)
		fmt.Println()
	}

	if appLayer := packet.ApplicationLayer(); appLayer != nil {
		fmt.Println("Application layer/Payload found.")
		payload := string(appLayer.Payload())
		fmt.Println(payload)

		if strings.Contains(payload, "HTTP") {
			fmt.Println("HTTP found!")
		}
	}

	if err := packet.ErrorLayer(); err != nil {
		fmt.Println("Error decoding some part of the packet:", err)
	}
}
