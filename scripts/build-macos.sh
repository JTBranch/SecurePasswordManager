#!/bin/bash
set -e

# Simple macOS build script
echo "Building Go Password Manager for macOS..."

# Set version info (use defaults if not provided)
VERSION=${1:-"development"}
COMMIT=${2:-"none"}
DATE=${3:-"unknown"}

echo "Version: ${VERSION}"
echo "Commit: ${COMMIT}"
echo "Date: ${DATE}"

# Create output directories
mkdir -p dist/macos-arm64
mkdir -p dist/macos-amd64

echo "Building macOS ARM64 (Apple Silicon)..."
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build \
  -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
  -o "dist/macos-arm64/go-password-manager" \
  ./cmd

echo "Building macOS AMD64 (Intel)..."
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build \
  -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
  -o "dist/macos-amd64/go-password-manager" \
  ./cmd

echo ""
echo "✓ macOS builds completed successfully!"
echo ""
echo "Built binaries:"
ls -la dist/macos-*/go-password-manager

echo ""
echo "Testing ARM64 binary:"
./dist/macos-arm64/go-password-manager --version

echo ""
echo "You can run the appropriate binary for your Mac:"
echo "  Apple Silicon (M1/M2/M3): ./dist/macos-arm64/go-password-manager"
echo "  Intel:                    ./dist/macos-amd64/go-password-manager"
