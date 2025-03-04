#!/bin/bash

# Exit on error
set -e

# Get build information
COMMIT_HASH=$(git rev-parse --short HEAD)
BUILD_DATE=$(date -u '+%Y-%m-%d %H:%M:%S UTC')

# Create build directory
mkdir -p build/Finicky.app/Contents/MacOS

# Build config API
cd config-api && npm run build && npm run generate-types && cd ..

export CGO_CFLAGS="-mmacosx-version-min=12.0"
export CGO_LDFLAGS="-mmacosx-version-min=12.0"

# Build the application
go build -C src \
    -ldflags "-X 'finicky/version.commitHash=${COMMIT_HASH}' -X 'finicky/version.buildDate=${BUILD_DATE}'" \
    -o ../build/Finicky.app/Contents/MacOS/Finicky

# Copy assets and install
cp -r assets/* build/Finicky.app/Contents/
rm -rf /Applications/Finicky.app
cp -r build/Finicky.app /Applications/

echo "Build complete ✨"