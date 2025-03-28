package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/lotekdan/go-browser-history/internal/config"
	"github.com/lotekdan/go-browser-history/internal/service"
)

// listenAndServe allows mocking http.ListenAndServe in tests
var listenAndServe = http.ListenAndServe

func Start(cfg *config.Config) error {
	srv := service.NewHistoryService()

	// Ensure handler is non-nil
	historyHandler := historyHandler(srv, cfg)
	if historyHandler == nil {
		return fmt.Errorf("history handler is nil")
	}

	// Register handler safely
	http.HandleFunc("/history", historyHandler)

	port := fmt.Sprintf(":%s", cfg.Port)
	return http.ListenAndServe(port, nil)
}

func historyHandler(srv service.HistoryService, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		query := r.URL.Query()
		browserParam := query.Get("browsers")
		daysParam := query.Get("days")
		startTimeParam := query.Get("start_time")
		endTimeParam := query.Get("end_time")

		// Clone config to avoid modifying the original
		localCfg := *cfg

		// Handle browsers
		var selectedBrowsers []string
		if browserParam != "" {
			selectedBrowsers = strings.Split(browserParam, ",")
		} else if localCfg.Browser != "" {
			selectedBrowsers = strings.Split(localCfg.Browser, ",")
		}

		// Handle days if provided
		if daysParam != "" {
			if days, err := strconv.Atoi(daysParam); err == nil && days > 0 {
				localCfg.HistoryDays = days
			} else {
				http.Error(w, "Invalid 'days' parameter", http.StatusBadRequest)
				return
			}
		}

		// Handle custom time range if provided
		if startTimeParam != "" && endTimeParam != "" {
			startTime, err1 := time.Parse(time.RFC3339, startTimeParam)
			endTime, err2 := time.Parse(time.RFC3339, endTimeParam)
			if err1 != nil || err2 != nil || startTime.After(endTime) {
				http.Error(w, "Invalid 'start_time' or 'end_time' format (use RFC3339)", http.StatusBadRequest)
				return
			}
			localCfg.StartTime = startTime
			localCfg.EndTime = endTime
		} else if startTimeParam == "" && endTimeParam == "" {
			// Use default time range based on days if no custom range is specified
			localCfg.EndTime = time.Now()
			localCfg.StartTime = localCfg.EndTime.AddDate(0, 0, -localCfg.HistoryDays)
		} else {
			http.Error(w, "Both 'start_time' and 'end_time' must be provided together", http.StatusBadRequest)
			return
		}

		entries, err := srv.GetHistory(&localCfg, selectedBrowsers)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(entries); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
