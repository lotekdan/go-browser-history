package history

import (
	"time"
)

// MockBrowser mocks the browser.Browser interface
type MockBrowser struct {
	HistoryPath       []string
	ExtractHistoryFn  func(dbPath string, startTime, endTime time.Time, verbose bool) ([]HistoryEntry, error)
	GetHistoryPathsFn func() ([]string, error)
}

func (m *MockBrowser) GetHistoryPaths() ([]string, error) {
	if m.GetHistoryPathsFn != nil {
		return m.GetHistoryPathsFn()
	}
	return m.HistoryPath, nil
}

func (m *MockBrowser) ExtractHistory(dbPath string, startTime, endTime time.Time, verbose bool) ([]HistoryEntry, error) {
	if m.ExtractHistoryFn != nil {
		return m.ExtractHistoryFn(dbPath, startTime, endTime, verbose)
	}
	return nil, nil
}
