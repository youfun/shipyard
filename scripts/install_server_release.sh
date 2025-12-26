#!/bin/bash
set -e

# Configuration
REPO="youfun/deployer-cli"
INSTALL_DIR="/usr/local/bin"
SERVICE_NAME="shipyard"
CONFIG_DIR="/etc/shipyard"
DB_DIR="/var/lib/shipyard"

# Temporary filenames for download
BINARY_NAME="shipyard-server-dl"
CLI_BINARY_NAME="shipyard-dl"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

log() {
    echo -e "${GREEN}[INFO] $1${NC}"
}

error() {
    echo -e "${RED}[ERROR] $1${NC}"
    exit 1
}

# Check for root
if [ "$EUID" -ne 0 ]; then
  error "Please run as root"
fi

# Detect Architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64)
        ARCH="arm64"
        ;;
    *)
        error "Unsupported architecture: $ARCH"
        ;;
esac

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
if [ "$OS" != "linux" ]; then
    error "This script is for Linux only"
fi

# Get Version
VERSION=$1
if [ -z "$VERSION" ]; then
    log "Fetching latest version..."
    VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$VERSION" ]; then
        error "Could not determine latest version"
    fi
    log "Latest version is $VERSION"
fi

# Construct Download URLs
# Assuming assets are named: shipyard-linux-amd64 and shipyard-server-linux-amd64
# Adjust this naming convention if your release assets differ
SERVER_ASSET="shipyard-server-${OS}-${ARCH}"
CLI_ASSET="shipyard-${OS}-${ARCH}"

# Check if binaries exist in release (simple check via curl head)
# Alternatively, just try to download
DOWNLOAD_URL_SERVER="https://github.com/$REPO/releases/download/$VERSION/$SERVER_ASSET"
DOWNLOAD_URL_CLI="https://github.com/$REPO/releases/download/$VERSION/$CLI_ASSET"

log "Downloading $SERVER_ASSET..."
curl -L -o "$INSTALL_DIR/$BINARY_NAME" "$DOWNLOAD_URL_SERVER" || error "Failed to download server binary"

log "Downloading $CLI_ASSET..."
curl -L -o "$INSTALL_DIR/$CLI_BINARY_NAME" "$DOWNLOAD_URL_CLI" || error "Failed to download CLI binary"

# Rename to standard names
mv "$INSTALL_DIR/$BINARY_NAME" "$INSTALL_DIR/shipyard-server"
mv "$INSTALL_DIR/$CLI_BINARY_NAME" "$INSTALL_DIR/shipyard"

chmod +x "$INSTALL_DIR/shipyard-server"
chmod +x "$INSTALL_DIR/shipyard"

log "Binaries installed to $INSTALL_DIR"

# Create Directories
mkdir -p "$CONFIG_DIR"
mkdir -p "$DB_DIR"

# Create Systemd Service
log "Creating systemd service..."

cat > /etc/systemd/system/${SERVICE_NAME}.service <<EOF
[Unit]
Description=Shipyard Server
After=network.target

[Service]
Type=simple
User=root
# Change User to a specific user if desired (recommended for security)
# User=shipyard
# Group=shipyard
ExecStart=$INSTALL_DIR/shipyard-server --port 8080 --config $CONFIG_DIR/shipyard.toml
WorkingDirectory=$DB_DIR
Restart=always
RestartSec=5
Environment="GIN_MODE=release"
# Add other environment variables here
# Environment="PORT=8080"

[Install]
WantedBy=multi-user.target
EOF

# Reload Systemd
systemctl daemon-reload

log "Service created. To start, run:"
log "systemctl enable --now $SERVICE_NAME"
log "Status check: systemctl status $SERVICE_NAME"

# Optional: Create default config if missing
if [ ! -f "$CONFIG_DIR/shipyard.toml" ]; then
    log "Creating default config at $CONFIG_DIR/shipyard.toml"
    touch "$CONFIG_DIR/shipyard.toml"
    # Populate with default values if needed
fi

log "Installation complete!"
