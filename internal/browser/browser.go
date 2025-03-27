package browser

import "time" // For handling time-related operations like timestamps

// HistoryEntry represents a single browser history entry.
type HistoryEntry struct {
	URL       string    // The URL of the visited webpage
	Title     string    // The title of the webpage
	Timestamp time.Time // The time the webpage was visited
}

// Browser defines the interface for different browser implementations.
// It specifies the methods that all browser types (e.g., Chrome, Firefox) must implement.
type Browser interface {
	// GetHistoryPath retrieves the file path to the browser's history database.
	// Returns the path as a string or an error if the path cannot be determined.
	GetHistoryPath() (string, error)

	// ExtractHistory extracts history entries from the specified database file.
	// It filters entries between startTime and endTime, returning them as a slice
	// of HistoryEntry or an error if extraction fails.
	ExtractHistory(dbPath string, startTime, endTime time.Time) ([]HistoryEntry, error)
}
