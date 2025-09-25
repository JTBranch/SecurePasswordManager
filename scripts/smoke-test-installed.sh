#!/usr/bin/env bash
# smoke-test-installed.sh
# Usage: smoke-test-installed.sh /path/to/installed/password-manager

set -euo pipefail

BIN_PATH=${1:-}
if [[ -z "$BIN_PATH" ]]; then
    echo "Usage: $0 /path/to/password-manager"
    exit 2
fi

if [[ ! -x "$BIN_PATH" ]]; then
    echo "Binary not found or not executable: $BIN_PATH"
    exit 1
fi

echo "Running smoke test: $BIN_PATH --print-env"
OUT=$("$BIN_PATH" --print-env || true)
echo "Output: $OUT"
if [[ "$OUT" == "production" ]]; then
    echo "OK: Installed binary reports production environment"
    exit 0
else
    echo "FAIL: Installed binary did not report production (got: '$OUT')"
    exit 3
fi
