package history

import (
	"time"

	"github.com/lotekdan/go-browser-history/internal/browser"
)

type OutputEntry struct {
	Timestamp  string `json:"timestamp"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	VisitCount int    `json:"visitcount"`
	Typed      int    `json:"typed"`
	VisitType  string `json:"visittype"`
	Browser    string `json:"browser"`
}

func ToOutputEntries(entries []browser.HistoryEntry, browserName string) []OutputEntry {
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
		})
	}
	return output
}
