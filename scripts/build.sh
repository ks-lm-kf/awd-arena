#!/bin/bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

info()  { echo -e "${BLUE}[INFO]${NC} $1"; }
ok()    { echo -e "${GREEN}[OK]${NC} $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
fail()  { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

echo -e "${GREEN}=== AWD Arena Platform - Build ===${NC}"

# Detect Go
if ! command -v go &> /dev/null; then
    fail "Go is not installed. Install from https://go.dev/dl/"
fi

GO_VERSION=$(go version | awk '{print $3}')
info "Go version: $GO_VERSION"

VERSION=${VERSION:-0.1.0}
BUILD_DIR="build"
DIST_DIR="dist"
LDFLAGS="-s -w -X main.version=${VERSION}"

rm -rf "$BUILD_DIR" "$DIST_DIR"
mkdir -p "$BUILD_DIR" "$DIST_DIR"

# Build Linux
info "Building Linux amd64..."
GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "$BUILD_DIR/awd-arena" ./cmd/server
GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "$BUILD_DIR/awd-cli" ./cmd/cli
GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "$BUILD_DIR/awd-migrator" ./cmd/migrator
ok "Linux build complete"

# Build Windows
info "Building Windows amd64..."
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "$BUILD_DIR/awd-arena.exe" ./cmd/server
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "$BUILD_DIR/awd-cli.exe" ./cmd/cli
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "$BUILD_DIR/awd-migrator.exe" ./cmd/migrator
ok "Windows build complete"

# Build frontend
FRONTEND_OK=false
if [ -d "web" ] && [ -f "web/package.json" ]; then
    info "Building frontend..."
    cd web
    if [ ! -d "node_modules" ]; then
        npm install
    fi
    if npm run build; then
        cp -r dist "$BUILD_DIR/web-dist"
        FRONTEND_OK=true
        ok "Frontend built"
    else
        warn "Frontend build failed, continuing without it"
    fi
    cd ..
else
    warn "No frontend directory found, skipping"
fi

# Package Linux
info "Packaging Linux tar.gz..."
LINUX_PKG="$BUILD_DIR/pkg-linux"
mkdir -p "$LINUX_PKG"
cp "$BUILD_DIR/awd-arena" "$BUILD_DIR/awd-cli" "$BUILD_DIR/awd-migrator" "$LINUX_PKG/"
cp -r configs "$LINUX_PKG/"
cp -r scripts "$LINUX_PKG/"
cp -r migrations "$LINUX_PKG/"
if [ "$FRONTEND_OK" = true ]; then cp -r "$BUILD_DIR/web-dist" "$LINUX_PKG/web-dist"; fi
tar -czf "$DIST_DIR/awd-arena-${VERSION}-linux-amd64.tar.gz" -C "$BUILD_DIR" "pkg-linux" --transform="s/^pkg-linux/awd-arena-${VERSION}/"
ok "Linux package: $DIST_DIR/awd-arena-${VERSION}-linux-amd64.tar.gz"

# Package Windows
info "Packaging Windows zip..."
WIN_PKG="$BUILD_DIR/pkg-windows"
mkdir -p "$WIN_PKG"
cp "$BUILD_DIR/awd-arena.exe" "$BUILD_DIR/awd-cli.exe" "$BUILD_DIR/awd-migrator.exe" "$WIN_PKG/"
cp -r configs "$WIN_PKG/"
cp -r scripts "$WIN_PKG/"
cp -r migrations "$WIN_PKG/"
if [ "$FRONTEND_OK" = true ]; then cp -r "$BUILD_DIR/web-dist" "$WIN_PKG/web-dist"; fi
(cd "$BUILD_DIR" && zip -q -r "../$DIST_DIR/awd-arena-${VERSION}-windows-amd64.zip" "pkg-windows" -x "*.DS_Store")
ok "Windows package: $DIST_DIR/awd-arena-${VERSION}-windows-amd64.zip"

echo -e "${GREEN}=== Build complete ===${NC}"
ls -lh "$DIST_DIR/"
