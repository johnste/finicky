#!/bin/bash
# Build Finicky for Linux into apps/finicky/build/linux/finicky.
#
# Needs Go, Node, and the GTK 3 and WebKitGTK 4.1 development files
# (Arch: gtk3 webkit2gtk-4.1; Debian/Ubuntu: libgtk-3-dev libwebkit2gtk-4.1-dev).
# Run ./scripts/install.sh once first to install the npm dependencies.

# Exit on error
set -e

# Build config-api
(
    cd packages/config-api
    npm run build
    npm run generate-types

    cd ../../
    cp packages/config-api/dist/finickyConfigAPI.js apps/finicky/src/assets/finickyConfigAPI.js
)

# Build finicky-ui
(
    cd packages/finicky-ui
    npm run build

    cd ../../
    rm -rf apps/finicky/src/assets/templates
    mkdir -p apps/finicky/src/assets/templates
    cp -r packages/finicky-ui/dist/* apps/finicky/src/assets/templates
)

# Get build information. There's no app bundle on Linux, so the version from
# Info.plist is built into the binary.
COMMIT_HASH=$(git rev-parse --short HEAD)
BUILD_DATE=$(date -u '+%Y-%m-%d %H:%M:%S UTC')
API_HOST=$(cat .env 2>/dev/null | grep API_HOST | cut -d '=' -f 2 || echo "")
VERSION=$(grep -A1 '<key>CFBundleVersion</key>' apps/finicky/assets/Info.plist | sed -n 's/.*<string>\(.*\)<\/string>.*/\1/p')

mkdir -p apps/finicky/build/linux

go build -C apps/finicky/src \
    -ldflags \
    "-X 'finicky/version.commitHash=${COMMIT_HASH}' \
    -X 'finicky/version.buildDate=${BUILD_DATE}' \
    -X 'finicky/version.apiHost=${API_HOST}' \
    -X 'finicky/version.appVersion=${VERSION}'" \
    -o ../build/linux/finicky

cp packages/config-api/dist/finicky.d.ts apps/finicky/build/linux/finicky.d.ts
cp packages/finicky-ui/public/finicky-icon.png apps/finicky/build/linux/finicky.png

echo "Build complete ✨ apps/finicky/build/linux/finicky"
