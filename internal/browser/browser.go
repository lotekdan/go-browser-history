package browser

import "time"

// HistoryEntry represents a single browser history entry
type HistoryEntry struct {
	URL       string
	Title     string
	Timestamp time.Time
}

// Browserම

// Browser defines the interface for different browser implementations
type Browser interface {
	GetHistoryPath() (string, error)
	ExtractHistory(dbPath string, startTime, endTime time.Time) ([]HistoryEntry, error)
}
