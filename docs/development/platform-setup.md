# Platform setup

The application uses `gopacket/pcap`, which requires each operating system's
native packet-capture runtime and permissions. The setup scripts build two
executables:

- `capture`: live capture and transformed NDJSON output without a UI.
- `networktrafficart`: live capture with the Ebiten display.

Run executables from the repository root so `.env` and the external assets are
available.

## Interface discovery

Interface selection is shared by both executables:

1. Use `CAPTURE_INTERFACE` or `--interface` when explicitly configured.
2. Otherwise, determine the local IPv4 address used by the default route.
3. Match that address to the interfaces returned by libpcap/Npcap.
4. Fall back to the first private, non-loopback IPv4 interface.

Explicit overrides accept either the pcap device name or its description. This
is especially useful on Windows, where names are commonly
`\Device\NPF_{GUID}` but descriptions are human-readable.

Inspect available values with:

```sh
./bin/capture --list-interfaces
```

## macOS

Requirements:

- Go
- Xcode command-line tools
- BPF device access

Setup:

```sh
./scripts/setup_macos.sh
```

The secure temporary permission route is:

```sh
./scripts/setup_macos.sh --grant-capture
```

It makes the current user the owner of the existing `/dev/bpf*` devices with
mode `600`. macOS recreates those permissions after reboot. For persistent
group-based access, install the ChmodBPF package distributed with Wireshark.

Do not use the legacy `chmod 666 /dev/bpf*` approach.

## Linux

Requirements:

- Go
- A C compiler and `pkg-config`
- libpcap development headers
- X11/OpenGL and audio development libraries for Ebiten
- `setcap` for non-root capture

Install supported distribution packages and build:

```sh
./scripts/setup_linux.sh --install-deps
```

Grant only the capabilities required by the built binaries:

```sh
./scripts/setup_linux.sh --grant-capture
```

Both may be combined:

```sh
./scripts/setup_linux.sh --install-deps --grant-capture
```

Capabilities are attached to the built files. Rebuilding or replacing a binary
may require running the grant step again.

## Windows

Requirements:

- Go
- Npcap runtime

Install Go:

```powershell
winget install --id GoLang.Go -e
```

Download and install the current Npcap runtime from
[npcap.com](https://npcap.com/#download). Npcap's default non-WinPcap
compatibility mode is appropriate; `gopacket` explicitly searches the Npcap
DLL directory.

Run setup:

```powershell
PowerShell -ExecutionPolicy Bypass -File .\scripts\setup_windows.ps1
```

The Windows implementation loads Npcap dynamically and does not require CGO, a
C compiler, or the Npcap SDK. The script downloads assets, builds both
executables, reports administrator-only Npcap installations, and lists the
interfaces visible to Npcap.

If Npcap was installed with administrator-only access, either run the
application elevated or reinstall Npcap without that restriction. Running
without unnecessary elevation is preferred.

## Asset sources

The setup scripts download:

- Natural Earth country boundaries used by the map display.
- A GeoLite2 City database from the same URL used by the existing project
  scripts.

These files remain external and ignored by Git. Review their licenses and
distribution requirements before packaging the application for others.

