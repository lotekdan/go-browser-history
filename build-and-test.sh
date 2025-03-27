#!/bin/bash

BIN_DIR="bin"
mkdir -p "$BIN_DIR"
VERSION="v1.0.0"
ORIGINAL_CC="$CC"

OS=$(uname | tr '[:upper:]' '[:lower:]')
echo "Building for $OS (amd64)..."
GOOS="$OS" GOARCH=amd64 CGO_ENABLED=1 go build -o "$BIN_DIR/go-browser-history-${OS}-amd64" -ldflags "-X main.Version=$VERSION" ./cmd/
if [ $? -ne 0 ]; then echo "Build failed" >&2; exit 1; fi

echo "Running unit tests..."
CGO_ENABLED=1 go test ./...
if [ $? -ne 0 ]; then echo "Unit tests failed" >&2; exit 1; fi

BINARY="$BIN_DIR/go-browser-history-${OS}-amd64"
if [ ! -f "$BINARY" ]; then echo "Error: Binary not found at $BINARY" >&2; exit 1; fi

echo "Testing version flag..."
OUTPUT=$($BINARY --version)
if echo "$OUTPUT" | grep -q "v1.0.0"; then echo "Version test passed: $OUTPUT"; else echo "Version test failed. Expected 'v1.0.0', got: $OUTPUT" >&2; exit 1; fi

echo "Testing all browsers (30 days)..."
OUTPUT=$($BINARY)
if [ -n "$OUTPUT" ]; then echo "All browsers test passed (output received)"; else echo "All browsers test failed (no output)" >&2; exit 1; fi

echo "Testing Chrome history (10 days)..."
OUTPUT=$($BINARY --days 10 --browser chrome)
if [ -n "$OUTPUT" ] && ! echo "$OUTPUT" | grep -q "\[edge\]" && ! echo "$OUTPUT" | grep -q "\[firefox\]"; then
    echo "Chrome-only test passed"
else
    echo "Chrome-only test failed (unexpected output or no output)" >&2; exit 1
fi

echo "Testing JSON output..."
OUTPUT=$($BINARY --days 5 --json)
if echo "$OUTPUT" | jq . >/dev/null 2>&1 || [ "$OUTPUT" = "[]" ]; then echo "JSON output test passed"; else echo "JSON output test failed (invalid JSON): $OUTPUT" >&2; exit 1; fi

export CC="$ORIGINAL_CC"
echo "Build and tests completed successfully. Binaries are in $BIN_DIR"