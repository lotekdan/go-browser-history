package main

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/lotekdan/go-browser-history/internal/config"
)

func TestVersion(t *testing.T) {
	if Version != "dev" {
		t.Errorf("Version = %q, want %q", Version, "dev")
	}
}

func TestRun_Success(t *testing.T) {
	// Redirect stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() {
		w.Close()
		os.Stderr = oldStderr
	}()

	// Mock serverStartFunc to succeed
	oldServerStart := serverStartFunc
	serverStartFunc = func(cfg *config.Config) error {
		if cfg.Mode != "api" {
			t.Errorf("cfg.Mode = %q, want %q", cfg.Mode, "api")
		}
		return nil
	}
	defer func() { serverStartFunc = oldServerStart }()

	// Run and check exit code
	exitCode := run()
	if exitCode != 0 {
		t.Errorf("run() = %d, want 0", exitCode)
	}

	// Check stderr is empty
	w.Close()
	var stderr bytes.Buffer
	stderr.ReadFrom(r)
	if stderr.Len() > 0 {
		t.Errorf("stderr = %q, want empty", stderr.String())
	}
}

func TestRun_ServerStartError(t *testing.T) {
	// Redirect stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() {
		w.Close()
		os.Stderr = oldStderr
	}()

	// Mock serverStartFunc to fail
	oldServerStart := serverStartFunc
	serverStartFunc = func(cfg *config.Config) error {
		return errors.New("server start failed")
	}
	defer func() { serverStartFunc = oldServerStart }()

	// Run and check exit code
	exitCode := run()
	if exitCode != 1 {
		t.Errorf("run() = %d, want 1", exitCode)
	}

	// Check stderr output
	w.Close()
	var stderr bytes.Buffer
	stderr.ReadFrom(r)
	expected := "Error starting API server: server start failed\n"
	if stderr.String() != expected {
		t.Errorf("stderr = %q, want %q", stderr.String(), expected)
	}
}
