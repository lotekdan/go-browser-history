package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lotekdan/go-browser-history/internal/history"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBrowser for testing
type MockBrowser struct {
	mock.Mock
}

func (m *MockBrowser) GetHistoryPaths() ([]history.HistoryPathEntry, error) {
	args := m.Called()
	return args.Get(0).([]history.HistoryPathEntry), args.Error(1)
}

func (m *MockBrowser) ExtractHistory(dbPath, profile string, startTime, endTime time.Time, debug bool) ([]history.HistoryEntry, error) {
	args := m.Called(dbPath, profile, startTime, endTime, debug)
	return args.Get(0).([]history.HistoryEntry), args.Error(1)
}

func TestGetBrowserHistory(t *testing.T) {
	t.Run("successful_history_retrieval", func(t *testing.T) {
		// Setup mock browser
		mockBrowser := new(MockBrowser)

		// Create a real temp file to avoid filesystem errors
		tempDir := os.TempDir()
		tempPath := filepath.Join(tempDir, "test-history.db")
		err := os.WriteFile(tempPath, []byte("mock data"), 0644)
		assert.NoError(t, err)
		defer os.Remove(tempPath) // Cleanup

		// Mock expectations
		mockBrowser.On("GetHistoryPaths").Return([]history.HistoryPathEntry{
			{Path: tempPath, ProfileName: ""},
		}, nil)
		expectedEntries := []history.HistoryEntry{
			{URL: "http://example.com"},
		}
		mockBrowser.On("ExtractHistory", mock.Anything, "", mock.Anything, mock.Anything, false).Return(expectedEntries, nil)

		// Execute
		history, err := GetBrowserHistory(mockBrowser, time.Time{}, time.Time{}, false)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedEntries, history, "history entries mismatch")
		mockBrowser.AssertExpectations(t)
	})

	t.Run("no_files_returns_empty_history", func(t *testing.T) {
		// Setup mock browser
		mockBrowser := new(MockBrowser)

		// Mock GetHistoryPaths to return no paths
		mockBrowser.On("GetHistoryPaths").Return([]history.HistoryPathEntry{}, nil)

		// Execute
		history, err := GetBrowserHistory(mockBrowser, time.Time{}, time.Time{}, false)

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, history, "expected empty history when no files")
		mockBrowser.AssertExpectations(t)
	})

	t.Run("error_in_extraction", func(t *testing.T) {
		// Setup mock browser
		mockBrowser := new(MockBrowser)

		// Create a real temp file
		tempDir := os.TempDir()
		tempPath := filepath.Join(tempDir, "test-history-err.db")
		err := os.WriteFile(tempPath, []byte("mock data"), 0644)
		assert.NoError(t, err)
		defer os.Remove(tempPath)

		// Mock expectations
		mockBrowser.On("GetHistoryPaths").Return([]history.HistoryPathEntry{
			{Path: tempPath, ProfileName: ""},
		}, nil)
		// Use typed nil for []history.HistoryEntry to avoid panic
		mockBrowser.On("ExtractHistory", mock.Anything, "", mock.Anything, mock.Anything, false).Return(([]history.HistoryEntry)(nil), fmt.Errorf("extraction error"))

		// Execute
		history, err := GetBrowserHistory(mockBrowser, time.Time{}, time.Time{}, false)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, history)
		assert.Equal(t, "extraction error", err.Error(), "error message mismatch")
		mockBrowser.AssertExpectations(t)
	})
}
