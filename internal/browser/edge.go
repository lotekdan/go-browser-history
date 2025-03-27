package browser

import (
	"fmt"     // For error messages
	"os"      // For environment variables
	"runtime" // For OS detection
	"time"    // For time range parameters
)

// EdgeBrowser implements the Browser interface for Microsoft Edge.
type EdgeBrowser struct{}

// NewEdgeBrowser creates a new instance of EdgeBrowser.
func NewEdgeBrowser() Browser {
	return &EdgeBrowser{}
}

// GetHistoryPath retrieves the path to Edge's history database file.
func (eb *EdgeBrowser) GetHistoryPath() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return os.Getenv("LOCALAPPDATA") + "\\Microsoft\\Edge\\User Data\\Default\\History", nil
	case "darwin":
		return os.Getenv("HOME") + "/Library/Application Support/Microsoft Edge/Default/History", nil
	case "linux":
		return os.Getenv("HOME") + "/.config/microsoft-edge/Default/History", nil
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// ExtractHistory extracts Edge history entries, delegating to ChromeBrowser due to shared schema.
func (eb *EdgeBrowser) ExtractHistory(historyDBPath string, startTime, endTime time.Time) ([]HistoryEntry, error) {
	chromeBrowser := &ChromeBrowser{}
	return chromeBrowser.ExtractHistory(historyDBPath, startTime, endTime)
}
