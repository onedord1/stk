#!/bin/bash
# Quick installation script for SysTask

set -e

VERSION="1.0.2"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.config/systask"

echo "🚀 Installing SysTask v${VERSION}"
echo "=================================="

# Check if running with sudo for system-wide install
if [ "$EUID" -eq 0 ]; then
    INSTALL_DIR="/usr/bin"
    CONFIG_DIR="/etc/systask"
    echo "📍 Installing to system directory: ${INSTALL_DIR}"
else
    echo "📍 Installing to user directory: ${INSTALL_DIR}"
    echo "   (Use 'sudo' for system-wide installation)"
fi

# Build binary
echo "🔨 Building SysTask..."
go build -o systask ./main.go

# Install binary
echo "📦 Installing binary..."
if [ "$EUID" -eq 0 ]; then
    install -Dm755 systask "${INSTALL_DIR}/systask"
else
    mkdir -p "${INSTALL_DIR}"
    cp systask "${INSTALL_DIR}/"
    chmod +x "${INSTALL_DIR}/systask"
fi

# Create config directory
echo "⚙️  Setting up configuration..."
mkdir -p "${CONFIG_DIR}"
if [ ! -f "${CONFIG_DIR}/config.yaml" ]; then
    cp config.yaml.example "${CONFIG_DIR}/config.yaml"
    echo "   Created: ${CONFIG_DIR}/config.yaml"
else
    echo "   Config already exists: ${CONFIG_DIR}/config.yaml"
fi

echo ""
echo "✅ Installation complete!"
echo ""
echo "📝 Next steps:"
echo "   1. Edit your config: nano ${CONFIG_DIR}/config.yaml"
echo "   2. Add your SSH hosts to the config"
echo "   3. Run: systask"
echo ""
echo "📚 For help, run: systask --help"
echo "📖 Documentation: https://github.com/onedord1/stk"
