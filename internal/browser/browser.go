package browser

import "time" // For timestamp handling

// HistoryEntry represents a single browser history entry.
type HistoryEntry struct {
	URL       string    // Webpage URL
	Title     string    // Webpage title
	Timestamp time.Time // Visit timestamp
}

// Browser defines the interface for browser-specific history extraction.
type Browser interface {
	// GetHistoryPath retrieves the path to the browser's history database.
	GetHistoryPath() (string, error)
	// ExtractHistory retrieves history entries from the database within the given time range.
	ExtractHistory(dbPath string, startTime, endTime time.Time) ([]HistoryEntry, error)
}
