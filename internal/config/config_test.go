package config

import (
	"reflect"
	"testing"
	"time"
)

func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()
	if cfg == nil {
		t.Fatal("NewDefaultConfig() returned nil")
	}

	now := time.Now()
	expected := &Config{
		HistoryDays: 30,
		Browser:     "",
		JSONOutput:  false,
		Mode:        "cli",
		Port:        "8080",
		Debug:       false,
		StartTime:   now.AddDate(0, 0, -30),
		EndTime:     now,
	}

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"HistoryDays", cfg.HistoryDays, expected.HistoryDays},
		{"Browser", cfg.Browser, expected.Browser},
		{"JSONOutput", cfg.JSONOutput, expected.JSONOutput},
		{"Mode", cfg.Mode, expected.Mode},
		{"Port", cfg.Port, expected.Port},
		{"Debug", cfg.Debug, expected.Debug},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.got, tt.want) {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}

	timeTolerance := time.Second
	if diff := cfg.StartTime.Sub(expected.StartTime); diff > timeTolerance || diff < -timeTolerance {
		t.Errorf("StartTime = %v, want %v (±%v)", cfg.StartTime, expected.StartTime, timeTolerance)
	}
	if diff := cfg.EndTime.Sub(expected.EndTime); diff > timeTolerance || diff < -timeTolerance {
		t.Errorf("EndTime = %v, want %v (±%v)", cfg.EndTime, expected.EndTime, timeTolerance)
	}
}

func TestConfigFields(t *testing.T) {
	cfg := NewDefaultConfig()

	if cfg.HistoryDays <= 0 {
		t.Errorf("HistoryDays should be positive, got %d", cfg.HistoryDays)
	}
	if cfg.Mode == "" {
		t.Errorf("Mode should not be empty, got %q", cfg.Mode)
	}
	if cfg.Port == "" {
		t.Errorf("Port should not be empty, got %q", cfg.Port)
	}
	if cfg.StartTime.After(cfg.EndTime) {
		t.Errorf("StartTime %v should not be after EndTime %v", cfg.StartTime, cfg.EndTime)
	}
}
