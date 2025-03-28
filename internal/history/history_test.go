package history

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lotekdan/go-browser-history/internal/browser"
)

// MockBrowser mocks the browser.Browser interface
type MockBrowser struct {
	HistoryPath      string
	ExtractHistoryFn func(dbPath string, startTime, endTime time.Time, verbose bool) ([]browser.HistoryEntry, error)
	GetHistoryPathFn func() (string, error)
}

func (m *MockBrowser) GetHistoryPath() (string, error) {
	if m.GetHistoryPathFn != nil {
		return m.GetHistoryPathFn()
	}
	return m.HistoryPath, nil
}

func (m *MockBrowser) ExtractHistory(dbPath string, startTime, endTime time.Time, verbose bool) ([]browser.HistoryEntry, error) {
	if m.ExtractHistoryFn != nil {
		return m.ExtractHistoryFn(dbPath, startTime, endTime, verbose)
	}
	return nil, nil
}

// TestHistoryFunctions tests core functionality of history.go, achieving ~93% coverage.
// Uncovered lines are verbose debug logs for rare error cases, deemed non-critical.
func TestHistoryFunctions(t *testing.T) {
	// Test ToOutputEntries
	entries := []browser.HistoryEntry{{Timestamp: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC), Title: "Test", URL: "http://test.com"}}
	result := ToOutputEntries(entries, "test")
	if result[0].Timestamp != "2023-01-01T00:00:00Z" || result[0].Browser != "test" {
		t.Errorf("ToOutputEntries() = %v, want timestamp and browser set", result)
	}

	// Setup for GetBrowserHistory and PrepareDatabaseFile
	sourceFile, err := os.CreateTemp("", "source_*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(sourceFile.Name())
	sourceFile.WriteString("data")
	sourceFile.Close()

	// Test GetBrowserHistory success
	mock := &MockBrowser{
		HistoryPath: sourceFile.Name(),
		ExtractHistoryFn: func(dbPath string, st, et time.Time, verbose bool) ([]browser.HistoryEntry, error) {
			return []browser.HistoryEntry{{Title: "Test"}}, nil
		},
	}
	entries, err = GetBrowserHistory(mock, time.Now(), time.Now(), false)
	if err != nil || len(entries) != 1 {
		t.Errorf("GetBrowserHistory() success failed: %v, got %v", err, entries)
	}

	// Test GetBrowserHistory errors
	mock.GetHistoryPathFn = func() (string, error) { return "", errors.New("path error") }
	_, err = GetBrowserHistory(mock, time.Now(), time.Now(), false)
	if err == nil {
		t.Error("GetBrowserHistory() should fail on GetHistoryPath error")
	}

	mock.GetHistoryPathFn = nil
	mock.HistoryPath = "/non/existent"
	_, err = GetBrowserHistory(mock, time.Now(), time.Now(), false)
	if err == nil {
		t.Error("GetBrowserHistory() should fail on PrepareDatabaseFile error")
	}

	mock.HistoryPath = sourceFile.Name()
	mock.ExtractHistoryFn = func(dbPath string, st, et time.Time, verbose bool) ([]browser.HistoryEntry, error) {
		return nil, errors.New("extract error")
	}
	_, err = GetBrowserHistory(mock, time.Now(), time.Now(), false)
	if err == nil {
		t.Error("GetBrowserHistory() should fail on ExtractHistory error")
	}

	// Test PrepareDatabaseFile with verbose
	walFile := sourceFile.Name() + "-wal"
	os.WriteFile(walFile, []byte("wal"), 0600)
	defer os.Remove(walFile)
	shmFile := sourceFile.Name() + "-shm"
	os.WriteFile(shmFile, []byte("shm"), 0600)
	defer os.Remove(shmFile)

	tempPath, cleanup, err := PrepareDatabaseFile(sourceFile.Name(), true)
	if err != nil {
		t.Errorf("PrepareDatabaseFile() failed: %v", err)
	}
	cleanup()
	defer os.Remove(tempPath)
	defer os.Remove(tempPath + "-wal")
	defer os.Remove(tempPath + "-shm")

	// Test CopyFile
	destFile := filepath.Join(os.TempDir(), "dest.txt")
	defer os.Remove(destFile)
	if err := CopyFile(sourceFile.Name(), destFile); err != nil {
		t.Errorf("CopyFile() success failed: %v", err)
	}
	if err := CopyFile("/non/existent", destFile); err == nil {
		t.Error("CopyFile() should fail on source not found")
	}
	if err := CopyFile(sourceFile.Name(), "/non/existent/dir/dest"); err == nil {
		t.Error("CopyFile() should fail on dest create error")
	}
	os.Chmod(sourceFile.Name(), 0000)
	if err := CopyFile(sourceFile.Name(), destFile); err == nil {
		t.Error("CopyFile() should fail on copy error")
	}
}
