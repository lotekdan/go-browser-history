package test

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/lotekdan/go-browser-history/internal/browser"
	_ "github.com/mattn/go-sqlite3"
)

func TestEdgeGetHistoryPath(t *testing.T) {
	b := browser.NewEdgeBrowser() // Updated to use imported package
	path, err := b.GetHistoryPath()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if path == "" {
		t.Error("Expected non-empty path")
	}
}

func TestEdgeExtractHistory(t *testing.T) {
	tempFile, err := os.CreateTemp("", "edge_history_test_*.sqlite")
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
        CREATE TABLE urls (
            url TEXT,
            title TEXT,
            last_visit_time INTEGER
        );
        INSERT INTO urls (url, title, last_visit_time) VALUES
            ('https://example.org', 'Example Org', ?);
    `, browser.TimeToChromeTime(time.Now())) // Edge uses Chrome’s format
	if err != nil {
		t.Fatalf("Failed to setup DB: %v", err)
	}

	b := browser.NewEdgeBrowser() // Updated to use imported package
	startTime := time.Now().AddDate(0, 0, -1)
	endTime := time.Now()

	entries, err := b.ExtractHistory(tempFile.Name(), startTime, endTime)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].URL != "https://example.org" {
		t.Errorf("Expected URL 'https://example.org', got %s", entries[0].URL)
	}
}
