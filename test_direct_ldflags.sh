#!/bin/bash
set -e

COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
VERSION="v1.0.0-test"

echo "Testing direct ldflags syntax with go build:"
echo "Building with version: ${VERSION}, commit: ${COMMIT}, date: ${DATE}"

# Try building with go build to test the direct ldflags syntax
go build -ldflags "-X go-password-manager/cmd.version=${VERSION} -X go-password-manager/cmd.commit=${COMMIT} -X go-password-manager/cmd.date=${DATE}" -o test_binary ./cmd

if [ $? -eq 0 ]; then
    echo "✓ Direct ldflags syntax is correct"
    rm -f test_binary
else
    echo "✗ Direct ldflags syntax failed"
    exit 1
fi
