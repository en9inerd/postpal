#!/usr/bin/env bash
set -euo pipefail

echo "=== PostPal Auth Testing Script ==="
echo ""

# Check if password is provided
if [ $# -lt 1 ]; then
    echo "Usage: $0 <password> [port]"
    echo ""
    echo "Example:"
    echo "  $0 my-secure-password"
    echo "  $0 my-secure-password 8080"
    exit 1
fi

PASSWORD="$1"
PORT="${2:-8000}"

echo "1. Generating password hash..."
HASH=$(go run scripts/generate-password-hash.go "$PASSWORD")
echo "   Hash: $HASH"
echo ""

echo "2. Generating session secret..."
SECRET=$(openssl rand -base64 32)
echo "   Secret: $SECRET"
echo ""

echo "3. Setting environment variables..."
export AUTH_PASSWORD_HASH="$HASH"
export AUTH_SESSION_SECRET="$SECRET"
export AUTH_SESSION_MAX_AGE="86400"
export APP_PORT="$PORT"
echo "   ✓ Environment variables set"
echo ""

echo "4. Testing configuration..."
if ! go run ./cmd/app --help > /dev/null 2>&1; then
    echo "   ✗ Failed to parse config"
    exit 1
fi
echo "   ✓ Configuration valid"
echo ""

echo "=== Ready to test ==="
echo ""
echo "Password: $PASSWORD"
echo "Port: $PORT"
echo ""
echo "Start server with:"
echo "  go run ./cmd/app"
echo ""
echo "Or with environment variables:"
echo "  AUTH_PASSWORD_HASH='$HASH' \\"
echo "  AUTH_SESSION_SECRET='$SECRET' \\"
echo "  APP_PORT=$PORT \\"
echo "  go run ./cmd/app"
echo ""
echo "Test URLs:"
echo "  Login: http://localhost:$PORT/login"
echo "  Protected: http://localhost:$PORT/ (will redirect to login)"
echo ""
