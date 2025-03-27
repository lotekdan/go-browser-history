package browser

import (
	"database/sql"  // For SQL database interactions
	"fmt"           // For formatted output and error messages
	"os"            // For file operations
	"path/filepath" // For path manipulation
	"regexp"        // For parsing profiles.ini
	"runtime"       // For OS detection
	"time"          // For time handling

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// profilePathRegex extracts profile paths from profiles.ini.
var profilePathRegex = regexp.MustCompile(`Path=(.+)$`)

// firefoxHistoryQuery retrieves Firefox history entries.
const firefoxHistoryQuery = `
    SELECT moz_places.url, moz_places.title, moz_historyvisits.visit_date
    FROM moz_places 
    JOIN moz_historyvisits ON moz_historyvisits.place_id = moz_places.id
    WHERE moz_historyvisits.visit_date >= ? AND moz_historyvisits.visit_date <= ?
    ORDER BY moz_historyvisits.visit_date DESC`

// FirefoxBrowser implements the Browser interface for Mozilla Firefox.
type FirefoxBrowser struct{}

// NewFirefoxBrowser creates a new instance of FirefoxBrowser.
func NewFirefoxBrowser() Browser {
	return &FirefoxBrowser{}
}

// GetHistoryPath retrieves the path to Firefox's history database file.
func (fb *FirefoxBrowser) GetHistoryPath() (string, error) {
	baseDir, err := fb.getFirefoxProfileBaseDir()
	if err != nil {
		return "", err
	}
	return fb.findHistoryDBPath(baseDir)
}

// getFirefoxProfileBaseDir determines the OS-specific base directory for Firefox profiles.
func (fb *FirefoxBrowser) getFirefoxProfileBaseDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return os.Getenv("APPDATA") + "\\Mozilla\\Firefox\\Profiles", nil
	case "darwin":
		return os.Getenv("HOME") + "/Library/Application Support/Firefox/Profiles", nil
	case "linux":
		return os.Getenv("HOME") + "/.mozilla/firefox", nil
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// findHistoryDBPath locates the places.sqlite file within the profile base directory.
func (fb *FirefoxBrowser) findHistoryDBPath(firefoxProfileDir string) (string, error) {
	profilesIniPath := filepath.Join(filepath.Dir(firefoxProfileDir), "profiles.ini")
	if iniData, err := os.ReadFile(profilesIniPath); err == nil {
		matches := profilePathRegex.FindAllStringSubmatch(string(iniData), -1)
		for _, profileMatch := range matches {
			if len(profileMatch) > 1 {
				profilePath := filepath.Join(filepath.Dir(firefoxProfileDir), profileMatch[1])
				historyDBPath := filepath.Join(profilePath, "places.sqlite")
				if _, err := os.Stat(historyDBPath); err == nil {
					return historyDBPath, nil
				}
			}
		}
	}

	profileDirPaths, err := filepath.Glob(firefoxProfileDir + "/*")
	if err != nil || len(profileDirPaths) == 0 {
		return "", fmt.Errorf("could not find any Firefox profile in %s", firefoxProfileDir)
	}
	for _, profileDirPath := range profileDirPaths {
		historyDBPath := filepath.Join(profileDirPath, "places.sqlite")
		if _, err := os.Stat(historyDBPath); err == nil {
			return historyDBPath, nil
		}
	}
	return "", fmt.Errorf("no valid Firefox profile with places.sqlite found in %s", firefoxProfileDir)
}

// ExtractHistory extracts Firefox history entries within the given time range.
func (fb *FirefoxBrowser) ExtractHistory(historyDBPath string, startTime, endTime time.Time) ([]HistoryEntry, error) {
	db, err := sql.Open("sqlite3", "file:"+historyDBPath+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("failed to open Firefox history database at %s: %v", historyDBPath, err)
	}
	defer db.Close()

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

	return entries, nil
}
