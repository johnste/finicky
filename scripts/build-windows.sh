#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "Building Finicky for Windows..."

# Build frontend assets (same as macOS)
(
    cd "$ROOT_DIR/packages/config-api"
    npm run build
    npm run generate-types
    cp dist/finickyConfigAPI.js "$ROOT_DIR/apps/finicky/src/assets/finickyConfigAPI.js"
)

(
    cd "$ROOT_DIR/packages/finicky-ui"
    npm run build
    mkdir -p "$ROOT_DIR/apps/finicky/src/assets/templates"
    cp -r dist/* "$ROOT_DIR/apps/finicky/src/assets/templates"
)

# Build info
COMMIT_HASH=$(git -C "$ROOT_DIR" rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_DATE=$(date -u '+%Y-%m-%d %H:%M:%S UTC')
API_HOST=$(cat "$ROOT_DIR/.env" 2>/dev/null | grep API_HOST | cut -d '=' -f 2 || echo "")

# Cross-compile for Windows
mkdir -p "$ROOT_DIR/apps/finicky/build/windows"

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
    go build -C "$ROOT_DIR/apps/finicky/src" \
    -ldflags \
    "-X 'finicky/version.commitHash=${COMMIT_HASH}' \
     -X 'finicky/version.buildDate=${BUILD_DATE}' \
     -X 'finicky/version.apiHost=${API_HOST}' \
     -H windowsgui" \
    -o ../build/windows/Finicky.exe

echo "Binary: apps/finicky/build/windows/Finicky.exe"
ls -lh "$ROOT_DIR/apps/finicky/build/windows/Finicky.exe"
file "$ROOT_DIR/apps/finicky/build/windows/Finicky.exe"

echo ""
echo "To build the installer, run Inno Setup on scripts/installer.iss"
echo "  iscc scripts/installer.iss"
echo ""
echo "Build complete."
