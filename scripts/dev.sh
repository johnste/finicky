#!/bin/bash
# Run Finicky in a specific test scenario without building.
#
# Usage:
#   ./scripts/dev.sh [scenario]
#
# Scenarios:
#   1  none   No JS config, no rules.json
#   2  js     JS config only
#   3  json   rules.json only
#   4  both   JS config + rules.json (default)

set -e

BINARY="apps/finicky/build/Finicky.app/Contents/MacOS/Finicky"
JS_CONFIG="$(pwd)/testdata/config.js"
JSON_RULES="$(pwd)/testdata/rules.json"
NO_RULES="/tmp/finicky-dev-no-rules-$(date +%s).json"  # unique per run, intentionally absent

case "$1" in
  0|normal)
    echo "Scenario 0: normal (auto-detect configs)"
    "$BINARY" --window
    ;;
  1|none)
    echo "Scenario 1: no JS config, no rules.json"
    "$BINARY" --window --no-config --rules "$NO_RULES"
    ;;
  2|js)
    echo "Scenario 2: JS config only"
    "$BINARY" --window --config "$JS_CONFIG" --rules "$NO_RULES"
    ;;
  3|json)
    echo "Scenario 3: rules.json only"
    "$BINARY" --window --no-config --rules "$JSON_RULES"
    ;;
  4|both)
    echo "Scenario 4: JS config + rules.json"
    "$BINARY" --window --config "$JS_CONFIG" --rules "$JSON_RULES"
    ;;
  *)
    echo "Usage: $0 <scenario>"
    echo ""
    echo "  0  normal  Auto-detect configs (default behavior)"
    echo "  1  none    No JS config, no rules.json"
    echo "  2  js      JS config only  ($JS_CONFIG)"
    echo "  3  json    rules.json only ($JSON_RULES)"
    echo "  4  both    JS config + rules.json"
    exit 0
    ;;
esac
