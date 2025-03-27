package browser

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

type EdgeBrowser struct{}

func NewEdgeBrowser() Browser {
	return &EdgeBrowser{}
}

func (e *EdgeBrowser) GetHistoryPath() (string, error) {
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

func (e *EdgeBrowser) ExtractHistory(dbPath string, startTime, endTime time.Time) ([]HistoryEntry, error) {
	// Edge uses the same database schema as Chrome
	chrome := &ChromeBrowser{}
	return chrome.ExtractHistory(dbPath, startTime, endTime)
}
