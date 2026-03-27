#Requires -Version 5.1
<#
.SYNOPSIS
    AWD Arena Platform - Build Script (Windows)
.DESCRIPTION
    Builds all platform binaries and packages them into zip files.
#>

$ErrorActionPreference = "Stop"

$VERSION = if ($env:VERSION) { $env:VERSION } else { "0.1.0" }
$BUILD_DIR = "build"
$DIST_DIR = "dist"

Write-Host "=== AWD Arena Platform - Build ($VERSION) ===" -ForegroundColor Green

# Clean
if (Test-Path $BUILD_DIR) { Remove-Item -Recurse -Force $BUILD_DIR }
if (Test-Path $DIST_DIR) { Remove-Item -Recurse -Force $DIST_DIR }
New-Item -ItemType Directory -Force -Path $BUILD_DIR, $DIST_DIR | Out-Null

$LDFLAGS = "-s -w -X main.version=$VERSION"

# Build Linux amd64
Write-Host "[INFO] Building Linux amd64..." -ForegroundColor Blue
$env:GOOS = "linux"; $env:GOARCH = "amd64"
go build -ldflags $LDFLAGS -o "$BUILD_DIR/awd-arena" ./cmd/server
go build -ldflags $LDFLAGS -o "$BUILD_DIR/awd-cli" ./cmd/cli
go build -ldflags $LDFLAGS -o "$BUILD_DIR/awd-migrator" ./cmd/migrator
Write-Host "[OK] Linux build complete" -ForegroundColor Green

# Build Windows amd64
Write-Host "[INFO] Building Windows amd64..." -ForegroundColor Blue
$env:GOOS = "windows"; $env:GOARCH = "amd64"
go build -ldflags $LDFLAGS -o "$BUILD_DIR/awd-arena.exe" ./cmd/server
go build -ldflags $LDFLAGS -o "$BUILD_DIR/awd-cli.exe" ./cmd/cli
go build -ldflags $LDFLAGS -o "$BUILD_DIR/awd-migrator.exe" ./cmd/migrator
Write-Host "[OK] Windows build complete" -ForegroundColor Green

# Build frontend
Write-Host "[INFO] Building frontend..." -ForegroundColor Blue
if (Test-Path "web\package.json") {
    Push-Location web
    if (-not (Test-Path "node_modules")) {
        npm install
        if ($LASTEXITCODE -ne 0) { Write-Host "[WARN] npm install failed, skipping frontend" -ForegroundColor Yellow; Pop-Location; $frontendOk = $false }
    }
    if ((Test-Path "node_modules") -and (-not $frontendOk)) {
        npm run build
        if ($LASTEXITCODE -eq 0) {
            Copy-Item -Recurse dist "$BUILD_DIR\web-dist"
            Write-Host "[OK] Frontend built" -ForegroundColor Green
            $frontendOk = $true
        } else {
            Write-Host "[WARN] Frontend build failed" -ForegroundColor Yellow
            $frontendOk = $false
        }
    }
    Pop-Location
} else {
    Write-Host "[WARN] No frontend directory found" -ForegroundColor Yellow
    $frontendOk = $false
}

# Package Windows
Write-Host "[INFO] Packaging Windows zip..." -ForegroundColor Blue
$winZip = "$DIST_DIR\awd-arena-$VERSION-windows-amd64.zip"
Compress-Archive -Path "$BUILD_DIR\awd-arena.exe", "$BUILD_DIR\awd-cli.exe", "$BUILD_DIR\awd-migrator.exe" -DestinationPath $winZip
if ($frontendOk) {
    # Append web-dist to zip
    $tempDir = "$BUILD_DIR\win-pkg"
    Copy-Item -Recurse $BUILD_DIR\web-dist "$tempDir\web-dist"
    Compress-Archive -Path "$BUILD_DIR\awd-arena.exe", "$BUILD_DIR\awd-cli.exe", "$BUILD_DIR\awd-migrator.exe", "$tempDir\web-dist" -DestinationPath $winZip -Force
    Remove-Item -Recurse -Force $tempDir
}
# Add configs, scripts, migrations to zip
$pkgRoot = "$BUILD_DIR\win-pkg"
New-Item -ItemType Directory -Force -Path $pkgRoot | Out-Null
Copy-Item -Recurse configs "$pkgRoot\configs"
Copy-Item -Recurse scripts "$pkgRoot\scripts"
Copy-Item -Recurse migrations "$pkgRoot\migrations"
if ($frontendOk) { Copy-Item -Recurse "$BUILD_DIR\web-dist" "$pkgRoot\web-dist" }
Compress-Archive -Path "$BUILD_DIR\awd-arena.exe", "$BUILD_DIR\awd-cli.exe", "$BUILD_DIR\awd-migrator.exe", "$pkgRoot\configs", "$pkgRoot\scripts", "$pkgRoot\migrations" -DestinationPath $winZip -Force
Remove-Item -Recurse -Force $pkgRoot
Write-Host "[OK] $winZip" -ForegroundColor Green

Write-Host ""
Write-Host "=== Build Complete ===" -ForegroundColor Green
Get-ChildItem $DIST_DIR | ForEach-Object { Write-Host "  $($_.Name) ($([math]::Round($_.Length/1MB, 1)) MB)" }
