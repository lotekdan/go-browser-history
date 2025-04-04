package browser

import (
	"database/sql"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestNewChromeBrowser(t *testing.T) {
	browser := NewChromeBrowser()
	if _, ok := browser.(*ChromeBrowser); !ok {
		t.Error("NewChromeBrowser should return a *ChromeBrowser")
	}
}

func TestChromeBrowser_GetHistoryPath(t *testing.T) {
	cb := &ChromeBrowser{}

	// Since we can't modify runtime.GOOS, we'll test the current OS only
	// For full OS testing, consider using build tags or a mock
	paths, err := cb.GetHistoryPath()
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		if err != nil {
			t.Errorf("Unexpected error on supported OS %s: %v", runtime.GOOS, err)
		}
		if len(paths) == 0 {
			t.Error("Expected at least one path, got none")
		}
	} else {
		if err == nil {
			t.Errorf("Expected error on unsupported OS %s, got nil", runtime.GOOS)
		}
	}
}

func TestChromeBrowser_ExtractHistory(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_history.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
        CREATE TABLE urls (
            url TEXT,
            title TEXT,
            last_visit_time INTEGER
        );
        INSERT INTO urls VALUES 
            ('https://test.com', 'Test', ?),
            ('https://example.com', 'Example', ?);
    `, TimeToChromeTime(time.Now().Add(-1*time.Hour)), TimeToChromeTime(time.Now()))
	if err != nil {
		t.Fatalf("Failed to setup test data: %v", err)
	}

	cb := &ChromeBrowser{}
	startTime := time.Now().Add(-2 * time.Hour)
	endTime := time.Now().Add(time.Hour)

	entries, err := cb.ExtractHistory(dbPath, startTime, endTime, false)
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

func TestTimeConversions(t *testing.T) {
	now := time.Now()
	chromeTime := TimeToChromeTime(now)
	converted := ChromeTimeToTime(chromeTime)

	if !now.Truncate(time.Microsecond).Equal(converted) {
		t.Errorf("Time conversion failed: %v != %v", now, converted)
	}
}
