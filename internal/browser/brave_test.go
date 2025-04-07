package browser

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lotekdan/go-browser-history/internal/history"
)

func TestNewBraveBrowser(t *testing.T) {
	browser := NewBraveBrowser()
	if _, ok := browser.(*BraveBrowser); !ok {
		t.Error("NewBraveBrowser should return a *BraveBrowser")
	}
}

func TestBraveBrowser_GetHistoryPath(t *testing.T) {
	tests := []struct {
		name           string
		setupProfile   bool
		expectPaths    bool
		expectedMinLen int
	}{
		{
			name:           "WithProfile",
			setupProfile:   true,
			expectPaths:    true,
			expectedMinLen: 1,
		},
		{
			name:           "NoProfile",
			setupProfile:   false,
			expectPaths:    false,
			expectedMinLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bb := &BraveBrowser{}
			tempDir := t.TempDir()

			var baseDir string
			switch runtime.GOOS {
			case "windows":
				baseDir = filepath.Join(tempDir, "BraveSoftware", "Brave-Browser", "User Data")
				os.Setenv("LOCALAPPDATA", tempDir)
			case "darwin":
				baseDir = filepath.Join(tempDir, "Library", "Application Support", "BraveSoftware", "Brave-Browser")
				os.Setenv("HOME", tempDir)
			case "linux":
				baseDir = filepath.Join(tempDir, ".config", "BraveSoftware", "Brave-Browser")
				os.Setenv("HOME", tempDir)
			default:
				t.Skipf("Skipping test on unsupported OS: %s", runtime.GOOS)
			}

			var historyPath string
			if tt.setupProfile {
				defaultDir := filepath.Join(baseDir, "Default")
				if err := os.MkdirAll(defaultDir, 0755); err != nil {
					t.Fatalf("Failed to create default dir: %v", err)
				}
				historyPath = filepath.Join(defaultDir, "History")
				file, err := os.Create(historyPath)

				if err != nil {
					t.Fatalf("Failed to create History file: %v", err)
				}
				file.Close()

				entries, err := os.ReadDir(baseDir)
				if err != nil {
					t.Fatalf("Failed to read baseDir: %v", err)
				}
				fmt.Printf("baseDir contents: %v\n", entries)
			}

			paths, err := bb.GetHistoryPaths()
			fmt.Printf("paths: %v, err: %v\n", paths, err)
			historyPaths := history.HistoryPathEntry{
				Profile: "Default",
				Path:    historyPath,
			}
			if err != nil && tt.expectPaths {
				t.Errorf("Unexpected error when expecting paths: %v", err)
			}
			if len(paths) < tt.expectedMinLen {
				t.Errorf("Expected at least %d paths, got %d", tt.expectedMinLen, len(paths))
			}
			if tt.setupProfile && len(paths) > 0 && paths[0] != historyPaths {
				t.Errorf("Expected path %s, got %s", historyPath, paths[0])
			}
			if !tt.setupProfile && err != nil && !os.IsNotExist(err) {
				t.Errorf("Expected nil or an error indicating path not exist, got %v", err)
			}
		})
	}
}
