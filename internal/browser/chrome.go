package browser

import (
	"database/sql"
	"fmt"
	"os"
	"runtime"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import sqlite3 driver
)

type ChromeBrowser struct{}

func NewChromeBrowser() Browser {
	return &ChromeBrowser{}
}

func (c *ChromeBrowser) GetHistoryPath() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return os.Getenv("LOCALAPPDATA") + "\\Google\\Chrome\\User Data\\Default\\History", nil
	case "darwin":
		return os.Getenv("HOME") + "/Library/Application Support/Google/Chrome/Default/History", nil
	case "linux":
		return os.Getenv("HOME") + "/.config/google-chrome/Default/History", nil
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func (c *ChromeBrowser) ExtractHistory(dbPath string, startTime, endTime time.Time) ([]HistoryEntry, error) {
	db, err := sql.Open("sqlite3", "file:"+dbPath+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	// Chrome timestamps are microseconds since 1601-01-01
	chromeStartTime := timeToChromeTime(startTime)
	chromeEndTime := timeToChromeTime(endTime)

	rows, err := db.Query(`
        SELECT url, title, last_visit_time 
        FROM urls 
        WHERE last_visit_time >= ? AND last_visit_time <= ?
        ORDER BY last_visit_time DESC`,
		chromeStartTime, chromeEndTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %v", err)
	}
	defer rows.Close()

	var entries []HistoryEntry
	for rows.Next() {
		var url, title string
		var timestamp int64
		if err := rows.Scan(&url, &title, &timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		entries = append(entries, HistoryEntry{
			URL:       url,
			Title:     title,
			Timestamp: chromeTimeToTime(timestamp),
		})
	}

	return entries, nil
}

// timeToChromeTime converts Go time.Time to Chrome timestamp
func timeToChromeTime(t time.Time) int64 {
	epochDiff := int64(11644473600000000) // Microseconds from 1601-01-01 to 1970-01-01
	return t.UnixMicro() + epochDiff
}

// chromeTimeToTime converts Chrome timestamp to Go time.Time
func chromeTimeToTime(chromeTimestamp int64) time.Time {
	epochDiff := int64(11644473600000000)
	unixMicro := chromeTimestamp - epochDiff
	return time.UnixMicro(unixMicro)
}
