package service_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lotekdan/go-browser-history/internal/browser"
	"github.com/lotekdan/go-browser-history/internal/config"
	"github.com/lotekdan/go-browser-history/internal/service"
)

// MockBrowser implements browser.Browser for testing
type MockBrowser struct {
	historyPath []string
	entries     []browser.HistoryEntry
	pathErr     error
	extractErr  error
}

func (m *MockBrowser) GetHistoryPath() ([]string, error) {
	return m.historyPath, m.pathErr
}

func (m *MockBrowser) ExtractHistory(dbPath string, startTime, endTime time.Time, debug bool) ([]browser.HistoryEntry, error) {
	return m.entries, m.extractErr
}

func TestGetHistory(t *testing.T) {
	// Create a temp file to simulate a valid DB path
	tempDir := t.TempDir()
	tempDBPath := filepath.Join(tempDir, "mock_history.db")
	if err := os.WriteFile(tempDBPath, []byte("mock data"), 0644); err != nil {
		t.Fatalf("Failed to create temp DB file: %v", err)
	}
	tempPath := []string{tempDBPath}

	tests := []struct {
		name             string
		selectedBrowsers []string
		mockBrowser      *MockBrowser
		wantEntries      int
		wantErr          bool
	}{
		{
			name:             "valid browser with no entries",
			selectedBrowsers: []string{"mock"},
			mockBrowser: &MockBrowser{
				historyPath: tempPath,
				entries:     []browser.HistoryEntry{},
				pathErr:     nil,
				extractErr:  nil,
			},
			wantEntries: 0,
			wantErr:     false,
		},
		{
			name:             "invalid browser",
			selectedBrowsers: []string{"invalid"},
			mockBrowser:      nil,
			wantEntries:      0,
			wantErr:          true,
		},
		{
			name:             "valid browser with entries",
			selectedBrowsers: []string{"mock"},
			mockBrowser: &MockBrowser{
				historyPath: tempPath,
				entries: []browser.HistoryEntry{
					{URL: "http://example.com", Title: "Example", Timestamp: time.Now()},
				},
				pathErr:    nil,
				extractErr: nil,
			},
			wantEntries: 1,
			wantErr:     false,
		},
		{
			name:             "error from GetHistoryPath",
			selectedBrowsers: []string{"mock"},
			mockBrowser: &MockBrowser{
				historyPath: tempPath,
				entries:     []browser.HistoryEntry{{URL: "http://example.com"}},
				pathErr:     os.ErrNotExist,
				extractErr:  nil,
			},
			wantEntries: 0,
			wantErr:     false, // fetchHistory continues on error, so no error returned
		},
		{
			name:             "error from ExtractHistory",
			selectedBrowsers: []string{"mock"},
			mockBrowser: &MockBrowser{
				historyPath: tempPath,
				entries:     nil,
				pathErr:     nil,
				extractErr:  os.ErrInvalid,
			},
			wantEntries: 0,
			wantErr:     false, // fetchHistory continues on error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup config
			endTime, _ := time.Parse("2006-01-02 15:04:05", "2025-03-28 23:59:59")
			cfg := &config.Config{
				EndTime:     endTime,
				HistoryDays: 28,
				Debug:       true,
			}

			// Setup service with mock
			browserMap := make(map[string]browser.Browser)
			if tt.mockBrowser != nil {
				browserMap["mock"] = tt.mockBrowser
			}
			s := service.NewHistoryService(browserMap)

			// Run GetHistory
			entries, err := s.GetHistory(cfg, tt.selectedBrowsers)

			// Debug output
			t.Logf("Returned entries: %v", entries)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("GetHistory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check entries
			if len(entries) != tt.wantEntries {
				t.Errorf("GetHistory() returned %d entries, want %d", len(entries), tt.wantEntries)
			}

			// Verify StartTime adjustment (only when no error)
			wantStartTime := endTime.AddDate(0, 0, -28)
			if !tt.wantErr && !cfg.StartTime.Equal(wantStartTime) {
				t.Errorf("StartTime = %v, want %v", cfg.StartTime, wantStartTime)
			}
			if tt.wantErr && !cfg.StartTime.IsZero() {
				t.Errorf("StartTime = %v, want zero value", cfg.StartTime)
			}
		})
	}
}
