#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

echo "=== Start Build Process (WSL Environment) ==="

# 1. Get Git Version (for version injection)
if GIT_VERSION=$(git rev-parse --short HEAD 2>/dev/null); then
    echo "Git Version: $GIT_VERSION"
else
    echo "WARNING: Could not get git version. Using 'dev'."
    GIT_VERSION="dev"
fi

# 2. Create output directory
mkdir -p build

# Set common build parameters
# -s -w: Strip symbol table and debug information (reduce size)
# -X: Inject version variable
LDFLAGS="-X main.Version=$GIT_VERSION -s -w"

# ==========================================
# Linux (amd64) Build Section
# ==========================================
echo -e "\n[1/6] Building shipyard for Linux (amd64)..."

echo -e "\n[2/6] Building shipyard-server for Linux (amd64)..."
# Note: This is the server Linux version build you requested
env CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o build/shipyard-server-linux-amd64 ./cmd/shipyard-server
echo "SUCCESS: build/shipyard-server-linux-amd64"

# echo -e "\n[3/6] Building shipyard-cli for Linux (amd64)..."
# env CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o build/shipyard-cli-linux-amd64 ./cmd/shipyard-cli
# echo "SUCCESS: build/shipyard-cli-linux-amd64"

# ==========================================
# Windows (amd64) Build Section - Cross Compilation
# ==========================================
echo -e "\n[4/6] Building shipyard for Windows (amd64)..."

# Check if MinGW cross-compiler is installed
if ! command -v x86_64-w64-mingw32-gcc &> /dev/null; then
    echo "--------------------------------------------------------"
    echo "Error: Windows cross-compiler (x86_64-w64-mingw32-gcc) not found."
    echo "Please run the following command to install:"
    echo "  sudo apt update && sudo apt install mingw-w64"
    echo "--------------------------------------------------------"
    exit 1
fi

# echo -e "\n[5/6] Building shipyard-server for Windows (amd64)..."
# env CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -ldflags "$LDFLAGS" -o build/shipyard-server-windows-amd64.exe ./cmd/shipyard-server
# echo "SUCCESS: build/shipyard-server-windows-amd64.exe"

echo -e "\n[6/6] Building shipyard-cli for Windows (amd64)..."
env CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -ldflags "$LDFLAGS" -o build/shipyard-cli-windows-amd64.exe ./cmd/shipyard-cli
echo "SUCCESS: build/shipyard-cli-windows-amd64.exe"

# ==========================================
echo -e "\n=== All build tasks completed ==="

# ==========================================
# UPX Compression Section (Optional)
# ==========================================
if command -v upx &> /dev/null; then
    echo -e "\n[UPX] UPX detected, starting binary compression..."
    for file in build/*; do
        if [ -f "$file" ]; then
            echo "Compressing: $file"
            upx --best --lzma -q "$file" || echo "Warning: Failed to compress $file, skipping"
        fi
    done
    echo "[UPX] Compression completed"
else
    echo -e "\n[UPX] UPX not detected, skipping compression step"
    echo "Tip: Install UPX to reduce binary size (debian): sudo apt install upx"
fi

echo -e "\nBuild Artifacts List:"
ls -lh build/