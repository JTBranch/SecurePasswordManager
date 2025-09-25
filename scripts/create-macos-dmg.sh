#!/usr/bin/env bash
set -euo pipefail

# create-macos-dmg.sh
# Usage: ./scripts/create-macos-dmg.sh [arch] [version]
#
# This script packages a built macOS binary (from dist/macos-<arch>/go-password-manager)
# into a minimal .app bundle and then creates a compressed .dmg suitable for distribution.
# It does not sign or notarize the app (you'll need to sign & notarize with Apple for App Store
# or Gatekeeper-friendly releases).

ARCH_ARG=${1:-}
VERSION=${2:-development}
# MODE: "dmg" (default) or "assemble" (only assemble app bundle into OUT_DIR)
MODE=${3:-dmg}

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DIST_DIR="$PROJECT_ROOT/dist"

detect_arch() {
  if [[ -n "$ARCH_ARG" ]]; then
    echo "$ARCH_ARG"
    return
  fi
  local machine
  machine=$(uname -m)
  if [[ "$machine" == "arm64" ]]; then
    echo "macos-arm64"
  elif [[ "$machine" == "x86_64" ]]; then
    echo "macos-amd64"
  else
    echo "macos-arm64"
  fi
}

ARCH_DIR=$(detect_arch)
BIN_SRC="$DIST_DIR/${ARCH_DIR}/go-password-manager"

if [[ ! -x "$BIN_SRC" ]]; then
  echo "Error: binary not found or not executable: $BIN_SRC" >&2
  echo "Run scripts/build-macos.sh first to build the binary." >&2
  exit 2
fi

APP_NAME="Password Manager"
APP_BUNDLE_NAME="PasswordManager.app"
APP_EXEC_NAME="password-manager"
BUNDLE_ID="com.jtbranch.gopasswordmanager"


OUT_DIR="$DIST_DIR/macos-dmg"
mkdir -p "$OUT_DIR"

STAGING_DIR=$(mktemp -d)
trap 'rm -rf "$STAGING_DIR"' EXIT

echo "Assembling .app bundle in staging area..."
APP_CONTENTS="$STAGING_DIR/$APP_BUNDLE_NAME/Contents"
APP_MACOS="$APP_CONTENTS/MacOS"
APP_RESOURCES="$APP_CONTENTS/Resources"
mkdir -p "$APP_MACOS" "$APP_RESOURCES"

# copy executable
cp "$BIN_SRC" "$APP_MACOS/$APP_EXEC_NAME"
chmod +x "$APP_MACOS/$APP_EXEC_NAME"

# copy icon if present in repo assets
ICNS_SRC="$PROJECT_ROOT/ui/assets/main-icon.icns"
if [[ -f "$ICNS_SRC" ]]; then
  cp "$ICNS_SRC" "$APP_RESOURCES/main-icon.icns"
fi

# Create Info.plist
cat > "$APP_CONTENTS/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleName</key>
  <string>${APP_NAME}</string>
  <key>CFBundleDisplayName</key>
  <string>${APP_NAME}</string>
  <key>CFBundleIdentifier</key>
  <string>${BUNDLE_ID}</string>
  <key>CFBundleExecutable</key>
  <string>${APP_EXEC_NAME}</string>
  <key>CFBundleIconFile</key>
  <string>main-icon</string>
  <key>LSUIElement</key>
  <false/>
</dict>
</plist>
PLIST

# If assemble-only mode is requested, copy assembled app to OUT_DIR and exit so workflow can sign it
ASSEMBLED_APP="$OUT_DIR/$APP_BUNDLE_NAME"
cp -R "$STAGING_DIR/$APP_BUNDLE_NAME" "$ASSEMBLED_APP"

if [[ "$MODE" == "assemble" ]]; then
  echo "Assembled app bundle available at: $ASSEMBLED_APP"
  exit 0
fi

echo "Preparing DMG contents (app + Applications symlink)..."
DMG_STAGING="$STAGING_DIR/dmg-root"
mkdir -p "$DMG_STAGING"
cp -R "$ASSEMBLED_APP" "$DMG_STAGING/"

# create Applications symlink inside DMG for user convenience
ln -s /Applications "$DMG_STAGING/Applications"

DMG_NAME="GoPasswordManager-${VERSION}-${ARCH_DIR}.dmg"
DMG_PATH="$OUT_DIR/$DMG_NAME"

echo "Creating compressed DMG: $DMG_PATH"
hdiutil create -volname "${APP_NAME}" -srcfolder "$DMG_STAGING" -ov -format UDZO "$DMG_PATH"

echo "DMG created: $DMG_PATH"
echo "Note: the DMG and .app are unsigned unless codesign + notarize steps are performed."

exit 0
