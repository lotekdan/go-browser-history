#!/bin/bash

# Ensure bin directory exists
BIN_DIR="bin"
mkdir -p "$BIN_DIR"

# Version info (optional, can be set via git tag or manually)
VERSION="v1.0.0"

# Build for Linux
echo "Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o "$BIN_DIR/go-browser-history-linux-amd64" -ldflags "-X main.Version=$VERSION" ./cmd/
if [ $? -ne 0 ]; then
    echo "Linux build failed" >&2
    exit 1
fi

# Build for macOS
echo "Building for macOS (amd64)..."
GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -o "$BIN_DIR/go-browser-history-darwin-amd64" -ldflags "-X main.Version=$VERSION" ./cmd/
if [ $? -ne 0 ]; then
    echo "macOS build failed" >&2
    exit 1
fi

# Build for Windows (requires cross-compiler, e.g., mingw-w64)
echo "Building for Windows (amd64)..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -o "$BIN_DIR/go-browser-history-windows-amd64.exe" -ldflags "-X main.Version=$VERSION" ./cmd/
if [ $? -ne 0 ]; then
    echo "Windows build failed (skipped if cross-compiler not set up)" >&2
fi

echo "Build completed. Binaries are in $BIN_DIR"