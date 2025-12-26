#!/bin/bash

# ==============================================================================
# Script Name: install-shipyard-server.sh
# Description: Download and install shipyard-server from GitHub Release to system
# Usage:       curl -fsSL https://raw.githubusercontent.com/youfun/shipyard/main/scripts/install-shipyard-server.sh | bash
#             or: ./install-shipyard-server.sh [version]
# ==============================================================================

set -e
set -o pipefail

# GitHub repository configuration
GITHUB_REPO="youfun/shipyard"  
BINARY_NAME="shipyard-server-linux-amd64"

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print information functions
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_prompt() {
    echo -e "${BLUE}[?]${NC} $1"
}

# --- Detect system architecture ---
detect_arch() {
    local arch=$(uname -m)
    case $arch in
        x86_64|amd64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        *)
            print_error "Unsupported architecture: $arch"
            exit 1
            ;;
    esac
}

# --- Get latest version or use specified version ---
get_version() {
    local version="${1:-latest}"
    if [ "$version" = "latest" ]; then
        print_info "Getting latest version information..."
        version=$(curl -fsSL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
        if [ -z "$version" ]; then
            print_error "Unable to get latest version information"
            exit 1
        fi
    fi
    echo "$version"
}

# --- Download binary file ---
download_binary() {
    local version=$1
    local download_url="https://github.com/${GITHUB_REPO}/releases/download/${version}/${BINARY_NAME}"
    
    print_info "Download URL: ${download_url}"
    print_info "Downloading ${BINARY_NAME}..."
    
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL -o "${BINARY_NAME}" "${download_url}"
    elif command -v wget >/dev/null 2>&1; then
        wget -q -O "${BINARY_NAME}" "${download_url}"
    else
        print_error "curl or wget is required to download files"
        exit 1
    fi
    
    if [ ! -f "${BINARY_NAME}" ]; then
        print_error "Download failed"
        exit 1
    fi
    
    chmod +x "${BINARY_NAME}"
    print_info "Download completed"
}

# --- Ask for installation mode ---
echo ""
echo "======================================"
echo "  Shipyard Server Installation Wizard"
echo "======================================"
echo ""
print_prompt "Please select installation mode:"
echo "  1) Standard Mode - Install to system directories (requires root privileges)"
echo "  2) Test Mode - Install to current directory (no root privileges required)"
echo ""
read -p "Please enter your choice [1/2]: " INSTALL_MODE

case "$INSTALL_MODE" in
    1)
        print_info "Selected: Standard Mode"
        IS_TEST_MODE=false
        ;;
    2)
        print_info "Selected: Test Mode"
        IS_TEST_MODE=true
        ;;
    *)
        print_error "Invalid choice, defaulting to Standard Mode"
        IS_TEST_MODE=false
        ;;
esac
echo ""

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Get version parameter
VERSION="${1:-latest}"

# Configuration variables
INSTALL_NAME="shipyard-server"
SERVICE_USER="shipyard"
SERVICE_GROUP="shipyard"

if [ "$IS_TEST_MODE" = true ]; then
    # Test mode: all paths under current directory
    INSTALL_DIR="${SCRIPT_DIR}"
    CONFIG_DIR="${SCRIPT_DIR}/config"
    DATA_DIR="${SCRIPT_DIR}/data"
    LOG_DIR="${SCRIPT_DIR}/logs"
    USE_SYSTEMD=false
    print_info "Test mode configuration:"
    echo "  - Install directory: ${INSTALL_DIR}"
    echo "  - Config directory: ${CONFIG_DIR}"
    echo "  - Data directory: ${DATA_DIR}"
    echo "  - Log directory: ${LOG_DIR}"
    echo "  - systemd service: No"
else
    # Standard mode: system directories
    INSTALL_DIR="/usr/local/bin"
    CONFIG_DIR="/etc/shipyard"
    DATA_DIR="/var/lib/shipyard"
    LOG_DIR="/var/log/shipyard"
    USE_SYSTEMD=true
    print_info "Standard mode configuration:"
    echo "  - Install directory: ${INSTALL_DIR}"
    echo "  - Config directory: ${CONFIG_DIR}"
    echo "  - Data directory: ${DATA_DIR}"
    echo "  - Log directory: ${LOG_DIR}"
    echo "  - systemd service: Yes"
fi
echo ""

# --- Permission check ---
if [ "$IS_TEST_MODE" = false ]; then
    if [ "$(id -u)" -ne 0 ]; then
        print_error "Standard mode requires root privileges to run"
        echo "Please use: sudo $0"
        exit 1
    fi
