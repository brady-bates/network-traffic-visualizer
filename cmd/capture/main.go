package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"networktrafficvisualizer/internal/capture"
	"networktrafficvisualizer/internal/captureoutput"
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
	flag.StringVar(&opts.deviceName, "interface", "", "capture interface name or description; defaults to the interface carrying the default IPv4 route")
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

	device, err := capture.FindDevice(opts.deviceName)
	if err != nil {
		return err
	}

	handle, err := capture.OpenDevice(device)
	if err != nil {
		return err
	}
	defer handle.Close()

	provider := capture.NewCaptureProvider(handle, device.Subnet)
	if opts.filter != "" {
		if err := provider.SetHandleBPFFilter(opts.filter); err != nil {
			return fmt.Errorf("set BPF filter: %w", err)
		}
	}

	fmt.Fprintf(diagnostics, "capturing on %s with IPv4 subnet %s; streaming packet metadata as NDJSON\n", device.DisplayName(), device.Subnet)
	go provider.StartPacketCapture(nil)
	return captureoutput.StreamNDJSON(ctx, output, provider.Data, opts.count)
}

func listInterfaces(output io.Writer) error {
	devices, err := capture.ListDevices()
	if err != nil {
		return err
	}

	for _, device := range devices {
		if device.Description == "" || device.Description == device.Name {
			fmt.Fprintln(output, device.Name)
		} else {
			fmt.Fprintf(output, "%s\n  name: %s\n", device.Description, device.Name)
		}
		for _, address := range device.Addresses {
			fmt.Fprintf(output, "  address: %s\n", address.IP)
		}
	}
	return nil
}
