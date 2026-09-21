#!/usr/bin/env bash

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GRANT_CAPTURE=false

if [[ "${1:-}" == "--grant-capture" ]]; then
  GRANT_CAPTURE=true
elif [[ -n "${1:-}" ]]; then
  echo "Usage: $0 [--grant-capture]" >&2
  exit 2
fi

cd "$ROOT"

if ! command -v go >/dev/null 2>&1; then
  echo "Go is required: https://go.dev/dl/" >&2
  exit 1
fi

if ! xcode-select -p >/dev/null 2>&1; then
  echo "Xcode command-line tools are required. Run: xcode-select --install" >&2
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
  sudo chown "$USER":staff /dev/bpf*
  sudo chmod 600 /dev/bpf*
  echo "Granted $USER temporary owner-only BPF access. macOS resets this after reboot."
elif [[ ! -r /dev/bpf0 ]]; then
  echo
  echo "Capture access is not configured."
  echo "Permanent route: install Wireshark's ChmodBPF package."
  echo "Temporary route: $0 --grant-capture"
fi

echo
echo "Setup complete."
echo "Test capture: ./bin/capture --count 10"
echo "Run display: ./bin/networktrafficvisualizer"
