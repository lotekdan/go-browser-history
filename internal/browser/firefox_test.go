package browser_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/lotekdan/go-browser-history/internal/browser"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestNewFirefoxBrowser(t *testing.T) {
	b := browser.NewFirefoxBrowser()
	assert.NotNil(t, b)
	_, ok := b.(*browser.FirefoxBrowser)
	assert.True(t, ok, "NewFirefoxBrowser should return a *FirefoxBrowser")
}

func TestFirefoxGetHistoryPath_Success(t *testing.T) {
	// Arrange
	fb := &browser.FirefoxBrowser{}

	// Create a temp dir with the correct Firefox structure
	tmpDir, err := os.MkdirTemp("", "firefox_test_*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set HOME/APPDATA to tmpDir and create .mozilla/firefox (or equivalent)
	var baseDir string
	switch runtime.GOOS {
	case "windows":
		t.Setenv("APPDATA", tmpDir)
		baseDir = filepath.Join(tmpDir, "Mozilla", "Firefox", "Profiles")
	case "darwin":
		t.Setenv("HOME", tmpDir)
		baseDir = filepath.Join(tmpDir, "Library", "Application Support", "Firefox", "Profiles")
	case "linux":
		t.Setenv("HOME", tmpDir)
		baseDir = filepath.Join(tmpDir, ".mozilla", "firefox")
	default:
		t.Skipf("unsupported OS: %s", runtime.GOOS)
	}
	assert.NoError(t, os.MkdirAll(baseDir, 0755))

	// Set up profiles.ini and places.sqlite
	profilesIni := filepath.Join(filepath.Dir(baseDir), "profiles.ini")
	profileDir := filepath.Join(baseDir, "testprofile.default")
	historyDB := filepath.Join(profileDir, "places.sqlite")
	assert.NoError(t, os.Mkdir(profileDir, 0755))
	assert.NoError(t, os.WriteFile(profilesIni, []byte("[Profile0]\nPath=testprofile.default"), 0644))
	assert.NoError(t, os.WriteFile(historyDB, []byte("mock db"), 0644))

	// Act
	path, err := fb.GetHistoryPath()

	// Assert
	assert.NoError(t, err)
	expectedPath := filepath.Join(baseDir, "testprofile.default", "places.sqlite")
	assert.Equal(t, expectedPath, path)
}

func TestFirefoxGetHistoryPath_NoProfiles(t *testing.T) {
	// Arrange
	fb := &browser.FirefoxBrowser{}
	tmpDir, err := os.MkdirTemp("", "firefox_test_*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Set HOME/APPDATA to an empty temp dir
	switch runtime.GOOS {
	case "windows":
		t.Setenv("APPDATA", tmpDir)
	case "darwin", "linux":
		t.Setenv("HOME", tmpDir)
	}

	// Act
	path, err := fb.GetHistoryPath()

	// Assert
	assert.Error(t, err)
	assert.Empty(t, path)
	assert.Contains(t, err.Error(), "could not find any Firefox profile")
}

func TestFirefoxExtractHistory_Success(t *testing.T) {
	// Arrange: Create a temporary SQLite database file
	tmpFile, err := os.CreateTemp("", "firefox_test_*.sqlite")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	db, err := sql.Open("sqlite3", tmpFile.Name())
	assert.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE moz_places (id INTEGER PRIMARY KEY, url TEXT, title TEXT);
		CREATE TABLE moz_historyvisits (id INTEGER PRIMARY KEY, place_id INTEGER, visit_date INTEGER);
		INSERT INTO moz_places (id, url, title) VALUES (1, 'http://example.com', 'Example');
		INSERT INTO moz_historyvisits (place_id, visit_date) VALUES (1, ?);
	`, time.Now().Add(-1*time.Hour).UnixMicro())
	assert.NoError(t, err)

	fb := &browser.FirefoxBrowser{}
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()

	// Act
	entries, err := fb.ExtractHistory(tmpFile.Name(), start, end, false)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "http://example.com", entries[0].URL)
	assert.Equal(t, "Example", entries[0].Title)
	assert.True(t, entries[0].Timestamp.After(start) && entries[0].Timestamp.Before(end))
}

func TestFirefoxExtractHistory_InvalidDB(t *testing.T) {
	// Arrange
	fb := &browser.FirefoxBrowser{}
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()

	// Act
	entries, err := fb.ExtractHistory("nonexistent.sqlite", start, end, true)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, entries)
	assert.Contains(t, err.Error(), "unable to open database file")
}
