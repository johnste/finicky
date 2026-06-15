#!/bin/bash

fd . --type f packages/config-api packages/finicky-ui apps/finicky \
  | entr -r sh -c 'clear && DEBUG=true ./scripts/build.sh && ./apps/finicky/build/Finicky.app/Contents/MacOS/Finicky && echo "$?"'
