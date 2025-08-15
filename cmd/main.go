package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"go-password-manager/internal/env"
	"go-password-manager/ui"
)

var (
	version = "development"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Handle version flag
	var showVersion = flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Go Password Manager %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Built: %s\n", date)
		os.Exit(0)
	}

	// Load environment configuration
	config, err := env.Load()
	if err != nil {
		log.Fatalf("Failed to load environment configuration: %v", err)
	}

	// Set version in config if available
	if version != "dev" {
		config.AppVersion = version
	}

	// Pass config to the UI
	app := ui.NewApp(config)
	app.Run()
}
