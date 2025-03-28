# build-and-test.ps1

# Ensure bin directory exists
$binDir = "bin"
if (-not (Test-Path $binDir)) {
    New-Item -ItemType Directory -Path $binDir | Out-Null
}

# Get the latest Git tag and build number
$latestTag = git describe --tags --abbrev=0 2>$null
if (-not $latestTag) {
    # Fallback if no tags exist
    $appVersion = "v0.0.0-0"
    Write-Host "Warning: No Git tags found, using fallback version: $appVersion" -ForegroundColor Yellow
} else {
    # Count commits since the latest tag
    $buildNumber = (git rev-list "$latestTag..HEAD" --count 2>$null)
    if (-not $buildNumber) {
        $buildNumber = "0"  # On the tag itself
    }
    $appVersion = "$latestTag-$buildNumber"
}

# Store original CC value
$originalCC = $env:CC

# Build for Windows
Write-Host "Building for Windows (amd64) with version $appVersion..."
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "1"
$env:CC = $null
go build -o "$binDir/go-browser-history-windows-amd64.exe" -ldflags "-X main.Version=$appVersion" ./cmd/
if ($LASTEXITCODE -ne 0) {
    Write-Host "Windows build failed for $binDir/go-browser-history-windows-amd64.exe" -ForegroundColor Red
    exit 1
}

# Run unit tests
Write-Host "Running unit tests..."
go test ./... -v --cover
#go test ./cmd/api -v --cover
#go test ./internal/browser -v --cover
#go test ./internal/browser -v --cover
#go test ./internal/config -v --cover
#go test ./internal/history -v --cover
#go test ./internal/server -v --cover
#go test ./internal/service -v --cover
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
if ($output -match [regex]::Escape($appVersion)) {
    Write-Host "Version test passed: $output" -ForegroundColor Green
} else {
    Write-Host "Version test failed. Expected '$appVersion', got: $output" -ForegroundColor Red
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
if ($output -and $output -notmatch "\[edge\]" -and $output -notmatch "\[firefox\]") {
    Write-Host "Chrome-only test passed" -ForegroundColor Green
} else {
    Write-Host "Chrome-only test failed (unexpected output or no output)" -ForegroundColor Red
    exit 1
}

# Test 4: JSON output
Write-Host "Testing JSON output..."
$output = & $binary --days 5 --json
try {
    $json = $output | ConvertFrom-Json
    if ($json -or $output -eq "[]") {
        Write-Host "JSON output test passed" -ForegroundColor Green
    } else {
        Write-Host "JSON output test failed (empty or invalid JSON)" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "JSON output test failed (invalid JSON): $output" -ForegroundColor Red
    exit 1
}

# Restore original CC value
$env:CC = $originalCC

Write-Host "Build and tests completed successfully. Binaries are in $binDir" -ForegroundColor Green