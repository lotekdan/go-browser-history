## Go-Browser-History


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
        

## Building and Testing
### Windows
```powershell
.\build-and-test.ps1
```

Builds the Windows binary, runs unit tests, and tests the binary.

macOS/Linux

bash

```bash
chmod +x build-and-test.sh
./build-and-test.sh
```

Builds the native binary, runs unit tests, and tests the binary. Install jq for JSON validation if needed (brew install jq or sudo apt-get install jq).


## Usage

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

Feel free to open issues or submit pull requests on GitHub!