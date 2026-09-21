package capture

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"math/rand"
	"net"
)

type Packet struct {
	SrcIP      net.IP
	DstIP      net.IP
	IsIncoming bool
}

func NewPacket(srcIP, dstIP net.IP) Packet {
	return Packet{
		SrcIP:      srcIP,
		DstIP:      dstIP,
		IsIncoming: rand.Intn(2) == 1,
	}
}

func NewPacketFromGopacket(packet gopacket.Packet, subnet *net.IPNet) Packet {
	var srcIP, dstIP net.IP

	// TODO improve handling of IPv6 rather than "normalizing" to IPv4
	switch layer := packet.NetworkLayer().(type) {
	case *layers.IPv4:
		srcIP = layer.SrcIP
		dstIP = layer.DstIP
	case *layers.IPv6:
		srcIP = normalizeIP(layer.SrcIP)
		dstIP = normalizeIP(layer.DstIP)
	default:
		fmt.Printf("Unknown layer type %s - check if layer type is valid before calling\n", packet.NetworkLayer().LayerType().String())
	}

	return Packet{
		srcIP,
		dstIP,
		subnet.Contains(dstIP),
	}
}

func IsValidLayerType(layer gopacket.Layer) bool {
	if layer == nil {
		return false
	}

	switch layer.(type) {
	case *layers.IPv4, *layers.IPv6:
		return true
	default:
		return false
	}
}

// TODO find a better way to convert IPv4 and IPv6 to a common representation
func iPv6toIPv4Format(ip net.IP) net.IP {
	return ip[12:16].To4()
}

func isIPv6(ip net.IP) bool {
	return len(ip) == net.IPv6len
}

func normalizeIP(ip net.IP) net.IP {
	if isIPv6(ip) {
		return iPv6toIPv4Format(ip)
	} else {
		return ip
	}
}
