package history

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/lotekdan/go-browser-history/internal/browser"
)

// GetBrowserHistory retrieves history entries for the specified browser and time range
func GetBrowserHistory(b browser.Browser, startTime, endTime time.Time) ([]browser.HistoryEntry, error) {
	originalPath, err := b.GetHistoryPath()
	if err != nil {
		return nil, err
	}

	dbPath, cleanup, err := prepareDatabaseFile(originalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare database file: %v", err)
	}
	defer cleanup()

	entries, err := b.ExtractHistory(dbPath, startTime, endTime)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// prepareDatabaseFile tries to use the original file or creates a copy if locked
func prepareDatabaseFile(originalPath string) (string, func(), error) {
	// Try to open the file with write access to detect locks
	f, err := os.OpenFile(originalPath, os.O_RDWR, 0666)
	if err == nil {
		f.Close()
		return originalPath, func() {}, nil
	}

	// If we can't open for writing, assume it's locked
	tempDir := os.TempDir()
	tempPath := filepath.Join(tempDir, "go-browser-history-"+filepath.Base(originalPath)+"-"+fmt.Sprintf("%d", time.Now().UnixNano()))

	err = copyFile(originalPath, tempPath)
	if err != nil {
		return "", func() {}, fmt.Errorf("failed to create temporary copy: %v", err)
	}

	cleanup := func() {
		if err := os.Remove(tempPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to remove temp file %s: %v\n", tempPath, err)
		}
	}

	return tempPath, cleanup, nil
}

// copyFile creates a copy of the source file at the destination path
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return destFile.Sync()
}
