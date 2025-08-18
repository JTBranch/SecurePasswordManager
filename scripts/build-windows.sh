#!/bin/bash
set -e

# Check for required parameters
if [ $# -ne 3 ]; then
    echo "Usage: $0 <version> <commit> <date>"
    exit 1
fi

VERSION="$1"
COMMIT="$2"
DATE="$3"

echo "Building Go Password Manager for Windows..."
echo "Version: ${VERSION}"
echo "Commit: ${COMMIT}"
echo "Date: ${DATE}"

# Create output directories
mkdir -p dist/windows-amd64

echo "Building Windows AMD64..."

# Try cross-compilation with proper app ID first
echo "Attempting cross-compilation with fyne-cross..."

# Install fyne-cross if not available
if ! command -v fyne-cross &> /dev/null; then
    echo "Installing fyne-cross..."
    go install github.com/fyne-io/fyne-cross@latest
    export PATH=$PATH:$(go env GOPATH)/bin
fi

# Use a proper app ID with dots
APP_ID="com.github.jtbranch.go-password-manager"

echo "Using app ID: ${APP_ID}"

# Try fyne-cross with proper app ID (without ldflags due to fyne-cross limitations)
echo "Building without version injection due to fyne-cross ldflags limitations..."
echo "Production mode will be detected via environment or default to production for releases"

if fyne-cross windows -arch=amd64 --app-id "${APP_ID}" ./cmd; then
    echo "✓ fyne-cross build successful"
    
    # Check what files were actually created
    echo "Checking fyne-cross output:"
    ls -la fyne-cross/dist/windows-amd64/
    
    # Extract the executable from the ZIP file that fyne-cross creates
    # Try different possible ZIP filenames
    ZIP_FILE=""
    if [ -f "fyne-cross/dist/windows-amd64/go-password-manager.exe.zip" ]; then
        ZIP_FILE="go-password-manager.exe.zip"
    elif [ -f "fyne-cross/dist/windows-amd64/SecurePasswordManager.exe.zip" ]; then
        ZIP_FILE="SecurePasswordManager.exe.zip"
    elif [ -f "fyne-cross/dist/windows-amd64/cmd.exe.zip" ]; then
        ZIP_FILE="cmd.exe.zip"
    else
        # Find any .zip file
        ZIP_FILE=$(ls fyne-cross/dist/windows-amd64/*.zip | head -1 | xargs basename)
    fi
    
    if [ -n "$ZIP_FILE" ]; then
        echo "Found ZIP file: $ZIP_FILE"
        echo "Extracting Windows executable from ZIP..."
        cd fyne-cross/dist/windows-amd64
        unzip -o "$ZIP_FILE"
        
        # Find the extracted .exe file
        EXE_FILE=$(ls *.exe | head -1)
        echo "Found executable: $EXE_FILE"
        cd ../../..
        
        # Move the binary to our expected location
        mkdir -p dist
        cp "fyne-cross/dist/windows-amd64/$EXE_FILE" dist/password-manager-windows-amd64.exe
    else
        echo "Error: No ZIP file found in fyne-cross output"
        exit 1
    fi
    
    echo "✓ Windows build completed successfully!"
    echo ""
    echo "Built binary:"
    ls -la dist/password-manager-windows-amd64.exe
    
else
    echo "⚠ fyne-cross failed"
    echo "Windows builds require a Windows environment for cross-compilation with CGO"
    echo "Skipping Windows build on this platform"
    exit 1
fi

echo ""
echo "Windows build process completed!"
