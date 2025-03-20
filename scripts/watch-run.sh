#!/bin/bash

fd . --type f packages/config-api packages/finicky-ui apps/finicky \
  | entr -r sh -c 'clear && DEBUG=true ./scripts/build.sh && ./build/Finicky.app/Contents/MacOS/Finicky && echo "$?"'
