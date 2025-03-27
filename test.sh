#!/bin/bash

# Run unit tests
echo "Running unit tests..."
CGO_ENABLED=1 go test ./...
if [ $? -ne 0 ]; then
    echo "Unit tests failed" >&2
    exit 1
fi

# Ensure build script exists
if [ ! -f "build.sh" ]; then
    echo "Error: build.sh not found" >&2
    exit 1
fi

# Run the build
echo "Building the project..."
./build.sh
if [ $? -ne 0 ]; then
    echo "Build failed" >&2
    exit 1
fi

# Binary path
OS=$(uname | tr '[:upper:]' '[:lower:]')
BINARY="bin/go-browser-history-${OS}-amd64"
if [ ! -f "$BINARY" ]; then
    echo "Error: Binary not found at $BINARY" >&2
    exit 1
fi

# Test 1: Version check
echo "Testing version flag..."
OUTPUT=$($BINARY --version)
if echo "$OUTPUT" | grep -q "v1.0.0"; then
    echo "Version test passed: $OUTPUT"
else
    echo "Version test failed. Expected 'v1.0.0', got: $OUTPUT" >&2
    exit 1
fi

# Test 2: All browsers, default 30 days
echo "Testing all browsers (30 days)..."
OUTPUT=$($BINARY)
if [ -n "$OUTPUT" ]; then
    echo "All browsers test passed (output received)"
else
    echo "All browsers test failed (no output)" >&2
    exit 1
fi

# Test 3: Chrome only, 10 days
echo "Testing Chrome history (10 days)..."
OUTPUT=$($BINARY --days 10 --browser chrome)
if [ -n "$OUTPUT" ] && ! echo "$OUTPUT" | grep -q "edge" && ! echo "$OUTPUT" | grep -q "firefox"; then
    echo "Chrome-only test passed"
else
    echo "Chrome-only test failed (unexpected output or no output): $OUTPUT" >&2
    exit 1
fi

# Test 4: JSON output
echo "Testing JSON output..."
OUTPUT=$($BINARY --days 5 --json)
if echo "$OUTPUT" | jq . >/dev/null 2>&1; then
    echo "JSON output test passed"
else
    echo "JSON output test failed (invalid JSON): $OUTPUT" >&2
    exit 1
fi

echo "All tests passed!"