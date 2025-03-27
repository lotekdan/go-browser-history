package main

import (
	"fmt"
	"os"

	"github.com/lotekdan/go-browser-history/internal/config"
	"github.com/lotekdan/go-browser-history/internal/server"
)

var Version string = "dev"

func main() {
	cfg := config.NewDefaultConfig()
	cfg.Mode = "api"

	if err := server.Start(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting API server: %v\n", err)
		os.Exit(1)
	}
}
