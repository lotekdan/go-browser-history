package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/lotekdan/go-browser-history/internal/config"
	"github.com/lotekdan/go-browser-history/internal/server"
	"github.com/lotekdan/go-browser-history/internal/service"
	"github.com/spf13/cobra"
)

var Version string = "dev"

func main() {
	cfg := config.NewDefaultConfig()
	var browsers []string
	var mode string

	rootCmd := &cobra.Command{
		Use:   "go-browser-history",
		Short: "Retrieve browser history from Chrome, Edge, or Firefox",
		Run: func(cmd *cobra.Command, args []string) {
			cfg.Browser = strings.Join(browsers, ",")
			switch mode {
			case "api":
				cfg.Mode = "api"
				if err := server.Start(cfg); err != nil {
					fmt.Fprintf(os.Stderr, "Critical error starting API: %v\n", err)
					os.Exit(1)
				}
			default: // CLI mode
				cfg.Mode = "cli"
				historyService := service.NewHistoryService()
				browserList := parseBrowsers(cfg.Browser)
				entries, err := historyService.GetHistory(cfg, browserList)
				if err != nil {
					if cfg.JSONOutput {
						fmt.Fprintf(os.Stderr, `{"error": "Failed to retrieve history: %v"}`, err)
					} else {
						fmt.Fprintf(os.Stderr, "Failed to retrieve history: %v\n", err)
					}
					os.Exit(1)
				}
				historyService.OutputResults(entries, cfg.JSONOutput, os.Stdout)
			}
		},
	}
	rootCmd.Flags().IntVarP(&cfg.HistoryDays, "days", "d", cfg.HistoryDays, "Number of days of history to retrieve")
	rootCmd.Flags().StringSliceVarP(&browsers, "browser", "b", nil, "Browser types (chrome, edge, firefox)")
	rootCmd.Flags().BoolVarP(&cfg.JSONOutput, "json", "j", false, "Output results in JSON format (CLI only)")
	rootCmd.Flags().StringVarP(&mode, "mode", "m", "cli", "Run mode: 'cli' (default) or 'api'")
	rootCmd.Flags().StringVarP(&cfg.Port, "port", "p", cfg.Port, "Port for API mode")
	rootCmd.Flags().BoolVarP(&cfg.Debug, "debug", "", false, "Enable debug logging")
	rootCmd.Version = Version

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Critical error: %v\n", err)
		os.Exit(1)
	}
}

func parseBrowsers(browserString string) []string {
	if browserString == "" {
		return nil
	}
	return strings.Split(browserString, ",")
}
