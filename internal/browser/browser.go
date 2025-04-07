package browser

import (
	"time"

	"github.com/lotekdan/go-browser-history/internal/history"
)

// Browser defines the interface for browser-specific history extraction.
type Browser interface {
	// GetHistoryPath retrieves the path to the browser's history database.
	GetHistoryPaths() ([]string, error)
	// ExtractHistory retrieves history entries from the database within the given time range, with optional verbose logging.
	ExtractHistory(dbPath string, startTime, endTime time.Time, verbose bool) ([]history.HistoryEntry, error)
}
