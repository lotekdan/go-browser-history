

## go-browser-history


`go-browser-history` is a command-line tool written in Go that retrieves browsing history from Google Chrome, Microsoft Edge, and Mozilla Firefox across Windows, macOS, and Linux. It supports filtering history by a specified number of days and can output results in either human-readable text or JSON format. The tool handles locked database files by creating temporary copies, making it robust even when browsers are running.

Features

-   Retrieve history from Chrome, Edge, and Firefox.
    
-   Filter by time range (e.g., last 30 days).
    
-   Output in text or JSON format.
    
-   Cross-platform support: Windows, macOS, Linux.
    
-   Handles locked history files gracefully.
    

Prerequisites

-   Go: Version 1.22 or higher installed.
    
-   GCC: Required for CGO (SQLite dependency).
    
    -   Windows: MinGW-w64 (e.g., via MSYS2: pacman -S mingw-w64-x86_64-gcc).
        
    -   macOS: Xcode command-line tools (xcode-select --install).
        
    -   Linux: GCC (sudo apt-get install build-essential on Ubuntu).
        

Building the Tool

Binaries are built into the bin/ directory with platform-specific names (e.g., go-browser-history-windows-amd64.exe).

Windows

1.  Ensure MinGW-w64’s gcc is in your PATH (test with gcc --version).
    
2.  Run the PowerShell build script:
    
    powershell
    
    ```powershell
    .\build.ps1
    ```
    
    -   Outputs: bin/go-browser-history-windows-amd64.exe.
        
    -   Cross-compilation for macOS/Linux requires additional setup (see below).
        

macOS/Linux

1.  Ensure GCC is installed.
    
2.  Make the Bash script executable:
    
    bash
    
    ```bash
    chmod +x build.sh
    ```
    
3.  Run:
    
    bash
    
    ```bash
    ./build.sh
    ```
    
    -   Outputs: bin/go-browser-history-<os>-amd64 (e.g., darwin-amd64 for macOS, linux-amd64 for Linux).
        
    -   Cross-compilation for Windows requires mingw-w64 (e.g., brew install mingw-w64 on macOS).
        

Cross-Compilation (Optional)

To build for all platforms from one machine:

-   Windows: Install osxcross for macOS and a Linux cross-compiler (e.g., x86_64-linux-gnu-gcc) for Linux.
    
-   macOS/Linux: Install mingw-w64 for Windows.
    
-   Update CC variables in the scripts with the correct cross-compiler paths.
    
-   Alternatively, use Docker:
    
    bash
    
    ```bash
    docker run --rm -v $(pwd):/go/src/project -w /go/src/project golang:cross bash -c "GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -o bin/go-browser-history-windows-amd64.exe ./cmd/ && GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o bin/go-browser-history-linux-amd64 ./cmd/ && GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -o bin/go-browser-history-darwin-amd64 ./cmd/"
    ```
    

Usage

Run the built binary with the following arguments:

bash

```bash
go-browser-history [flags]
```

Supported Arguments

```text
  -b, --browser string   Browser type (chrome, edge, firefox). Leave empty for all browsers
  -d, --days int         Number of days of history to retrieve (default 30)
  -h, --help             Show help message
  -j, --json             Output results in JSON format
      --version          Show version (set during build, e.g., v1.0.0)
```

Examples

-   Get history from all browsers for the last 10 days (text output):
    
    bash
    
    ```bash
    go-browser-history --days 10
    ```
    
-   Get Chrome history for the last 30 days in JSON:
    
    bash
    
    ```bash
    go-browser-history --browser chrome --json
    ```
    
-   Show version:
    
    bash
    
    ```bash
    go-browser-history --version
    ```
    

Notes

-   Browser State: Close browsers before running to avoid locked database issues, though the tool will attempt to copy locked files if needed.
    
-   Permissions: Ensure the tool has read access to browser history files (e.g., %LOCALAPPDATA%\Google\Chrome\User Data\Default\History on Windows).
    
-   Dependencies: Managed via go.mod. Run go mod tidy to ensure all are fetched.
    

Contributing

Feel free to use how you see fit. Another prompt based project.