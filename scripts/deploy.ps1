#Requires -Version 5.1
<#
.SYNOPSIS
    AWD Arena Platform - Windows One-Click Deploy
.DESCRIPTION
    Builds and starts the AWD Arena server on Windows.
#>

$ErrorActionPreference = "Stop"

Write-Host "=== AWD Arena Platform - Deploy ===" -ForegroundColor Green

# 1. Check Go
Write-Host "[INFO] Checking Go installation..." -ForegroundColor Blue
$goCmd = Get-Command go -ErrorAction SilentlyContinue
if (-not $goCmd) {
    Write-Host "[ERROR] Go is not installed. Download from https://go.dev/dl/" -ForegroundColor Red
    exit 1
}
$goVersion = & go version
Write-Host "[OK] $goVersion" -ForegroundColor Green

# 2. Check port 8080
Write-Host "[INFO] Checking port 8080..." -ForegroundColor Blue
$portInUse = Get-NetTCPConnection -LocalPort 8080 -ErrorAction SilentlyContinue
if ($portInUse) {
    Write-Host "[ERROR] Port 8080 is already in use by PID $($portInUse.OwningProcess)" -ForegroundColor Red
    exit 1
}
Write-Host "[OK] Port 8080 is available" -ForegroundColor Green

# 3. Create data directory
Write-Host "[INFO] Creating data directory..." -ForegroundColor Blue
New-Item -ItemType Directory -Force -Path "data" | Out-Null
Write-Host "[OK] data/ ready" -ForegroundColor Green

# 4. Build server
Write-Host "[INFO] Building awd-arena.exe..." -ForegroundColor Blue
go build -o awd-arena.exe ./cmd/server
if ($LASTEXITCODE -ne 0) { Write-Host "[ERROR] Build failed" -ForegroundColor Red; exit 1 }
Write-Host "[OK] awd-arena.exe built" -ForegroundColor Green

# 5. Build CLI
Write-Host "[INFO] Building awd-cli.exe..." -ForegroundColor Blue
go build -o awd-cli.exe ./cmd/cli
if ($LASTEXITCODE -ne 0) { Write-Host "[ERROR] CLI build failed" -ForegroundColor Red; exit 1 }
Write-Host "[OK] awd-cli.exe built" -ForegroundColor Green

# 6. Check frontend
Write-Host "[INFO] Checking frontend build..." -ForegroundColor Blue
if (Test-Path "web\dist\index.html") {
    Write-Host "[OK] Frontend already built (web/dist/)" -ForegroundColor Green
} else {
    Write-Host "[WARN] Frontend not built. Run 'cd web && npm install && npm run build' to enable web UI" -ForegroundColor Yellow
}

# 7. Start server
Write-Host "[INFO] Starting AWD Arena server..." -ForegroundColor Blue
$serverProcess = Start-Process -FilePath ".\awd-arena.exe" -PassThru
Start-Sleep -Seconds 2

if (-not $serverProcess.HasExited) {
    Write-Host "[OK] Server started (PID: $($serverProcess.Id))" -ForegroundColor Green
} else {
    Write-Host "[ERROR] Server failed to start" -ForegroundColor Red
    exit 1
}

# 8. Output info
Write-Host ""
Write-Host "=== Deployment Complete ===" -ForegroundColor Green
Write-Host "  Management:  " -NoNewline; Write-Host "http://localhost:8080" -ForegroundColor Blue
Write-Host "  Default:     admin / admin123" -ForegroundColor Cyan
Write-Host "  Stop server: Stop-Process -Id $($serverProcess.Id)" -ForegroundColor DarkGray
