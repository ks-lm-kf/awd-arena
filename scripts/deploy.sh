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

echo -e "${GREEN}=== AWD Arena Platform - Deploy ===${NC}"

# Check port 8080
check_port() {
    if ss -tlnp 2>/dev/null | grep -q ":$1 "; then
        fail "Port $1 is already in use"
    fi
    ok "Port $1 is available"
}

check_port 8080

# Ensure data directory exists
mkdir -p data

# Build server
info "Building AWD Arena..."
go build -o build/awd-arena ./cmd/server
go build -o build/awd-cli ./cmd/cli
ok "Build complete"

# Start server
info "Starting AWD Arena server..."
./build/awd-arena &
SERVER_PID=$!
sleep 1

if kill -0 $SERVER_PID 2>/dev/null; then
    ok "Server started (PID: $SERVER_PID)"
else
    fail "Server failed to start"
fi

echo -e "${GREEN}=== Deployment complete ===${NC}"
echo -e "  Management: ${BLUE}http://localhost:8080${NC}"
echo -e "  Default: admin / admin123"
echo -e "  Stop: kill $SERVER_PID"
