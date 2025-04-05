package browser

import (
	"database/sql"
	"os"
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
	tempDir := t.TempDir()

	var baseDir string
	switch runtime.GOOS {
	case "windows":
		baseDir = filepath.Join(tempDir, "Google", "Chrome", "User Data")
		os.Setenv("LOCALAPPDATA", tempDir)
	case "darwin":
		baseDir = filepath.Join(tempDir, "Library", "Application Support", "Google", "Chrome")
		os.Setenv("HOME", tempDir)
	case "linux":
		baseDir = filepath.Join(tempDir, ".config", "google-chrome")
		os.Setenv("HOME", tempDir)
	default:
		t.Skipf("Skipping test on unsupported OS: %s", runtime.GOOS)
	}

	defaultDir := filepath.Join(baseDir, "Default")
	if err := os.MkdirAll(defaultDir, 0755); err != nil {
		t.Fatalf("Failed to create default dir: %v", err)
	}
	historyPath := filepath.Join(defaultDir, "History")
	file, err := os.Create(historyPath)
	if err != nil {
		t.Fatalf("Failed to create History file: %v", err)
	}
	file.Close()

	paths, err := cb.GetHistoryPath()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(paths) == 0 {
		t.Error("Expected at least one path, got none")
	}
	if paths[0] != historyPath {
		t.Errorf("Expected path %s, got %s", historyPath, paths[0])
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
            id INTEGER PRIMARY KEY,
            url TEXT,
            title TEXT,
            visit_count INTEGER,
            typed_count INTEGER,
            last_visit_time INTEGER
        );
        CREATE TABLE visits (
            id INTEGER PRIMARY KEY,
            url INTEGER,
            visit_time INTEGER,
            transition INTEGER
        );
        INSERT INTO urls VALUES 
            (1, 'https://test.com', 'Test', 2, 1, ?),
            (2, 'https://example.com', 'Example', 1, 0, ?);
        INSERT INTO visits VALUES 
            (1, 1, ?, 1),
            (2, 2, ?, 8);
    `, TimeToChromeTime(time.Now().Add(-1*time.Hour)), TimeToChromeTime(time.Now()),
		TimeToChromeTime(time.Now().Add(-1*time.Hour)), TimeToChromeTime(time.Now()))
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
