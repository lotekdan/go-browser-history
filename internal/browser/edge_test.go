package browser_test

import (
	"database/sql"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/lotekdan/go-browser-history/internal/browser"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestNewEdgeBrowser(t *testing.T) {
	// Act
	b := browser.NewEdgeBrowser()

	// Assert
	assert.NotNil(t, b)
	_, ok := b.(*browser.EdgeBrowser)
	assert.True(t, ok, "NewEdgeBrowser should return a *EdgeBrowser")
}

func TestEdgeGetHistoryPath_Success(t *testing.T) {
	// Arrange
	eb := &browser.EdgeBrowser{}
	t.Setenv("HOME", "/home/test")
	t.Setenv("LOCALAPPDATA", "C:\\Users\\Test\\AppData\\Local")

	// Act & Assert based on OS
	path, err := eb.GetHistoryPath()
	assert.NoError(t, err)
	switch runtime.GOOS {
	case "windows":
		assert.Equal(t, "C:\\Users\\Test\\AppData\\Local\\Microsoft\\Edge\\User Data\\Default\\History", path)
	case "darwin":
		assert.Equal(t, "/home/test/Library/Application Support/Microsoft Edge/Default/History", path)
	case "linux":
		assert.Equal(t, "/home/test/.config/microsoft-edge/Default/History", path)
	default:
		t.Skipf("unsupported OS for this test: %s", runtime.GOOS)
	}
}

func TestEdgeGetHistoryPath_UnsupportedOS(t *testing.T) {
	// Arrange
	eb := &browser.EdgeBrowser{}

	// Act
	path, err := eb.GetHistoryPath()

	// Assert
	if runtime.GOOS != "windows" && runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		assert.Error(t, err)
		assert.Empty(t, path)
		assert.Contains(t, err.Error(), "unsupported operating system")
	} else {
		t.Log("Skipping unsupported OS test on supported platform")
	}
}

func TestEdgeExtractHistory_Success(t *testing.T) {
	// Arrange: Create a temporary SQLite database file
	tmpFile, err := os.CreateTemp("", "edge_test_*.db")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	db, err := sql.Open("sqlite3", tmpFile.Name())
	assert.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE urls (url TEXT, title TEXT, last_visit_time INTEGER);
		INSERT INTO urls VALUES ('http://example.com', 'Example', ?);
	`, browser.TimeToChromeTime(time.Now().Add(-1*time.Hour)))
	assert.NoError(t, err)

	eb := &browser.EdgeBrowser{}
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()

	// Act
	entries, err := eb.ExtractHistory(tmpFile.Name(), start, end, false)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "http://example.com", entries[0].URL)
	assert.Equal(t, "Example", entries[0].Title)
	assert.True(t, entries[0].Timestamp.After(start) && entries[0].Timestamp.Before(end))
}

func TestEdgeExtractHistory_InvalidDB(t *testing.T) {
	// Arrange
	eb := &browser.EdgeBrowser{}
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()

	// Act
	entries, err := eb.ExtractHistory("nonexistent.db", start, end, true)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, entries)
	assert.Contains(t, err.Error(), "unable to open database file") // Matches SQLite driver error
}
