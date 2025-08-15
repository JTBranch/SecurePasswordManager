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

# Add fyne-cross to path if not already there
export PATH=$PATH:$(go env GOPATH)/bin

# Construct the ldflags string using the -X=key=value format
LD_FLAGS="-X=go-password-manager/cmd.version=${VERSION} -X=go-password-manager/cmd.commit=${COMMIT} -X=go-password-manager/cmd.date=${DATE}"

echo "Using LD_FLAGS: ${LD_FLAGS}"

# Build for all targets, assuming fyne-cross and dependencies are installed
fyne-cross linux -arch=amd64 --app-id go-password-manager --ldflags "${LD_FLAGS}" ./cmd
fyne-cross darwin -arch=arm64 --app-id go-password-manager --ldflags "${LD_FLAGS}" ./cmd
fyne-cross windows -arch=amd64 --app-id go-password-manager --ldflags "${LD_FLAGS}" ./cmd

echo "Build process completed successfully."

# The binaries will be in the fyne-cross/bin directory
# We can move them to dist for consistency
echo "Moving binaries to dist/ directory..."
mkdir -p dist
mv fyne-cross/bin/linux-amd64/go-password-manager dist/password-manager-linux-amd64
mv fyne-cross/bin/darwin-arm64/go-password-manager dist/password-manager-macos-arm64
mv fyne-cross/bin/windows-amd64/go-password-manager.exe dist/password-manager-windows-amd64.exe

# Make binaries executable
echo "Setting executable permissions..."
chmod +x dist/password-manager-linux-amd64
chmod +x dist/password-manager-macos-arm64

echo "Artifacts are ready in dist/ directory."
