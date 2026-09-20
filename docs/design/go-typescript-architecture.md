# Go Collector and TypeScript Display Architecture

**Status:** Accepted  
**Date:** 2026-09-20  
**Implementation:** Pending

## Context

Network Traffic Art needs two capabilities with different technical demands:

1. Efficient, privileged packet capture and reliable interpretation of network
   traffic.
2. A flexible visual layer for experimenting with layouts, animation, and
   interactive controls.

Go and `gopacket` are a good fit for capture and long-running data processing,
but the current native rendering layer makes visual iteration unnecessarily
difficult. Moving all processing into a browser would instead expose raw packet
data to a higher-overhead runtime and duplicate network logic in the UI.

## Decision

The application will use a lightweight Go collector and a TypeScript display.
They will communicate through a versioned stream of aggregated flow summaries.

```text
libpcap
   |
   v
Go capture -> decode -> normalize -> aggregate -> privacy filter
                                                   |
                                                   v
                                      versioned flow summaries
                                                   |
                                             WebSocket
                                                   |
                                                   v
                         TypeScript scene model -> Canvas/WebGL display
```

Raw packets and payloads must not cross the collector/display boundary.

## Responsibilities

### Go collector

The collector owns network semantics:

- Open and manage live capture devices.
- Support recorded PCAP replay through the same capture interface.
- Decode link, network, and transport layers.
- Determine local and remote endpoints and traffic direction.
- Classify protocols and common service categories.
- Aggregate packets into bounded, expiring flows.
- Cache optional GeoIP, ASN, and service enrichment.
- Apply privacy and redaction rules.
- Publish versioned snapshots or deltas at a controlled cadence.
- Report dropped packets, queue pressure, and collector health.

The collector must not contain display coordinates, colors, animation state, or
other visual decisions.

### TypeScript display

The display owns visual semantics:

- Convert flow summaries into a stable scene model.
- Select layouts, positions, colors, paths, and visual scales.
- Animate direction, volume, activity, and expiration.
- Maintain transitions between collector updates.
- Provide configuration and diagnostic overlays.
- Render with Canvas or WebGL.

The display must not decode packets or independently redefine network
direction, protocol, or service classifications.

## Runtime boundary

The initial implementation will use:

- HTTP for the compiled frontend assets.
- WebSocket for collector updates and health events.
- JSON messages while the contract is small and update rates are modest.
- A schema version on every connection or message envelope.
- A default update cadence of 5–10 collector updates per second.
- A display loop independent of the collector cadence, normally capped at
  30 frames per second on low-power hardware.

The Go binary should embed the compiled frontend using `go:embed`. Node.js is a
development and build dependency, not a production runtime dependency.

The transport contract must remain independent of a particular TypeScript
framework or rendering library. React or Svelte may provide the application
shell, while a Canvas/WebGL renderer such as PixiJS can own the animated scene.

## Deployment modes

The same contract supports three modes:

### Single-device display

The Go collector serves the embedded frontend on loopback. A local kiosk browser
or packaged web view displays it.

### Remote collector

A collector running on a router, gateway, or mirrored network interface sends
flow summaries to a separate display device. Rendering failures cannot affect
routing or capture.

### Recorded replay

A PCAP file produces the same normalized events as live capture. This is the
default development and visual-regression path and does not require elevated
capture permissions.

## Security and privacy

- Bind to loopback by default.
- Run capture with the narrowest OS capabilities available.
- Never require world-writable packet-capture devices.
- Do not transmit packet payloads.
- Treat IP addresses, DNS names, device identifiers, and traffic history as
  sensitive data.
- Make retention and redaction explicit collector policies.
- Require authentication and encrypted transport before enabling remote access
  outside a trusted, isolated network.
- Serve a restrictive Content Security Policy with the embedded frontend.

## Performance constraints

- All queues, caches, and active-flow collections must be bounded.
- Aggregate before serialization; do not send one message per packet.
- Coalesce updates when the display cannot keep up.
- Use logarithmic or capped visual scales so traffic spikes do not create
  unbounded scene objects.
- Cache enrichment results outside the rendering loop.
- Measure collector CPU, display CPU/GPU, memory, dropped packets, and update
  latency on the intended hardware.

Binary serialization may replace JSON only if profiling demonstrates that JSON
is a meaningful bottleneck.

## Initial message shape

The exact schema will be finalized during implementation, but the boundary
should resemble:

```json
{
  "schemaVersion": 1,
  "observedAt": "2026-09-20T05:00:00Z",
  "flows": [
    {
      "id": "stable-flow-id",
      "direction": "outbound",
      "remoteAddress": "203.0.113.10",
      "protocol": "tcp",
      "service": "https",
      "packetsPerSecond": 42,
      "bytesPerSecond": 16384,
      "firstSeen": "2026-09-20T04:59:51Z",
      "lastSeen": "2026-09-20T05:00:00Z",
      "location": {
        "countryCode": "US",
        "latitude": 44.98,
        "longitude": -93.27
      }
    }
  ]
}
```

This is a flow summary, not a persisted event format. Fields that are not
required by the selected visual modes should be omitted.

## Migration sequence

1. Extract capture behind live and PCAP-replay interfaces.
2. Define normalized packet observations and add decoder tests.
3. Add a bounded rolling flow aggregator with deterministic tests.
4. Define and test the versioned WebSocket contract.
5. Build one TypeScript vertical slice that renders replayed flows.
6. Embed the compiled frontend in the Go binary.
7. Add kiosk-mode packaging and measure it on target hardware.
8. Add remote-collector support only after the local mode is stable.

The existing native display can remain as a diagnostic client during migration,
but new visual features should target the TypeScript display.

## Open implementation choices

- Application-shell framework, if any.
- Canvas 2D versus WebGL renderer.
- Flow key and aggregation-window definitions.
- GeoIP/ASN data source and update process.
- Kiosk browser versus packaged web view.
- Authentication mechanism for remote collectors.
- Final target hardware and resource budgets.

