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

# Create a launcher script
cat > "$INSTALL_DIR/launch-password-manager.sh" << 'EOF'
#!/bin/bash
cd "$(dirname "$0")"
./password-manager
EOF
chmod +x "$INSTALL_DIR/launch-password-manager.sh"

# Create desktop shortcut (optional)
DESKTOP_FILE="$HOME/Desktop/Password Manager.command"
cat > "$DESKTOP_FILE" << EOF
#!/bin/bash
cd "$INSTALL_DIR"
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
