#!/bin/bash

# ==============================================================================
# Script Name: install_caddy_universal.sh
# Description: Download and install the latest Caddy from GitHub Releases on mainstream Linux systems.
#              This script auto-detects the package manager (APT, YUM/DNF, Pacman) and installs required dependencies.
# Author:     Gemini
# Date:       2025-08-30
# ==============================================================================

# Exit immediately if a command exits with a non-zero status, and propagate the failure to the parent shell
set -e
set -o pipefail

# --- Permission check ---
if [ "$(id -u)" -ne 0 ]; then
  echo "Error: This script must be run as root." >&2
  echo "Try: 'sudo ./install_caddy_universal.sh'." >&2
  exit 1
fi

# --- Dependency installation ---
echo ">>> Step 1: Detect package manager and install dependencies (curl)..."
PKG_MANAGER=""
INSTALL_CMD=""

if command -v apt-get &> /dev/null; then
  PKG_MANAGER="apt"
  INSTALL_CMD="apt-get install -y"
  echo "Detected Debian/Ubuntu (apt). Updating package list..."
  apt-get update
elif command -v dnf &> /dev/null; then
  PKG_MANAGER="dnf"
  INSTALL_CMD="dnf install -y"
  echo "Detected Fedora/RHEL/AlmaLinux/Rocky (dnf)..."
  # Check specific distribution
  if [ -f /etc/os-release ]; then
    source /etc/os-release
    echo "Detected distribution: $NAME $VERSION_ID"
  fi
elif command -v pacman &> /dev/null; then
  PKG_MANAGER="pacman"
  # Pacman requires non-interactive arguments
  INSTALL_CMD="pacman -S --noconfirm"
  echo "Detected Arch Linux (pacman)..."
else
  echo "Error: Unknown package manager. Please install 'curl' manually and re-run the script." >&2
  exit 1
fi

# Check and install curl
if ! command -v curl &> /dev/null; then
  echo "Installing 'curl'..."
  ${INSTALL_CMD} curl
else
  echo "'curl' is already installed."
fi

# --- Detect system architecture and build download URL ---
echo ""
echo ">>> Step 2: Detect system architecture..."

case "$(uname -m)" in
  x86_64) ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  armv6l | armv7l) ARCH="armv7" ;;
  *)
    echo "Error: Unsupported architecture: $(uname -m)" >&2
    exit 1
    ;;
esac
echo "Detected architecture: $ARCH"
echo ">>> Building download URL..."
DOWNLOAD_URL="https://github.com/openmindw/caddy-cloudflare/releases/download/latest/caddy-cloudflare-linux-${ARCH}.tar.gz"
echo "Download URL: ${DOWNLOAD_URL}"

# --- download and install ---
echo ""
echo ">>> Step 3: Download and install Caddy (Cloudflare) binary..."
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

curl -sSL "$DOWNLOAD_URL" -o "caddy.tar.gz"

# List archive contents for debugging
echo "Archive contents:"
tar -tzf "caddy.tar.gz"

# Extract all contents
tar -xzf "caddy.tar.gz"

# Find caddy executable
CADDY_BINARY=""
if [ -f "caddy" ]; then
    CADDY_BINARY="caddy"
elif [ -f "caddy-linux-${ARCH}" ]; then
    CADDY_BINARY="caddy-linux-${ARCH}"
elif [ -f "caddy-cloudflare" ]; then
    CADDY_BINARY="caddy-cloudflare"
else
    # check bin 
    CADDY_BINARY=$(find . -type f -executable -name "*caddy*" | head -1)
    if [ -z "$CADDY_BINARY" ]; then
        echo "Error: caddy executable not found in archive." >&2
        echo "Archive contents:"
        ls -la
        exit 1
    fi
    CADDY_BINARY=$(basename "$CADDY_BINARY")
fi

echo "Found Caddy executable: $CADDY_BINARY"
echo "Moving Caddy to /usr/local/bin/..."
mv "$CADDY_BINARY" /usr/local/bin/caddy
chown root:root /usr/local/bin/caddy
chmod +x /usr/local/bin/caddy

cd ..
rm -rf "$TMP_DIR"

# --- Create Caddy user and directories ---
echo ""
echo ">>> Step 4: Create Caddy user, group and required directories..."
if ! getent group caddy > /dev/null; then
  echo "Creating 'caddy' group..."
  groupadd --system caddy
fi
if ! id caddy > /dev/null 2>&1; then
  echo "Creating 'caddy' user..."
  useradd --system --gid caddy  --shell /usr/sbin/nologin caddy
fi

mkdir -p /etc/caddy
chown -R root:caddy /etc/caddy
mkdir -p /var/lib/caddy
chown -R caddy:caddy /var/lib/caddy


# Added: create configuration directory and fix permissions
mkdir -p /home/caddy/.config/caddy
chown -R caddy:caddy /home/caddy


# --- Configure systemd service ---
echo ""
echo ">>> Step 5: Configure systemd service for Caddy..."
cat <<EOF > /etc/systemd/system/caddy.service
[Unit]
Description=Caddy
Documentation=https://caddyserver.com/docs/
After=network.target network-online.target
Requires=network-online.target

[Service]
Type=notify
User=caddy
Group=caddy
ExecStart=/usr/local/bin/caddy run --environ --config /etc/caddy/caddy.json --resume
ExecReload=/usr/local/bin/caddy reload --config /etc/caddy/caddy.json --resume
TimeoutStopSec=5s
LimitNOFILE=1048576
LimitNPROC=512
PrivateTmp=true
ProtectSystem=full
AmbientCapabilities=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
EOF

# --- Create default configuration file ---
echo ""
echo ">>> Step 6: Create a default /etc/caddy/caddy.json..."
if [ ! -f /etc/caddy/caddy.json ]; then
cat <<EOF > /etc/caddy/caddy.json
{
  "apps": {
    "http": {
      "servers": {
        "srv0": {
          "listen": [
            ":80",
            ":443"
          ],
          "protocols": [
            "h1",
            "h2"
          ],
          "routes": []
        }
      }
    },
    "tls": {
      "automation": {}
    }
  }
}
EOF
chown caddy:caddy /etc/caddy/caddy.json
chmod 644 /etc/caddy/caddy.json
fi

# --- Start Caddy service ---
echo ""
echo ">>> Step 7: Reload systemd and start Caddy service..."
systemctl daemon-reload
systemctl enable --now caddy

# --- Verify installation ---
echo ""
echo ">>> Step 8: Verify installation..."
sleep 2 # wait a moment to ensure the service is fully started.

if ! command -v caddy &> /dev/null; then
    echo "Error: Caddy installation failed, command not found." >&2
    exit 1
fi

echo "Caddy installed successfully!"
caddy version

echo ""
echo "Caddy service status:"
systemctl status caddy --no-pager | cat
echo ""
echo -e "\033[32mInstallation complete!\033[0m"
echo "You can test the default page by visiting http://<your_server_ip>."
echo "Configuration file is located at /etc/caddy/caddy.json."
echo "After updating the configuration, run 'sudo systemctl reload caddy' to apply changes."

exit 0