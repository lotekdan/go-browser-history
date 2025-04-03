package browser

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HistoryEntry represents a single browser history entry.
type HistoryEntry struct {
	URL       string
	Title     string
	Timestamp time.Time
}

// Browser defines the interface for browser-specific history extraction.
type Browser interface {
	// GetHistoryPath retrieves the path to the browser's history database.
	GetHistoryPath() (string, error)
	// ExtractHistory retrieves history entries from the database within the given time range, with optional verbose logging.
	ExtractHistory(dbPath string, startTime, endTime time.Time, verbose bool) ([]HistoryEntry, error)
}

func GetBrowserProfilePaths(dir string, browserType string) ([]string, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, os.ErrInvalid
	}

	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	entries, err := f.ReadDir(-1)
	if err != nil {
		return nil, err
	}

	var profilePaths []string

	switch browserType {
	case "firefox":
		//do something
	default:
		for _, entry := range entries {
			if entry.IsDir() {
				if strings.Contains(strings.ToLower(entry.Name()), "profile") {
					fullPath := filepath.Join(dir, entry.Name())
					profilePaths = append(profilePaths, fullPath)
				}
			}
		}

		if len(profilePaths) == 0 {
			return nil, os.ErrNotExist
		}

	}
	return profilePaths, nil
}
