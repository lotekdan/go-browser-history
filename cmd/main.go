package main

import (
	"encoding/json" // For JSON encoding/decoding
	"fmt"           // For formatted output
	"os"            // For operating system interactions like exiting on error
	"time"          // For handling time-related operations

	"github.com/lotekdan/go-browser-history/internal/browser" // Browser-specific logic
	"github.com/lotekdan/go-browser-history/internal/history" // History retrieval logic
	"github.com/spf13/cobra"                                  // CLI framework
)

// Version holds the application version, defaults to "dev" if not set during build.
var Version string = "dev"

// OutputEntry defines the structure for a history entry in the output.
type OutputEntry struct {
	Timestamp string `json:"timestamp"` // ISO 8601 formatted visit time
	Title     string `json:"title"`     // Webpage title
	URL       string `json:"url"`       // Webpage URL
	Browser   string `json:"browser"`   // Source browser name
}

// NewRootCmd creates and configures the root CLI command.
func NewRootCmd() *cobra.Command {
	var historyDays int    // Days of history to retrieve
	var browserName string // Specific browser to query (optional)
	var jsonOutput bool    // Toggle JSON output format

	cmd := &cobra.Command{
		Use:     "go-browser-history",                                     // CLI command name
		Short:   "Retrieve browser history from Chrome, Edge, or Firefox", // Brief description
		Version: Version,                                                  // App version
		Run: func(cmd *cobra.Command, args []string) { // Command execution logic
			endTime := time.Now()
			startTime := endTime.AddDate(0, 0, -historyDays)
			browsers := initializeBrowsers()
			historyEntries := fetchHistory(browsers, browserName, startTime, endTime, jsonOutput)
			outputResults(historyEntries, jsonOutput)
		},
	}

	// Define command-line flags
	cmd.Flags().IntVarP(&historyDays, "days", "d", 30, "Number of days of history to retrieve")
	cmd.Flags().StringVarP(&browserName, "browser", "b", "", "Browser type (chrome, edge, firefox). Leave empty for all")
	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output results in JSON format")

	return cmd
}

// initializeBrowsers sets up the supported browser implementations.
func initializeBrowsers() map[string]browser.Browser {
	browsers := make(map[string]browser.Browser)
	browsers["chrome"] = browser.NewChromeBrowser()
	browsers["edge"] = browser.NewEdgeBrowser()
	browsers["firefox"] = browser.NewFirefoxBrowser()
	return browsers
}

// fetchHistory retrieves history entries from selected browsers.
func fetchHistory(browsers map[string]browser.Browser, browserName string, startTime, endTime time.Time, jsonOutput bool) []OutputEntry {
	var selectedBrowsers []string
	if browserName == "" {
		selectedBrowsers = []string{"chrome", "edge", "firefox"} // Default to all browsers
	} else if _, exists := browsers[browserName]; !exists {
		if !jsonOutput {
			fmt.Println("Invalid browser type. Use: chrome, edge, or firefox")
		}
		return nil
	} else {
		selectedBrowsers = []string{browserName}
	}

	var historyEntries []OutputEntry
	for _, name := range selectedBrowsers {
		browserImpl := browsers[name]
		historyDBPath, err := browserImpl.GetHistoryPath()
		if err != nil {
			if !jsonOutput {
				fmt.Printf("Error finding %s history file: %v\n", name, err)
			}
			continue
		}
		if !jsonOutput && len(selectedBrowsers) > 1 {
			fmt.Printf("Using %s database path: %s\n", name, historyDBPath)
		}

		entries, err := history.GetBrowserHistory(browserImpl, startTime, endTime)
		if err != nil {
			if !jsonOutput {
				fmt.Printf("Error retrieving %s history (may be locked or inaccessible): %v\n", name, err)
			}
			continue
		}

		for _, entry := range entries {
			historyEntries = append(historyEntries, OutputEntry{
				Timestamp: entry.Timestamp.Format(time.RFC3339),
				Title:     entry.Title,
				URL:       entry.URL,
				Browser:   name,
			})
		}
	}
	return historyEntries
}

// outputResults displays the history entries in the requested format.
func outputResults(historyEntries []OutputEntry, jsonOutput bool) {
	if len(historyEntries) == 0 {
		if !jsonOutput {
			fmt.Println("No history entries found for the specified time range across all selected browsers.")
		} else {
			fmt.Println("[]")
		}
		return
	}

	if jsonOutput {
		jsonData, err := json.MarshalIndent(historyEntries, "", "  ")
		if err != nil {
			fmt.Printf("Error formatting JSON: %v\n", err)
			return
		}
		fmt.Println(string(jsonData))
	} else {
		for _, entry := range historyEntries {
			pageTitle := entry.Title
			if pageTitle == "" {
				pageTitle = "(no title)"
			}
			fmt.Printf("%-30s %-50s (%s) [%s]\n", entry.Timestamp, pageTitle, entry.URL, entry.Browser)
		}
	}
}

// main executes the root command and handles errors.
func main() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
