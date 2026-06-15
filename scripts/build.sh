#!/bin/bash

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
    mkdir -p apps/finicky/src/assets/templates
    cp -r packages/finicky-ui/dist/* apps/finicky/src/assets/templates
)

# Get build information
COMMIT_HASH=$(git rev-parse --short HEAD)
BUILD_DATE=$(date -u '+%Y-%m-%d %H:%M:%S UTC')
API_HOST=$(cat .env 2>/dev/null | grep API_HOST | cut -d '=' -f 2 || echo "")

copy_assets() {
    local APP_NAME=$1
    cp packages/config-api/dist/finicky.d.ts apps/finicky/build/${APP_NAME}/Contents/Resources/finicky.d.ts
    rsync -a --exclude='menu.iconset' apps/finicky/assets/ apps/finicky/build/${APP_NAME}/Contents/
}

build_arch() {
    local ARCH=$1
    local APP_NAME="Finicky-${ARCH}.app"

    mkdir -p apps/finicky/build/${APP_NAME}/Contents/MacOS
    mkdir -p apps/finicky/build/${APP_NAME}/Contents/Resources

    if [ "$ARCH" = "amd64" ]; then
        export GOARCH=amd64
        export CGO_CFLAGS="-target x86_64-apple-macos12.0"
        export CGO_LDFLAGS="-target x86_64-apple-macos12.0"
    else
        export GOARCH=arm64
        export CGO_CFLAGS="-mmacosx-version-min=12.0"
        export CGO_LDFLAGS="-mmacosx-version-min=12.0"
    fi

    export CGO_ENABLED=1
    export CC=clang

    go build -C apps/finicky/src \
        -ldflags \
        "-X 'finicky/version.commitHash=${COMMIT_HASH}' \
        -X 'finicky/version.buildDate=${BUILD_DATE}' \
        -X 'finicky/version.apiHost=${API_HOST}'" \
        -o ../build/${APP_NAME}/Contents/MacOS/Finicky
}

if [ "${BUILD_UNIVERSAL:-0}" = "1" ]; then
    build_arch arm64
    build_arch amd64

    APP_NAME="Finicky.app"
    mkdir -p apps/finicky/build/${APP_NAME}/Contents/MacOS
    mkdir -p apps/finicky/build/${APP_NAME}/Contents/Resources

    lipo -create \
        apps/finicky/build/Finicky-arm64.app/Contents/MacOS/Finicky \
        apps/finicky/build/Finicky-amd64.app/Contents/MacOS/Finicky \
        -output apps/finicky/build/${APP_NAME}/Contents/MacOS/Finicky

    lipo -info apps/finicky/build/${APP_NAME}/Contents/MacOS/Finicky
    copy_assets ${APP_NAME}

elif [ -n "$BUILD_TARGET_ARCH" ]; then
    # Single-arch CI build (legacy/fallback)
    APP_NAME="Finicky-${BUILD_TARGET_ARCH}.app"
    build_arch ${BUILD_TARGET_ARCH}
    copy_assets ${APP_NAME}

else
    # Local build — native arch only
    APP_NAME="Finicky.app"
    # Rename arch build to plain Finicky.app
    build_arch arm64

    rm -r apps/finicky/build/${APP_NAME}
    mv apps/finicky/build/Finicky-arm64.app apps/finicky/build/${APP_NAME}
    copy_assets ${APP_NAME}

    # Replace existing app
    rm -rf /Applications/Finicky.app
    cp -r apps/finicky/build/Finicky.app /Applications/
fi

echo "Build complete ✨"
