package csv

import (
	"encoding/csv"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"log"
	"networktrafficart/internal/util"
	"os"
	"reflect"
)

type packetRecord struct {
	Timestamp int64
	Length    int32

	SrcMAC    string
	DstMAC    string
	EtherType string

	SrcIP      string
	DstIP      string
	IPVersion  int32
	TTL        int32
	IPProtocol string

	SrcPort  int32
	DstPort  int32
	Protocol string

	TCPSyn bool
	TCPAck bool
	TCPFin bool
	TCPRst bool
	TCPPsh bool

	DNSQuery        string
	DNSIsResponse   bool
	DNSResponseCode int32
}

func writeCSVHeader(writer *csv.Writer) error {
	return writer.Write(reflectPacketRecord())
}

func appendPacketToCSV(writer *csv.Writer, packet gopacket.Packet) error {
	return writer.Write(newPacketRecord(packet).toStringArray())
}

func StreamToCSV(shutDown *util.ShutdownContext, packetOut <-chan gopacket.Packet, filename string) {
	var file *os.File
	var err error

	fileExists, _ := util.FileExists(filename)

	if fileExists {
		if file, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644); err != nil {
			log.Fatal(err)
		}
	} else {
		file, err = os.Create(filename)
	}

	writer := csv.NewWriter(file)

	if !fileExists {
		if err = writeCSVHeader(writer); err != nil {
			log.Fatal(err)
		}
		writer.Flush()
	}

	for {
		select {
		case packet := <-packetOut:
			if err = appendPacketToCSV(writer, packet); err != nil {
				log.Fatal("Failed to append to file: ", err)
			}
			writer.Flush()
		case <-shutDown.Context.Done():
			fmt.Println("Shutdown signal received")
			writer.Flush()
			if err = file.Sync(); err != nil {
				log.Println(err)
			}
			if err = file.Close(); err != nil {
				log.Println(err)
			}
			return
		}
	}
}

func newPacketRecord(packet gopacket.Packet) packetRecord {
	r := packetRecord{
		Timestamp: packet.Metadata().Timestamp.Unix(),
		Length:    int32(packet.Metadata().Length),
	}

	if eth, ok := packet.LinkLayer().(*layers.Ethernet); ok {
		r.SrcMAC = eth.SrcMAC.String()
		r.DstMAC = eth.DstMAC.String()
		r.EtherType = eth.EthernetType.String()
	}

	if ip4 := packet.Layer(layers.LayerTypeIPv4); ip4 != nil {
		v4 := ip4.(*layers.IPv4)
		r.SrcIP = v4.SrcIP.String()
		r.DstIP = v4.DstIP.String()
		r.IPVersion = 4
		r.TTL = int32(v4.TTL)
		r.IPProtocol = v4.Protocol.String()
	}

	if ip6 := packet.Layer(layers.LayerTypeIPv6); ip6 != nil {
		v6 := ip6.(*layers.IPv6)
		r.SrcIP = v6.SrcIP.String()
		r.DstIP = v6.DstIP.String()
		r.IPVersion = 6
		r.TTL = int32(v6.HopLimit)
		r.IPProtocol = v6.NextHeader.String()
	}

	if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
		tcp := tcpLayer.(*layers.TCP)
		r.SrcPort = int32(tcp.SrcPort)
		r.DstPort = int32(tcp.DstPort)
		r.Protocol = "TCP"
		r.TCPSyn = tcp.SYN
		r.TCPAck = tcp.ACK
		r.TCPFin = tcp.FIN
		r.TCPRst = tcp.RST
		r.TCPPsh = tcp.PSH
	}

	if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
		udp := udpLayer.(*layers.UDP)
		r.SrcPort = int32(udp.SrcPort)
		r.DstPort = int32(udp.DstPort)
		r.Protocol = "UDP"
	}

	if dnsLayer := packet.Layer(layers.LayerTypeDNS); dnsLayer != nil {
		dns := dnsLayer.(*layers.DNS)
		r.DNSIsResponse = dns.QR
		r.DNSResponseCode = int32(dns.ResponseCode)
		if len(dns.Questions) > 0 {
			r.DNSQuery = string(dns.Questions[0].Name)
		}
	}

	return r
}

func (r packetRecord) toStringArray() []string {
	return []string{
		fmt.Sprintf("%d", r.Timestamp),
		fmt.Sprintf("%d", r.Length),
		r.SrcMAC,
		r.DstMAC,
		r.EtherType,
		r.SrcIP,
		r.DstIP,
		fmt.Sprintf("%d", r.IPVersion),
		fmt.Sprintf("%d", r.TTL),
		r.IPProtocol,
		fmt.Sprintf("%d", r.SrcPort),
		fmt.Sprintf("%d", r.DstPort),
		r.Protocol,
		fmt.Sprintf("%t", r.TCPSyn),
		fmt.Sprintf("%t", r.TCPAck),
		fmt.Sprintf("%t", r.TCPFin),
		fmt.Sprintf("%t", r.TCPRst),
		fmt.Sprintf("%t", r.TCPPsh),
		r.DNSQuery,
		fmt.Sprintf("%t", r.DNSIsResponse),
		fmt.Sprintf("%d", r.DNSResponseCode),
	}
}

func reflectPacketRecord() []string {
	t := reflect.TypeOf(packetRecord{})
	var headers []string
	for i := 0; i < t.NumField(); i++ {
		headers = append(headers, t.Field(i).Name)
	}
	return headers
}
