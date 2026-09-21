# Network Traffic Art

Network Traffic Art captures local network metadata and renders remote
connections as a live visual display.

## Quick setup

Run the setup script for the target operating system from the repository root.
The scripts create `.env`, download the map and GeoIP assets, download Go
modules, and build the capture and display executables under `bin/`.

### macOS

```sh
./scripts/setup_macos.sh --grant-capture
./bin/capture --count 10
./bin/networktrafficvisualizer
```

`--grant-capture` applies owner-only BPF permissions until the next reboot. For
permanent access, install Wireshark's ChmodBPF package.

### Linux

```sh
./scripts/setup_linux.sh --install-deps --grant-capture
./bin/capture --count 10
./bin/networktrafficvisualizer
```

The setup script supports Debian/Ubuntu, Fedora, and Arch package managers.
The display requires a graphical X11 or XWayland session.

### Windows

Install these prerequisites first:

1. [Go](https://go.dev/dl/)
2. [Npcap runtime](https://npcap.com/#download)

Then run:

```powershell
PowerShell -ExecutionPolicy Bypass -File .\scripts\setup_windows.ps1
.\bin\capture.exe --count 10
.\bin\networktrafficvisualizer.exe
```

The Windows build is pure Go. It does not require GCC, CGO, or the Npcap SDK;
Npcap is loaded dynamically when the executable starts.

## Interface selection

Both executables automatically select the interface carrying the default IPv4
route. This works with Unix interface names such as `en0` and `eth0`, as well as
Windows Npcap names such as `\Device\NPF_{GUID}`.

List every interface:

```sh
./bin/capture --list-interfaces
```

Override selection for the capture command:

```sh
./bin/capture --interface en0
```

Override selection for the display by setting `CAPTURE_INTERFACE` in `.env`.
The override may be the pcap interface name or its displayed description.

## Headless data output

The capture command runs the real capture and transformation path without
starting Ebiten:

```sh
./bin/capture --count 100
./bin/capture --filter "tcp port 443"
```

Transformed events are written to standard output as NDJSON. Diagnostics are
written to standard error.

See [platform setup](docs/development/platform-setup.md) and
[headless capture development](docs/development/headless-capture.md) for
additional details.

