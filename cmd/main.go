package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/lotekdan/go-browser-history/internal/browser"
	"github.com/lotekdan/go-browser-history/internal/history"
	"github.com/spf13/cobra"
)

var Version string = "dev"

type OutputEntry struct {
	Timestamp string `json:"timestamp"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Browser   string `json:"browser"`
}

func NewRootCmd() *cobra.Command {
	var days int
	var browserType string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:     "go-browser-history",
		Short:   "Retrieve browser history from Chrome, Edge, or Firefox",
		Version: Version,
		Run: func(cmd *cobra.Command, args []string) {
			endTime := time.Now()
			startTime := endTime.AddDate(0, 0, -days)

			browsers := make(map[string]browser.Browser)
			browsers["chrome"] = browser.NewChromeBrowser()
			browsers["edge"] = browser.NewEdgeBrowser()
			browsers["firefox"] = browser.NewFirefoxBrowser()

			var selectedBrowsers []string
			if browserType == "" {
				selectedBrowsers = []string{"chrome", "edge", "firefox"}
			} else {
				if _, exists := browsers[browserType]; !exists {
					fmt.Println("Invalid browser type. Use: chrome, edge, or firefox")
					return
				}
				selectedBrowsers = []string{browserType}
			}

			var allEntries []OutputEntry
			for _, bType := range selectedBrowsers {
				b := browsers[bType]
				dbPath, err := b.GetHistoryPath()
				if err != nil {
					if !jsonOutput {
						fmt.Printf("Error finding %s history file: %v\n", bType, err)
					}
					continue
				}
				// Only print DB path when querying all browsers and not in JSON mode
				if !jsonOutput && len(selectedBrowsers) > 1 {
					fmt.Printf("Using %s database path: %s\n", bType, dbPath)
				}

				entries, err := history.GetBrowserHistory(b, startTime, endTime)
				if err != nil {
					if !jsonOutput {
						fmt.Printf("Error retrieving %s history (may be locked or inaccessible): %v\n", bType, err)
					}
					continue
				}

				for _, entry := range entries {
					allEntries = append(allEntries, OutputEntry{
						Timestamp: entry.Timestamp.Format(time.RFC3339),
						Title:     entry.Title,
						URL:       entry.URL,
						Browser:   bType,
					})
				}
			}

			if len(allEntries) == 0 {
				if !jsonOutput {
					fmt.Println("No history entries found for the specified time range across all selected browsers.")
				} else {
					fmt.Println("[]")
				}
				return
			}

			if jsonOutput {
				jsonData, err := json.MarshalIndent(allEntries, "", "  ")
				if err != nil {
					fmt.Printf("Error formatting JSON: %v\n", err)
					return
				}
				fmt.Println(string(jsonData))
			} else {
				for _, entry := range allEntries {
					title := entry.Title
					if title == "" {
						title = "(no title)"
					}
					fmt.Printf("%-30s %-50s (%s) [%s]\n",
						entry.Timestamp,
						title,
						entry.URL,
						entry.Browser)
				}
			}
		},
	}

	cmd.Flags().IntVarP(&days, "days", "d", 30, "Number of days of history to retrieve")
	cmd.Flags().StringVarP(&browserType, "browser", "b", "", "Browser type (chrome, edge, firefox). Leave empty for all browsers")
	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output results in JSON format")

	return cmd
}

func main() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
