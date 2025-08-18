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

echo "Starting Linux build process..."
echo "Version: ${VERSION}"
echo "Commit: ${COMMIT}"
echo "Date: ${DATE}"

# Install necessary tools
echo "Installing tools..."
echo "Installing fyne-cross for Linux builds..."
go install github.com/fyne-io/fyne-cross@latest

# Add fyne-cross to path if not already there
export PATH=$PATH:$(go env GOPATH)/bin

echo "Building binaries without version injection (ldflags issue with fyne-cross)..."
echo "Version info will show defaults: version=${VERSION}, commit=${COMMIT}, date=${DATE}"

# Build for Linux without ldflags to avoid fyne-cross issues
echo "Building Linux binary..."
echo "Note: fyne-cross doesn't support ldflags, so production mode detection will use environment defaults"
fyne-cross linux -arch=amd64 --app-id com.github.jtbranch.go-password-manager ./cmd

echo "Build process completed successfully."

# The binaries will be in the fyne-cross/bin directory
# We can move them to dist for consistency
echo "Moving binaries to dist/ directory..."
mkdir -p dist

# Check if the binary exists and move it
if [ -f "fyne-cross/bin/linux-amd64/cmd" ]; then
    mv fyne-cross/bin/linux-amd64/cmd dist/password-manager-linux-amd64
elif [ -f "fyne-cross/bin/linux-amd64/go-password-manager" ]; then
    mv fyne-cross/bin/linux-amd64/go-password-manager dist/password-manager-linux-amd64
else
    echo "Error: Could not find Linux binary in expected location"
    ls -la fyne-cross/bin/linux-amd64/
    exit 1
fi

# Make binaries executable
echo "Setting executable permissions..."
chmod +x dist/password-manager-linux-amd64

echo "Artifacts are ready in dist/ directory (Linux only)."
