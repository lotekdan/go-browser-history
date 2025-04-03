package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/lotekdan/go-browser-history/internal/config"
	"github.com/lotekdan/go-browser-history/internal/service"
	"github.com/spf13/cobra"
)

func NewRootCmd(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "go-browser-history",
		Short: "Retrieve browser history from Chrome, Edge, or Firefox",
		Run: func(cmd *cobra.Command, args []string) {
			historyService := service.NewHistoryService()
			if err := runCLI(historyService, cfg); err != nil {
				if cfg.JSONOutput {
					fmt.Fprintf(os.Stderr, `{"error": "CLI execution failed: %v"}`, err)
				} else {
					fmt.Fprintf(os.Stderr, "CLI execution failed: %v\n", err)
				}
				os.Exit(1)
			}
		},
	}
	// No need to redefine flags here; they’re inherited from the root command in cmd/main.go
	return cmd
}

func runCLI(srv *service.HistoryService, cfg *config.Config) error {
	browsers := parseBrowsers(cfg.Browser)
	entries, err := srv.GetHistory(cfg, browsers)
	if err != nil {
		return err
	}
	srv.OutputResults(entries, cfg.JSONOutput, os.Stdout)
	return nil
}

func parseBrowsers(browserString string) []string {
	if browserString == "" {
		return nil
	}
	return strings.Split(browserString, ",")
}
