package main

import (
	"encoding/json" // For JSON encoding/decoding
	"fmt"           // For formatted output
	"os"            // For operating system interactions like exiting on error
	"time"          // For handling time-related operations

	"github.com/lotekdan/go-browser-history/internal/browser" // Internal package for browser-specific logic
	"github.com/lotekdan/go-browser-history/internal/history" // Internal package for history retrieval
	"github.com/spf13/cobra"                                  // For command-line interface framework
)

// Version holds the application version, defaults to "dev" if not set during build.
var Version string = "dev"

// OutputEntry defines the structure for a single history entry in the output.
type OutputEntry struct {
	Timestamp string `json:"timestamp"` // ISO 8601 formatted timestamp of the visit
	Title     string `json:"title"`     // Title of the webpage
	URL       string `json:"url"`       // URL of the visited page
	Browser   string `json:"browser"`   // Name of the browser the entry came from
}

// NewRootCmd creates and configures the root command for the CLI application.
func NewRootCmd() *cobra.Command {
	// Declare flags as local variables to be bound to the command
	var days int           // Number of days to look back in history
	var browserType string // Specific browser to query (optional)
	var jsonOutput bool    // Flag to toggle JSON output format

	// Define the root command with its usage, description, and version
	cmd := &cobra.Command{
		Use:     "go-browser-history",                                     // Command name in CLI
		Short:   "Retrieve browser history from Chrome, Edge, or Firefox", // Brief description
		Version: Version,                                                  // Application version
		Run: func(cmd *cobra.Command, args []string) { // Main logic executed when command runs
			// Calculate time range for history retrieval
			endTime := time.Now()                     // Current time as the end of the range
			startTime := endTime.AddDate(0, 0, -days) // Start time based on days flag

			// Initialize supported browsers with their implementations
			browsers := make(map[string]browser.Browser)
			browsers["chrome"] = browser.NewChromeBrowser()   // Chrome browser instance
			browsers["edge"] = browser.NewEdgeBrowser()       // Edge browser instance
			browsers["firefox"] = browser.NewFirefoxBrowser() // Firefox browser instance

			// Determine which browsers to query
			var selectedBrowsers []string
			if browserType == "" {
				// If no browser specified, query all supported browsers
				selectedBrowsers = []string{"chrome", "edge", "firefox"}
			} else {
				// Validate the specified browser type
				if _, exists := browsers[browserType]; !exists {
					fmt.Println("Invalid browser type. Use: chrome, edge, or firefox")
					return
				}
				selectedBrowsers = []string{browserType} // Use only the specified browser
			}

			// Collect all history entries from selected browsers
			var allEntries []OutputEntry
			for _, bType := range selectedBrowsers {
				b := browsers[bType] // Get the browser instance

				// Retrieve the history database file path
				dbPath, err := b.GetHistoryPath()
				if err != nil {
					// Skip this browser if path retrieval fails, log error unless in JSON mode
					if !jsonOutput {
						fmt.Printf("Error finding %s history file: %v\n", bType, err)
					}
					continue
				}

				// Display database path when querying multiple browsers in non-JSON mode
				if !jsonOutput && len(selectedBrowsers) > 1 {
					fmt.Printf("Using %s database path: %s\n", bType, dbPath)
				}

				// Fetch history entries for the specified time range
				entries, err := history.GetBrowserHistory(b, startTime, endTime)
				if err != nil {
					// Skip this browser if history retrieval fails, log error unless in JSON mode
					if !jsonOutput {
						fmt.Printf("Error retrieving %s history (may be locked or inaccessible): %v\n", bType, err)
					}
					continue
				}

				// Convert retrieved entries to output format
				for _, entry := range entries {
					allEntries = append(allEntries, OutputEntry{
						Timestamp: entry.Timestamp.Format(time.RFC3339), // Standardize timestamp format
						Title:     entry.Title,                          // Page title
						URL:       entry.URL,                            // Page URL
						Browser:   bType,                                // Browser name
					})
				}
			}

			// Handle case where no entries are found
			if len(allEntries) == 0 {
				if !jsonOutput {
					fmt.Println("No history entries found for the specified time range across all selected browsers.")
				} else {
					fmt.Println("[]") // Empty JSON array for consistency
				}
				return
			}

			// Output results based on format flag
			if jsonOutput {
				// Serialize entries to indented JSON
				jsonData, err := json.MarshalIndent(allEntries, "", "  ")
				if err != nil {
					fmt.Printf("Error formatting JSON: %v\n", err)
					return
				}
				fmt.Println(string(jsonData)) // Print JSON output
			} else {
				// Print entries in a human-readable table format
				for _, entry := range allEntries {
					title := entry.Title
					if title == "" {
						title = "(no title)" // Substitute for empty titles
					}
					// Format output with fixed-width columns for readability
					fmt.Printf("%-30s %-50s (%s) [%s]\n",
						entry.Timestamp,
						title,
						entry.URL,
						entry.Browser)
				}
			}
		},
	}

	// Bind flags to the command with defaults and descriptions
	cmd.Flags().IntVarP(&days, "days", "d", 30, "Number of days of history to retrieve")
	cmd.Flags().StringVarP(&browserType, "browser", "b", "", "Browser type (chrome, edge, firefox). Leave empty for all browsers")
	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output results in JSON format")

	return cmd // Return the configured command
}

// main is the entry point of the program, executing the root command.
func main() {
	// Execute the root command and handle any errors
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Println(err) // Print error message
		os.Exit(1)       // Exit with failure status
	}
}
