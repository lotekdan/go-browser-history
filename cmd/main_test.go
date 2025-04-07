package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/lotekdan/go-browser-history/internal/config"
	"github.com/lotekdan/go-browser-history/internal/history"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockHistoryService mocks the HistoryService interface
type MockHistoryService struct {
	mock.Mock
}

func (m *MockHistoryService) GetHistory(cfg *config.Config, selectedBrowsers []string) ([]history.OutputEntry, error) {
	args := m.Called(cfg, selectedBrowsers)
	return args.Get(0).([]history.OutputEntry), args.Error(1)
}

func (m *MockHistoryService) OutputResults(entries []history.OutputEntry, cfg *config.Config, writer io.Writer) {
	m.Called(entries, cfg, writer)
}

// SetupRootCmd creates a testable rootCmd with mocked dependencies
func setupRootCmd(t *testing.T, mockService *MockHistoryService) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	cfg := config.NewDefaultConfig()
	var browsers []string
	var mode string

	rootCmd := &cobra.Command{
		Use:   "go-browser-history",
		Short: "Retrieve browser history",
		Run: func(cmd *cobra.Command, args []string) {
			cfg.Browser = strings.Join(browsers, ",")
			cfg.Mode = mode
			switch mode {
			case "api":
				fmt.Fprintf(os.Stderr, "API mode not tested here\n")
			default: // CLI mode
				historyService := mockService
				browserList := parseBrowsers(cfg.Browser)
				entries, err := historyService.GetHistory(cfg, browserList)
				if err != nil {
					if cfg.JSONOutput {
						fmt.Fprintf(os.Stderr, `{"error": "Failed to retrieve history: %v"}`, err)
					} else {
						fmt.Fprintf(os.Stderr, "Failed to retrieve history: %v\n", err)
					}
					return // Avoid os.Exit in test
				}
				historyService.OutputResults(entries, cfg, os.Stdout)
			}
		},
	}

	// Define flags matching main.go
	rootCmd.Flags().IntVarP(&cfg.HistoryDays, "days", "d", cfg.HistoryDays, "Number of days of history to retrieve")
	rootCmd.Flags().StringSliceVarP(&browsers, "browser", "b", nil, "Browser types (chrome, edge, brave, firefox)")
	rootCmd.Flags().BoolVarP(&cfg.JSONOutput, "json", "j", false, "Output results in JSON format (CLI only)")
	rootCmd.Flags().BoolVar(&cfg.PrettyPrint, "pretty", false, "For JSON output providing a pretty print format")
	rootCmd.Flags().StringVarP(&mode, "mode", "m", "cli", "Run mode: 'cli' (default) or 'api'")
	rootCmd.Flags().StringVarP(&cfg.Port, "port", "p", cfg.Port, "Port for API mode")
	rootCmd.Flags().BoolVarP(&cfg.Debug, "debug", "", false, "Enable debug logging")

	// Redirect stdout and stderr using pipes
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	stdoutPipe, stdoutWriter, err := os.Pipe()
	assert.NoError(t, err)
	stderrPipe, stderrWriter, err := os.Pipe()
	assert.NoError(t, err)

	osStdout := os.Stdout
	osStderr := os.Stderr
	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter
	t.Cleanup(func() {
		os.Stdout = osStdout
		os.Stderr = osStderr
	})

	// Use WaitGroup to synchronize output capture
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(stdout, stdoutPipe)
	}()
	go func() {
		defer wg.Done()
		_, _ = io.Copy(stderr, stderrPipe)
	}()

	return rootCmd, stdout, stderr
}

