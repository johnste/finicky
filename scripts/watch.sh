#!/bin/bash

fd . --type f packages/config-api packages/finicky-ui apps/finicky \
  | entr -r sh -c 'clear && ./scripts/build.sh && killall Finicky && echo "$?"'
