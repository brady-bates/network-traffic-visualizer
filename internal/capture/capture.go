package capture

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"log"
	"net"
	"strings"
)

type Capture struct {
	Handle      *pcap.Handle
	Packets     chan Packet
	localSubnet *net.IPNet
}

type Device struct {
	Name        string
	Description string
	IPv4        net.IP
	Subnet      *net.IPNet
}

func NewCaptureProvider(handle *pcap.Handle, subnet *net.IPNet) *Capture {
	return &Capture{
		Handle:      handle,
		Packets:     make(chan Packet, 50000),
		localSubnet: subnet,
	}
}

func (c *Capture) StartPacketCapture(packetIn chan<- gopacket.Packet) {
	source := gopacket.NewPacketSource(c.Handle, c.Handle.LinkType())

	for packet := range source.Packets() {
		if packetIn != nil {
			select {
			case packetIn <- packet:
			default:
			}
		}

		if IsValidLayerType(packet.NetworkLayer()) {
			select {
			case c.Packets <- NewPacketFromGopacket(packet, c.localSubnet):
			default:
				log.Println("Dropped packet (channel full)")
			}
		} else {
			var layerType string
			if packet.NetworkLayer() == nil {
				layerType = "nil"
			} else {
				layerType = packet.NetworkLayer().LayerType().String()
			}
			log.Printf("Dropped packet (invalid network layer type %s)\n", layerType)
		}
	}
}

func (c *Capture) SetHandleBPFFilter(filter string) error {
	return c.Handle.SetBPFFilter(filter)
}

func FindDevice(preferred string) (Device, error) {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		return Device{}, err
	}

	return selectDevice(devices, preferred, defaultRouteIPv4())
}

func selectDevice(devices []pcap.Interface, preferred string, routedIP net.IP) (Device, error) {
	candidates := make([]Device, 0, len(devices))
	for _, device := range devices {
		candidate, ok := deviceIPv4(device)
		if !ok {
			continue
		}
		candidates = append(candidates, candidate)
	}

	if preferred != "" {
		for _, candidate := range candidates {
			if strings.EqualFold(candidate.Name, preferred) ||
				(candidate.Description != "" && strings.EqualFold(candidate.Description, preferred)) {
				return candidate, nil
			}
		}

		return Device{}, fmt.Errorf("capture interface %q not found or has no non-loopback IPv4 address", preferred)
	}

	if routedIP != nil {
		for _, candidate := range candidates {
			if candidate.IPv4.Equal(routedIP) {
				return candidate, nil
			}
		}
	}

	for _, candidate := range candidates {
		if candidate.IPv4.IsPrivate() {
			return candidate, nil
		}
	}

	if len(candidates) > 0 {
		return candidates[0], nil
	}

	return Device{}, fmt.Errorf("could not find a capture interface with a non-loopback IPv4 address")
}

func defaultRouteIPv4() net.IP {
	connection, err := net.Dial("udp4", "192.0.2.1:9")
	if err != nil {
		return nil
	}
	defer connection.Close()

	address, ok := connection.LocalAddr().(*net.UDPAddr)
	if !ok {
		return nil
	}
	return address.IP.To4()
}

func deviceIPv4(device pcap.Interface) (Device, bool) {
	for _, address := range device.Addresses {
		ipv4 := address.IP.To4()
		if ipv4 == nil || ipv4.IsLoopback() {
			continue
		}

		mask := address.Netmask
		if len(mask) == 0 {
			mask = net.CIDRMask(32, 32)
		}

		return Device{
			Name:        device.Name,
			Description: device.Description,
			IPv4:        ipv4,
			Subnet: &net.IPNet{
				IP:   ipv4.Mask(mask),
				Mask: mask,
			},
		}, true
	}

	return Device{}, false
}

func ListDevices() ([]pcap.Interface, error) {
	return pcap.FindAllDevs()
}

func (d Device) DisplayName() string {
	if d.Description == "" || d.Description == d.Name {
		return d.Name
	}
	return fmt.Sprintf("%s (%s)", d.Description, d.Name)
}

func OpenDevice(d Device) (*pcap.Handle, error) {
	if d.Name == "" {
		return nil, fmt.Errorf("capture interface name is empty")
	}

	handle, err := pcap.OpenLive(d.Name, 65536, true, pcap.BlockForever)
	if err != nil {
		return nil, fmt.Errorf("open capture interface %s: %w", d.DisplayName(), err)
	}
	return handle, nil
}
