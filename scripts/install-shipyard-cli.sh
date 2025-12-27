#!/bin/bash

# ==============================================================================
# Script Name: install-shipyard-cli.sh
# Description: Auto-detect system and download install shipyard-cli from GitHub Release
# Usage:       curl -fsSL https://raw.githubusercontent.com/youfun/shipyard/main/scripts/install-shipyard-cli.sh | bash
#             or: ./install-shipyard-cli.sh [version]
# ==============================================================================

set -e
set -o pipefail

# GitHub repository configuration
GITHUB_REPO="youfun/shipyard" 

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
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

print_success() {
    echo -e "${CYAN}[SUCCESS]${NC} $1"
}

# --- Show welcome information ---
echo ""
echo "======================================"
echo "  Shipyard CLI Installer"
echo "======================================"
echo ""

# --- Detect operating system ---
detect_os() {
    local os=$(uname -s)
    case $os in
        Linux*)
            echo "linux"
            ;;
        Darwin*)
            echo "darwin"
            ;;
        *)
            print_error "Unsupported operating system: $os"
            print_error "Only Linux and macOS are supported"
            exit 1
            ;;
    esac
}

# --- Detect architecture ---
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

# --- Get version ---
get_version() {
    local version="${1:-latest}"
    if [ "$version" = "latest" ]; then
        print_info "Getting latest version information..." >&2
        version=$(curl -fsSL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
        if [ -z "$version" ]; then
            print_error "Unable to get latest version information" >&2
            exit 1
        fi
    fi
    echo "$version"
}

# --- Download binary file ---
download_binary() {
    local os=$1
    local arch=$2
    local version=$3
    local binary_name="shipyard-cli-${os}-${arch}"
    
    # Windows requires .exe suffix
    if [ "$os" = "windows" ]; then
        binary_name="${binary_name}.exe"
    fi
    
    local download_url="https://github.com/${GITHUB_REPO}/releases/download/${version}/${binary_name}"
    
    print_info "Download URL: ${download_url}"
    print_info "Downloading ${binary_name}..."
    
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL -o "shipyard-cli" "${download_url}"
    elif command -v wget >/dev/null 2>&1; then
        wget -q -O "shipyard-cli" "${download_url}"
    else
        print_error "curl or wget is required to download files"
        exit 1
    fi
    
    if [ ! -f "shipyard-cli" ]; then
        print_error "Download failed"
        exit 1
    fi
    
    chmod +x "shipyard-cli"
    print_info "Download completed"
}

# --- Detect installation directory ---
get_install_dir() {
    # Check if has root privileges
    if [ "$(id -u)" -eq 0 ]; then
        echo "/usr/local/bin"
    else
        # Non-root user install to user directory
        local user_bin="$HOME/.local/bin"
        mkdir -p "$user_bin"
        echo "$user_bin"
    fi
}

# --- Check PATH ---
check_path() {
    local install_dir=$1
    if [[ ":$PATH:" != *":$install_dir:"* ]]; then
        print_warn "$install_dir is not in PATH"
        print_warn "Please add the following to your shell configuration file (~/.bashrc, ~/.zshrc, etc):"
        echo ""
        echo "    export PATH=\"\$PATH:$install_dir\""
        echo ""
    fi
}

# --- Main installation process ---
main() {
    # Get version parameter
    local version="${1:-latest}"
    
    # Detect system information
    print_info "Detecting system information..."
    local os=$(detect_os)
    local arch=$(detect_arch)
    print_info "Operating system: $os"
    print_info "Architecture: $arch"
    
    # Get version
    version=$(get_version "$version")
    print_info "Version: $version"
    echo ""
    
    # Create temporary directory
    local tmp_dir=$(mktemp -d)
    trap "rm -rf ${tmp_dir}" EXIT
    cd "${tmp_dir}"
    
    # Download binary file
    download_binary "$os" "$arch" "$version"
    
    # Determine installation directory
    local install_dir=$(get_install_dir)
    print_info "Installation directory: $install_dir"
    
    # Install
    print_info "Installing..."
    if [ "$(id -u)" -eq 0 ]; then
        # Root user install directly
        mv "shipyard-cli" "${install_dir}/shipyard-cli"
        chown root:root "${install_dir}/shipyard-cli"
        chmod 755 "${install_dir}/shipyard-cli"
    else
        # Non-root user
        if [ -w "$install_dir" ]; then
            mv "shipyard-cli" "${install_dir}/shipyard-cli"
        else
            print_warn "No write permission: $install_dir"
            print_info "Trying to use sudo..."
            sudo mv "shipyard-cli" "${install_dir}/shipyard-cli"
            sudo chmod 755 "${install_dir}/shipyard-cli"
        fi
    fi
    
    # Verify installation
    if command -v shipyard-cli >/dev/null 2>&1; then
        print_success "Installation successful!"
        echo ""
        echo "======================================"
        print_info "Version information:"
        shipyard-cli --version || true
        echo ""
        print_info "Getting started:"
        echo "  1. Login to server:"
        echo "     shipyard-cli login --endpoint http://your-server:15678"
        echo ""
        echo "  2. Test connection:"
        echo "     shipyard-cli ping"
        echo ""
        echo "  3. View help:"
        echo "     shipyard-cli --help"
        echo "======================================"
    else
        print_success "Installation completed: ${install_dir}/shipyard-cli"
        check_path "$install_dir"
        echo ""
        print_info "Reload shell configuration or restart terminal to use shipyard-cli command"
    fi
}

# Run main function
main "$@"
