package history

import (
	"time"
)

type HistoryPathEntry struct {
	Profile     string
	ProfileName string
	Path        string
}

// HistoryEntry represents a single browser history entry.
type HistoryEntry struct {
	URL        string
	Title      string
	VisitCount int
	Typed      int
	VisitType  string
	Timestamp  time.Time
	Profile    string
}

type OutputEntry struct {
	Timestamp  string `json:"timestamp"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	VisitCount int    `json:"visitCount"`
	Typed      int    `json:"typed"`
	VisitType  string `json:"visitType"`
	Browser    string `json:"browser"`
	Profile    string `json:"profile"`
}

func ToOutputEntries(entries []HistoryEntry, browserName string) []OutputEntry {
	var output []OutputEntry
	for _, entry := range entries {
		output = append(output, OutputEntry{
			Timestamp:  entry.Timestamp.Format(time.RFC3339),
			Title:      entry.Title,
			URL:        entry.URL,
			VisitCount: entry.VisitCount,
			Typed:      entry.Typed,
			VisitType:  entry.VisitType,
			Browser:    browserName,
			Profile:    entry.Profile,
		})
	}
	return output
}
