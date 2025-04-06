package config

import "time"

type Config struct {
	HistoryDays int
	Browser     string
	JSONOutput  bool
	PrettyPrint bool
	Mode        string
	Port        string
	Debug       bool // New field for debug logging
	StartTime   time.Time
	EndTime     time.Time
}

func NewDefaultConfig() *Config {
	now := time.Now()
	return &Config{
		HistoryDays: 30,
		Mode:        "cli",
		Port:        "8080",
		Debug:       false, // Debug off by default
		StartTime:   now.AddDate(0, 0, -30),
		EndTime:     now,
	}
}
