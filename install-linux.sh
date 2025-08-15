#!/bin/bash

# Go Password Manager Installer for Linux
# This script automatically downloads and installs the latest version

set -e

echo "🔐 Go Password Manager Installer"
echo "================================"
echo ""

# Create installation directory
INSTALL_DIR="$HOME/.local/bin"
mkdir -p "$INSTALL_DIR"

# Ensure the directory is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$HOME/.bashrc"
    echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$HOME/.zshrc" 2>/dev/null || true
fi

echo "📥 Downloading latest version..."

# Get latest release info
LATEST_RELEASE=$(curl -s https://api.github.com/repos/JTBranch/SecurePasswordManager/releases/latest)
DOWNLOAD_URL=$(echo "$LATEST_RELEASE" | grep "browser_download_url.*password-manager-linux-amd64" | cut -d '"' -f 4)
VERSION=$(echo "$LATEST_RELEASE" | grep '"tag_name"' | cut -d '"' -f 4)

if [[ -z "$DOWNLOAD_URL" ]]; then
    echo "❌ Failed to find download URL for Linux binary"
    exit 1
fi

echo "📦 Installing version $VERSION..."

# Download the binary
curl -L -o "$INSTALL_DIR/password-manager" "$DOWNLOAD_URL"
chmod +x "$INSTALL_DIR/password-manager"

# Create desktop entry if we're in a desktop environment
if [[ -n "$XDG_CURRENT_DESKTOP" ]]; then
    DESKTOP_DIR="$HOME/.local/share/applications"
    mkdir -p "$DESKTOP_DIR"
    
    cat > "$DESKTOP_DIR/password-manager.desktop" << EOF
[Desktop Entry]
Name=Go Password Manager
Comment=Secure password management application
Exec=$INSTALL_DIR/password-manager
Icon=applications-utilities
Terminal=false
Type=Application
Categories=Utility;Security;
EOF
fi

echo ""
echo "🎉 Installation Complete!"
echo ""
echo "📂 Installed to: $INSTALL_DIR/password-manager"
if [[ -n "$XDG_CURRENT_DESKTOP" ]]; then
    echo "🖥️  Desktop entry created (check your applications menu)"
fi
echo ""
echo "🚀 To run the app:"
echo "   1. Open terminal and run: password-manager"
echo "   2. Or find 'Go Password Manager' in your applications menu"
echo ""
echo "📋 Your passwords will be stored in: ~/.config/password-manager/secrets.json"
echo ""
echo "⚠️  You may need to restart your terminal or run: source ~/.bashrc"
echo ""
echo "Thank you for using Go Password Manager! 🔐"
