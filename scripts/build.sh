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

echo "Installing tools..."
# Only install fyne-cross if we're not on macOS (since we'll use native builds there)
if [[ "$OSTYPE" != "darwin"* ]]; then
  echo "Installing fyne-cross for Linux/Windows builds..."
  go install github.com/fyne-io/fyne-cross@latest
fi

# We don't need the regular fyne CLI since we're using direct go build for macOS
# and fyne-cross handles packaging for other platforms

# Add fyne-cross to path if not already there
export PATH=$PATH:$(go env GOPATH)/bin

echo "Building binaries without version injection (ldflags issue with fyne-cross)..."
echo "Version info will show defaults: version=${VERSION}, commit=${COMMIT}, date=${DATE}"

# Build for all targets without ldflags to avoid fyne-cross issues
echo "Building Linux binary..."
fyne-cross linux -arch=amd64 --app-id go-password-manager ./cmd

echo "Building macOS binary..."
# Build for macOS (requires macOS development environment)
if [[ "$OSTYPE" == "darwin"* ]]; then
  # We're on macOS, can build directly with CGO
  echo "Building macOS ARM64..."
  mkdir -p fyne-cross/bin/darwin-arm64
  CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o "fyne-cross/bin/darwin-arm64/go-password-manager" ./cmd
  
  echo "Building macOS AMD64..."
  mkdir -p fyne-cross/bin/darwin-amd64
  CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o "fyne-cross/bin/darwin-amd64/go-password-manager" ./cmd
  
  echo "✓ macOS builds completed"
else
  echo "Warning: macOS builds require macOS host (skipping)"
fi

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
