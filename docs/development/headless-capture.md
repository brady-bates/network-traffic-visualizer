# Headless capture development

The capture pipeline can run without Ebiten, map assets, GeoIP data, or a
graphical environment. It writes one packet-metadata event per line as
newline-delimited JSON (NDJSON).

The command always uses live `pcap` capture and the same packet transformation
used by the application. It does not start or depend on the renderer.

List available interfaces:

```sh
go run ./cmd/capture --list-interfaces
```

Capture from the first non-loopback interface with an IPv4 address:

```sh
go run ./cmd/capture
```

Select an interface and optional complete BPF filter:

```sh
go run ./cmd/capture --interface en0 --filter "tcp port 443"
```

Live capture still requires the operating system to grant access to the capture
device. Do not make BPF devices world-writable.

## Output

Machine-readable NDJSON is written to standard output. Diagnostics are written
to standard error, so output can be redirected or piped safely:

```sh
go run ./cmd/capture --interface en0 --count 100 > packets.ndjson
```

Each line uses a versioned envelope:

```json
{"schemaVersion":1,"observedAt":"2026-09-20T05:00:00Z","sourceAddress":"192.0.2.1","destinationAddress":"198.51.100.2","direction":"outbound"}
```

This command is a packet-capture development surface, not the future
collector-to-display transport. The display transport will receive bounded,
aggregated flow summaries rather than one event per packet.

