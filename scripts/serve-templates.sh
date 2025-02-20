#!/bin/bash

# Get the directory of this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# Calculate the templates directory path relative to this script
TEMPLATES_DIR="$SCRIPT_DIR/../src/assets/templates"

# Create a temporary directory for development
TMP_DIR="/tmp/finicky-dev-templates"
mkdir -p "$TMP_DIR"

# Copy and modify the template files for development
echo "Preparing templates for development..."
cp "$TEMPLATES_DIR"/* "$TMP_DIR/"

# Replace finicky-assets://local/ with ./ in the copied files
sed -i '' 's|finicky-assets://local/|./|g' "$TMP_DIR"/*.html

# Create a Python server script
cat > "$TMP_DIR/server.py" << 'EOF'
import http.server
import socketserver

class DevHandler(http.server.SimpleHTTPRequestHandler):
    def end_headers(self):
        self.send_header('Cache-Control', 'no-cache, no-store, must-revalidate')
        self.send_header('Pragma', 'no-cache')
        self.send_header('Expires', '0')
        super().end_headers()

PORT = 8000
Handler = DevHandler

with socketserver.TCPServer(("", PORT), Handler) as httpd:
    print(f"Server running at http://localhost:{PORT}")
    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        print("\nShutting down server...")
        httpd.shutdown()
EOF

echo "Starting development server for templates..."
echo "Templates directory: $TMP_DIR"
echo "Access the templates at: http://localhost:8000"
echo "Press Ctrl+C to stop the server"

# Change to the temporary directory and start the Python server
cd "$TMP_DIR" && python3 server.py