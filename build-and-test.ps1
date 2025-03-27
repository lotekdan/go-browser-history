# build-and-test.ps1

# Ensure bin directory exists
$binDir = "bin"
if (-not (Test-Path $binDir)) {
    New-Item -ItemType Directory -Path $binDir | Out-Null
}

# Version info (optional, can be set via git tag or manually)
$version = "v1.0.0"

# Store original CC value
$originalCC = $env:CC

# Build for Windows (use default Windows GCC)
Write-Host "Building for Windows (amd64)..."
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "1"
$env:CC = $null  # Use default Windows compiler
go build -o "$binDir/go-browser-history-windows-amd64.exe" -ldflags "-X main.Version=$version" ./cmd/
if ($LASTEXITCODE -ne 0) {
    Write-Host "Windows build failed" -ForegroundColor Red
    exit 1
}

# Run unit tests
Write-Host "Running unit tests..."
go test ./...
if ($LASTEXITCODE -ne 0) {
    Write-Host "Unit tests failed" -ForegroundColor Red
    exit 1
}

# Binary path
$binary = "$binDir/go-browser-history-windows-amd64.exe"
if (-not (Test-Path $binary)) {
    Write-Host "Error: Binary not found at $binary" -ForegroundColor Red
    exit 1
}

# Test 1: Version check
Write-Host "Testing version flag..."
$output = & $binary --version
if ($output -match "v1.0.0") {
    Write-Host "Version test passed: $output" -ForegroundColor Green
} else {
    Write-Host "Version test failed. Expected 'v1.0.0', got: $output" -ForegroundColor Red
    exit 1
}

# Test 2: All browsers, default 30 days
Write-Host "Testing all browsers (30 days)..."
$output = & $binary
if ($output) {
    Write-Host "All browsers test passed (output received)" -ForegroundColor Green
} else {
    Write-Host "All browsers test failed (no output)" -ForegroundColor Red
    exit 1
}

# Test 3: Chrome only, 10 days
Write-Host "Testing Chrome history (10 days)..."
$output = & $binary --days 10 --browser chrome
if ($output -and $output -notmatch "edge" -and $output -notmatch "firefox") {
    Write-Host "Chrome-only test passed" -ForegroundColor Green
} else {
    Write-Host "Chrome-only test failed (unexpected output or no output): $output" -ForegroundColor Red
    exit 1
}

# Test 4: JSON output
Write-Host "Testing JSON output..."
$output = & $binary --days 5 --json
try {
    $json = $output | ConvertFrom-Json
    if ($json -or $output -eq "[]") {  # Allow empty array as valid JSON
        Write-Host "JSON output test passed" -ForegroundColor Green
    } else {
        Write-Host "JSON output test failed (empty or invalid JSON): $output" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "JSON output test failed (invalid JSON): $output" -ForegroundColor Red
    exit 1
}

# Optional: Cross-compilation builds (commented out unless cross-compilers are set up)
<#
# Check for macOS cross-compiler
$macCompiler = "x86_64-apple-darwin21.6-clang"
if (Get-Command $macCompiler -ErrorAction SilentlyContinue) {
    Write-Host "Building for macOS (amd64)..."
    $env:GOOS = "darwin"
    $env:GOARCH = "amd64"
    $env:CC = $macCompiler
    go build -o "$binDir/go-browser-history-darwin-amd64" -ldflags "-X main.Version=$version" ./cmd/
    if ($LASTEXITCODE -ne 0) {
        Write-Host "macOS build failed" -ForegroundColor Red
    }
} else {
    Write-Host "Skipping macOS build (cross-compiler $macCompiler not found)" -ForegroundColor Yellow
}

# Check for Linux cross-compiler
$linuxCompiler = "x86_64-linux-gnu-gcc"
if (Get-Command $linuxCompiler -ErrorAction SilentlyContinue) {
    Write-Host "Building for Linux (amd64)..."
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    $env:CC = $linuxCompiler
    go build -o "$binDir/go-browser-history-linux-amd64" -ldflags "-X main.Version=$version" ./cmd/
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Linux build failed" -ForegroundColor Red
    }
} else {
    Write-Host "Skipping Linux build (cross-compiler $linuxCompiler not found)" -ForegroundColor Yellow
}
#>

# Restore original CC value
$env:CC = $originalCC

Write-Host "Build and tests completed successfully. Binaries are in $binDir" -ForegroundColor Green