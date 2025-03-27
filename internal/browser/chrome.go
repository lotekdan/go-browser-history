package browser

import (
	"database/sql" // For SQL database interactions
	"fmt"          // For formatted output and error messages
	"os"           // For operating system interactions like environment variables
	"runtime"      // For determining the operating system
	"time"         // For handling time-related operations

	_ "github.com/mattn/go-sqlite3" // Import sqlite3 driver anonymously to register it with database/sql
)

// ChromeBrowser implements the Browser interface for Google Chrome.
type ChromeBrowser struct{}

// NewChromeBrowser creates a new instance of ChromeBrowser.
func NewChromeBrowser() Browser {
	return &ChromeBrowser{} // Return a pointer to a new ChromeBrowser struct
}

// GetHistoryPath retrieves the path to Chrome's history database file based on the OS.
func (c *ChromeBrowser) GetHistoryPath() (string, error) {
	switch runtime.GOOS { // Determine the operating system
	case "windows":
		// Windows path using LOCALAPPDATA environment variable
		return os.Getenv("LOCALAPPDATA") + "\\Google\\Chrome\\User Data\\Default\\History", nil
	case "darwin":
		// macOS path using HOME environment variable
		return os.Getenv("HOME") + "/Library/Application Support/Google/Chrome/Default/History", nil
	case "linux":
		// Linux path using HOME environment variable
		return os.Getenv("HOME") + "/.config/google-chrome/Default/History", nil
	default:
		// Return an error for unsupported operating systems
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// ExtractHistory extracts Chrome history entries from the database within the given time range.
func (c *ChromeBrowser) ExtractHistory(dbPath string, startTime, endTime time.Time) ([]HistoryEntry, error) {
	// Open the SQLite database in read-only mode to avoid locking issues
	db, err := sql.Open("sqlite3", "file:"+dbPath+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err) // Wrap error with context
	}
	defer db.Close() // Ensure the database connection is closed when done

	// Convert Go time.Time to Chrome's timestamp format (microseconds since 1601-01-01)
	chromeStartTime := timeToChromeTime(startTime)
	chromeEndTime := timeToChromeTime(endTime)

	// Query the database for history entries within the time range, sorted by most recent
	rows, err := db.Query(`
        SELECT url, title, last_visit_time 
        FROM urls 
        WHERE last_visit_time >= ? AND last_visit_time <= ?
        ORDER BY last_visit_time DESC`,
		chromeStartTime, chromeEndTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %v", err) // Wrap error with context
	}
	defer rows.Close() // Ensure query results are closed when done

	// Collect history entries from the query results
	var entries []HistoryEntry
	for rows.Next() { // Iterate over each row
		var url, title string
		var timestamp int64
		// Scan row data into variables
		if err := rows.Scan(&url, &title, &timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err) // Wrap error with context
		}
		// Convert Chrome timestamp back to Go time.Time and create entry
		entries = append(entries, HistoryEntry{
			URL:       url,
			Title:     title,
			Timestamp: chromeTimeToTime(timestamp),
		})
	}

	return entries, nil // Return the collected history entries
}

// timeToChromeTime converts Go time.Time to Chrome timestamp.
// Chrome uses microseconds since 1601-01-01 as its epoch.
func timeToChromeTime(t time.Time) int64 {
	epochDiff := int64(11644473600000000) // Microseconds from 1601-01-01 to 1970-01-01
	return t.UnixMicro() + epochDiff      // Convert Unix microseconds to Chrome's epoch
}

// chromeTimeToTime converts Chrome timestamp to Go time.Time.
// Reverses the conversion from Chrome's epoch to Unix epoch.
func chromeTimeToTime(chromeTimestamp int64) time.Time {
	epochDiff := int64(11644473600000000)    // Microseconds from 1601-01-01 to 1970-01-01
	unixMicro := chromeTimestamp - epochDiff // Convert Chrome microseconds to Unix microseconds
	return time.UnixMicro(unixMicro)         // Create time.Time from Unix microseconds
}
