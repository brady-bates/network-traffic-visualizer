package capture

import (
	"github.com/google/gopacket/pcap"
	"net"
	"testing"
)

func TestSelectDeviceUsesInterfaceOnDefaultRoute(t *testing.T) {
	devices := []pcap.Interface{
		testInterface("docker0", "Container network", "172.17.0.1", 16),
		testInterface("wifi0", "Wi-Fi", "192.168.1.25", 24),
	}

	device, err := selectDevice(devices, "", net.ParseIP("192.168.1.25"))
	if err != nil {
		t.Fatal(err)
	}
	if device.Name != "wifi0" {
		t.Fatalf("Name = %q, want %q", device.Name, "wifi0")
	}
	if device.Subnet.String() != "192.168.1.0/24" {
		t.Fatalf("Subnet = %q, want %q", device.Subnet, "192.168.1.0/24")
	}
}

func TestSelectDeviceUsesExplicitName(t *testing.T) {
	devices := []pcap.Interface{
		testInterface("ethernet0", "Ethernet", "10.0.0.5", 24),
		testInterface("wifi0", "Wi-Fi", "192.168.1.25", 24),
	}

	device, err := selectDevice(devices, "ethernet0", net.ParseIP("192.168.1.25"))
	if err != nil {
		t.Fatal(err)
	}
	if device.Name != "ethernet0" {
		t.Fatalf("Name = %q, want %q", device.Name, "ethernet0")
	}
}

func TestSelectDeviceUsesWindowsDescription(t *testing.T) {
	devices := []pcap.Interface{
		testInterface(`\Device\NPF_{A1B2C3}`, "Intel Ethernet Controller", "10.0.0.5", 24),
	}

	device, err := selectDevice(devices, "intel ethernet controller", nil)
	if err != nil {
		t.Fatal(err)
	}
	if device.Name != `\Device\NPF_{A1B2C3}` {
		t.Fatalf("Name = %q, want Windows NPF device name", device.Name)
	}
}

func TestSelectDeviceRejectsUnknownOverride(t *testing.T) {
	devices := []pcap.Interface{
		testInterface("wifi0", "Wi-Fi", "192.168.1.25", 24),
	}

	if _, err := selectDevice(devices, "missing", nil); err == nil {
		t.Fatal("selectDevice returned nil error for an unknown interface")
	}
}

func testInterface(name, description, address string, prefix int) pcap.Interface {
	return pcap.Interface{
		Name:        name,
		Description: description,
		Addresses: []pcap.InterfaceAddress{
			{
				IP:      net.ParseIP(address),
				Netmask: net.CIDRMask(prefix, 32),
			},
		},
	}
}
