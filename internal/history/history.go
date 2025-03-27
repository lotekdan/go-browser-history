package history

import (
	"fmt"           // For formatted output and error messages
	"io"            // For file copying operations
	"os"            // For operating system interactions like file handling
	"path/filepath" // For manipulating file paths
	"time"          // For generating unique temp file names

	"github.com/lotekdan/go-browser-history/internal/browser" // Browser-specific logic
)

// GetBrowserHistory retrieves history entries for the specified browser and time range.
func GetBrowserHistory(browserImpl browser.Browser, startTime, endTime time.Time) ([]browser.HistoryEntry, error) {
	sourceDBPath, err := browserImpl.GetHistoryPath()
	if err != nil {
		return nil, err
	}

	historyDBPath, cleanup, err := PrepareDatabaseFile(sourceDBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare database file at %s: %v", sourceDBPath, err)
	}
	defer cleanup()

	entries, err := browserImpl.ExtractHistory(historyDBPath, startTime, endTime)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// PrepareDatabaseFile tries to use the original file or creates a copy if locked.
func PrepareDatabaseFile(sourceDBPath string) (string, func(), error) {
	// Attempt to open the file with write access to detect locks
	f, err := os.OpenFile(sourceDBPath, os.O_RDWR, 0666)
	if err == nil {
		f.Close()
		return sourceDBPath, func() {}, nil // No cleanup needed for original file
	}

	// Create a unique temporary file path
	tempDir := os.TempDir()
	tempDBPath := filepath.Join(tempDir, fmt.Sprintf("go-browser-history-%s-%d", filepath.Base(sourceDBPath), time.Now().UnixNano()))

	// Copy the original file to the temporary location
	if err := CopyFile(sourceDBPath, tempDBPath); err != nil {
		return "", func() {}, fmt.Errorf("failed to create temporary copy of %s: %v", sourceDBPath, err)
	}

	// Define cleanup function to remove the temporary file
	cleanup := func() {
		if err := os.Remove(tempDBPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to remove temp file %s: %v\n", tempDBPath, err)
		}
	}

	return tempDBPath, cleanup, nil
}

// CopyFile creates a copy of the source file at the destination path.
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %v", src, err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %v", dst, err)
	}
	defer destFile.Close()

	if _, err = io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy from %s to %s: %v", src, dst, err)
	}

	return destFile.Sync() // Ensure all data is written to disk
}
