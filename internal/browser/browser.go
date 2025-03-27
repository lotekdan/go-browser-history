package browser

import "time"

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
