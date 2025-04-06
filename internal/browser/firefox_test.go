package browser

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-ini/ini"
	_ "github.com/mattn/go-sqlite3"
)

func TestNewFirefoxBrowser(t *testing.T) {
	browser := NewFirefoxBrowser()
	if _, ok := browser.(*FirefoxBrowser); !ok {
		t.Error("NewFirefoxBrowser should return a *FirefoxBrowser")
	}
}

func TestFirefoxBrowser_GetHistoryPath(t *testing.T) {
	fb := &FirefoxBrowser{}
	tempDir := t.TempDir()

	profilesIni := filepath.Join(tempDir, "profiles.ini")
	cfg := ini.Empty()
	section, _ := cfg.NewSection("Profile0")
	section.NewKey("Path", "testprofile")
	section.NewKey("IsRelative", "1")
	cfg.SaveTo(profilesIni)

	profileDir := filepath.Join(tempDir, "testprofile")
	if err := os.Mkdir(profileDir, 0755); err != nil {
		t.Fatalf("Failed to create profile dir: %v", err)
	}

	placesPath := filepath.Join(profileDir, "places.sqlite")
	file, err := os.Create(placesPath)
	if err != nil {
		t.Fatalf("Failed to create places.sqlite: %v", err)
	}
	file.Close() // Explicitly close the file to avoid lock

	paths, err := fb.getPaths(tempDir)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(paths) != 1 {
		t.Errorf("Expected 1 path, got %d", len(paths))
	}
	if paths[0] != placesPath {
		t.Errorf("Expected path %s, got %s", placesPath, paths[0])
	}
}

func TestFirefoxBrowser_ExtractHistory(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_places.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
        CREATE TABLE moz_places (
            id INTEGER PRIMARY KEY, 
            url TEXT, 
            title TEXT, 
            visit_count INTEGER, 
            typed INTEGER
        );
        CREATE TABLE moz_historyvisits (
            id INTEGER PRIMARY KEY, 
            place_id INTEGER, 
            visit_date INTEGER, 
            visit_type INTEGER
        );
        INSERT INTO moz_places VALUES (1, 'https://test.com', 'Test', 2, 1);
        INSERT INTO moz_places VALUES (2, 'https://example.com', NULL, 1, 0);
        INSERT INTO moz_historyvisits VALUES (1, 1, ?, 1);
        INSERT INTO moz_historyvisits VALUES (2, 2, ?, 2);
    `, time.Now().Add(-1*time.Hour).UnixMicro(), time.Now().UnixMicro())
	if err != nil {
		t.Fatalf("Failed to setup test data: %v", err)
	}

	fb := &FirefoxBrowser{}
	startTime := time.Now().Add(-2 * time.Hour)
	endTime := time.Now().Add(time.Hour)

	entries, err := fb.ExtractHistory(dbPath, startTime, endTime, false)
	if err != nil {
		t.Fatalf("ExtractHistory failed: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}

	for _, entry := range entries {
		if entry.URL == "" {
			t.Error("Expected non-empty URL")
		}
		if entry.Timestamp.Before(startTime) || entry.Timestamp.After(endTime) {
			t.Errorf("Timestamp %v outside of range %v - %v", entry.Timestamp, startTime, endTime)
		}
	}
}
