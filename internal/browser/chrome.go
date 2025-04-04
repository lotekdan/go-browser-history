package browser

import (
	"database/sql" // For SQL database interactions
	"fmt"          // For formatted output and error messages
	"os"           // For environment variables
	"path/filepath"
	"runtime" // For OS detection
	"strings"
	"time" // For time conversions

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// chromeHistoryQuery is the SQL query for retrieving Chrome history entries.
const chromeHistoryQuery = `
    SELECT url, title, last_visit_time 
    FROM urls 
    WHERE last_visit_time >= ? AND last_visit_time <= ?
    ORDER BY last_visit_time DESC`

// ChromeBrowser implements the Browser interface for Google Chrome.
type ChromeBrowser struct{}

// NewChromeBrowser creates a new instance of ChromeBrowser.
func NewChromeBrowser() Browser {
	return &ChromeBrowser{}
}

// GetHistoryPath retrieves the path to Chrome's history database file.
func (cb *ChromeBrowser) GetHistoryPath() (string, error) {
	switch runtime.GOOS {
	case "windows":
		fmt.Println(cb.GetHistoryPaths(os.Getenv("LOCALAPPDATA") + "\\Google\\Chrome\\User Data\\"))
		return os.Getenv("LOCALAPPDATA") + "\\Google\\Chrome\\User Data\\Default\\History", nil
	case "darwin":
		return os.Getenv("HOME") + "/Library/Application Support/Google/Chrome/Default/History", nil
	case "linux":
		fmt.Println(cb.GetHistoryPaths(os.Getenv("HOME") + "/.config/google-chrome/"))
		return os.Getenv("HOME") + "/.config/google-chrome/Default/History", nil
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// GetBrowserProfilePaths gets a collection of browser profile history paths.
func (cb *ChromeBrowser) GetHistoryPaths(dir string) ([]string, error) {
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

	for _, entry := range entries {
		if entry.IsDir() {
			if strings.Contains(strings.ToLower(entry.Name()), "profile") ||
				strings.Contains(strings.ToLower(entry.Name()), "default") {
				fullPath := filepath.Join(dir, entry.Name())
				profilePaths = append(profilePaths, fullPath)
			}
		}
	}
	if len(profilePaths) == 0 {
		return nil, os.ErrNotExist
	}

	return profilePaths, nil
}

// ExtractHistory extracts Chrome history entries within the given time range.
func (cb *ChromeBrowser) ExtractHistory(historyDBPath string, startTime, endTime time.Time, verbose bool) ([]HistoryEntry, error) {
	db, err := sql.Open("sqlite3", "file:"+historyDBPath+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("failed to open Chrome history database at %s: %v", historyDBPath, err)
	}
	defer db.Close()

	chromeStartTime := TimeToChromeTime(startTime)
	chromeEndTime := TimeToChromeTime(endTime)

	rows, err := db.Query(chromeHistoryQuery, chromeStartTime, chromeEndTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query Chrome history from %s: %v", historyDBPath, err)
	}
	defer rows.Close()

	var entries []HistoryEntry
	for rows.Next() {
		var pageURL, pageTitle string
		var visitTimestamp int64
		if err := rows.Scan(&pageURL, &pageTitle, &visitTimestamp); err != nil {
			return nil, fmt.Errorf("failed to scan Chrome history row from %s: %v", historyDBPath, err)
		}
		entries = append(entries, HistoryEntry{
			URL:       pageURL,
			Title:     pageTitle,
			Timestamp: ChromeTimeToTime(visitTimestamp),
		})
	}

	return entries, nil
}

// TimeToChromeTime converts Go time.Time to Chrome's timestamp format (microseconds since 1601-01-01).
func TimeToChromeTime(t time.Time) int64 {
	const epochDiff = 11644473600000000 // Microseconds from 1601-01-01 to 1970-01-01
	return t.UnixMicro() + epochDiff
}

// ChromeTimeToTime converts Chrome's timestamp to Go time.Time.
func ChromeTimeToTime(chromeTimestamp int64) time.Time {
	const epochDiff = 11644473600000000
	return time.UnixMicro(chromeTimestamp - epochDiff)
}
