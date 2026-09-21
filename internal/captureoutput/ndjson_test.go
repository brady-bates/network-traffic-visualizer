package captureoutput

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"networktrafficvisualizer/internal/capture"
	"strings"
	"testing"
	"time"
)

func TestNewPacketEvent(t *testing.T) {
	observedAt := time.Date(2026, time.September, 20, 5, 0, 0, 0, time.FixedZone("test", -5*60*60))
	event := NewPacketEvent(capture.PacketData{
		SrcIP:      net.ParseIP("192.0.2.1"),
		DstIP:      net.ParseIP("198.51.100.2"),
		IsIncoming: true,
	}, observedAt)

	if event.SchemaVersion != SchemaVersion {
		t.Fatalf("SchemaVersion = %d, want %d", event.SchemaVersion, SchemaVersion)
	}
	if event.ObservedAt != observedAt.UTC() {
		t.Fatalf("ObservedAt = %s, want %s", event.ObservedAt, observedAt.UTC())
	}
	if event.SourceAddress != "192.0.2.1" {
		t.Fatalf("SourceAddress = %q, want %q", event.SourceAddress, "192.0.2.1")
	}
	if event.DestinationAddress != "198.51.100.2" {
		t.Fatalf("DestinationAddress = %q, want %q", event.DestinationAddress, "198.51.100.2")
	}
	if event.Direction != "inbound" {
		t.Fatalf("Direction = %q, want %q", event.Direction, "inbound")
	}
}

func TestStreamNDJSONStopsAtLimit(t *testing.T) {
	packets := make(chan capture.PacketData, 3)
	for i := 0; i < 3; i++ {
		packets <- capture.PacketData{
			SrcIP: net.IPv4(192, 0, 2, byte(i+1)),
			DstIP: net.IPv4(198, 51, 100, 1),
		}
	}

	var output bytes.Buffer
	if err := StreamNDJSON(context.Background(), &output, packets, 2); err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("wrote %d lines, want 2", len(lines))
	}

	for _, line := range lines {
		var event PacketEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatalf("invalid JSON %q: %v", line, err)
		}
		if event.SchemaVersion != SchemaVersion {
			t.Fatalf("SchemaVersion = %d, want %d", event.SchemaVersion, SchemaVersion)
		}
	}
}
