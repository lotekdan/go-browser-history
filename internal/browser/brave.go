package browser

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

// BraveBrowser implements the Browser interface for Brave Browser.
type BraveBrowser struct{}

// NewBraveBrowser creates a new instance of BraveBrowser.
func NewBraveBrowser() Browser {
	return &BraveBrowser{}
}

// GetHistoryPath retrieves the path to Brave's history database file.
func (bb *BraveBrowser) GetHistoryPath() ([]string, error) {
	switch runtime.GOOS {
	case "windows":
		return bb.GetHistoryPaths(os.Getenv("LOCALAPPDATA") + "\\BraveSoftware\\Brave-Browser\\User Data")
	case "darwin":
		return bb.GetHistoryPaths(os.Getenv("HOME") + "/Library/Application Support/BraveSoftware/Brave-Browser")
	case "linux":
		return bb.GetHistoryPaths(os.Getenv("HOME") + "/.config/BraveSoftware/Brave-Browser")
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// ExtractHistory extracts Brave history entries, delegating to ChromeBrowser due to shared schema.
func (bb *BraveBrowser) ExtractHistory(historyDBPath string, startTime, endTime time.Time, verbose bool) ([]HistoryEntry, error) {
	chromeBrowser := &ChromeBrowser{}
	return chromeBrowser.ExtractHistory(historyDBPath, startTime, endTime, verbose)
}

// GetHistoryPaths gets a collection of browser profile history paths, delegating to ChromeBrowser due to shared schema.
func (bb *BraveBrowser) GetHistoryPaths(dir string) ([]string, error) {
	chromeBrowser := &ChromeBrowser{}
	return chromeBrowser.GetHistoryPaths(dir)
}
