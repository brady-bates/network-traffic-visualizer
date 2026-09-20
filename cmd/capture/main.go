package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/gopacket/pcap"
	"io"
	"log"
	"net"
	"networktrafficart/internal/capture"
	"networktrafficart/internal/captureoutput"
	"os"
	"os/signal"
	"syscall"
)

type options struct {
	deviceName string
	filter     string
	count      int
	list       bool
}

func main() {
	opts := parseOptions()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, os.Stdout, os.Stderr, opts); err != nil {
		log.Fatal(err)
	}
}

func parseOptions() options {
	var opts options
	flag.StringVar(&opts.deviceName, "interface", "", "capture interface; defaults to the first non-loopback IPv4 interface")
	flag.StringVar(&opts.filter, "filter", "", "optional complete BPF capture filter")
	flag.IntVar(&opts.count, "count", 0, "stop after this many events; zero streams until interrupted")
	flag.BoolVar(&opts.list, "list-interfaces", false, "list available capture interfaces and exit")
	flag.Parse()
	return opts
}

func run(ctx context.Context, output, diagnostics io.Writer, opts options) error {
	if opts.count < 0 {
		return fmt.Errorf("count must be zero or greater")
	}
	if opts.list {
		return listInterfaces(output)
	}

	deviceName, subnet, err := captureDevice(opts.deviceName)
	if err != nil {
		return err
	}

	handle, err := pcap.OpenLive(deviceName, 65536, true, pcap.BlockForever)
	if err != nil {
		return fmt.Errorf("open capture interface %s: %w", deviceName, err)
	}
	defer handle.Close()

	provider := capture.NewCaptureProvider(handle, subnet)
	if opts.filter != "" {
		if err := provider.SetHandleBPFFilter(opts.filter); err != nil {
			return fmt.Errorf("set BPF filter: %w", err)
		}
	}

	fmt.Fprintf(diagnostics, "capturing on %s (%s); streaming packet metadata as NDJSON\n", deviceName, subnet)
	go provider.StartPacketCapture(nil)
	return captureoutput.StreamNDJSON(ctx, output, provider.Data, opts.count)
}

func captureDevice(deviceName string) (string, *net.IPNet, error) {
	if deviceName == "" {
		return capture.GetDefaultCaptureDevice()
	}

	subnet, err := capture.GetInterfaceIPv4SubnetRange(deviceName)
	if err != nil {
		return "", nil, err
	}
	return deviceName, subnet, nil
}

func listInterfaces(output io.Writer) error {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		return err
	}

	for _, device := range devices {
		fmt.Fprintln(output, device.Name)
		for _, address := range device.Addresses {
			fmt.Fprintf(output, "  %s\n", address.IP)
		}
	}
	return nil
}
