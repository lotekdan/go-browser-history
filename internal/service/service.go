package service

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/lotekdan/go-browser-history/internal/browser"
	"github.com/lotekdan/go-browser-history/internal/config"
	"github.com/lotekdan/go-browser-history/internal/history"
	"github.com/lotekdan/go-browser-history/internal/utils"
)

// Define the HistoryService interface
type HistoryService interface {
	GetHistory(cfg *config.Config, selectedBrowsers []string) ([]history.OutputEntry, error)
	OutputResults(entries []history.OutputEntry, cfg *config.Config, writer io.Writer)
}

// Concrete implementation of HistoryService
type historyService struct {
	browserMap map[string]browser.Browser
}

// Ensure historyService implements the interface
var _ HistoryService = (*historyService)(nil)

// Constructor function
func NewHistoryService(browserMap map[string]browser.Browser) HistoryService {
	if browserMap == nil {
		browserMap = initializeBrowsers()
	}
	return &historyService{
		browserMap: browserMap,
	}
}

func initializeBrowsers() map[string]browser.Browser {
	return map[string]browser.Browser{
		"chrome":  browser.NewChromeBrowser(),
		"edge":    browser.NewEdgeBrowser(),
		"firefox": browser.NewFirefoxBrowser(),
		"brave":   browser.NewBraveBrowser(),
	}
}

// Implement GetHistory method
func (s *historyService) GetHistory(cfg *config.Config, selectedBrowsers []string) ([]history.OutputEntry, error) {
	browserList := s.resolveBrowsers(selectedBrowsers)
	if len(browserList) == 0 {
		return nil, fmt.Errorf("no valid browsers specified")
	}

	cfg.StartTime = cfg.EndTime.AddDate(0, 0, -cfg.HistoryDays)
	return s.fetchHistory(cfg, browserList)
}

func (s *historyService) resolveBrowsers(selectedBrowsers []string) []string {
	if len(selectedBrowsers) == 0 {
		return []string{"chrome", "edge", "brave", "firefox"}
	}
	var validBrowsers []string
	for _, name := range selectedBrowsers {
		if _, exists := s.browserMap[name]; exists {
			validBrowsers = append(validBrowsers, name)
		}
	}
	return validBrowsers
}

func (s *historyService) fetchHistory(cfg *config.Config, browsers []string) ([]history.OutputEntry, error) {
	var entries []history.OutputEntry
	for _, name := range browsers {
		browserImpl := s.browserMap[name]
		historyDBPaths, err := browserImpl.GetHistoryPaths()
		if err != nil {
			if shouldLog(cfg) {
				fmt.Fprintf(os.Stderr, "Debug: Error finding %s history file: %v\n", name, err)
			}
			continue
		}
		if shouldLog(cfg) && len(browsers) > 1 {
			fmt.Fprintf(os.Stderr, "Debug: Using %s database path: %s\n", name, historyDBPaths)
		}

		browserEntries, err := utils.GetBrowserHistory(browserImpl, cfg.StartTime, cfg.EndTime, shouldLog(cfg))
		if err != nil {
			if shouldLog(cfg) {
				fmt.Fprintf(os.Stderr, "Debug: Error retrieving %s history: %v\n", name, err)
			}
			continue
		}
		entries = append(entries, history.ToOutputEntries(browserEntries, name)...)
	}
	return entries, nil
}

// Implement OutputResults method
func (s *historyService) OutputResults(entries []history.OutputEntry, cfg *config.Config, writer io.Writer) {
	if cfg.JSONOutput {
		if cfg.PrettyPrint {
			jsonData, err := json.MarshalIndent(entries, "", "  ")
			if err != nil {
				fmt.Fprintln(writer, "[]") // Include newline on error
				return
			}
			fmt.Fprintln(writer, string(jsonData))
			return
		}
		jsonData, err := json.Marshal(entries)
		if err != nil {
			fmt.Fprintln(writer, "[]") // Include newline on error
			return
		}
		fmt.Fprintln(writer, string(jsonData)) // Use Fprintln to add newline
		return
	}

	if len(entries) == 0 {
		fmt.Fprintln(writer, "No history entries found.")
		return
	}

	for _, entry := range entries {
		title := entry.Title
		if title == "" {
			title = "(no title)"
		}
		fmt.Fprintf(writer, "%-30s %-50s (%s) [%d] [%d] [%s] [%s]\n",
			entry.Timestamp,
			title,
			entry.URL,
			entry.VisitCount,
			entry.Typed,
			entry.VisitType,
			entry.Browser)
	}
}

func shouldLog(cfg *config.Config) bool {
	return cfg.Debug || cfg.Mode == "api" // Log if --debug is set or in API mode
}
