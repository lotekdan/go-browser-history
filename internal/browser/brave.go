package browser

import (
	"fmt"     // For error messages
	"os"      // For environment variables
	"runtime" // For OS detection
	"time"    // For time range parameters
)

// BraveBrowser implements the Browser interface for Microsoft Edge.
type BraveBrowser struct{}

// NewBraveBrowser creates a new instance of BraveBrowser.
func NewBraveBrowser() Browser {
	return &BraveBrowser{}
}

// GetHistoryPath retrieves the path to Edge's history database file.
func (bb *BraveBrowser) GetHistoryPath() ([]string, error) {
	switch runtime.GOOS {
	case "windows":
		paths, _ := bb.GetHistoryPaths(os.Getenv("LOCALAPPDATA") + "\\BraveSoftware\\Brave-Browser\\User Data")
		return paths, nil
	case "darwin":
		paths, _ := bb.GetHistoryPaths(os.Getenv("HOME") + "/Library/Application Support/BraveSoftware/Brave-Browser")
		return paths, nil
	case "linux":
		paths, _ := bb.GetHistoryPaths(os.Getenv("HOME") + "/.config/BraveSoftware/Brave-Browser")
		return paths, nil
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// ExtractHistory extracts Edge history entries, delegating to ChromeBrowser due to shared schema.
func (bb *BraveBrowser) ExtractHistory(historyDBPath string, startTime, endTime time.Time, verbose bool) ([]HistoryEntry, error) {
	chromeBrowser := &ChromeBrowser{}
	return chromeBrowser.ExtractHistory(historyDBPath, startTime, endTime, verbose)
}

// GetBrowserProfilePaths gets a collection of browser profile history paths, delegating to ChromeBrowser due to shared schema.
func (bb *BraveBrowser) GetHistoryPaths(dir string) ([]string, error) {
	chromeBrowser := &ChromeBrowser{}
	return chromeBrowser.GetHistoryPaths(dir)
}
