package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
)

var (
	pcapFile    string        = "test.pcap"
	deviceName  string        = "lo0"
	snapLen     uint32        = 1024
	promiscuous bool          = false
	timeout     time.Duration = -1 * time.Second
	packetCount int           = 0
)

func main() {
	// capture(true)
	readCapFile()
}

func readCapFile() {
	// Open file instead of device
	handle, err := pcap.OpenOffline(pcapFile)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// Loop through packets in file
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	fmt.Println("Read packets:")
	for packet := range packetSource.Packets() {
		fmt.Println(packet)
	}
}

func capture(isLocal bool) {
	// Open output pcap file and write header
	f, err := os.Create(pcapFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	w := pcapgo.NewWriter(f)
	if isLocal {
		w.WriteFileHeader(snapLen, layers.LinkTypeLoop)
	} else {
		w.WriteFileHeader(snapLen, layers.LinkTypeEthernet)
	}

	// Open the device for capturing
	handle, err := pcap.OpenLive(deviceName, int32(snapLen), promiscuous, timeout)
	if err != nil {
		fmt.Printf("Error opening device %s: %v\n", deviceName, err)
		os.Exit(1)
	}
	defer handle.Close()

	filter := "tcp and port 8080"
	if err := handle.SetBPFFilter(filter); err != nil {
		log.Fatal(err)
	}

	// Start processing packets
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	fmt.Println("Pcap packet:")
	for packet := range packetSource.Packets() {
		fmt.Println(packet)
		w.WritePacket(packet.Metadata().CaptureInfo, packet.Data())
		packetCount++

		if packetCount > 100 {
			break
		}
	}
}
