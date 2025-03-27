package history

import (
	"fmt"           // For formatted output and error messages
	"io"            // For input/output operations like file copying
	"os"            // For operating system interactions like file handling
	"path/filepath" // For manipulating file paths
	"time"          // For handling time-related operations

	"github.com/lotekdan/go-browser-history/internal/browser" // Internal package for browser-specific logic
)

// GetBrowserHistory retrieves history entries for the specified browser and time range.
func GetBrowserHistory(b browser.Browser, startTime, endTime time.Time) ([]browser.HistoryEntry, error) {
	// Fetch the original history database file path from the browser implementation
	originalPath, err := b.GetHistoryPath()
	if err != nil {
		return nil, err // Return early if path retrieval fails
	}

	// Prepare the database file, handling potential locks by creating a copy if needed
	dbPath, cleanup, err := prepareDatabaseFile(originalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare database file: %v", err) // Wrap error with context
	}
	// Ensure cleanup is called to remove any temporary files when done
	defer cleanup()

	// Extract history entries from the database within the specified time range
	entries, err := b.ExtractHistory(dbPath, startTime, endTime)
	if err != nil {
		return nil, err // Return early if extraction fails
	}

	return entries, nil // Return the retrieved history entries
}

// prepareDatabaseFile tries to use the original file or creates a copy if locked.
func prepareDatabaseFile(originalPath string) (string, func(), error) {
	// Attempt to open the file with read/write access to check if it's locked
	f, err := os.OpenFile(originalPath, os.O_RDWR, 0666)
	if err == nil {
		// If successful, the file is not locked; use it directly
		f.Close()
		return originalPath, func() {}, nil // No-op cleanup function
	}

	// If opening fails (likely due to a lock), create a temporary copy
	tempDir := os.TempDir() // Get the system's temporary directory
	// Construct a unique temporary file name using the original file's base name and a timestamp
	tempPath := filepath.Join(tempDir, "go-browser-history-"+filepath.Base(originalPath)+"-"+fmt.Sprintf("%d", time.Now().UnixNano()))

	// Copy the original file to the temporary location
	err = copyFile(originalPath, tempPath)
	if err != nil {
		return "", func() {}, fmt.Errorf("failed to create temporary copy: %v", err) // Wrap error with context
	}

	// Define a cleanup function to remove the temporary file
	cleanup := func() {
		if err := os.Remove(tempPath); err != nil {
			// Log a warning if cleanup fails, but don't interrupt execution
			fmt.Fprintf(os.Stderr, "Warning: failed to remove temp file %s: %v\n", tempPath, err)
		}
	}

	return tempPath, cleanup, nil // Return the temp path and cleanup function
}

// copyFile creates a copy of the source file at the destination path.
func copyFile(src, dst string) error {
	// Open the source file for reading
	sourceFile, err := os.Open(src)
	if err != nil {
		return err // Return early if source file can't be opened
	}
	defer sourceFile.Close() // Ensure source file is closed when done

	// Create the destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return err // Return early if destination file can't be created
	}
	defer destFile.Close() // Ensure destination file is closed when done

	// Copy contents from source to destination
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err // Return early if copying fails
	}

	// Sync the destination file to ensure all data is written to disk
	return destFile.Sync()
}
