$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

function Fail([string]$Message) {
    Write-Error $Message
    exit 1
}

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Fail "Go is required. Install it with 'winget install --id GoLang.Go -e' or from https://go.dev/dl/."
}

$NpcapInstall = Join-Path $env:ProgramFiles "Npcap\NPFInstall.exe"
if (-not (Test-Path $NpcapInstall)) {
    Fail "Npcap is required. Install the current Npcap runtime from https://npcap.com/#download."
}

$env:CGO_ENABLED = "0"

if (-not (Test-Path ".env")) {
    Copy-Item ".env.example" ".env"
    Write-Host "Created .env from .env.example"
}

New-Item -ItemType Directory -Force "assets\map", "assets\geolitedb", "bin" | Out-Null

if (-not (Test-Path "assets\map\map.geojson")) {
    Invoke-WebRequest `
        -Uri "https://d2ad6b4ur7yvpq.cloudfront.net/naturalearth-3.3.0/ne_50m_admin_0_countries_lakes.geojson" `
        -OutFile "assets\map\map.geojson"
}

if (-not (Test-Path "assets\geolitedb\GeoLite2-City.mmdb")) {
    Invoke-WebRequest `
        -Uri "https://git.io/GeoLite2-City.mmdb" `
        -OutFile "assets\geolitedb\GeoLite2-City.mmdb"
}

go mod download
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

go build -o "bin\capture.exe" .\cmd\capture
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

go build -o "bin\networktrafficvisualizer.exe" .\cmd\networktrafficvisualizer
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

$NpcapParameters = Get-ItemProperty `
    -Path "HKLM:\SYSTEM\CurrentControlSet\Services\npcap\Parameters" `
    -ErrorAction SilentlyContinue
if ($NpcapParameters -and $NpcapParameters.AdminOnly -eq 1) {
    Write-Warning "Npcap is configured for administrator-only capture. Run the app elevated or reinstall Npcap without that restriction."
}

Write-Host ""
Write-Host "Available capture interfaces:"
& ".\bin\capture.exe" --list-interfaces
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host ""
Write-Host "Setup complete."
Write-Host "Test capture: .\bin\capture.exe --count 10"
Write-Host "Run display: .\bin\networktrafficvisualizer.exe"
