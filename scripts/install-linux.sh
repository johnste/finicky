#!/bin/bash
# Install the Linux build of Finicky for the current user, and register it
# as a browser so it can be chosen as the default.
#
# Usage:
#   ./scripts/build-linux.sh
#   ./scripts/install-linux.sh [--set-default]
#
# Installs to ~/.local/bin/finicky and ~/.local/share/applications/finicky.desktop.
# --set-default also makes Finicky the default web browser.

set -e

BUILD_DIR=apps/finicky/build/linux
BIN_DIR="$HOME/.local/bin"
DATA_DIR="${XDG_DATA_HOME:-$HOME/.local/share}"
APP_DIR="$DATA_DIR/applications"
ICON_DIR="$DATA_DIR/icons/hicolor/512x512/apps"

if [ ! -x "$BUILD_DIR/finicky" ]; then
    echo "No build found; run ./scripts/build-linux.sh first" >&2
    exit 1
fi

mkdir -p "$BIN_DIR" "$APP_DIR" "$ICON_DIR"
install -m 755 "$BUILD_DIR/finicky" "$BIN_DIR/finicky"
install -m 644 "$BUILD_DIR/finicky.png" "$ICON_DIR/finicky.png"

cat > "$APP_DIR/finicky.desktop" <<DESKTOP
[Desktop Entry]
Type=Application
Name=Finicky
GenericName=Web Browser
Comment=Always open the right browser
Exec="$BIN_DIR/finicky" %u
Icon=finicky
Terminal=false
Categories=Network;WebBrowser;
MimeType=x-scheme-handler/http;x-scheme-handler/https;x-scheme-handler/finicky;text/html;
StartupNotify=false
DESKTOP

update-desktop-database "$APP_DIR" 2>/dev/null || true

echo "Installed Finicky to $BIN_DIR/finicky"

if [ "$1" = "--set-default" ]; then
    xdg-settings set default-web-browser finicky.desktop
    echo "Finicky is now the default web browser"
else
    echo "To make it the default browser: xdg-settings set default-web-browser finicky.desktop"
fi
