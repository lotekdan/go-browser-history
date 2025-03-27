package browser

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type FirefoxBrowser struct{}

func NewFirefoxBrowser() Browser {
	return &FirefoxBrowser{}
}

func (f *FirefoxBrowser) GetHistoryPath() (string, error) {
	var profileDir string
	switch runtime.GOOS {
	case "windows":
		profileDir = os.Getenv("APPDATA") + "\\Mozilla\\Firefox\\Profiles"
	case "darwin":
		profileDir = os.Getenv("HOME") + "/Library/Application Support/Firefox/Profiles"
	case "linux":
		profileDir = os.Getenv("HOME") + "/.mozilla/firefox"
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	// Read profiles.ini to find the default profile
	iniPath := filepath.Join(filepath.Dir(profileDir), "profiles.ini")
	if _, err := os.Stat(iniPath); err == nil {
		iniData, err := os.ReadFile(iniPath)
		if err == nil {
			re := regexp.MustCompile(`Path=(.+)$`)
			matches := re.FindAllStringSubmatch(string(iniData), -1)
			for _, match := range matches {
				if len(match) > 1 {
					profilePath := filepath.Join(filepath.Dir(profileDir), match[1])
					dbPath := filepath.Join(profilePath, "places.sqlite")
					if _, err := os.Stat(dbPath); err == nil {
						return dbPath, nil
					}
				}
			}
		}
	}

	// Fallback: Look for any profile with places.sqlite
	profiles, err := filepath.Glob(profileDir + "/*")
	if err != nil || len(profiles) == 0 {
		return "", fmt.Errorf("could not find any Firefox profile")
	}
	for _, profile := range profiles {
		dbPath := profile + "/places.sqlite"
		if _, err := os.Stat(dbPath); err == nil {
			return dbPath, nil
		}
	}
	return "", fmt.Errorf("no valid Firefox profile with places.sqlite found")
}

func (f *FirefoxBrowser) ExtractHistory(dbPath string, startTime, endTime time.Time) ([]HistoryEntry, error) {
	db, err := sql.Open("sqlite3", "file:"+dbPath+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	rows, err := db.Query(`
        SELECT moz_places.url, moz_places.title, moz_historyvisits.visit_date
        FROM moz_places 
        JOIN moz_historyvisits ON moz_historyvisits.place_id = moz_places.id
        WHERE moz_historyvisits.visit_date >= ? AND moz_historyvisits.visit_date <= ?
        ORDER BY moz_historyvisits.visit_date DESC`,
		startTime.UnixMicro(), endTime.UnixMicro())
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %v", err)
	}
	defer rows.Close()

	var entries []HistoryEntry
	for rows.Next() {
		var url string
		var title sql.NullString // Use sql.NullString to handle NULL
		var timestamp int64
		if err := rows.Scan(&url, &title, &timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		// Use title.String if valid, empty string if NULL
		titleStr := ""
		if title.Valid {
			titleStr = title.String
		}
		entries = append(entries, HistoryEntry{
			URL:       url,
			Title:     titleStr,
			Timestamp: time.UnixMicro(timestamp),
		})
	}

	return entries, nil
}
