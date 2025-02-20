#!/bin/bash

# Get the directory of this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Watch template files and restart the server on changes
find "$SCRIPT_DIR/../src/assets/templates" -type f \( -name "*.html" -o -name "*.css" -o -name "*.js" \) | entr -r "$SCRIPT_DIR/serve-templates.sh"