else
    print_warn "Running in test mode, will skip operations requiring root privileges"
fi

# --- Download binary file ---
print_info "Preparing to download shipyard-server..."
ARCH=$(detect_arch)
BINARY_NAME="shipyard-server-linux-${ARCH}"
VERSION=$(get_version "$VERSION")
print_info "Version: ${VERSION}"
print_info "Architecture: ${ARCH}"

# Temporary download directory
TMP_DIR=$(mktemp -d)
trap "rm -rf ${TMP_DIR}" EXIT
cd "${TMP_DIR}"

download_binary "$VERSION"

# --- Check binary file ---
print_info "Verifying binary file..."
BINARY_PATH="${TMP_DIR}/${BINARY_NAME}"

if [ ! -f "$BINARY_PATH" ]; then
    print_error "Verification failed: binary file does not exist"
    exit 1
fi

if [ ! -x "$BINARY_PATH" ]; then
    print_warn "Binary file is not executable, adding execute permission..."
    chmod +x "$BINARY_PATH"
fi

# Show version information
print_info "Binary file version information:"
"$BINARY_PATH" --version || print_warn "Unable to get version information"
echo ""

# --- Create system user and group ---
if [ "$IS_TEST_MODE" = false ]; then
    print_info "Creating system user and group..."
    if ! getent group "$SERVICE_GROUP" > /dev/null 2>&1; then
        print_info "Creating group: $SERVICE_GROUP"
        groupadd --system "$SERVICE_GROUP"
    else
        print_info "Group already exists: $SERVICE_GROUP"
    fi

    if ! id "$SERVICE_USER" > /dev/null 2>&1; then
        print_info "Creating user: $SERVICE_USER"
        useradd --system --gid "$SERVICE_GROUP" \
            --home-dir "$DATA_DIR" \
            --shell /usr/sbin/nologin \
            --comment "Shipyard Server Service User" \
            "$SERVICE_USER"
    else
        print_info "User already exists: $SERVICE_USER"
    fi
else
    print_info "Test mode: skipping user and group creation"
fi

# --- Install binary file ---
if [ "$IS_TEST_MODE" = true ]; then
    print_info "Test mode: binary file is already in current directory"
    # Ensure executable
    chmod +x "$BINARY_PATH"
    print_info "Set executable permission: ${BINARY_PATH}"
else
    print_info "Installing binary file to ${INSTALL_DIR}..."
    cp "$BINARY_PATH" "${INSTALL_DIR}/${INSTALL_NAME}"
    chown root:root "${INSTALL_DIR}/${INSTALL_NAME}"
    chmod 755 "${INSTALL_DIR}/${INSTALL_NAME}"
    print_info "Installed: ${INSTALL_DIR}/${INSTALL_NAME}"
fi

# --- Create directories ---
print_info "Creating required directories..."
mkdir -p "$CONFIG_DIR"
mkdir -p "$DATA_DIR"
mkdir -p "$LOG_DIR"

if [ "$IS_TEST_MODE" = false ]; then
    chown -R root:${SERVICE_GROUP} "$CONFIG_DIR"
    chmod 750 "$CONFIG_DIR"
    
    chown -R ${SERVICE_USER}:${SERVICE_GROUP} "$DATA_DIR"
    chmod 750 "$DATA_DIR"
    
    chown -R ${SERVICE_USER}:${SERVICE_GROUP} "$LOG_DIR"
    chmod 750 "$LOG_DIR"
else
    chmod 750 "$CONFIG_DIR"
    chmod 750 "$DATA_DIR"
    chmod 750 "$LOG_DIR"
fi

print_info "Directories created:"
echo "  - Config directory: $CONFIG_DIR"
echo "  - Data directory: $DATA_DIR"
echo "  - Log directory: $LOG_DIR"

# --- Create environment configuration file ---
print_info "Creating environment configuration file..."
ENV_FILE="${CONFIG_DIR}/.env"
if [ ! -f "$ENV_FILE" ]; then
    cat > "$ENV_FILE" <<EOF
# Shipyard Server Environment Configuration
# Important: Please change JWT_SECRET to your own random key

# JWT Secret (must be set, used for token signing)
# Generate random key: openssl rand -base64 32
JWT_SECRET=PLEASE_CHANGE_THIS_TO_RANDOM_SECRET

# Database type: sqlite, postgres, turso
DB_TYPE=sqlite

# SQLite database path (when using sqlite)
DB_PATH=${DATA_DIR}/deploy.db

# PostgreSQL connection string (when using postgres)
# DATABASE_URL=postgres://user:password@localhost:5432/shipyard?sslmode=disable

