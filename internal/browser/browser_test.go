package browser

import (
	"testing"
	"time"
)

func TestHistoryEntry(t *testing.T) {
	entry := HistoryEntry{
		URL:       "https://example.com",
		Title:     "Example",
		Timestamp: time.Now(),
	}

	if entry.URL != "https://example.com" {
		t.Errorf("Expected URL to be 'https://example.com', got %s", entry.URL)
	}
	if entry.Title != "Example" {
		t.Errorf("Expected Title to be 'Example', got %s", entry.Title)
	}
	if entry.Timestamp.IsZero() {
		t.Error("Expected Timestamp to be non-zero")
	}
}
