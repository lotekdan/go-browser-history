# build.ps1

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

# Check for macOS cross-compiler
$macCompiler = "x86_64-apple-darwin21.6-clang"
if (Get-Command $macCompiler -ErrorAction SilentlyContinue) {
    Write-Host "Building for macOS (amd64)..."
    $env:GOOS = "darwin"
    $env:GOARCH = "amd64"
    $env:CGO_ENABLED = "1"
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
    $env:CGO_ENABLED = "1"
    $env:CC = $linuxCompiler
    go build -o "$binDir/go-browser-history-linux-amd64" -ldflags "-X main.Version=$version" ./cmd/
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Linux build failed" -ForegroundColor Red
    }
} else {
    Write-Host "Skipping Linux build (cross-compiler $linuxCompiler not found)" -ForegroundColor Yellow
}

# Restore original CC value
$env:CC = $originalCC

Write-Host "Build completed. Binaries are in $binDir" -ForegroundColor Green