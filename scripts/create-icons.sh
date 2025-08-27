#!/usr/bin/env bash
set -euo pipefail

# Compute paths relative to script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$SCRIPT_DIR/.."
SRC="$ROOT_DIR/apps/finicky/assets/template-icon.pdf"
ICONSET_DIR="$ROOT_DIR/apps/finicky/assets/menu.iconset"
OUTPUT_ICNS="$ROOT_DIR/apps/finicky/assets/menu-bar.icns"

# Remove old iconset and recreate
rm -rf "$ICONSET_DIR"
mkdir -p "$ICONSET_DIR"

# Generate PNG slices at various sizes and @2x
for SIZE in 16 20 32 128 256 512; do
  sips -s format png -z $SIZE $SIZE "$SRC" --out "$ICONSET_DIR/icon_${SIZE}x${SIZE}.png"
  sips -s format png -z $((SIZE*2)) $((SIZE*2)) "$SRC" --out "$ICONSET_DIR/icon_${SIZE}x${SIZE}@2x.png"
done

# Build the ICNS file
iconutil -c icns "$ICONSET_DIR" -o "$OUTPUT_ICNS"