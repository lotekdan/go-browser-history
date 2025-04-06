package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lotekdan/go-browser-history/internal/config"
	"github.com/lotekdan/go-browser-history/internal/history"
	"github.com/lotekdan/go-browser-history/internal/service"
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

func (m *mockHistoryService) OutputResults(entries []history.OutputEntry, cfg *config.Config, writer io.Writer) {
	if cfg.JSONOutput {
		jsonData, err := json.Marshal(entries)
		if err != nil {
			fmt.Fprintln(writer, "[]") // Return empty JSON array on error
			return
		}
		fmt.Fprintln(writer, string(jsonData))
	} else {
		if len(entries) == 0 {
			fmt.Fprintln(writer, "No history entries found.")
			return
		}
		for _, entry := range entries {
			title := entry.Title
			if title == "" {
				title = "(no title)"
			}
			fmt.Fprintf(writer, "%-30s %-50s (%s) [%s]\n", entry.Timestamp, title, entry.URL, entry.Browser)
		}
	}
}

func TestStart(t *testing.T) {
	// Reset http.DefaultServeMux to avoid duplicate route panic
	http.DefaultServeMux = new(http.ServeMux)

	cfg := &config.Config{Port: "8080"}
	srv := service.NewHistoryService(nil)
	handler := historyHandler(srv, cfg)

	// Ensure handler is set up correctly
	if handler == nil {
		t.Fatalf("listenAndServe handler is nil, want non-nil")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/history", handler)

	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Get(server.URL + "/history")
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()
}

func TestStart_ListenAndServeError(t *testing.T) {
	cfg := &config.Config{Port: "8080"}
	srv := service.NewHistoryService(nil)
	handler := historyHandler(srv, cfg)

	if handler == nil {
		t.Fatalf("listenAndServe handler is nil, want non-nil")
	}

	// Use a new ServeMux to prevent duplicate route panic
	mux := http.NewServeMux()
	mux.HandleFunc("/history", handler)

	server := httptest.NewServer(mux) // Create a new test server with this mux
	defer server.Close()

	resp, err := http.Get(server.URL + "/history")
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()
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
			return []history.OutputEntry{
				{Timestamp: "2023-01-01T00:00:00Z", Title: "Test", URL: "http://test.com", Browser: "chrome"},
			}, nil
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
