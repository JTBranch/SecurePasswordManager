#!/bin/bash
set -e

# Check for required arguments
if [ "$#" -ne 3 ]; then
    echo "Usage: $0 <version> <commit> <date>"
    exit 1
fi

VERSION=$1
COMMIT=$2
DATE=$3

echo "Starting build process..."
echo "Version: ${VERSION}"
echo "Commit: ${COMMIT}"
echo "Date: ${DATE}"

echo "Installing Fyne CLI..."
go install fyne.io/fyne/v2/cmd/fyne@latest

# Add fyne-cross to path if not already there
export PATH=$PATH:$(go env GOPATH)/bin

echo "Building binaries without version injection (ldflags issue with fyne-cross)..."
echo "Version info will show defaults: version=${VERSION}, commit=${COMMIT}, date=${DATE}"

# Build for all targets without ldflags to avoid fyne-cross issues
echo "Building Linux binary..."
fyne-cross linux -arch=amd64 --app-id go-password-manager ./cmd

echo "Building macOS binary..."
# Skip macOS builds for now due to SDK path requirements in fyne-cross
# TODO: Configure macOS SDK properly for Darwin builds
echo "Skipping macOS build temporarily due to SDK configuration requirements"
# fyne-cross darwin -arch=arm64 --app-id go-password-manager --macosx-version-min=10.15 ./cmd

echo "Building Windows binary..."
fyne-cross windows -arch=amd64 --app-id go-password-manager ./cmd

echo "Build process completed successfully."

# The binaries will be in the fyne-cross/bin directory
# We can move them to dist for consistency
echo "Moving binaries to dist/ directory..."
mkdir -p dist
mv fyne-cross/bin/linux-amd64/go-password-manager dist/password-manager-linux-amd64
# Skip macOS binary (not built)
mv fyne-cross/bin/windows-amd64/go-password-manager.exe dist/password-manager-windows-amd64.exe

# Make binaries executable
echo "Setting executable permissions..."
chmod +x dist/password-manager-linux-amd64

echo "Artifacts are ready in dist/ directory (Linux and Windows only)."
