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

# Download configs archive from release if available
CONFIGS_URL=$(echo "$LATEST_RELEASE" | grep "browser_download_url.*configs-.*zip" | cut -d '"' -f 4)
if [[ -n "$CONFIGS_URL" ]]; then
    echo "📁 Downloading configs..."
    curl -L -o "$INSTALL_DIR/configs.zip" "$CONFIGS_URL"
    if command -v unzip >/dev/null 2>&1; then
        unzip -o "$INSTALL_DIR/configs.zip" -d "$INSTALL_DIR"
        rm "$INSTALL_DIR/configs.zip"
    else
        echo "⚠️  unzip not found; created configs.zip in install dir. Please extract it to $INSTALL_DIR"
    fi
else
    echo "📁 No configs archive in release; fetching default config from repository..."
    mkdir -p "$INSTALL_DIR/configs"
    curl -sL https://raw.githubusercontent.com/JTBranch/SecurePasswordManager/main/configs/default.yaml -o "$INSTALL_DIR/configs/default.yaml" || true
fi

# Create desktop entry if we're in a desktop environment
if [[ -n "$XDG_CURRENT_DESKTOP" ]]; then
    DESKTOP_DIR="$HOME/.local/share/applications"
    mkdir -p "$DESKTOP_DIR"

    # Try to download a PNG icon and save next to the binary
    ICON_URL=$(echo "$LATEST_RELEASE" | grep "browser_download_url.*main-icon.*\\.png" | cut -d '"' -f 4)
    ICON_PATH="$INSTALL_DIR/main-icon.png"
    if [[ -n "$ICON_URL" ]]; then
        curl -L -o "$ICON_PATH" "$ICON_URL" || true
    else
        curl -sL https://raw.githubusercontent.com/JTBranch/SecurePasswordManager/main/ui/assets/main-icon.png -o "$ICON_PATH" || true
    fi

    cat > "$DESKTOP_DIR/password-manager.desktop" << EOF
[Desktop Entry]
Name=Go Password Manager
Comment=Secure password management application
Exec=env GO_PASSWORD_MANAGER_ENV=production $INSTALL_DIR/password-manager
Icon=$ICON_PATH
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
