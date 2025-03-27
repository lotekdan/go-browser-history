package test

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/lotekdan/go-browser-history/internal/browser"
	_ "github.com/mattn/go-sqlite3"
)

func TestFirefoxGetHistoryPath(t *testing.T) {
	b := browser.NewFirefoxBrowser() // Updated to use imported package
	path, err := b.GetHistoryPath()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if path == "" {
		t.Error("Expected non-empty path")
	}
}

func TestFirefoxExtractHistory(t *testing.T) {
	tempFile, err := os.CreateTemp("", "firefox_history_test_*.sqlite")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	db, err := sql.Open("sqlite3", tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
        CREATE TABLE moz_places (
            id INTEGER PRIMARY KEY,
            url TEXT,
            title TEXT
        );
        CREATE TABLE moz_historyvisits (
            place_id INTEGER,
            visit_date INTEGER
        );
        INSERT INTO moz_places (id, url, title) VALUES (1, 'https://firefox.com', 'Firefox');
        INSERT INTO moz_historyvisits (place_id, visit_date) VALUES (1, ?);
    `, time.Now().UnixMicro()) // Firefox uses Unix microseconds
	if err != nil {
		t.Fatalf("Failed to setup DB: %v", err)
	}

	b := browser.NewFirefoxBrowser() // Updated to use imported package
	startTime := time.Now().AddDate(0, 0, -1)
	endTime := time.Now()

	entries, err := b.ExtractHistory(tempFile.Name(), startTime, endTime)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].URL != "https://firefox.com" {
		t.Errorf("Expected URL 'https://firefox.com', got %s", entries[0].URL)
	}
}
