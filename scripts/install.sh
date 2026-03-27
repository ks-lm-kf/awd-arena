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

echo -e "${GREEN}=== AWD Arena Platform - Install ===${NC}"

# Detect OS
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
else
    fail "Unsupported OS"
fi
info "Detected OS: $OS"

# Install Docker
install_docker() {
    if command -v docker &> /dev/null; then
        ok "Docker already installed"
        return
    fi
    info "Installing Docker..."
    if [ "$OS" = "ubuntu" ] || [ "$OS" = "debian" ]; then
        apt-get update
        apt-get install -y ca-certificates curl gnupg
        install -m 0755 -d /etc/apt/keyrings
        curl -fsSL https://download.docker.com/linux/$OS/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
        echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/$OS $(. /etc/os-release && echo $VERSION_CODENAME) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
        apt-get update
        apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
    elif [ "$OS" = "centos" ] || [ "$OS" = "rhel" ]; then
        yum install -y yum-utils
        yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
        yum install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
    fi
    systemctl enable docker
    systemctl start docker
    ok "Docker installed"
}

install_docker

# Install Docker Compose (if not included)
if ! command -v docker compose &> /dev/null; then
    warn "Docker Compose not found, installing plugin..."
    if [ "$OS" = "ubuntu" ] || [ "$OS" = "debian" ]; then
        apt-get install -y docker-compose-plugin
    fi
fi

# Copy binary
if [ -f "build/awd-arena" ]; then
    info "Installing AWD Arena binary..."
    cp build/awd-arena /usr/local/bin/awd-arena
    chmod +x /usr/local/bin/awd-arena
    ok "Binary installed to /usr/local/bin/awd-arena"
fi

# Create systemd service (optional)
if [ "$1" = "--systemd" ]; then
    info "Creating systemd service..."
    cat > /etc/systemd/system/awd-arena.service << 'EOF'
[Unit]
Description=AWD Arena Platform
After=docker.service
Requires=docker.service

[Service]
Type=simple
ExecStart=/usr/local/bin/awd-arena server --config /etc/awd-arena/config.yaml
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
    mkdir -p /etc/awd-arena
    cp configs/config.yaml /etc/awd-arena/config.yaml
    systemctl daemon-reload
    systemctl enable awd-arena
    ok "Systemd service created. Run: systemctl start awd-arena"
fi

echo -e "${GREEN}=== Installation complete ===${NC}"
