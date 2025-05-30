
#!/bin/bash

# This script builds the browser extension for Firefox and Chrome.

# firefox
mkdir -p build/firefox
cp *.js *.png build/firefox/
cp manifest-firefox.json build/firefox/manifest.json
zip -r build/firefox.zip -j build/firefox/*

# chrome
mkdir -p build/chrome
cp *.js *.png build/chrome/
cp manifest-chrome.json build/chrome/manifest.json
zip -r build/chrome.zip -j build/chrome/*

# clean up
rm -rf build/firefox build/chrome