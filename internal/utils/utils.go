package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/lotekdan/go-browser-history/internal/browser"
	"github.com/lotekdan/go-browser-history/internal/history"
)

// GetBrowserHistory retrieves history from the browsers profiles and pulls them into a final result.
func GetBrowserHistory(browserImpl browser.Browser, startTime, endTime time.Time, verbose bool) ([]history.HistoryEntry, error) {
	sourceDBPaths, err := browserImpl.GetHistoryPaths()
	if err != nil {
		return nil, err // Return error silently unless logged elsewhere
	}

	var history []history.HistoryEntry
	for _, sourceDBPath := range sourceDBPaths {
		historyDBPath, cleanup, err := PrepareDatabaseFile(sourceDBPath, verbose)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare database file at %s: %v", sourceDBPath, err)
		}
		defer cleanup()

		entries, err := browserImpl.ExtractHistory(historyDBPath, startTime, endTime, verbose)
		if err != nil {
			return nil, err
		}

		history = append(history, entries...)
	}
	return history, nil
}

func PrepareDatabaseFile(sourceDBPath string, verbose bool) (string, func(), error) {
	tempDir := os.TempDir()
	tempBaseName := fmt.Sprintf("go-browser-history-%s-%d", filepath.Base(sourceDBPath), time.Now().UnixNano())
	tempDBPath := filepath.Join(tempDir, tempBaseName)

	if err := CopyFile(sourceDBPath, tempDBPath); err != nil {
		return "", func() {}, fmt.Errorf("failed to copy database %s to %s: %v", sourceDBPath, tempDBPath, err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Debug: Attempting to copy %s to %s.\n", sourceDBPath, tempDBPath)
	}

	for _, suffix := range []string{"-wal", "-shm"} {
		srcExtra := sourceDBPath + suffix
		dstExtra := tempDBPath + suffix
		if _, err := os.Stat(srcExtra); err == nil {
			if err := CopyFile(srcExtra, dstExtra); err != nil && verbose {
				fmt.Fprintf(os.Stderr, "Debug: Warning: failed to copy %s to %s: %v\n", srcExtra, dstExtra, err)
			}
		}
	}

	cleanup := func() {
		for _, file := range []string{tempDBPath, tempDBPath + "-wal", tempDBPath + "-shm"} {
			if err := os.Remove(file); err != nil && !os.IsNotExist(err) && verbose {
				fmt.Fprintf(os.Stderr, "Debug: Warning: failed to remove temp file %s: %v\n", file, err)
			}
		}
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Debug: Using temp database file: %s\n", tempDBPath)
	}
	return tempDBPath, cleanup, nil
}

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

	return destFile.Sync()
}
