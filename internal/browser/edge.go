package browser

import (
	"fmt"     // For error messages
	"os"      // For environment variables
	"runtime" // For OS detection
	"time"    // For time range parameters

	"github.com/lotekdan/go-browser-history/internal/history"
)

// EdgeBrowser implements the Browser interface for Microsoft Edge.
type EdgeBrowser struct{}

// NewEdgeBrowser creates a new instance of EdgeBrowser.
func NewEdgeBrowser() Browser {
	return &EdgeBrowser{}
}

// GetHistoryPath retrieves the path to Edge's history database file.
func (eb *EdgeBrowser) GetHistoryPaths() ([]history.HistoryPathEntry, error) {
	switch runtime.GOOS {
	case "windows":
		return eb.getPaths(os.Getenv("LOCALAPPDATA") + "\\Microsoft\\Edge\\User Data")
	case "darwin":
		return eb.getPaths(os.Getenv("HOME") + "/Library/Application Support/Microsoft Edge")
	case "linux":
		return eb.getPaths(os.Getenv("HOME") + "/.config/microsoft-edge/")
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// ExtractHistory extracts Edge history entries, delegating to ChromeBrowser due to shared schema.
func (eb *EdgeBrowser) ExtractHistory(historyDBPath, profile string, startTime, endTime time.Time, verbose bool) ([]history.HistoryEntry, error) {
	chromeBrowser := &ChromeBrowser{}
	return chromeBrowser.ExtractHistory(historyDBPath, profile, startTime, endTime, verbose)
}

// GetBrowserProfilePaths gets a collection of browser profile history paths, delegating to ChromeBrowser due to shared schema.
func (eb *EdgeBrowser) getPaths(dir string) ([]history.HistoryPathEntry, error) {
	chromeBrowser := &ChromeBrowser{}
	return chromeBrowser.getPaths(dir)
}
