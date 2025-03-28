package main_test

import (
	"testing"

	"github.com/lotekdan/go-browser-history/internal/server"
	"github.com/stretchr/testify/assert"
)

func TestRunServer(t *testing.T) {
	srv := server.NewServer(nil)
	done := make(chan error)
	go func() {
		err := srv.Run(":8080")
		done <- err
	}()
	err := <-done
	assert.Error(t, err) // Port binding will fail in test env
}