# Turso configuration (when using turso)
# TURSO_DATABASE_URL=libsql://your-db.turso.io
# TURSO_AUTH_TOKEN=your-auth-token

# Server port
SERVER_PORT=8080

# Log level: debug, info, warn, error
LOG_LEVEL=info
EOF
    if [ "$IS_TEST_MODE" = false ]; then
        chown root:${SERVICE_GROUP} "$ENV_FILE"
        chmod 640 "$ENV_FILE"
    else
        chmod 640 "$ENV_FILE"
    fi
    print_warn "Created environment configuration file: $ENV_FILE"
    print_warn "Please edit this file and set JWT_SECRET!"
else
    print_info "Environment configuration file already exists: $ENV_FILE"
fi

# --- Create systemd service file ---
if [ "$USE_SYSTEMD" = true ]; then
    print_info "Creating systemd service..."
    SERVICE_FILE="/etc/systemd/system/shipyard-server.service"
cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=Shipyard Server - Automated Deployment Tool Server
Documentation=https://github.com/yourusername/shipyard
After=network.target network-online.target
Requires=network-online.target

[Service]
Type=simple
User=${SERVICE_USER}
Group=${SERVICE_GROUP}
WorkingDirectory=${DATA_DIR}
EnvironmentFile=${CONFIG_DIR}/.env

# Start command
ExecStart=${INSTALL_DIR}/${INSTALL_NAME} --port \${SERVER_PORT:-8080}

# Restart policy
Restart=on-failure
RestartSec=5s

# Resource limits
LimitNOFILE=65536
LimitNPROC=4096

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=${DATA_DIR} ${LOG_DIR}

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=shipyard-server

[Install]
WantedBy=multi-user.target
EOF

    chown root:root "$SERVICE_FILE"
    chmod 644 "$SERVICE_FILE"
    print_info "Created service file: $SERVICE_FILE"

    # --- Reload systemd ---
    print_info "Reloading systemd configuration..."
    systemctl daemon-reload

    # --- Enable service (but don't auto-start) ---
    print_info "Enabling shipyard-server service..."
    systemctl enable shipyard-server
else
    print_info "Test mode: skipping systemd service creation"
fi

# --- Complete ---
echo ""
print_info "=============================================="
print_info "Shipyard Server Installation Successful!"
print_info "=============================================="
echo ""
echo "Installation Information:"
if [ "$IS_TEST_MODE" = true ]; then
    echo "  - Installation mode: Test Mode"
    echo "  - Binary file: ${BINARY_PATH}"
else
    echo "  - Installation mode: Standard Mode"
    echo "  - Executable file: ${INSTALL_DIR}/${INSTALL_NAME}"
fi
echo "  - Config directory: ${CONFIG_DIR}"
echo "  - Data directory: ${DATA_DIR}"
echo "  - Log directory: ${LOG_DIR}"
if [ "$USE_SYSTEMD" = true ]; then
    echo "  - systemd service: shipyard-server.service"
fi
echo ""
echo "Next Steps:"
echo ""
echo -e "${YELLOW}1. Edit environment configuration file (required):${NC}"
if [ "$IS_TEST_MODE" = true ]; then
    echo "   nano ${CONFIG_DIR}/.env"
else
    echo "   sudo nano ${CONFIG_DIR}/.env"
fi
echo "   ${RED}Please change JWT_SECRET to a random key!${NC}"
echo "   Generate key: openssl rand -base64 32"
echo ""
if [ "$IS_TEST_MODE" = true ]; then
    echo -e "${YELLOW}2. Start service (test mode):${NC}"
    echo "   cd ${SCRIPT_DIR}"
    echo "   ./${BINARY_NAME} --port 8080"
    echo ""
    echo -e "${YELLOW}3. Or use environment variables:${NC}"
    echo "   export \$(cat ${CONFIG_DIR}/.env | xargs)"
    echo "   ./${BINARY_NAME}"
else
    echo -e "${YELLOW}2. Start service:${NC}"
    echo "   sudo systemctl start shipyard-server"
    echo ""
    echo -e "${YELLOW}3. Check service status:${NC}"
    echo "   sudo systemctl status shipyard-server"
    echo ""
    echo -e "${YELLOW}4. View logs:${NC}"
    echo "   sudo journalctl -u shipyard-server -f"
    echo ""
    echo -e "${YELLOW}5. Enable auto-start on boot (optional):${NC}"
    echo "   sudo systemctl enable shipyard-server"
fi
echo ""
echo -e "${GREEN}Access Web UI: http://localhost:8080${NC}"
echo ""

exit 0
