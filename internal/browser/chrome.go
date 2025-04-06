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
	SELECT
		url.url,
		url.title,
		url.visit_count,
		url.typed_count,
		CASE (visit.transition & 0xFF)
			WHEN 0 THEN 'LINK'
			WHEN 1 THEN 'TYPED'
			WHEN 2 THEN 'AUTO_BOOKMARK'
			WHEN 3 THEN 'AUTO_SUBFRAME'
			WHEN 4 THEN 'MANUAL_SUBFRAME'
			WHEN 5 THEN 'GENERATED'
			WHEN 6 THEN 'AUTO_TOPLEVEL'
			WHEN 7 THEN 'FORM_SUBMIT'
			WHEN 8 THEN 'RELOAD'
			WHEN 9 THEN 'KEYWORD'
			WHEN 10 THEN 'KEYWORD_GENERATED'
			ELSE 'UNKNOWN_CORE (' || (visit.transition & 0xFF) || ')'
		END || 
		CASE
			WHEN (visit.transition & 0x10000000) = 0x10000000 THEN ' (REDIRECT)'
			WHEN (visit.transition & 0x40000000) = 0x40000000 THEN ' (CLIENT_REDIRECT)'
			WHEN (visit.transition & 0x80000000) = 0x80000000 THEN ' (SERVER_REDIRECT)'
			WHEN (visit.transition & 0x01000000) = 0x01000000 THEN ' (FORWARD_BACK)'
			WHEN (visit.transition & 0x02000000) = 0x02000000 THEN ' (FROM_ADDRESS_BAR)'
			ELSE ''
		END AS transition_desc,
		visit.visit_time
	FROM urls url
	JOIN visits visit ON visit.url = url.id
	WHERE last_visit_time >= ? AND last_visit_time <= ?
	ORDER BY visit.visit_time DESC;`

// ChromeBrowser implements the Browser interface for Google Chrome.
type ChromeBrowser struct{}

// NewChromeBrowser creates a new instance of ChromeBrowser.
func NewChromeBrowser() Browser {
	return &ChromeBrowser{}
}

// GetHistoryPath retrieves collection of paths to Chrome's history database file.
func (cb *ChromeBrowser) GetHistoryPath() ([]string, error) {
	switch runtime.GOOS {
	case "windows":
		paths, _ := cb.GetHistoryPaths(os.Getenv("LOCALAPPDATA") + "\\Google\\Chrome\\User Data")
		return paths, nil
	case "darwin":
		paths, _ := cb.GetHistoryPaths(os.Getenv("HOME") + "/Library/Application Support/Google/Chrome")
		return paths, nil
	case "linux":
		paths, _ := cb.GetHistoryPaths(os.Getenv("HOME") + "/.config/google-chrome")
		return paths, nil
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
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
			if strings.HasPrefix(strings.ToLower(entry.Name()), "profile") ||
				strings.Contains(strings.ToLower(entry.Name()), "default") {
				fullPath := filepath.Join(dir, entry.Name(), "History")
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
		var pageURL, pageTitle, pageVisitType string
		var pageVisitCount, pageTyped int
		var visitTimestamp int64
		//var ProfileName string
		if err := rows.Scan(&pageURL,
			&pageTitle,
			&pageVisitCount,
			&pageTyped,
			&pageVisitType,
			&visitTimestamp); err != nil {
			return nil, fmt.Errorf("failed to scan Chrome history row from %s: %v", historyDBPath, err)
		}
		entries = append(entries, HistoryEntry{
			URL:        pageURL,
			Title:      pageTitle,
			VisitCount: pageVisitCount,
			Typed:      pageTyped,
			VisitType:  pageVisitType,
			Timestamp:  ChromeTimeToTime(visitTimestamp),
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
