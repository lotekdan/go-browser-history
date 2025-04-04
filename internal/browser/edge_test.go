package browser

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNewEdgeBrowser(t *testing.T) {
	browser := NewEdgeBrowser()
	if _, ok := browser.(*EdgeBrowser); !ok {
		t.Error("NewEdgeBrowser should return a *EdgeBrowser")
	}
}

func TestEdgeBrowser_GetHistoryPath(t *testing.T) {
	eb := &EdgeBrowser{}
	tempDir := t.TempDir()

	var baseDir string
	switch runtime.GOOS {
	case "windows":
		baseDir = filepath.Join(tempDir, "Microsoft", "Edge", "User Data")
		os.Setenv("LOCALAPPDATA", tempDir)
	case "darwin":
		baseDir = filepath.Join(tempDir, "Library", "Application Support", "Microsoft Edge")
		os.Setenv("HOME", tempDir)
	case "linux":
		baseDir = filepath.Join(tempDir, ".config", "microsoft-edge")
		os.Setenv("HOME", tempDir)
	default:
		t.Skipf("Skipping test on unsupported OS: %s", runtime.GOOS)
	}

	defaultDir := filepath.Join(baseDir, "Default")
	if err := os.MkdirAll(defaultDir, 0755); err != nil {
		t.Fatalf("Failed to create default dir: %v", err)
	}
	historyPath := filepath.Join(defaultDir, "History")
	file, err := os.Create(historyPath)
	if err != nil {
		t.Fatalf("Failed to create History file: %v", err)
	}
	file.Close()

	paths, err := eb.GetHistoryPath()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("Expected at least one path, got none")
	}
	if paths[0] != historyPath {
		t.Errorf("Expected path %s, got %s", historyPath, paths[0])
	}
}
