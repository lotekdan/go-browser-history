package test

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/lotekdan/go-browser-history/internal/browser"
	"github.com/lotekdan/go-browser-history/internal/history"
	_ "github.com/mattn/go-sqlite3" // For SQLite in-memory DB
)

type mockBrowser struct {
	dbPath string // Add field to customize path
}

func (m *mockBrowser) GetHistoryPath() (string, error) {
	return m.dbPath, nil // Use configurable path
}

func (m *mockBrowser) ExtractHistory(dbPath string, start, end time.Time) ([]browser.HistoryEntry, error) {
	return []browser.HistoryEntry{{URL: "http://example.com"}}, nil
}

func TestGetBrowserHistory(t *testing.T) {
	// Create a temporary SQLite file
	tempFile, err := os.CreateTemp("", "mock_history_test_*.sqlite")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Set up a minimal SQLite database (optional, depending on ExtractHistory needs)
	db, err := sql.Open("sqlite3", tempFile.Name())
	if err != nil {
		t.Fatalf("failed to open DB: %v", err)
	}
	defer db.Close()

	// Minimal schema to satisfy potential ExtractHistory calls (adjust as needed)
	_, err = db.Exec("CREATE TABLE urls (url TEXT)") // Example schema
	if err != nil {
		t.Fatalf("failed to setup DB: %v", err)
	}

	mock := &mockBrowser{dbPath: tempFile.Name()} // Pass temp file to mock
	entries, err := history.GetBrowserHistory(mock, time.Now().Add(-24*time.Hour), time.Now())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(entries) != 1 || entries[0].URL != "http://example.com" {
		t.Errorf("unexpected entries: %v", entries)
	}
}

func TestPrepareDatabaseFile_Unlocked(t *testing.T) {
	tempFile, err := os.CreateTemp("", "testdb")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())
	path, cleanup, err := history.PrepareDatabaseFile(tempFile.Name())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer cleanup()
	if path != tempFile.Name() {
		t.Errorf("expected path %s, got %s", tempFile.Name(), path)
	}
}

func TestPrepareDatabaseFile_Locked(t *testing.T) {
	path, cleanup, err := history.PrepareDatabaseFile("nonexistent.db")
	if err == nil {
		t.Fatal("expected error for locked/nonexistent file")
	}
	cleanup()
	if path != "" {
		t.Errorf("expected empty path on error, got %s", path)
	}
}

func TestCopyFile(t *testing.T) {
	src, err := os.CreateTemp("", "src")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(src.Name())
	_, err = src.WriteString("test data")
	if err != nil {
		t.Fatal(err)
	}
	dst := src.Name() + ".copy"
	err = history.CopyFile(src.Name(), dst)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer os.Remove(dst)
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "test data" {
		t.Errorf("expected 'test data', got '%s'", data)
	}
}
