package browser

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/go-ini/ini"
	_ "github.com/mattn/go-sqlite3"
)

var profilePathRegex = regexp.MustCompile(`Path=(.+)$`)

const firefoxHistoryQuery = `
    SELECT moz_places.url, moz_places.title, moz_historyvisits.visit_date
    FROM moz_places 
    JOIN moz_historyvisits ON moz_historyvisits.place_id = moz_places.id
    WHERE moz_historyvisits.visit_date >= ? AND moz_historyvisits.visit_date <= ?
    ORDER BY moz_historyvisits.visit_date DESC`

type FirefoxBrowser struct{}

func NewFirefoxBrowser() Browser {
	return &FirefoxBrowser{}
}

func (fb *FirefoxBrowser) GetHistoryPath() ([]string, error) {
	baseDir, err := fb.getFirefoxProfileBaseDir()
	if err != nil {
		return nil, err
	}
	return fb.GetHistoryPaths(baseDir)
}

func (fb *FirefoxBrowser) getFirefoxProfileBaseDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return os.Getenv("APPDATA") + "\\Mozilla\\Firefox\\Profiles\\", nil
	case "darwin":
		return os.Getenv("HOME") + "/Library/Application Support/Firefox/Profiles/", nil
	case "linux":
		return os.Getenv("HOME") + "/.mozilla/firefox/", nil
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// GetBrowserProfilePaths gets a collection of browser profile history paths.
func (fb *FirefoxBrowser) GetHistoryPaths(dir string) ([]string, error) {
	profileIniFile := filepath.Join(dir, "profiles.ini")
	cfg, err := ini.Load(profileIniFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load profiles.ini: %w", err)
	}

	var profiles []string

	for _, section := range cfg.Sections() {
		if !strings.HasPrefix(section.Name(), "Profile") {
			continue
		}

		path := section.Key("Path").String()
		isRelative := section.Key("IsRelative").MustInt(0) == 1

		if path == "" {
			continue
		}

		if isRelative {
			path = filepath.Join(dir, path, "places.sqlite")
		}

		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			continue
		}
		profiles = append(profiles, path)
	}
	return profiles, nil
}

func (fb *FirefoxBrowser) ExtractHistory(historyDBPath string, startTime, endTime time.Time, verbose bool) ([]HistoryEntry, error) {
	db, err := sql.Open("sqlite3", "file:"+historyDBPath+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("failed to open Firefox history database at %s: %v", historyDBPath, err)
	}
	defer db.Close()

	if verbose {
		fmt.Fprintf(os.Stderr, "Debug: Querying Firefox history from %s, start: %v, end: %v\n", historyDBPath, startTime, endTime)
		fmt.Fprintf(os.Stderr, "Debug: Query params: start=%d, end=%d\n", startTime.UnixMicro(), endTime.UnixMicro())
	}
	rows, err := db.Query(firefoxHistoryQuery, startTime.UnixMicro(), endTime.UnixMicro())
	if err != nil {
		return nil, fmt.Errorf("failed to query Firefox history from %s: %v", historyDBPath, err)
	}
	defer rows.Close()

	var entries []HistoryEntry
	for rows.Next() {
		var pageURL string
		var pageTitle sql.NullString
		var visitTimestamp int64
		if err := rows.Scan(&pageURL, &pageTitle, &visitTimestamp); err != nil {
			return nil, fmt.Errorf("failed to scan Firefox history row from %s: %v", historyDBPath, err)
		}
		title := ""
		if pageTitle.Valid {
			title = pageTitle.String
		}
		entries = append(entries, HistoryEntry{
			URL:       pageURL,
			Title:     title,
			Timestamp: time.UnixMicro(visitTimestamp),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating Firefox history rows from %s: %v", historyDBPath, err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Debug: Retrieved %d entries from Firefox\n", len(entries))
		if len(entries) == 0 {
			fmt.Fprintf(os.Stderr, "Debug: Warning: No entries found. Database may be empty, history not flushed, or time range incorrect.\n")
		}
	}
	return entries, nil
}
