#!/usr/bin/env bash
set -euo pipefail

# Define destination directory
DEST_DIR="ui/static/js"
DEST_FILE="$DEST_DIR/htmx.min.js"
URL="https://cdn.jsdelivr.net/npm/htmx.org@latest/dist/htmx.min.js"

# Ensure destination directory exists
mkdir -p "$DEST_DIR"

# Fetch latest htmx and overwrite the local copy
echo "Downloading latest htmx from $URL ..."
curl -fsSL "$URL" -o "$DEST_FILE"

echo "Updated htmx at $DEST_FILE"
