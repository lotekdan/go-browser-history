package main

import (
	"bytes"
	"os"
	"testing"
)

func TestRootCmd(t *testing.T) {
	// Redirect stdout for testing
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		w.Close()
		os.Stdout = old
	}()

	// Create and execute the command with invalid browser
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--browser", "invalid"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error from execution, got %v", err)
	}

	// Capture output
	w.Close()
	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}
	output := buf.String()

	if !bytes.Contains([]byte(output), []byte("Invalid browser type")) {
		t.Errorf("Expected invalid browser message, got: %s", output)
	}
}
