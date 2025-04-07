package service

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/lotekdan/go-browser-history/internal/browser"
	"github.com/lotekdan/go-browser-history/internal/config"
	"github.com/lotekdan/go-browser-history/internal/history"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBrowser is a minimal mock for NewHistoryService
type MockBrowser struct {
	mock.Mock
}

func (m *MockBrowser) GetHistoryPaths() ([]history.HistoryPathEntry, error) {
	return nil, nil
}

func (m *MockBrowser) ExtractHistory(dbPath, profile string, startTime, endTime time.Time, debug bool) ([]history.HistoryEntry, error) {
	return nil, nil
}

func TestHistoryService(t *testing.T) {
	// Setup mock browser with default browsers
	browserMap := map[string]browser.Browser{
		"chrome":  new(MockBrowser),
		"edge":    new(MockBrowser),
		"firefox": new(MockBrowser),
		"brave":   new(MockBrowser),
		"mock":    new(MockBrowser),
	}

	// Create service instance
	service := NewHistoryService(browserMap)
	hs := service.(*historyService)

	// Sample config
	cfg := &config.Config{
		Debug:       false,
		JSONOutput:  false,
		PrettyPrint: false,
	}

	t.Run("resolveBrowsers_Default", func(t *testing.T) {
		browsers := hs.resolveBrowsers(nil)
		assert.Equal(t, []string{"chrome", "edge", "brave", "firefox"}, browsers)
	})

	t.Run("resolveBrowsers_Selected", func(t *testing.T) {
		selected := []string{"chrome", "mock", "invalid"}
		browsers := hs.resolveBrowsers(selected)
		assert.ElementsMatch(t, []string{"chrome", "mock"}, browsers) // Order doesn’t matter
	})

	t.Run("OutputResults_Text", func(t *testing.T) {
		entries := []history.OutputEntry{
			{
				Timestamp:  "2025-04-06T12:00:00Z",
				URL:        "https://example.com",
				Title:      "Example",
				VisitCount: 1,
				Typed:      0,
				VisitType:  "",
				Browser:    "mock",
				Profile:    "default",
			},
		}

		var buf bytes.Buffer
		service.OutputResults(entries, cfg, &buf)

		// Match actual output: 20 chars + 11 spaces (as observed)
		expected := "2025-04-06T12:00:00Z           Example                                            (https://example.com) [1] [0] [] [mock] [default]\n"
		assert.Equal(t, expected, buf.String())
	})

	t.Run("OutputResults_JSON", func(t *testing.T) {
		cfg.JSONOutput = true
		entries := []history.OutputEntry{
			{
				Timestamp:  "2025-04-06T12:00:00Z",
				URL:        "https://example.com",
				Title:      "Example",
				VisitCount: 1,
				Typed:      0,
				VisitType:  "",
				Browser:    "mock",
				Profile:    "default",
			},
		}

		var buf bytes.Buffer
		service.OutputResults(entries, cfg, &buf)

		var result []history.OutputEntry
		err := json.Unmarshal(buf.Bytes()[:buf.Len()-1], &result)
		assert.NoError(t, err)
		assert.Equal(t, entries, result)
	})

	t.Run("OutputResults_NoEntries", func(t *testing.T) {
		cfg.JSONOutput = false
		var buf bytes.Buffer
		service.OutputResults([]history.OutputEntry{}, cfg, &buf)
		assert.Equal(t, "No history entries found.\n", buf.String())
	})
}

func TestNewHistoryService(t *testing.T) {
	t.Run("WithNilBrowserMap", func(t *testing.T) {
		service := NewHistoryService(nil)
		assert.NotNil(t, service)
		hs, ok := service.(*historyService)
		assert.True(t, ok)
		assert.Len(t, hs.browserMap, 4)
		assert.Contains(t, hs.browserMap, "chrome")
		assert.Contains(t, hs.browserMap, "edge")
		assert.Contains(t, hs.browserMap, "firefox")
		assert.Contains(t, hs.browserMap, "brave")
	})

	t.Run("WithCustomBrowserMap", func(t *testing.T) {
		customMap := map[string]browser.Browser{
			"custom": new(MockBrowser),
		}
		service := NewHistoryService(customMap)
		hs, ok := service.(*historyService)
		assert.True(t, ok)
		assert.Equal(t, customMap, hs.browserMap)
	})
}
