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

func TestNewChromeBrowser(t *testing.T) {
	b := browser.NewChromeBrowser()
	assert.NotNil(t, b)
	_, ok := b.(*browser.ChromeBrowser)
	assert.True(t, ok, "NewChromeBrowser should return a *ChromeBrowser")
}

func TestChromeGetHistoryPath_Success(t *testing.T) {
	cb := &browser.ChromeBrowser{}
	t.Setenv("HOME", "/home/test")
	t.Setenv("LOCALAPPDATA", "C:\\Users\\Test\\AppData\\Local")
	path, err := cb.GetHistoryPath()
	assert.NoError(t, err)
	switch runtime.GOOS {
	case "windows":
		assert.Equal(t, "C:\\Users\\Test\\AppData\\Local\\Google\\Chrome\\User Data\\Default\\History", path)
	case "darwin":
		assert.Equal(t, "/home/test/Library/Application Support/Google/Chrome/Default/History", path)
	case "linux":
		assert.Equal(t, "/home/test/.config/google-chrome/Default/History", path)
	default:
		t.Skipf("unsupported OS for this test: %s", runtime.GOOS)
	}
}

func TestChromeGetHistoryPath_UnsupportedOS(t *testing.T) {
	cb := &browser.ChromeBrowser{}
	path, err := cb.GetHistoryPath()
	if runtime.GOOS != "windows" && runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		assert.Error(t, err)
		assert.Empty(t, path)
		assert.Contains(t, err.Error(), "unsupported operating system")
	} else {
		t.Log("Skipping unsupported OS test on supported platform")
	}
}

func TestChromeExtractHistory_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "chrome_test_*.db")
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

	cb := &browser.ChromeBrowser{}
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()

	entries, err := cb.ExtractHistory(tmpFile.Name(), start, end, false)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "http://example.com", entries[0].URL)
	assert.Equal(t, "Example", entries[0].Title)
	assert.True(t, entries[0].Timestamp.After(start) && entries[0].Timestamp.Before(end))
}

func TestChromeExtractHistory_InvalidDB(t *testing.T) {
	cb := &browser.ChromeBrowser{}
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()
	entries, err := cb.ExtractHistory("nonexistent.db", start, end, true)
	assert.Error(t, err)
	assert.Nil(t, entries)
	assert.Contains(t, err.Error(), "unable to open database file") // Match SQLite driver error
}

func TestTimeToChromeTime(t *testing.T) {
	tm := time.Unix(0, 0)
	chromeTime := browser.TimeToChromeTime(tm)
	const epochDiff int64 = 11644473600000000 // Explicitly int64
	assert.Equal(t, epochDiff, chromeTime)
}

func TestChromeTimeToTime(t *testing.T) {
	const epochDiff int64 = 11644473600000000
	tm := browser.ChromeTimeToTime(epochDiff)
	assert.Equal(t, time.Unix(0, 0), tm)
}
