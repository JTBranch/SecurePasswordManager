#!/bin/bash
set -e

COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
VERSION="v1.0.0-debug"

echo "=== DEBUGGING LDFLAGS ==="
echo "VERSION: '${VERSION}'"
echo "COMMIT: '${COMMIT}'"
echo "DATE: '${DATE}'"

# Test different approaches
echo ""
echo "=== Testing approach 1: Direct string ==="
LDFLAGS_1="-X go-password-manager/cmd.version=${VERSION}"
echo "LDFLAGS_1: '${LDFLAGS_1}'"
go build -ldflags "${LDFLAGS_1}" -o test1 ./cmd && echo "✓ Approach 1 works" && rm -f test1 || echo "✗ Approach 1 failed"

echo ""
echo "=== Testing approach 2: Multiple -X flags ==="
LDFLAGS_2="-X go-password-manager/cmd.version=${VERSION} -X go-password-manager/cmd.commit=${COMMIT}"
echo "LDFLAGS_2: '${LDFLAGS_2}'"
go build -ldflags "${LDFLAGS_2}" -o test2 ./cmd && echo "✓ Approach 2 works" && rm -f test2 || echo "✗ Approach 2 failed"

echo ""
echo "=== Testing approach 3: All three flags ==="
LDFLAGS_3="-X go-password-manager/cmd.version=${VERSION} -X go-password-manager/cmd.commit=${COMMIT} -X go-password-manager/cmd.date=${DATE}"
echo "LDFLAGS_3: '${LDFLAGS_3}'"
go build -ldflags "${LDFLAGS_3}" -o test3 ./cmd && echo "✓ Approach 3 works" && rm -f test3 || echo "✗ Approach 3 failed"

echo ""
echo "=== Testing what fyne-cross would see ==="
echo "Command that would be passed to fyne-cross:"
echo "fyne-cross linux -arch=amd64 --app-id go-password-manager --ldflags \"${LDFLAGS_3}\" ./cmd"
