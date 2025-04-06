package browser

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/go-ini/ini"
	_ "github.com/mattn/go-sqlite3"
)

const firefoxHistoryQuery = `
	SELECT
		moz_places.url, 
		moz_places.title, 
		moz_places.visit_count, 
		moz_places.typed, 
		CASE moz_historyvisits.visit_type
        	WHEN 1 THEN 'TRANSITION_LINK'
        	WHEN 2 THEN 'TRANSITION_TYPED'
        	WHEN 3 THEN 'TRANSITION_BOOKMARK'
			WHEN 4 THEN 'TRANSITION_EMBED'
			WHEN 5 THEN 'TRANSITION_REDIRECT_PERMANENT'
			WHEN 6 THEN 'TRANSITION_REDIRECT_TEMPORARY'
			WHEN 7 THEN 'TRANSITION_DOWNLOAD'
			WHEN 8 THEN 'TRANSITION_FRAMED_LINK'
			WHEN 9 THEN 'TRANSITION_RELOAD'
        ELSE 'UNKNOWN (' || moz_historyvisits.visit_type || ')'
		END AS visit_day_desc, 
		moz_historyvisits.visit_date 
		FROM moz_places
    	JOIN moz_historyvisits ON moz_historyvisits.place_id = moz_places.id
		WHERE moz_historyvisits.visit_date >= ? AND moz_historyvisits.visit_date <= ?
    ORDER BY moz_historyvisits.visit_date DESC`

type FirefoxBrowser struct{}

func NewFirefoxBrowser() Browser {
	return &FirefoxBrowser{}
}

// GetHistoryPath retrieves collection of paths to Firefox's history database file.
func (fb *FirefoxBrowser) GetHistoryPaths() ([]string, error) {
	baseDir, err := fb.getFirefoxProfileBaseDir()
	if err != nil {
		return nil, err
	}
	return fb.getPaths(baseDir)
}

func (fb *FirefoxBrowser) getFirefoxProfileBaseDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return os.Getenv("APPDATA") + "\\Mozilla\\Firefox\\", nil
	case "darwin":
		return os.Getenv("HOME") + "/Library/Application Support/Firefox/", nil
	case "linux":
		return os.Getenv("HOME") + "/.mozilla/firefox/", nil
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// GetBrowserProfilePaths gets a collection of browser profile history paths.
func (fb *FirefoxBrowser) getPaths(dir string) ([]string, error) {
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

// ExtractHistory gets records from the defined history db and date range.
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
		var pageURL, pageVisitType string
		var pageVisitCount, pageTyped int
		var pageTitle sql.NullString
		var visitTimestamp int64
		//var ProfileName string
		if err := rows.Scan(
			&pageURL,
			&pageTitle,
			&pageVisitCount,
			&pageTyped,
			&pageVisitType,
			&visitTimestamp); err != nil {
			return nil, fmt.Errorf("failed to scan Firefox history row from %s: %v", historyDBPath, err)
		}
		title := ""
		if pageTitle.Valid {
			title = pageTitle.String
		}
		entries = append(entries, HistoryEntry{
			URL:        pageURL,
			Title:      title,
			VisitCount: pageVisitCount,
			Typed:      pageTyped,
			VisitType:  pageVisitType,
			Timestamp:  time.UnixMicro(visitTimestamp),
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
