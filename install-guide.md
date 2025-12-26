# Shipyard Installation Scripts

This directory contains automated installation scripts for the Shipyard Server and CLI.

## 📦 Available Scripts

### 1. install-shipyard-cli.sh

Automatically detects the system and installs the shipyard-cli client from GitHub Release.

**Features:**
- ✅ Automatically detects OS (Linux/macOS)
- ✅ Automatically detects architecture (amd64/arm64)
- ✅ Downloads the latest version or a specific version from GitHub Release
- ✅ Smartly chooses the installation directory (root user installs to `/usr/local/bin`, regular user installs to `~/.local/bin`)
- ✅ Automatically adds execution permissions
- ✅ Verifies installation and provides a usage guide

**Usage:**

```bash
# One-line installation of the latest version (Recommended)
curl -fsSL https://raw.githubusercontent.com/youfun/shipyard/main/scripts/install-shipyard-cli.sh | bash



# Or download and run locally
wget https://raw.githubusercontent.com/youfun/shipyard/main/scripts/install-shipyard-cli.sh
chmod +x install-shipyard-cli.sh
./install-shipyard-cli.sh           # Install latest version
```

**Installation Location:**
- Root User: `/usr/local/bin/shipyard-cli`
- Regular User: `~/.local/bin/shipyard-cli`

**First Time Use:**

```bash
# Check version
shipyard-cli --version

# Log in to the server
shipyard-cli login --endpoint http://your-server:8080

# Test connection
shipyard-cli ping
```

---

### 2. install-shipyard-server.sh

Downloads and installs shipyard-server from GitHub Release to the system, supporting both standard and test modes.

**Features:**
- ✅ Automatically detects system architecture (amd64/arm64)
- ✅ Downloads the latest version or a specific version from GitHub Release
- ✅ Two installation modes:
  - **Standard Mode**: Installs to system directories, creates a systemd service
  - **Test Mode**: Installs to the current directory, requires no root privileges
- ✅ Automatically creates system user and group (Standard Mode)
- ✅ Generates environment configuration templates
- ✅ Configures systemd service (Standard Mode)

**Usage:**

```bash
# One-line installation of the latest version (Standard Mode, requires sudo)
curl -fsSL https://raw.githubusercontent.com/YOUR_ORG/deployer/main/scripts/install-shipyard-server.sh | sudo bash

# Install a specific version
curl -fsSL https://raw.githubusercontent.com/YOUR_ORG/deployer/main/scripts/install-shipyard-server.sh | sudo bash -s v1.0.0

# Or download and run locally
wget https://raw.githubusercontent.com/YOUR_ORG/deployer/main/scripts/install-shipyard-server.sh
chmod +x install-shipyard-server.sh
sudo ./install-shipyard-server.sh           # Install latest version
sudo ./install-shipyard-server.sh v1.0.0    # Install specific version
```

**Installation Mode Selection:**

Upon running the script, you will be prompted to choose:
1. **Standard Mode** - Install to system directories (requires root privileges)
   - Binary: `/usr/local/bin/shipyard-server`
   - Config Directory: `/etc/shipyard`
   - Data Directory: `/var/lib/shipyard`
   - Log Directory: `/var/log/shipyard`
   - systemd Service: `shipyard-server.service`

2. **Test Mode** - Install to current directory (no root privileges required)
   - All files are in the current directory
   - Does not create a systemd service
   - Suitable for testing and development

**Configuration and Startup (Standard Mode):**

```bash
# 1. Edit configuration file (Required)
sudo nano /etc/shipyard/.env

# 2. Generate JWT Secret
openssl rand -base64 32

# 3. Start the service
sudo systemctl start shipyard-server

# 4. Check status
sudo systemctl status shipyard-server

# 5. View logs
sudo journalctl -u shipyard-server -f

# 6. Enable auto-start on boot (Optional)
sudo systemctl enable shipyard-server
```

**Configuration and Startup (Test Mode):**

```bash
# 1. Edit configuration file
nano config/.env

# 2. Start the service
export $(cat config/.env | xargs)
./shipyard-server-linux-amd64 --port 8080
```

---

## 🔧 Configuration File Guide

An `.env` configuration file is automatically generated after installation. Main configuration items:

```bash
# JWT Secret (Must Change!)
JWT_SECRET=PLEASE_CHANGE_THIS_TO_RANDOM_SECRET

# Database Type
DB_TYPE=sqlite

# SQLite Database Path
DB_PATH=/var/lib/shipyard/deploy.db

# Server Port
SERVER_PORT=8080

# Log Level
LOG_LEVEL=info
```

**Important:** Please ensure you change `JWT_SECRET` to a random key!

---

## 📝 Updating Repository URL in Scripts

Before use, please update the `GITHUB_REPO` variable in the scripts to your actual repository URL:

```bash
# Change this line
GITHUB_REPO="YOUR_ORG/deployer"

# To
GITHUB_REPO="yourusername/deployer"
```

Or provide the correct download link when creating a GitHub Release.

---

## 🐛 Troubleshooting

### Download Failure

```bash
# Check network connection
curl -I https://github.com

# Manual download
wget https://github.com/YOUR_ORG/deployer/releases/download/v1.0.0/shipyard-cli-linux-amd64
chmod +x shipyard-cli-linux-amd64
sudo mv shipyard-cli-linux-amd64 /usr/local/bin/shipyard-cli
```

### PATH Issues

If the command is not found after installing to `~/.local/bin`:

```bash
# Add to PATH (bash)
echo 'export PATH="$PATH:$HOME/.local/bin"' >> ~/.bashrc
source ~/.bashrc

# Add to PATH (zsh)
echo 'export PATH="$PATH:$HOME/.local/bin"' >> ~/.zshrc
source ~/.zshrc
```

### Permission Issues

```bash
# Standard mode requires sudo
sudo ./install-shipyard-server.sh

# Or use test mode (no sudo required)
./install-shipyard-server.sh  # Select option 2 (Test Mode)
```

---

## 📚 More Information

- [Main README](README.md)
- [GitHub Releases](https://github.com/YOUR_ORG/deployer/releases)
- [Issue Tracker](https://github.com/YOUR_ORG/deployer/issues)
