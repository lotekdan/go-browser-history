package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/lotekdan/go-browser-history/internal/config"
	"github.com/lotekdan/go-browser-history/internal/history"
)

// mockHistoryService implements service.HistoryService
type mockHistoryService struct {
	getHistoryFunc func(cfg *config.Config, selectedBrowsers []string) ([]history.OutputEntry, error)
}

func (m *mockHistoryService) GetHistory(cfg *config.Config, selectedBrowsers []string) ([]history.OutputEntry, error) {
	if m.getHistoryFunc != nil {
		return m.getHistoryFunc(cfg, selectedBrowsers)
	}
	return nil, nil
}

func (m *mockHistoryService) OutputResults(entries []history.OutputEntry, jsonOutput bool, writer io.Writer) {
	// Stub implementation to satisfy service.HistoryService interface
}

func TestStart(t *testing.T) {
	// Mock listenAndServe
	oldListenAndServe := listenAndServe
	listenAndServe = func(addr string, handler http.Handler) error {
		if addr != ":8080" {
			t.Errorf("listenAndServe addr = %q, want %q", addr, ":8080")
		}
		if handler == nil {
			t.Error("listenAndServe handler is nil, want non-nil")
		}
		return nil
	}
	defer func() { listenAndServe = oldListenAndServe }()

	cfg := &config.Config{Port: "8080"}
	err := Start(cfg)
	if err != nil {
		t.Errorf("Start() returned error: %v", err)
	}
}

func TestStart_ListenAndServeError(t *testing.T) {
	// Mock listenAndServe to fail
	oldListenAndServe := listenAndServe
	listenAndServe = func(addr string, handler http.Handler) error {
		return errors.New("listen failed")
	}
	defer func() { listenAndServe = oldListenAndServe }()

	cfg := &config.Config{Port: "8080"}
	err := Start(cfg)
	if err == nil || err.Error() != "listen failed" {
		t.Errorf("Start() error = %v, want %q", err, "listen failed")
	}
}

func TestHistoryHandler_Success(t *testing.T) {
	srv := &mockHistoryService{
		getHistoryFunc: func(cfg *config.Config, selectedBrowsers []string) ([]history.OutputEntry, error) {
			return []history.OutputEntry{
				{Timestamp: "2023-01-01T00:00:00Z", Title: "Test", URL: "http://test.com", Browser: "chrome"},
			}, nil
		},
	}
	cfg := &config.Config{
		HistoryDays: 30,
		Port:        "8080",
		EndTime:     time.Now(),
	}

	req, _ := http.NewRequest("GET", "/history", nil)
	rr := httptest.NewRecorder()
	handler := historyHandler(srv, cfg)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("handler returned status %d, want %d", rr.Code, http.StatusOK)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var entries []history.OutputEntry
	json.Unmarshal(rr.Body.Bytes(), &entries)
	if len(entries) != 1 || entries[0].Title != "Test" {
		t.Errorf("response = %v, want one entry with Title 'Test'", entries)
	}
}

func TestHistoryHandler_BrowsersParam(t *testing.T) {
	srv := &mockHistoryService{
		getHistoryFunc: func(cfg *config.Config, selectedBrowsers []string) ([]history.OutputEntry, error) {
			if !reflect.DeepEqual(selectedBrowsers, []string{"chrome", "firefox"}) {
				t.Errorf("selectedBrowsers = %v, want %v", selectedBrowsers, []string{"chrome", "firefox"})
			}
			return []history.OutputEntry{}, nil
		},
	}
	cfg := &config.Config{HistoryDays: 30, EndTime: time.Now()}

	req, _ := http.NewRequest("GET", "/history?browsers=chrome,firefox", nil)
	rr := httptest.NewRecorder()
	handler := historyHandler(srv, cfg)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("handler returned status %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestHistoryHandler_DaysParamInvalid(t *testing.T) {
	srv := &mockHistoryService{}
	cfg := &config.Config{HistoryDays: 30, EndTime: time.Now()}

	req, _ := http.NewRequest("GET", "/history?days=invalid", nil)
	rr := httptest.NewRecorder()
	handler := historyHandler(srv, cfg)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("handler returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
	if body := rr.Body.String(); body != "Invalid 'days' parameter\n" {
		t.Errorf("response body = %q, want %q", body, "Invalid 'days' parameter\n")
	}
}

func TestHistoryHandler_TimeRangeInvalid(t *testing.T) {
	srv := &mockHistoryService{}
	cfg := &config.Config{HistoryDays: 30, EndTime: time.Now()}

	req, _ := http.NewRequest("GET", "/history?start_time=invalid&end_time=2023-01-01T00:00:00Z", nil)
	rr := httptest.NewRecorder()
	handler := historyHandler(srv, cfg)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("handler returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
	if body := rr.Body.String(); body != "Invalid 'start_time' or 'end_time' format (use RFC3339)\n" {
		t.Errorf("response body = %q, want %q", body, "Invalid 'start_time' or 'end_time' format (use RFC3339)\n")
	}
}

func TestHistoryHandler_PartialTimeRange(t *testing.T) {
	srv := &mockHistoryService{}
	cfg := &config.Config{HistoryDays: 30, EndTime: time.Now()}

	req, _ := http.NewRequest("GET", "/history?start_time=2023-01-01T00:00:00Z", nil)
	rr := httptest.NewRecorder()
	handler := historyHandler(srv, cfg)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("handler returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
	if body := rr.Body.String(); body != "Both 'start_time' and 'end_time' must be provided together\n" {
		t.Errorf("response body = %q, want %q", body, "Both 'start_time' and 'end_time' must be provided together\n")
	}
}

func TestHistoryHandler_GetHistoryError(t *testing.T) {
	srv := &mockHistoryService{
		getHistoryFunc: func(cfg *config.Config, selectedBrowsers []string) ([]history.OutputEntry, error) {
			return nil, errors.New("history fetch failed")
		},
	}
	cfg := &config.Config{HistoryDays: 30, EndTime: time.Now()}

	req, _ := http.NewRequest("GET", "/history", nil)
	rr := httptest.NewRecorder()
	handler := historyHandler(srv, cfg)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("handler returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
	if body := rr.Body.String(); body != "history fetch failed\n" {
		t.Errorf("response body = %q, want %q", body, "history fetch failed\n")
	}
}
