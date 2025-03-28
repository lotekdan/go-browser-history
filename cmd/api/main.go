package main

import (
	"fmt"
	"os"

	"github.com/lotekdan/go-browser-history/internal/config"
	"github.com/lotekdan/go-browser-history/internal/server"
)

var Version string = "dev"

// serverStartFunc allows mocking server.Start in tests
var serverStartFunc = server.Start

func main() {
	os.Exit(run())
}

func run() int {
	cfg := config.NewDefaultConfig()
	cfg.Mode = "api"

	if err := serverStartFunc(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting API server: %v\n", err)
		return 1
	}
	return 0
}
