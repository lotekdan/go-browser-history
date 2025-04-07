package browser

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/lotekdan/go-browser-history/internal/history"
)

// BraveBrowser implements the Browser interface for Brave Browser.
type BraveBrowser struct{}

// NewBraveBrowser creates a new instance of BraveBrowser.
func NewBraveBrowser() Browser {
	return &BraveBrowser{}
}

// GetHistoryPath retrieves the path to Brave's history database file.
func (bb *BraveBrowser) GetHistoryPaths() ([]history.HistoryPathEntry, error) {
	switch runtime.GOOS {
	case "windows":
		return bb.getPaths(os.Getenv("LOCALAPPDATA") + "\\BraveSoftware\\Brave-Browser\\User Data")
	case "darwin":
		return bb.getPaths(os.Getenv("HOME") + "/Library/Application Support/BraveSoftware/Brave-Browser")
	case "linux":
		return bb.getPaths(os.Getenv("HOME") + "/.config/BraveSoftware/Brave-Browser")
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// ExtractHistory extracts Brave history entries, delegating to ChromeBrowser due to shared schema.
func (bb *BraveBrowser) ExtractHistory(historyDBPath, profile string, startTime, endTime time.Time, verbose bool) ([]history.HistoryEntry, error) {
	chromeBrowser := &ChromeBrowser{}
	return chromeBrowser.ExtractHistory(historyDBPath, profile, startTime, endTime, verbose)
}

// GetHistoryPaths gets a collection of browser profile history paths, delegating to ChromeBrowser due to shared schema.
func (bb *BraveBrowser) getPaths(dir string) ([]history.HistoryPathEntry, error) {
	chromeBrowser := &ChromeBrowser{}
	return chromeBrowser.getPaths(dir)
}
