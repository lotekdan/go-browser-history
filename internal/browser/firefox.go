package browser

import (
	"database/sql"  // For SQL database interactions
	"fmt"           // For formatted output and error messages
	"os"            // For operating system interactions like file reading
	"path/filepath" // For manipulating file paths
	"regexp"        // For regular expression parsing of profiles.ini
	"runtime"       // For determining the operating system
	"time"          // For handling time-related operations

	_ "github.com/mattn/go-sqlite3" // Import sqlite3 driver anonymously to register it with database/sql
)

// FirefoxBrowser implements the Browser interface for Mozilla Firefox.
type FirefoxBrowser struct{}

// NewFirefoxBrowser creates a new instance of FirefoxBrowser.
func NewFirefoxBrowser() Browser {
	return &FirefoxBrowser{} // Return a pointer to a new FirefoxBrowser struct
}

// GetHistoryPath retrieves the path to Firefox's history database file (places.sqlite).
func (f *FirefoxBrowser) GetHistoryPath() (string, error) {
	var profileDir string
	// Determine the Firefox profile directory based on the operating system
	switch runtime.GOOS {
	case "windows":
		// Windows path using APPDATA environment variable
		profileDir = os.Getenv("APPDATA") + "\\Mozilla\\Firefox\\Profiles"
	case "darwin":
		// macOS path using HOME environment variable
		profileDir = os.Getenv("HOME") + "/Library/Application Support/Firefox/Profiles"
	case "linux":
		// Linux path using HOME environment variable
		profileDir = os.Getenv("HOME") + "/.mozilla/firefox"
	default:
		// Return an error for unsupported operating systems
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	// Read profiles.ini to find the default profile
	iniPath := filepath.Join(filepath.Dir(profileDir), "profiles.ini")
	if _, err := os.Stat(iniPath); err == nil { // Check if profiles.ini exists
		iniData, err := os.ReadFile(iniPath)
		if err == nil { // Successfully read the file
			// Use regex to extract profile paths from profiles.ini
			re := regexp.MustCompile(`Path=(.+)$`)
			matches := re.FindAllStringSubmatch(string(iniData), -1)
			for _, match := range matches {
				if len(match) > 1 { // Ensure there's a captured group
					profilePath := filepath.Join(filepath.Dir(profileDir), match[1])
					dbPath := filepath.Join(profilePath, "places.sqlite")
					if _, err := os.Stat(dbPath); err == nil { // Check if places.sqlite exists
						return dbPath, nil // Return the first valid database path
					}
				}
			}
		}
	}

	// Fallback: Search for any profile directory containing places.sqlite
	profiles, err := filepath.Glob(profileDir + "/*")
	if err != nil || len(profiles) == 0 {
		return "", fmt.Errorf("could not find any Firefox profile") // No profiles found
	}
	for _, profile := range profiles {
		dbPath := profile + "/places.sqlite"
		if _, err := os.Stat(dbPath); err == nil { // Check if places.sqlite exists
			return dbPath, nil // Return the first valid database path
		}
	}
	// No valid profile with places.sqlite found
	return "", fmt.Errorf("no valid Firefox profile with places.sqlite found")
}

// ExtractHistory extracts Firefox history entries from the database within the given time range.
func (f *FirefoxBrowser) ExtractHistory(dbPath string, startTime, endTime time.Time) ([]HistoryEntry, error) {
	// Open the SQLite database in read-only mode to avoid locking issues
	db, err := sql.Open("sqlite3", "file:"+dbPath+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err) // Wrap error with context
	}
	defer db.Close() // Ensure the database connection is closed when done

	// Query joins moz_places and moz_historyvisits to get visit details, sorted by most recent
	rows, err := db.Query(`
        SELECT moz_places.url, moz_places.title, moz_historyvisits.visit_date
        FROM moz_places 
        JOIN moz_historyvisits ON moz_historyvisits.place_id = moz_places.id
        WHERE moz_historyvisits.visit_date >= ? AND moz_historyvisits.visit_date <= ?
        ORDER BY moz_historyvisits.visit_date DESC`,
		startTime.UnixMicro(), endTime.UnixMicro()) // Firefox uses microseconds since Unix epoch
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %v", err) // Wrap error with context
	}
	defer rows.Close() // Ensure query results are closed when done

	// Collect history entries from the query results
	var entries []HistoryEntry
	for rows.Next() { // Iterate over each row
		var url string
		var title sql.NullString // Handle potentially NULL titles with sql.NullString
		var timestamp int64
		// Scan row data into variables
		if err := rows.Scan(&url, &title, &timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err) // Wrap error with context
		}
		// Convert NULL title to empty string if invalid, use string value if valid
		titleStr := ""
		if title.Valid {
			titleStr = title.String
		}
		// Create and append the history entry
		entries = append(entries, HistoryEntry{
			URL:       url,
			Title:     titleStr,
			Timestamp: time.UnixMicro(timestamp), // Convert microseconds to time.Time
		})
	}

	return entries, nil // Return the collected history entries
}
