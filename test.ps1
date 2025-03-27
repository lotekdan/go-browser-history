# test.ps1

# Run unit tests
Write-Host "Running unit tests..."
$env:CGO_ENABLED = "1"
go test ./...
if ($LASTEXITCODE -ne 0) {
    Write-Host "Unit tests failed" -ForegroundColor Red
    exit 1
}

# Ensure build script exists
if (-not (Test-Path "build.ps1")) {
    Write-Host "Error: build.ps1 not found" -ForegroundColor Red
    exit 1
}

# Run the build
Write-Host "Building the project..."
.\build.ps1
if ($LASTEXITCODE -ne 0) {
    Write-Host "Build failed" -ForegroundColor Red
    exit 1
}

# Binary path
$binary = "bin/go-browser-history-windows-amd64.exe"
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
    if ($json) {
        Write-Host "JSON output test passed" -ForegroundColor Green
    } else {
        Write-Host "JSON output test failed (empty or invalid JSON)" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "JSON output test failed (invalid JSON): $output" -ForegroundColor Red
    exit 1
}

Write-Host "All tests passed!" -ForegroundColor Green