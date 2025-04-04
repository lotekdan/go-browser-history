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
func (eb *EdgeBrowser) GetHistoryPath() ([]string, error) {
	switch runtime.GOOS {
	case "windows":
		paths, _ := eb.GetHistoryPaths(os.Getenv("LOCALAPPDATA") + "\\Microsoft\\Edge\\User Data")
		return paths, nil
	case "darwin":
		paths, _ := eb.GetHistoryPaths(os.Getenv("HOME") + "/Library/Application Support/Microsoft Edge")
		return paths, nil
	case "linux":
		paths, _ := eb.GetHistoryPaths(os.Getenv("HOME") + "/.config/microsoft-edge/")
		return paths, nil
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// ExtractHistory extracts Edge history entries, delegating to ChromeBrowser due to shared schema.
func (eb *EdgeBrowser) ExtractHistory(historyDBPath string, startTime, endTime time.Time, verbose bool) ([]HistoryEntry, error) {
	chromeBrowser := &ChromeBrowser{}
	return chromeBrowser.ExtractHistory(historyDBPath, startTime, endTime, verbose)
}

// GetBrowserProfilePaths gets a collection of browser profile history paths, delegating to ChromeBrowser due to shared schema.
func (eb *EdgeBrowser) GetHistoryPaths(dir string) ([]string, error) {
	chromeBrowser := &ChromeBrowser{}
	return chromeBrowser.GetHistoryPaths(dir)
}
