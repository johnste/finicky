#!/bin/bash

mkdir -p build/Finicky.app/Contents/MacOS
cd config-api && npm run build && npm run generate-types && cd ..
go build -C src -o ../build/Finicky.app/Contents/MacOS/Finicky
cp -r assets/* build/Finicky.app/Contents/
rm -r /Applications/Finicky.app
cp -r build/Finicky.app /Applications/
echo "Build complete ✨"