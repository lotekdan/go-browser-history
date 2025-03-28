#!/bin/bash

BIN_DIR="bin"
mkdir -p "$BIN_DIR"
ORIGINAL_CC="$CC"

# Get the latest Git tag and build number
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null)
if [ -z "$LATEST_TAG" ]; then
    # Fallback if no tags exist
    APP_VERSION="v0.0.0-0"
    echo "Warning: No Git tags found, using fallback version: $APP_VERSION" >&2
else
    # Count commits since the latest tag
    BUILD_NUMBER=$(git rev-list "$LATEST_TAG..HEAD" --count 2>/dev/null)
    if [ -z "$BUILD_NUMBER" ]; then
        BUILD_NUMBER="0"  # On the tag itself
    fi
    APP_VERSION="$LATEST_TAG-$BUILD_NUMBER"
fi

OS=$(uname | tr '[:upper:]' '[:lower:]')
echo "Building for $OS (amd64) with version $APP_VERSION..."
GOOS="$OS" GOARCH=amd64 CGO_ENABLED=1 go build -o "$BIN_DIR/go-browser-history-${OS}-amd64" -ldflags "-X main.Version=$APP_VERSION" ./cmd/
if [ $? -ne 0 ]; then echo "Build failed for $BIN_DIR/go-browser-history-${OS}-amd64" >&2; exit 1; fi

echo "Running unit tests..."
CGO_ENABLED=1 go test ./... -v --cover
#CGO_ENABLED=1 go test ./cmd/api -v --cover
#CGO_ENABLED=1 go test ./internal/browser -v --cover
#CGO_ENABLED=1 go test ./internal/browser -v --cover
#CGO_ENABLED=1 go test ./internal/config -v --cover
#CGO_ENABLED=1 go test ./internal/history -v --cover
#CGO_ENABLED=1 go test ./internal/server -v --cover
#CGO_ENABLED=1 go test ./internal/service -v --cover
if [ $? -ne 0 ]; then echo "Unit tests failed" >&2; exit 1; fi

BINARY="$BIN_DIR/go-browser-history-${OS}-amd64"
if [ ! -f "$BINARY" ]; then echo "Error: Binary not found at $BINARY" >&2; exit 1; fi

echo "Testing version flag..."
OUTPUT=$($BINARY --version)
if echo "$OUTPUT" | grep -q "$APP_VERSION"; then 
    echo "Version test passed: $OUTPUT"
else 
    echo "Version test failed. Expected '$APP_VERSION', got: $OUTPUT" >&2; exit 1
fi

echo "Testing all browsers (30 days)..."
OUTPUT=$($BINARY)
if [ -n "$OUTPUT" ]; then 
    echo "All browsers test passed (output received)"
else 
    echo "All browsers test failed (no output)" >&2; exit 1
fi

echo "Testing Chrome history (10 days)..."
OUTPUT=$($BINARY --days 10 --browser chrome)
if [ -n "$OUTPUT" ] && ! echo "$OUTPUT" | grep -q "\[edge\]" && ! echo "$OUTPUT" | grep -q "\[firefox\]"; then
    echo "Chrome-only test passed"
else
    echo "Chrome-only test failed (unexpected output or no output)" >&2; exit 1
fi

echo "Testing JSON output..."
OUTPUT=$($BINARY --days 5 --json)
if echo "$OUTPUT" | jq . >/dev/null 2>&1 || [ "$OUTPUT" = "[]" ]; then 
    echo "JSON output test passed"
else 
    echo "JSON output test failed (invalid JSON): $OUTPUT" >&2; exit 1
fi

export CC="$ORIGINAL_CC"
echo "Build and tests completed successfully. Binaries are in $BIN_DIR"