#!/bin/bash

# Go Password Manager Installer for macOS
# This script automatically downloads and installs the latest version

set -e

echo "🔐 Go Password Manager Installer"
echo "================================"
echo ""

# Detect macOS architecture
ARCH=$(uname -m)
if [[ "$ARCH" == "arm64" ]]; then
    BINARY_NAME="go-password-manager-macos-arm64"
    echo "✅ Detected: Apple Silicon Mac (M1/M2/M3)"
elif [[ "$ARCH" == "x86_64" ]]; then
    BINARY_NAME="go-password-manager-macos-amd64"
    echo "✅ Detected: Intel Mac"
else
    echo "❌ Unsupported architecture: $ARCH"
    exit 1
fi

# Create installation directory
INSTALL_DIR="$HOME/Applications/PasswordManager"
mkdir -p "$INSTALL_DIR"

echo "📥 Downloading latest version..."

# Get latest release info
LATEST_RELEASE=$(curl -s https://api.github.com/repos/JTBranch/SecurePasswordManager/releases/latest)
DOWNLOAD_URL=$(echo "$LATEST_RELEASE" | grep "browser_download_url.*$BINARY_NAME" | cut -d '"' -f 4)
VERSION=$(echo "$LATEST_RELEASE" | grep '"tag_name"' | cut -d '"' -f 4)

if [[ -z "$DOWNLOAD_URL" ]]; then
    echo "❌ Failed to find download URL for $BINARY_NAME"
    exit 1
fi

echo "📦 Installing version $VERSION..."

# Download the binary
curl -L -o "$INSTALL_DIR/password-manager" "$DOWNLOAD_URL"
chmod +x "$INSTALL_DIR/password-manager"

# Attempt to download configs archive if present in release assets and extract next to binary
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
    # Fallback: try to fetch a configs directory from the repo (raw) as a last resort
    echo "📁 No configs archive in release; fetching default configs from repository..."
    mkdir -p "$INSTALL_DIR/configs"
    curl -sL https://raw.githubusercontent.com/JTBranch/SecurePasswordManager/main/configs/default.yaml -o "$INSTALL_DIR/configs/default.yaml" || true
fi

# Create a launcher script
cat > "$INSTALL_DIR/launch-password-manager.sh" << 'EOF'
#!/bin/bash
cd "$(dirname "$0")"
export GO_PASSWORD_MANAGER_ENV=production
./password-manager
EOF
chmod +x "$INSTALL_DIR/launch-password-manager.sh"

# Create desktop shortcut (optional)
DESKTOP_FILE="$HOME/Desktop/Password Manager.command"
cat > "$DESKTOP_FILE" << EOF
#!/bin/bash
cd "$INSTALL_DIR"
export GO_PASSWORD_MANAGER_ENV=production
./password-manager
EOF
chmod +x "$DESKTOP_FILE"

echo ""
echo "🎉 Installation Complete!"
echo ""
echo "📂 Installed to: $INSTALL_DIR"
echo "🖥️  Desktop shortcut: Password Manager.command"
echo ""
echo "🚀 To run the app:"
echo "   1. Double-click 'Password Manager.command' on your Desktop"
echo "   2. Or run: $INSTALL_DIR/password-manager"
echo ""
echo "📋 Your passwords will be stored in: $INSTALL_DIR/secrets.json"
echo ""
echo "Thank you for using Go Password Manager! 🔐"
