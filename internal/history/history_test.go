package history

import (
	"os"
	"testing"
	"time"

	"github.com/lotekdan/go-browser-history/internal/browser"
)

type mockBrowser struct {
	path string
}

func (m *mockBrowser) GetHistoryPath() (string, error) {
	return m.path, nil
}

func (m *mockBrowser) ExtractHistory(dbPath string, startTime, endTime time.Time) ([]browser.HistoryEntry, error) {
	return []browser.HistoryEntry{
		{URL: "https://example.com", Title: "Example", Timestamp: time.Now()},
	}, nil
}

func TestGetBrowserHistory(t *testing.T) {
	tempFile, err := os.CreateTemp("", "mock_history_test_*.sqlite")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	b := &mockBrowser{path: tempFile.Name()}
	startTime := time.Now().AddDate(0, 0, -1)
	endTime := time.Now()

	entries, err := GetBrowserHistory(b, startTime, endTime)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
}

func TestPrepareDatabaseFile(t *testing.T) {
	// Test with accessible file
	tempFile, err := os.CreateTemp("", "test_db_*.sqlite")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	path, cleanup, err := prepareDatabaseFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if path != tempFile.Name() {
		t.Errorf("Expected original path %s, got %s", tempFile.Name(), path)
	}
	cleanup()

	// Test with "locked" file (make it read-only to simulate lock)
	lockedFile, err := os.CreateTemp("", "locked_db_*.sqlite")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(lockedFile.Name())

	// Make the file read-only
	err = os.Chmod(lockedFile.Name(), 0444) // Read-only
	if err != nil {
		t.Fatalf("Failed to set read-only: %v", err)
	}
	defer os.Chmod(lockedFile.Name(), 0666) // Restore permissions for cleanup

	path, cleanup, err = prepareDatabaseFile(lockedFile.Name())
	if err != nil {
		t.Fatalf("Expected no error with copy, got %v", err)
	}
	if path == lockedFile.Name() {
		t.Errorf("Expected copy path, got original %s", path)
	}
	cleanup()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("Expected temp file %s to be deleted, got %v", path, err)
	}
}

func TestCopyFile(t *testing.T) {
	src, err := os.CreateTemp("", "src_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}
	defer os.Remove(src.Name())

	_, err = src.Write([]byte("test data"))
	if err != nil {
		t.Fatalf("Failed to write to source: %v", err)
	}

	dst, err := os.CreateTemp("", "dst_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create dest file: %v", err)
	}
	defer os.Remove(dst.Name())

	err = copyFile(src.Name(), dst.Name())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	content, err := os.ReadFile(dst.Name())
	if err != nil {
		t.Fatalf("Failed to read dest file: %v", err)
	}
	if string(content) != "test data" {
		t.Errorf("Expected 'test data', got %s", string(content))
	}
}
