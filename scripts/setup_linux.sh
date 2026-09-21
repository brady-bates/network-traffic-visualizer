#!/usr/bin/env bash

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INSTALL_DEPS=false
GRANT_CAPTURE=false

for argument in "$@"; do
  case "$argument" in
    --install-deps) INSTALL_DEPS=true ;;
    --grant-capture) GRANT_CAPTURE=true ;;
    *)
      echo "Usage: $0 [--install-deps] [--grant-capture]" >&2
      exit 2
      ;;
  esac
done

install_dependencies() {
  if command -v apt-get >/dev/null 2>&1; then
    sudo apt-get update
    sudo apt-get install -y \
      build-essential pkg-config curl libpcap-dev libcap2-bin \
      libgl1-mesa-dev libx11-dev libxcursor-dev libxrandr-dev \
      libxinerama-dev libxi-dev libxxf86vm-dev libasound2-dev libudev-dev
  elif command -v dnf >/dev/null 2>&1; then
    sudo dnf install -y \
      gcc gcc-c++ pkgconf-pkg-config curl libpcap-devel libcap \
      mesa-libGL-devel libX11-devel libXcursor-devel libXrandr-devel \
      libXinerama-devel libXi-devel libXxf86vm-devel alsa-lib-devel \
      systemd-devel
  elif command -v pacman >/dev/null 2>&1; then
    sudo pacman -S --needed \
      base-devel pkgconf curl libpcap libcap mesa libx11 libxcursor \
      libxrandr libxinerama libxi libxxf86vm alsa-lib systemd-libs
  else
    echo "Unsupported package manager. Install a C compiler, pkg-config, libpcap development headers, libcap, X11/OpenGL, and ALSA development libraries." >&2
    exit 1
  fi
}

if [[ "$INSTALL_DEPS" == true ]]; then
  install_dependencies
fi

cd "$ROOT"

for command in go curl pkg-config; do
  if ! command -v "$command" >/dev/null 2>&1; then
    echo "$command is required. Re-run with --install-deps where supported." >&2
    exit 1
  fi
done

if ! pkg-config --exists libpcap; then
  echo "libpcap development files are missing. Re-run with --install-deps." >&2
  exit 1
fi

if [[ ! -f .env ]]; then
  cp .env.example .env
  echo "Created .env from .env.example"
fi

mkdir -p assets/map assets/geolitedb bin

if [[ ! -f assets/map/map.geojson ]]; then
  curl --fail --location \
    --output assets/map/map.geojson \
    "https://d2ad6b4ur7yvpq.cloudfront.net/naturalearth-3.3.0/ne_50m_admin_0_countries_lakes.geojson"
fi

if [[ ! -f assets/geolitedb/GeoLite2-City.mmdb ]]; then
  curl --fail --location \
    --output assets/geolitedb/GeoLite2-City.mmdb \
    "https://git.io/GeoLite2-City.mmdb"
fi

go mod download
go build -o bin/capture ./cmd/capture
go build -o bin/networktrafficvisualizer ./cmd/networktrafficvisualizer

if [[ "$GRANT_CAPTURE" == true ]]; then
  if ! command -v setcap >/dev/null 2>&1; then
    echo "setcap is required. Install libcap2-bin or your distribution's libcap package." >&2
    exit 1
  fi
  sudo setcap cap_net_raw,cap_net_admin=eip bin/capture bin/networktrafficvisualizer
  echo "Granted packet-capture capabilities to the built binaries."
else
  echo
  echo "Capture permissions are not changed automatically."
  echo "Re-run with --grant-capture to grant capabilities to the built binaries."
fi

echo
echo "Setup complete."
echo "Test capture: ./bin/capture --count 10"
echo "Run display from a graphical session: ./bin/networktrafficvisualizer"
