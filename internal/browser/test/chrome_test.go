package test // Separate package from browser

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/lotekdan/go-browser-history/internal/browser" // Import parent package
	_ "github.com/mattn/go-sqlite3"
)

func TestChromeGetHistoryPath(t *testing.T) {
	b := browser.NewChromeBrowser() // Fully qualified name
	path, err := b.GetHistoryPath()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if path == "" {
		t.Error("Expected non-empty path")
	}
}

func TestChromeExtractHistory(t *testing.T) {
	tempFile, err := os.CreateTemp("", "chrome_history_test_*.sqlite")
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
            ('https://example.com', 'Example', ?);
    `, browser.TimeToChromeTime(time.Now())) // Use exported function
	if err != nil {
		t.Fatalf("Failed to setup DB: %v", err)
	}

	b := browser.NewChromeBrowser() // Fully qualified name
	startTime := time.Now().AddDate(0, 0, -1)
	endTime := time.Now()

	entries, err := b.ExtractHistory(tempFile.Name(), startTime, endTime)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].URL != "https://example.com" {
		t.Errorf("Expected URL 'https://example.com', got %s", entries[0].URL)
	}
}