func TestMainCmd(t *testing.T) {
	t.Run("parseBrowsers", func(t *testing.T) {
		tests := []struct {
			input    string
			expected []string
		}{
			{"", nil},
			{"chrome", []string{"chrome"}},
			{"chrome,firefox", []string{"chrome", "firefox"}},
		}
		for _, tt := range tests {
			result := parseBrowsers(tt.input)
			assert.Equal(t, tt.expected, result, "parseBrowsers(%q)", tt.input)
		}
	})

	t.Run("CLI_Success", func(t *testing.T) {
		mockService := new(MockHistoryService)
		rootCmd, stdout, stderr := setupRootCmd(t, mockService)

		// Mock GetHistory
		entries := []history.OutputEntry{
			{
				Timestamp: "2025-04-06T12:00:00Z",
				URL:       "https://example.com",
				Title:     "Example",
				Browser:   "chrome",
			},
		}
		mockService.On("GetHistory", mock.Anything, []string{"chrome"}).Return(entries, nil)
		mockService.On("OutputResults", entries, mock.Anything, os.Stdout).Run(func(args mock.Arguments) {
			writer := args.Get(2).(*os.File)
			_, _ = writer.WriteString("2025-04-06T12:00:00Z           Example                                            (https://example.com) [0] [0] [] [chrome] []\n")
		})

		// Set flags and execute
		rootCmd.SetArgs([]string{"--browser", "chrome", "--mode", "cli"})
		err := rootCmd.Execute()
		assert.NoError(t, err)

		// Close writers to flush pipes and wait for output
		os.Stdout.Close()
		os.Stderr.Close()
		wg := &sync.WaitGroup{}
		wg.Add(2)
		go func() { defer wg.Done(); io.Copy(stdout, os.Stdout) }()
		go func() { defer wg.Done(); io.Copy(stderr, os.Stderr) }()
		wg.Wait()

		assert.Equal(t, "2025-04-06T12:00:00Z           Example                                            (https://example.com) [0] [0] [] [chrome] []\n", stdout.String())
		assert.Empty(t, stderr.String())
		mockService.AssertExpectations(t)
	})

	t.Run("CLI_Error", func(t *testing.T) {
		mockService := new(MockHistoryService)
		rootCmd, stdout, stderr := setupRootCmd(t, mockService)

		// Mock GetHistory to return an error
		mockService.On("GetHistory", mock.Anything, []string{"firefox"}).Return(([]history.OutputEntry)(nil), errors.New("history error"))

		// Set flags and execute
		rootCmd.SetArgs([]string{"--browser", "firefox", "--mode", "cli"})
		err := rootCmd.Execute()
		assert.NoError(t, err)

		// Close writers to flush pipes and wait for output
		os.Stdout.Close()
		os.Stderr.Close()
		wg := &sync.WaitGroup{}
		wg.Add(2)
		go func() { defer wg.Done(); io.Copy(stdout, os.Stdout) }()
		go func() { defer wg.Done(); io.Copy(stderr, os.Stderr) }()
		wg.Wait()

		assert.Empty(t, stdout.String())
		assert.Equal(t, "Failed to retrieve history: history error\n", stderr.String())
		mockService.AssertExpectations(t)
	})

	t.Run("CLI_JSON_Error", func(t *testing.T) {
		mockService := new(MockHistoryService)
		rootCmd, stdout, stderr := setupRootCmd(t, mockService)

		// Mock GetHistory to return an error with JSON output
		mockService.On("GetHistory", mock.Anything, []string{"firefox"}).Return(([]history.OutputEntry)(nil), errors.New("history error"))

		// Set flags and execute
		rootCmd.SetArgs([]string{"--browser", "firefox", "--mode", "cli", "--json"})
		err := rootCmd.Execute()
		assert.NoError(t, err)

		// Close writers to flush pipes and wait for output
		os.Stdout.Close()
		os.Stderr.Close()
		wg := &sync.WaitGroup{}
		wg.Add(2)
		go func() { defer wg.Done(); io.Copy(stdout, os.Stdout) }()
		go func() { defer wg.Done(); io.Copy(stderr, os.Stderr) }()
		wg.Wait()

		assert.Empty(t, stdout.String())
		assert.Equal(t, `{"error": "Failed to retrieve history: history error"}`, stderr.String())
		mockService.AssertExpectations(t)
	})
}
