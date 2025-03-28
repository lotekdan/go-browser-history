package browser_test

import (
	"testing"
	"time"

	"github.com/lotekdan/go-browser-history/internal/browser"
	"github.com/stretchr/testify/assert"
)

// mockBrowser implements the Browser interface for testing.
type mockBrowser struct{}

func (m *mockBrowser) GetHistoryPath() (string, error) {
	return "testdata/mock_history.db", nil
}

func (m *mockBrowser) ExtractHistory(dbPath string, startTime, endTime time.Time, verbose bool) ([]browser.HistoryEntry, error) {
	return []browser.HistoryEntry{
		{URL: "http://example.com", Title: "Example", Timestamp: time.Now()},
	}, nil
}

func TestGetHistoryPath_Success(t *testing.T) {
	// Arrange
	b := &mockBrowser{}

	// Act
	path, err := b.GetHistoryPath()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "testdata/mock_history.db", path)
}

func TestExtractHistory_Success(t *testing.T) {
	// Arrange
	b := &mockBrowser{}
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()

	// Act
	entries, err := b.ExtractHistory("testdata/mock_history.db", start, end, false)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "http://example.com", entries[0].URL)
	assert.Equal(t, "Example", entries[0].Title)
	assert.False(t, entries[0].Timestamp.IsZero())
}

func TestExtractHistory_EmptyResult(t *testing.T) {
	// Arrange
	b := &mockBrowserEmpty{}
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()

	// Act
	entries, err := b.ExtractHistory("testdata/empty.db", start, end, true)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, entries)
}

// mockBrowserEmpty returns an empty history for edge case testing.
type mockBrowserEmpty struct{}

func (m *mockBrowserEmpty) GetHistoryPath() (string, error) {
	return "testdata/empty.db", nil
}

func (m *mockBrowserEmpty) ExtractHistory(dbPath string, startTime, endTime time.Time, verbose bool) ([]browser.HistoryEntry, error) {
	return []browser.HistoryEntry{}, nil
}
