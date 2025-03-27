package browser

import (
	"fmt"     // For formatted error messages
	"os"      // For operating system interactions like environment variables
	"runtime" // For determining the operating system
	"time"    // For handling time-related operations
)

// EdgeBrowser implements the Browser interface for Microsoft Edge.
type EdgeBrowser struct{}

// NewEdgeBrowser creates a new instance of EdgeBrowser.
func NewEdgeBrowser() Browser {
	return &EdgeBrowser{} // Return a pointer to a new EdgeBrowser struct
}

// GetHistoryPath retrieves the path to Edge's history database file based on the OS.
func (e *EdgeBrowser) GetHistoryPath() (string, error) {
	switch runtime.GOOS { // Determine the operating system
	case "windows":
		// Windows path using LOCALAPPDATA environment variable
		return os.Getenv("LOCALAPPDATA") + "\\Microsoft\\Edge\\User Data\\Default\\History", nil
	case "darwin":
		// macOS path using HOME environment variable
		return os.Getenv("HOME") + "/Library/Application Support/Microsoft Edge/Default/History", nil
	case "linux":
		// Linux path using HOME environment variable
		return os.Getenv("HOME") + "/.config/microsoft-edge/Default/History", nil
	default:
		// Return an error for unsupported operating systems
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// ExtractHistory extracts Edge history entries from the database within the given time range.
// Edge shares the same database schema as Chrome, so it delegates to ChromeBrowser's implementation.
func (e *EdgeBrowser) ExtractHistory(dbPath string, startTime, endTime time.Time) ([]HistoryEntry, error) {
	chrome := &ChromeBrowser{} // Create a ChromeBrowser instance for shared logic
	// Delegate to Chrome's ExtractHistory since the database format is identical
	return chrome.ExtractHistory(dbPath, startTime, endTime)
}
